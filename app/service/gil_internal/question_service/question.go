// 对接题库平台 API
package question_service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gil_teacher/app/conf"
	"gil_teacher/app/consts"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/model/api"
	"gil_teacher/app/model/itl"
	"gil_teacher/app/service/gil_internal/admin_service"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Client 题库 HTTP 客户端
type Client struct {
	httpClient  *http.Client
	host        string
	log         *logger.ContextLogger
	adminClient *admin_service.AdminClient
}

// NewClient 创建题库客户端
func NewClient(c *conf.Conf, log *logger.ContextLogger, adminClient *admin_service.AdminClient) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: consts.QuestionAPIDefaultTimeout,
		},
		host:        c.QuestionAPI.Host,
		log:         log,
		adminClient: adminClient,
	}
}

// doRequest 执行 HTTP 请求
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	url := c.host + path

	var bodyReader io.Reader
	var bodyStr string
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("序列化请求体失败: %v", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
		bodyStr = string(jsonBody)
	}

	c.log.Info(ctx, "题库 API 请求 - 方法: %s, URL: %s, 请求体: %s", method, url, bodyStr)

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("执行请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应体失败: %v", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d: %s", resp.StatusCode, string(respBody))
	}

	// 打印响应数据（限制长度）
	const maxRespLength = 1000 // 最大响应长度
	respStr := string(respBody)
	if len(respStr) > maxRespLength {
		respStr = respStr[:maxRespLength] + "... (响应已截断，总长度: " + fmt.Sprintf("%d", len(respBody)) + " 字节)"
	}
	c.log.Debug(ctx, "题库 API 响应: %s", respStr)

	// 解析响应
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("解析响应体失败: %v", err)
		}
	}

	return nil
}

// GetBizTreeList 获取业务树列表 - 分为章节类型业务树和知识点类型业务树
func (c *Client) GetBizTreeList(ctx context.Context, bizTreeType, phase, subject int64) ([]itl.BizTreeInfo, error) {
	req := &itl.ListBizTreeRequestBody{
		BizTreeType: bizTreeType,
		PhaseList:   []int64{phase},
		SubjectList: []int64{subject},
		Page:        1,
		PageSize:    10000, // 这里需要拿到全部数据，但是题库分页了
	}

	result := &itl.ListBizTreeResponseBody{}

	err := c.doRequest(
		ctx,
		consts.QuestionAPIBizTreeList.Method,
		consts.QuestionAPIBizTreeList.Path,
		req,
		&result,
	)
	if err != nil {
		c.log.Error(ctx, "获取业务树列表失败: %v", err)
		return nil, err
	}

	if result.Code != 0 {
		c.log.Error(ctx, "获取业务树列表失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
		return nil, fmt.Errorf("获取业务树列表失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
	}

	return result.Data.BizTreeInfoList, nil
}

// GetBizTreeDetail 获取业务树详情
func (c *Client) GetBizTreeDetail(ctx context.Context, bizTreeID int64) (*itl.BizTreeEntity, error) {
	req := &itl.GetBizTreeDetailRequestQuery{
		BizTreeID: bizTreeID,
	}

	result := &itl.GetBizTreeDetailResponseBody{}

	path := fmt.Sprintf("%s?bizTreeId=%d", consts.QuestionAPIBizTreeDetail.Path, req.BizTreeID)
	err := c.doRequest(
		ctx,
		consts.QuestionAPIBizTreeDetail.Method,
		path,
		nil,
		&result,
	)
	if err != nil {
		c.log.Error(ctx, "获取业务树详情失败: %v", err)
		return nil, err
	}

	if result.Code != 0 {
		c.log.Error(ctx, "获取业务树详情失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
		// 题库这里的数据不存在时，返回的 http 状态码是 200，message 是 record not found
		if strings.Contains(result.Message, "record not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("获取业务树详情失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
	}

	return &result.Data, nil
}

// GetBizTreeLeafNodes 获取业务树某个节点下的所有叶子节点
func (c *Client) GetBizTreeLeafNodes(ctx context.Context, bizTreeID int64, bizTreeNodeID int64) ([]itl.BizTreeLeafNode, error) {
	bizTreeInfo, err := c.GetBizTreeDetail(ctx, bizTreeID)
	if err != nil {
		c.log.Error(ctx, "获取业务树详情失败: %v", err)
		return nil, err
	}

	// 遍历业务树 bizTreeNodeID 节点下的所有叶子节点
	leafNodes := []itl.BizTreeLeafNode{}

	if bizTreeInfo == nil || bizTreeInfo.BizTreeDetail == nil {
		c.log.Warn(ctx, "获取业务树详情失败: 业务树数据为空, bizTreeID: %d, bizTreeNodeID: %d", bizTreeID, bizTreeNodeID)
		return leafNodes, nil
	}

	// 查找指定节点ID的节点
	targetNode := findNodeByID(bizTreeInfo.BizTreeDetail, bizTreeNodeID)
	if targetNode == nil {
		return leafNodes, nil
	}

	// 收集所有叶子节点
	collectLeafNodes(targetNode, &leafNodes)

	return leafNodes, nil
}

// findNodeByID 根据节点ID查找业务树节点
func findNodeByID(root *itl.BizTreeNodeEntity, nodeID int64) *itl.BizTreeNodeEntity {
	if root == nil {
		return nil
	}

	// 如果当前节点就是目标节点，直接返回
	if root.BizTreeNodeId == nodeID {
		return root
	}

	// 递归查找子节点
	for _, child := range root.BizTreeNodeChildren {
		if found := findNodeByID(child, nodeID); found != nil {
			return found
		}
	}

	return nil
}

// collectLeafNodes 收集指定节点下的所有叶子节点
func collectLeafNodes(node *itl.BizTreeNodeEntity, leafNodes *[]itl.BizTreeLeafNode) {
	if node == nil {
		return
	}

	// 如果节点没有子节点，则为叶子节点
	if len(node.BizTreeNodeChildren) == 0 {
		leafNode := itl.BizTreeLeafNode{
			BizTreeNodeId:   node.BizTreeNodeId,
			BizTreeNodeName: node.BizTreeNodeName,
		}
		*leafNodes = append(*leafNodes, leafNode)
		return
	}

	// 递归处理所有子节点
	for _, child := range node.BizTreeNodeChildren {
		collectLeafNodes(child, leafNodes)
	}
}

// CheckQuestionSetExistByIDs 根据id检查题集列表，这个接口没有返回题目数据
func (c *Client) CheckQuestionSetExistByIDs(ctx context.Context, ids []int64) ([]itl.QuestionSetShelfInfo, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	req := &itl.CheckQuestionSetExistByIDsRequestBody{
		QuestionSetIds: ids,
	}

	result := &itl.CheckQuestionSetExistByIDsResponseBody{}

	err := c.doRequest(ctx, consts.QuestionAPIPracticeListByID.Method, consts.QuestionAPIPracticeListByID.Path, req, &result)
	if err != nil {
		c.log.Error(ctx, "根据ID获取巩固练习列表失败: %v", err)
		return nil, err
	}

	return result.Data.QuestionSetList, nil
}

// GetPracticeInfoByBizTreeNodeID 获取业务树叶子节点巩固练习信息
func (c *Client) GetPracticeInfoByBizTreeNodeID(ctx context.Context, bizTreeNodeID int64) (*itl.QuestionSetStableInfo, error) {
	result := &itl.GetQustionSetInfoResponseBody{}

	path := fmt.Sprintf(
		"%s?bizTreeNodeId=%d&sceneCategory=%d",
		consts.QuestionAPIGetQuestionSetInfo.Path,
		bizTreeNodeID,
		consts.QuestionSceneCategoryPractice,
	)
	err := c.doRequest(ctx, consts.QuestionAPIGetQuestionSetInfo.Method, path, nil, &result)
	if err != nil {
		c.log.Error(ctx, "获取巩固练习信息失败: %v", err)
		return nil, err
	}

	if result.Code != 0 {
		// 题集如果不存在返回了错误码，而不是空数据
		if result.Code == consts.QuestionAPICodeQuestionSetNotExist {
			return nil, nil
		}
		c.log.Error(ctx, "获取巩固练习信息失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
		return nil, fmt.Errorf("获取巩固练习信息失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
	}

	return &result.Data, nil
}

// GetQuestionSetByID 通过题集ID获取题集信息
func (c *Client) GetQuestionSetByID(ctx context.Context, questionSetID int64) (*itl.QuestionSetStableInfo, error) {
	result := &itl.GetQustionSetInfoResponseBody{}

	path := fmt.Sprintf(
		"%s?questionSetId=%d",
		consts.QuestionAPIGetQuestionSetInfo.Path,
		questionSetID,
	)
	err := c.doRequest(ctx, consts.QuestionAPIGetQuestionSetInfo.Method, path, nil, &result)
	if err != nil {
		c.log.Error(ctx, "获取题集信息失败: %v", err)
		return nil, err
	}

	if result.Code != 0 {
		// 题集如果不存在返回了错误码，而不是空数据
		if result.Code == consts.QuestionAPICodeQuestionSetNotExist {
			return nil, nil
		}
		c.log.Error(ctx, "获取题集信息失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
		return nil, fmt.Errorf("获取题集信息失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
	}

	// 额外处理题目的标签
	for _, questionGroup := range result.Data.QuestionGroupStableInfoList {
		if questionGroup != nil && questionGroup.QuestionInfoList != nil {
			for _, question := range questionGroup.QuestionInfoList {
				if question != nil {
					c.handleQuestionTags(ctx, question.QuestionInfo)
				}
			}
		}
	}

	return &result.Data, nil
}

// GetQuestionSetListByIDs 通过题集ID批量获取题集列表
func (c *Client) GetQuestionSetListByIDs(ctx context.Context, questionSetIDs []int64) ([]*itl.QuestionSetStableInfo, error) {
	if len(questionSetIDs) == 0 {
		return nil, nil
	}

	// 创建结果切片，预分配容量
	questionSetList := make([]*itl.QuestionSetStableInfo, len(questionSetIDs))

	// questionSetResult 题集结果
	type questionSetResult struct {
		index int
		info  *itl.QuestionSetStableInfo
	}

	// 创建结果通道
	resultChan := make(chan questionSetResult, len(questionSetIDs))
	// 创建信号量控制并发数
	sem := make(chan struct{}, 10)

	// 使用互斥锁和共享变量跟踪错误
	var mu sync.Mutex
	var hasErr bool

	// 使用WaitGroup等待所有goroutine完成
	var wg sync.WaitGroup

	// 遍历所有题集ID，获取题集信息
	for i, questionSetID := range questionSetIDs {
		wg.Add(1)
		go func(idx int, questionSetID int64) {
			defer wg.Done()
			sem <- struct{}{}        // 获取信号量
			defer func() { <-sem }() // 释放信号量

			// 检查是否已有错误
			mu.Lock()
			defer mu.Unlock()
			if hasErr {
				return
			}

			// 获取单个题集信息
			questionSetInfo, err := c.GetQuestionSetByID(ctx, questionSetID)
			if err != nil {
				c.log.Error(ctx, "获取题集信息失败 questionSetID: %d, err: %v", questionSetID, err)
				hasErr = true
				return
			}

			resultChan <- questionSetResult{
				index: idx,
				info:  questionSetInfo,
			}
		}(i, questionSetID)
	}

	// 启动一个goroutine关闭结果通道
	wg.Wait()
	close(resultChan)

	// 检查是否发生错误
	if hasErr {
		return nil, fmt.Errorf("批量获取题集信息失败")
	}

	// 收集结果
	for result := range resultChan {
		questionSetList[result.index] = result.info
	}

	return questionSetList, nil
}

// GetPracticeListByBizTreeNodeIDs 批量获取巩固练习信息，只有最底层叶子节点有巩固练习，返回结果按照入参顺序返回
func (c *Client) GetPracticeListByBizTreeNodeIDs(ctx context.Context, bizTreeNodeIDs []int64) ([]*itl.QuestionSetStableInfo, error) {
	if len(bizTreeNodeIDs) == 0 {
		return nil, nil
	}

	// 创建结果切片，预分配容量
	practiceInfoList := make([]*itl.QuestionSetStableInfo, len(bizTreeNodeIDs))

	// practiceResult 巩固练习结果
	type practiceResult struct {
		index int
		info  *itl.QuestionSetStableInfo
	}

	// 创建结果通道
	resultChan := make(chan practiceResult, len(bizTreeNodeIDs))
	// 创建信号量控制并发数
	sem := make(chan struct{}, 10)

	// 使用互斥锁和共享变量跟踪错误
	var mu sync.Mutex
	var hasErr bool

	// 使用WaitGroup等待所有goroutine完成
	var wg sync.WaitGroup

	// 遍历所有节点，获取巩固练习信息
	for i, bizTreeNodeID := range bizTreeNodeIDs {
		wg.Add(1)
		go func(idx int, nodeID int64) {
			defer wg.Done()
			sem <- struct{}{}        // 获取信号量
			defer func() { <-sem }() // 释放信号量

			// 检查是否已有错误
			mu.Lock()
			defer mu.Unlock()
			if hasErr {
				return
			}

			// 获取单个节点的巩固练习信息
			practiceInfo, err := c.GetPracticeInfoByBizTreeNodeID(ctx, nodeID)
			if err != nil {
				c.log.Error(ctx, "获取巩固练习信息失败 bizTreeNodeID: %d, err: %v", nodeID, err)
				hasErr = true
				return
			}

			resultChan <- practiceResult{
				index: idx,
				info:  practiceInfo,
			}
		}(i, bizTreeNodeID)
	}

	// 启动一个goroutine关闭结果通道
	wg.Wait()
	close(resultChan)

	// 检查是否发生错误
	if hasErr {
		return nil, fmt.Errorf("批量获取巩固练习信息失败")
	}

	// 收集结果
	for result := range resultChan {
		practiceInfoList[result.index] = result.info
	}

	return practiceInfoList, nil
}

// GetQuestionEnums 获取题目查询枚举值
func (c *Client) GetQuestionEnums(ctx context.Context) (*itl.QuestionEnumsData, error) {
	result := &itl.QuestionEnumsResponseBody{}

	err := c.doRequest(ctx, consts.QuestionAPIQuestionEnums.Method, consts.QuestionAPIQuestionEnums.Path, nil, &result)
	if err != nil {
		c.log.Error(ctx, "获取题目查询枚举值失败: %v", err)
		return nil, err
	}

	if result.Code != 0 {
		c.log.Error(ctx, "获取题目查询枚举值失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
		return nil, fmt.Errorf("获取题目查询枚举值失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
	}

	return &result.Data, nil
}

// GetQuestionList 获取题目列表
func (c *Client) GetQuestionList(ctx context.Context, req *api.GetQuestionListRequest) (*itl.QuestionListOutput, error) {
	reqBody := &itl.QuestionListRequestBody{
		PhaseList:         []int64{req.Phase},
		SubjectList:       []int64{req.Subject},
		BizTreeNodeIds:    req.BizTreeNodeIds,
		Keyword:           req.Keyword,
		QuestionType:      req.QuestionType,
		QuestionDifficult: req.QuestionDifficult,
		QuestionYears:     req.QuestionYears,
		Page:              req.Page,
		PageSize:          req.PageSize,
		Sort:              req.Sort,
	}
	result := &itl.QuestionListResponseBody{}

	err := c.doRequest(ctx, consts.QuestionAPIQuestionList.Method, consts.QuestionAPIQuestionList.Path, reqBody, &result)
	if err != nil {
		c.log.Error(ctx, "获取题目列表失败: %v", err)
		return nil, err
	}

	if result.Code != 0 {
		c.log.Error(ctx, "获取题目列表失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
		return nil, fmt.Errorf("获取题目列表失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
	}

	if result.Data.Questions == nil {
		return &result.Data, nil
	}

	// 额外处理题目的标签
	for _, question := range result.Data.Questions {
		c.handleQuestionTags(ctx, question)
	}

	return &result.Data, nil
}

// GetQuestionDetail 获取题目详情
func (c *Client) GetQuestionDetail(ctx context.Context, questionID string) (*itl.Question, error) {
	req := &itl.GetQuestionDetailRequestQuery{
		QuestionId: questionID,
	}

	result := &itl.GetQuestionDetailResponseBody{}

	path := fmt.Sprintf("%s?questionId=%s", consts.QuestionAPIQuestionDetail.Path, req.QuestionId)
	err := c.doRequest(ctx, consts.QuestionAPIQuestionDetail.Method, path, nil, &result)
	if err != nil {
		c.log.Error(ctx, "获取题目详情失败: %v", err)
		return nil, err
	}

	if result.Code != 0 {
		c.log.Error(ctx, "获取题目详情失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
		return nil, fmt.Errorf("获取题目详情失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
	}

	return result.Data, nil
}

// GetQuestionListByID 根据id获取题目列表
func (c *Client) GetQuestionListByID(ctx context.Context, ids []string, needContent bool) ([]*itl.Question, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	needContentInt := 0
	if needContent {
		needContentInt = 1
	}
	req := &itl.GetQuestionListByIDRequestBody{
		QuestionIdList: ids,
		NeedContent:    int64(needContentInt),
	}

	result := &itl.GetQuestionListByIDResponseBody{}

	err := c.doRequest(ctx, consts.QuestionAPIQuestionListByID.Method, consts.QuestionAPIQuestionListByID.Path, req, &result)
	if err != nil {
		c.log.Error(ctx, "根据ID获取题目列表失败: %v", err)
		return nil, err
	}

	if result.Code != 0 {
		c.log.Error(ctx, "根据ID获取题目列表失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
		return nil, fmt.Errorf("根据ID获取题目列表失败: 错误码: %d, 错误信息: %s", result.Code, result.Message)
	}

	// 检查QuestionList是否为空
	if result.Data.QuestionList == nil {
		return []*itl.Question{}, nil
	}

	// 额外处理题目的标签
	if needContent {
		for _, question := range result.Data.QuestionList {
			if question != nil {
				c.handleQuestionTags(ctx, question)
			}
		}
	}

	return result.Data.QuestionList, nil
}

// CheckResourceExist 检查资源是否存在
func (c *Client) CheckResourceExist(ctx context.Context, resources []api.TaskResource) bool {
	questionIDs := make([]string, 0, len(resources))
	for _, resource := range resources {
		if resource.ResourceType == consts.RESOURCE_TYPE_QUESTION {
			questionIDs = append(questionIDs, resource.ResourceID)
		}
	}

	questionSetIDs := make([]int64, 0, len(resources))
	for _, resource := range resources {
		if resource.ResourceType == consts.RESOURCE_TYPE_PRACTICE {
			id, err := strconv.ParseInt(resource.ResourceID, 10, 64)
			if err != nil {
				c.log.Error(ctx, "巩固练习ID格式错误")
				return false
			}
			questionSetIDs = append(questionSetIDs, id)
		}
	}

	// 如果两种资源都为空，直接返回 true
	if len(questionIDs) == 0 && len(questionSetIDs) == 0 {
		return true
	}

	questionChan := make(chan []*itl.Question, 1)
	practiceChan := make(chan []itl.QuestionSetShelfInfo, 1)

	// 只有当题目ID列表不为空时才检查题目
	if len(questionIDs) > 0 {
		go func() {
			questionList, err := c.GetQuestionListByID(ctx, questionIDs, false)
			if err != nil {
				c.log.Error(ctx, "获取题目列表失败: %v", err)
				questionChan <- nil
				return
			}
			questionChan <- questionList
		}()
	} else {
		questionChan <- []*itl.Question{}
	}

	// 只有当练习ID列表不为空时才检查练习
	if len(questionSetIDs) > 0 {
		go func() {
			practiceList, err := c.CheckQuestionSetExistByIDs(ctx, questionSetIDs)
			if err != nil {
				c.log.Error(ctx, "获取巩固练习列表失败: %v", err)
				practiceChan <- nil
				return
			}
			practiceChan <- practiceList
		}()
	} else {
		practiceChan <- []itl.QuestionSetShelfInfo{}
	}

	questionList := <-questionChan
	practiceList := <-practiceChan

	if questionList == nil || practiceList == nil {
		return false
	}

	// 需要再过滤一遍已上架的题集，教师端不使用未上架的题集
	onShelfPracticeCount := 0
	for _, practice := range practiceList {
		if practice.ShelfStatus == consts.QuestionSetShelfStatusOnShelf {
			onShelfPracticeCount++
		}
	}

	if len(questionList) != len(questionIDs) || onShelfPracticeCount != len(questionSetIDs) {
		return false
	}

	return true
}

// 获取题目和巩固练习列表
func (c *Client) GetResources(ctx context.Context, questionsIDs []string, practiceIDs []int64) (map[string][]*itl.Question, map[int64][]*itl.Question, int64, error) {
	questionList, err := c.GetQuestionListByID(ctx, questionsIDs, true)
	if err != nil {
		c.log.Error(ctx, "获取题目列表失败: %v", err)
		return nil, nil, 0, err
	}

	practiceList, err := c.GetQuestionSetListByIDs(ctx, practiceIDs)
	if err != nil {
		c.log.Error(ctx, "获取巩固练习列表失败: %v", err)
		return nil, nil, 0, err
	}

	resourceQuestions := make(map[string][]*itl.Question, 0) // 资源ID -> 题目列表
	resourcePractices := make(map[int64][]*itl.Question, 0)  // 题集ID -> 题目列表

	for _, question := range questionList {
		if _, ok := resourceQuestions[question.QuestionId]; !ok {
			resourceQuestions[question.QuestionId] = make([]*itl.Question, 0)
		}
		resourceQuestions[question.QuestionId] = append(resourceQuestions[question.QuestionId], question)
	}

	for _, practice := range practiceList {
		if _, ok := resourcePractices[practice.QuestionSetId]; !ok {
			resourcePractices[practice.QuestionSetId] = make([]*itl.Question, 0)
		}
		for _, questionGroup := range practice.QuestionGroupStableInfoList {
			if _, ok := resourcePractices[practice.QuestionSetId]; !ok {
				resourcePractices[practice.QuestionSetId] = make([]*itl.Question, 0)
			}

			for _, question := range questionGroup.QuestionInfoList {
				resourcePractices[practice.QuestionSetId] = append(resourcePractices[practice.QuestionSetId], question.QuestionInfo)
			}
		}
	}

	return resourceQuestions, resourcePractices, int64(len(questionList) + len(practiceList)), nil
}

// handleQuestionTags 处理 questionTags 字段
func (c *Client) handleQuestionTags(ctx context.Context, question *itl.Question) error {
	if question == nil {
		return nil
	}

	// 初始化标签列表
	question.QuestionTags = []string{}

	// 检查QuestionInfoEntity是否为nil
	if question.QuestionInfoEntity != nil && question.QuestionContentEntity != nil {
		// 题目标签的格式为：【年·省份·原始名称，题型，题目难度】
		tag1Arr := []string{}
		// 更早的年份不展示
		if question.QuestionYear > consts.GilQuestionYearBeforeCondition {
			tag1Arr = append(tag1Arr, fmt.Sprintf("%d年", question.QuestionYear))
		}
		if question.ProvinceCode != 0 {
			provinceName, _ := c.adminClient.GetProvinceNameByCode(ctx, question.ProvinceCode)
			if provinceName != "" {
				tag1Arr = append(tag1Arr, provinceName)
			}
		}
		if question.QuestionContentFormat != nil && question.QuestionContentFormat.QuestionOriginName != "" {
			tag1Arr = append(tag1Arr, question.QuestionContentFormat.QuestionOriginName)
		}
		if tag1 := strings.Join(tag1Arr, "·"); tag1 != "" {
			question.QuestionTags = append(question.QuestionTags, tag1)
		}
		if tag2 := consts.GetQuestionTypeName(question.QuestionType); tag2 != "" {
			question.QuestionTags = append(question.QuestionTags, tag2)
		}
		if tag3 := consts.GetQuestionDifficultName(question.QuestionDifficult); tag3 != "" {
			question.QuestionTags = append(question.QuestionTags, tag3)
		}
	}
	return nil
}
