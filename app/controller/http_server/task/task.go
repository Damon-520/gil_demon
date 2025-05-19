package controller_task

import (
	"gil_teacher/app/consts"
	"gil_teacher/app/controller/http_server/response"
	"gil_teacher/app/core/logger"
	dao_task "gil_teacher/app/dao/task"
	"gil_teacher/app/middleware"
	"gil_teacher/app/model/api"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/model/itl"
	"gil_teacher/app/service/gil_internal/admin_service"
	"gil_teacher/app/service/gil_internal/question_service"
	"gil_teacher/app/service/task_service"
	"gil_teacher/app/third_party/volc_ai"
	"gil_teacher/app/utils"

	"github.com/gin-gonic/gin"
)

// TaskController 任务控制器
type TaskController struct {
	log                 *logger.ContextLogger
	taskService         *task_service.TaskService
	taskResourceService *task_service.TaskResourceService
	questionAPI         *question_service.Client
	ucenterService      *admin_service.UcenterClient
	teacherMiddleware   *middleware.TeacherMiddleware
	volcAI              *volc_ai.Client
}

// NewTaskController 创建任务控制器实例
func NewTaskController(
	log *logger.ContextLogger,
	taskService *task_service.TaskService,
	taskResourceService *task_service.TaskResourceService,
	questionAPI *question_service.Client,
	ucenterService *admin_service.UcenterClient,
	teacherMiddleware *middleware.TeacherMiddleware,
	volcAI *volc_ai.Client,
) *TaskController {
	return &TaskController{
		log:                 log,
		taskService:         taskService,
		taskResourceService: taskResourceService,
		questionAPI:         questionAPI,
		ucenterService:      ucenterService,
		teacherMiddleware:   teacherMiddleware,
		volcAI:              volcAI,
	}
}

// GetTaskType 获取老师拥有的学科及任务类型
func (c *TaskController) GetTaskType(ctx *gin.Context) {
	// 只有学科教师和班主任有权限创建任务，学科教师限制了具体的学科，班主任则具备全部学科的权限
	// 从中间件获取老师的学段、学科
	phase := c.teacherMiddleware.ExtractTeacherPhase(ctx)
	subjects := c.teacherMiddleware.GetTaskCreationSubjects(ctx)
	c.log.Debug(ctx, "获取老师拥有的学科及任务类型: %v, %v", phase, subjects)

	// 根据学段学科获取所有学科的任务类型
	var res api.GetTaskTypeResponse

	if len(subjects) == 0 {
		response.Success(ctx, &res)
		return
	}

	// 获取学科老师具备的学科列表
	subjectTeacherSubjects := c.teacherMiddleware.GetSubjectTeacherSubjects(ctx)
	c.log.Debug(ctx, "老师作为学科老师具备的学科列表: %v", subjectTeacherSubjects)

	// 返回结果需要将学科老师的学科排到最前面
	subjects = append(subjectTeacherSubjects, subjects...)
	// 去重
	subjects = utils.RemoveDuplicateInt64(subjects)
	c.log.Debug(ctx, "老师创建任务时拥有权限的学科列表: %v", subjects)

	// 按照特定顺序遍历配置查找匹配的学段学科下的任务类型
	for _, subject := range subjects {
		for _, phaseInfo := range consts.AllPhaseSubjectTaskType {
			if phaseInfo.Key == phase {
				for _, subjectInfo := range phaseInfo.Subjects {
					if subjectInfo.Key == subject {
						res.SubjectTaskTypes = append(res.SubjectTaskTypes, api.SubjectTaskType{
							SubjectKey:  subjectInfo.Key,
							SubjectName: subjectInfo.Value,
							TaskTypes:   subjectInfo.TaskTypes,
						})
					}
				}
			}
		}
	}

	response.Success(ctx, &res)
}

// GetTaskByID 根据 ID 查询单个任务，前端复制任务时用
func (c *TaskController) GetTaskByID(ctx *gin.Context) {
	// 获取请求参数
	taskID := utils.Atoi64(ctx.Query("taskId"))
	if taskID == 0 {
		response.Err(ctx, response.ERR_INVALID_TASK)
		return
	}
	teacherID := c.teacherMiddleware.ExtractTeacherID(ctx)

	res := api.GetTaskDetailByIDResponse{}
	// 查询任务
	task, err := c.taskService.GetTaskByIDAndCreatorID(ctx, taskID, teacherID)
	if err != nil {
		c.log.Error(ctx, "获取任务失败: %v", err)
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}
	if task == nil {
		response.ParamError(ctx, response.ERR_INVALID_TASK)
		return
	}
	res.Task = task
	// 查询任务的资源
	resources, err := c.taskResourceService.GetTaskResourcesByTaskID(ctx, taskID)
	if err != nil {
		c.log.Error(ctx, "获取任务资源失败: %v", err)
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}
	res.Resources = resources
	// 查询任务的分配
	assigns, err := c.taskService.GetTaskAssignsByTaskIDs(ctx, []int64{taskID})
	if err != nil {
		c.log.Error(ctx, "获取任务分配失败: %v", err)
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}
	res.StudentGroups = assigns
	response.Success(ctx, res)
}

// GetStudentTaskList 查询学生的任务列表
func (c *TaskController) GetStudentTaskList(ctx *gin.Context) {
	// 验证请求参数
	var reqBody api.GetStudentTaskListRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		c.log.Error(ctx, "绑定请求参数失败: %v", err)
		response.ParamError(ctx)
		return
	}
	if err := reqBody.Validate(); err != nil {
		c.log.Error(ctx, "验证请求参数失败: %v", err)
		response.ParamError(ctx, *err)
		return
	}

	// 查询学生的任务列表
	tasks, total, err := c.taskService.GetStudentTaskList(ctx, reqBody.StudentTaskListQuery)
	if err != nil {
		c.log.Error(ctx, "获取学生任务列表失败: %v", err)
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}
	taskIDs := make([]int64, 0, len(tasks))
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.Task.TaskID)
	}

	// 查询任务资源
	resources, _, err := c.taskResourceService.GetTaskResourcesByTaskIDs(ctx, taskIDs)
	if err != nil {
		c.log.Error(ctx, "获取任务资源失败: %v", err)
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}

	result := api.GetStudentTaskListResponse{
		List: make([]struct {
			*dao_task.TaskAndTaskAssign
			Resources []*dao_task.TaskResource `json:"resources"`
		}, len(tasks)),
		ApiPageResponse: &consts.ApiPageResponse{
			Total:    total,
			Page:     reqBody.StudentTaskListQuery.Page,
			PageSize: reqBody.StudentTaskListQuery.PageSize,
		},
	}
	for i, task := range tasks {
		result.List[i].TaskAndTaskAssign = task
		resources := resources[task.Task.TaskID]
		result.List[i].Resources = make([]*dao_task.TaskResource, 0, len(resources))
		for _, resource := range resources {
			result.List[i].Resources = append(result.List[i].Resources, resource)
		}
	}

	response.Success(ctx, &result)
}

// CreateTask 创建任务
func (c *TaskController) CreateTask(ctx *gin.Context) {
	// 验证请求参数
	var reqBody api.CreateTaskRequestBody
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		c.log.Error(ctx, "绑定请求参数失败: %v", err)
		response.ParamError(ctx)
		return
	}
	if err := (&reqBody).Validate(); err != nil {
		c.log.Error(ctx, "验证请求参数失败: %v", err)
		response.ParamError(ctx, *err)
		return
	}

	// 从中间件获取学段、学校 ID 和老师 ID
	reqBody.Phase = c.teacherMiddleware.ExtractTeacherPhase(ctx)
	reqBody.SchoolID = c.teacherMiddleware.ExtractSchoolID(ctx)
	teacherID := c.teacherMiddleware.ExtractTeacherID(ctx)
	reqBody.CreatorID = teacherID
	reqBody.UpdaterID = teacherID
	c.log.Debug(ctx, "创建任务请求参数: %v", reqBody)

	// 检查教师是否具备学科的权限
	if !c.teacherMiddleware.HasTaskCreationSubjectPermission(ctx, reqBody.Subject) {
		response.Forbidden(ctx)
		return
	}

	// 检查教师是否具备班级的权限
	classIDs := []int64{}
	for _, group := range reqBody.StudentGroups {
		if group.GroupType == consts.TASK_GROUP_TYPE_CLASS && group.GroupID > 0 {
			classIDs = append(classIDs, group.GroupID)
		}
	}
	if !c.teacherMiddleware.TeacherHasClassPermission(ctx, classIDs...) {
		response.Forbidden(ctx)
		return
	}

	// 类型为班级时，处理 StudentIDs
	classStudentMap, err := c.ucenterService.GetClassStudent(ctx, reqBody.SchoolID, classIDs)
	if err != nil {
		c.log.Error(ctx, "获取班级下的学生ID失败: %v", err)
		response.Err(ctx, response.ERR_GIL_ADMIN)
		return
	}
	for i, group := range reqBody.StudentGroups {
		if group.GroupType == consts.TASK_GROUP_TYPE_CLASS && group.GroupID > 0 {
			classInfo, ok := classStudentMap[group.GroupID]
			if !ok {
				c.log.Error(ctx, "班级 %d 学生不存在", group.GroupID)
				response.Err(ctx, response.ERR_GIL_ADMIN)
				return
			}
			studentIDs := make([]int64, 0, len(classInfo.Students))
			for _, student := range classInfo.Students {
				studentIDs = append(studentIDs, student.ID)
			}
			reqBody.StudentGroups[i].StudentIDs = studentIDs
		}
	}

	// 内容平台检查资源是否存在
	if !c.questionAPI.CheckResourceExist(ctx, reqBody.Resources) {
		c.log.Warn(ctx, "请求的内容平台资源不存在")
		response.ParamError(ctx, response.ERR_INVALID_RESOURCE)
		return
	}

	// CQC 检查内容是否合规
	ok, err := c.volcAI.CQC(ctx, reqBody.TaskName+","+reqBody.TeacherComment)
	if err != nil {
		response.Err(ctx, response.ERR_VOLC_AI)
		return
	}
	if !ok {
		response.ParamError(ctx, response.ERR_CQC)
		return
	}

	// 如果是课程任务，则记录最近使用过的业务树
	if reqBody.TaskType == consts.TASK_TYPE_COURSE {
		if err := c.taskService.RecordLastUsedBizTree(ctx, teacherID, reqBody.Subject, reqBody.BizTreeID); err != nil {
			c.log.Error(ctx, "记录最近使用过的业务树失败: %v", err) // 失败了不影响创建任务
		}
	}

	// 创建任务
	err = c.taskService.CreateTask(ctx, &reqBody)
	if err != nil {
		c.log.Error(ctx, "创建任务失败: %v", err)
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}

	response.Success(ctx, nil)
}

// DeleteTask 删除任务
func (c *TaskController) DeleteTask(ctx *gin.Context) {
	// 验证请求参数
	var reqBody api.DeleteTaskRequestBody
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		response.ParamError(ctx)
		return
	}

	// 从中间件获取老师 ID
	teacherID := c.teacherMiddleware.ExtractTeacherID(ctx)

	// 删除任务，只能删除自己创建的任务
	err := c.taskService.DeleteTask(ctx, reqBody.TaskIDs, teacherID)
	if err != nil {
		c.log.Error(ctx, "删除任务失败: %v", err)
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}

	response.Success(ctx, nil)
}

// UpdateTask 更新任务
func (c *TaskController) UpdateTask(ctx *gin.Context) {
	// 验证请求参数
	var reqBody api.UpdateTaskRequestBody
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		response.ParamError(ctx)
		return
	}
	if err := (&reqBody).Validate(); err != nil {
		c.log.Error(ctx, "验证请求参数失败: %v", err)
		response.ParamError(ctx, *err)
		return
	}

	// 从中间件获取老师 ID
	teacherID := c.teacherMiddleware.ExtractTeacherID(ctx)

	// CQC 检查内容是否合规
	ok, err := c.volcAI.CQC(ctx, reqBody.TaskName+","+reqBody.TeacherComment)
	if err != nil {
		response.Err(ctx, response.ERR_VOLC_AI)
		return
	}
	if !ok {
		response.ParamError(ctx, response.ERR_CQC)
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if reqBody.TaskName != "" {
		updates["task_name"] = reqBody.TaskName
	}
	if reqBody.TeacherComment != "" {
		updates["teacher_comment"] = reqBody.TeacherComment
	}

	// 更新任务
	err = c.taskService.UpdateTask(ctx, reqBody.TaskID, teacherID, updates)
	if err != nil {
		c.log.Error(ctx, "更新任务失败: %v", err)
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}

	response.Success(ctx, nil)
}

// UpdateTaskAssign 更新任务分配
func (c *TaskController) UpdateTaskAssign(ctx *gin.Context) {
	// 验证请求参数
	var reqBody api.UpdateTaskAssignRequestBody
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		response.ParamError(ctx)
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if reqBody.StartTime > 0 {
		updates["start_time"] = reqBody.StartTime
	}
	if reqBody.Deadline > 0 {
		updates["deadline"] = reqBody.Deadline
	}

	// 更新任务分配
	err := c.taskService.UpdateTaskAssign(ctx, reqBody.AssignID, updates)
	if err != nil {
		c.log.Error(ctx, "更新任务分配失败: %v", err)
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}

	response.Success(ctx, nil)
}

// DeleteTaskAssign 删除任务分配
func (c *TaskController) DeleteTaskAssign(ctx *gin.Context) {
	// 验证请求参数
	var reqBody api.DeleteTaskAssignRequestBody
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		c.log.Warn(ctx, "绑定请求参数失败: %v", err)
		response.ParamError(ctx)
		return
	}
	if err := reqBody.Validate(); err != nil {
		c.log.Warn(ctx, "验证请求参数失败: %v", err)
		response.ParamError(ctx, *err)
		return
	}

	// 从中间件获取老师 ID
	teacherID := c.teacherMiddleware.ExtractTeacherID(ctx)

	// 先查询任务基本信息，检查任务是否存在
	task, err := c.taskService.GetTaskByIDAndCreatorID(ctx, reqBody.TaskID, teacherID)
	if err != nil {
		c.log.Error(ctx, "获取任务基本信息失败: %v", err)
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}
	if task == nil {
		c.log.Warn(ctx, "任务不存在")
		response.ParamError(ctx, response.ERR_INVALID_TASK)
		return
	}

	// 删除任务分配，如果分配的对象全部被删除则任务也一并删除
	err = c.taskService.DeleteTaskAssign(ctx, reqBody.TaskID, reqBody.AssignIDs, teacherID)
	if err != nil {
		c.log.Error(ctx, "删除任务分配失败: %v", err)
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}

	response.Success(ctx, nil)
}

// GetKnowledgeTreeList 获取知识点树列表
func (c *TaskController) GetKnowledgeTreeList(ctx *gin.Context) {
	// 获取 Query 请求参数
	subject := utils.Atoi64(ctx.Query("subject")) // 学科
	if !consts.SubjectExists(subject) {
		response.ParamError(ctx, response.ERR_SUBJECT)
		return
	}

	// 中间件获取老师的学科学段
	phase := c.teacherMiddleware.ExtractTeacherPhase(ctx)

	// 检查教师是否具备学科的权限
	if !c.teacherMiddleware.HasTaskCreationSubjectPermission(ctx, subject) {
		response.Forbidden(ctx)
		return
	}

	// 根据学科学段从题库获取知识点类型的业务树列表
	result, err := c.questionAPI.GetBizTreeList(ctx, consts.QuestionBizTreeTypeKnowledgePoint, phase, subject)
	if err != nil {
		c.log.Error(ctx, "获取基础树列表失败: %v", err)
		response.Err(ctx, response.ERR_GIL_QUESTION)
		return
	}

	response.Success(ctx, result)
}

// GetChapterBizTreeList 获取教材类型业务树列表
func (c *TaskController) GetChapterBizTreeList(ctx *gin.Context) {
	// 获取 Query 请求参数
	subject := utils.Atoi64(ctx.Query("subject")) // 学科
	if !consts.SubjectExists(subject) {
		response.ParamError(ctx, response.ERR_SUBJECT)
		return
	}

	// 先从运营平台获取学校的学科教材
	phase := c.teacherMiddleware.ExtractTeacherPhase(ctx)
	schoolID := c.teacherMiddleware.ExtractSchoolID(ctx)
	materialList, err := c.ucenterService.GetSchoolMaterial(ctx, schoolID)
	if err != nil {
		c.log.Error(ctx, "获取业务树列表失败: %v", err)
		response.Err(ctx, *err)
		return
	}
	// 记录教材版本
	materialMap := make(map[int64]struct{})
	for _, material := range materialList {
		for _, m := range material.Materials {
			materialMap[m] = struct{}{}
		}
	}

	// 再去题库获取业务树列表
	bizTreeList, err2 := c.questionAPI.GetBizTreeList(ctx, consts.QuestionBizTreeTypeAll, phase, subject)
	if err2 != nil {
		c.log.Error(ctx, "获取业务树列表失败: %v", err2)
		response.Err(ctx, response.ERR_GIL_QUESTION)
		return
	}

	// 获取老师最近使用过的业务树
	teacherID := c.teacherMiddleware.ExtractTeacherID(ctx)
	lastUsedBizTreeID, redisErr := c.taskService.GetLastUsedBizTree(ctx, teacherID, subject)
	if redisErr != nil {
		c.log.Error(ctx, "获取最近使用过的业务树失败: %v", redisErr) // 失败了不影响业务树列表
	}

	// 最后根据学科和学校的教材版本过滤业务树，将最近使用过的业务树放到数组的首位
	res := []itl.BizTreeInfo{}
	for _, bizTree := range bizTreeList {
		// 过滤学段学科、业务树类型为章节
		if bizTree.Phase == phase && bizTree.Subject == subject && bizTree.BizTreeType == consts.QuestionBizTreeTypeChapter {
			// 教材版本为学校的教材版本
			if _, ok := materialMap[bizTree.Material]; ok {
				// 如果业务树ID是最近使用过的业务树，则放到数组的首位
				if bizTree.BizTreeId == lastUsedBizTreeID {
					res = append([]itl.BizTreeInfo{bizTree}, res...)
				} else {
					res = append(res, bizTree)
				}
			}
		}
	}

	response.Success(ctx, res)
}

// GetBizTreeDetail 获取业务树详情
func (c *TaskController) GetBizTreeDetail(ctx *gin.Context) {
	bizTreeID := utils.Atoi64(ctx.Query("bizTreeId"))
	if bizTreeID == 0 {
		response.ParamError(ctx)
		return
	}

	// 根据业务树 ID 从题库获取业务树详情
	result, err := c.questionAPI.GetBizTreeDetail(ctx, bizTreeID)
	if err != nil {
		c.log.Error(ctx, "获取业务树详情失败: %v", err)
		response.Err(ctx, response.ERR_GIL_QUESTION)
		return
	}

	response.Success(ctx, result)
}

// GetAICourseAndPracticeList 获取业务树节点对应的 AI 课和巩固练习列表
func (c *TaskController) GetAICourseAndPracticeList(ctx *gin.Context) {
	// 获取 Query 请求参数
	subject := utils.Atoi64(ctx.Query("subject"))             // 学科
	bizTreeID := utils.Atoi64(ctx.Query("bizTreeId"))         // 业务树 ID
	bizTreeNodeID := utils.Atoi64(ctx.Query("bizTreeNodeId")) // 业务树节点 ID
	if !consts.SubjectExists(subject) {
		response.ParamError(ctx, response.ERR_SUBJECT)
		return
	}
	if bizTreeID == 0 || bizTreeNodeID == 0 {
		response.ParamError(ctx, response.ERR_BIZ_TREE)
		return
	}

	// 检查教师是否具备学科的权限
	if !c.teacherMiddleware.HasTaskCreationSubjectPermission(ctx, subject) {
		response.Forbidden(ctx)
		return
	}

	// 获取业务树节点 ID 下的所有叶子节点
	leafNodes, err := c.questionAPI.GetBizTreeLeafNodes(ctx, bizTreeID, bizTreeNodeID)
	if err != nil {
		c.log.Error(ctx, "获取业务树叶子节点失败: %v", err)
		response.Err(ctx, response.ERR_GIL_QUESTION)
		return
	}

	// 批量获取叶子节点的巩固练习信息
	bizTreeLeafNodeIDs := make([]int64, 0, len(leafNodes))
	for _, leafNode := range leafNodes {
		bizTreeLeafNodeIDs = append(bizTreeLeafNodeIDs, leafNode.BizTreeNodeId)
	}
	c.log.Debug(ctx, "业务树节点 %d 下的所有叶子节点: %v", bizTreeNodeID, bizTreeLeafNodeIDs)
	practiceInfoList, err := c.questionAPI.GetPracticeListByBizTreeNodeIDs(ctx, bizTreeLeafNodeIDs)
	if err != nil {
		c.log.Error(ctx, "获取巩固练习列表失败: %v", err)
		response.Err(ctx, response.ERR_GIL_QUESTION)
		return
	}

	// 构建结果，包含所有叶子节点，但只在有有效练习信息时才包含 Practice 字段
	res := api.GetAICourseAndPracticeListResponse{}
	practiceIds := []string{}
	for i, leafNode := range leafNodes {
		// 创建基本项，包含 BizTreeNodeID 和 BizTreeNodeName
		item := api.AICourseAndPracticeInfo{
			BizTreeNodeID:   leafNode.BizTreeNodeId,
			BizTreeNodeName: leafNode.BizTreeNodeName,
		}

		// 如果有有效的练习信息，则添加 Practice 字段
		if i < len(practiceInfoList) && practiceInfoList[i] != nil && practiceInfoList[i].QuestionSetId != 0 {
			practiceId := utils.I64ToStr(practiceInfoList[i].QuestionSetId)
			practiceIds = append(practiceIds, practiceId)

			item.Practice = api.Practice{
				ID: practiceInfoList[i].QuestionSetId,
			}
		}

		res = append(res, item)
	}

	// 如果没有有效的练习信息，直接返回结果
	if len(practiceIds) == 0 {
		response.Success(ctx, res)
		return
	}

	// 获取每个巩固练习已布置的班级ID
	teacherID := c.teacherMiddleware.ExtractTeacherID(ctx)
	req := &dto.TaskResourceAssignedClassIDsRequest{
		TeacherID:    teacherID,
		Subject:      subject,
		TaskType:     consts.TASK_TYPE_COURSE,
		ResourceType: consts.RESOURCE_TYPE_PRACTICE,
		ResourceIDs:  practiceIds,
		GroupType:    consts.TASK_GROUP_TYPE_CLASS,
	}
	resourceID2ClassIDMap, err := c.taskService.GetResourceAssignedClassIDs(ctx, req)
	if err != nil {
		c.log.Error(ctx, "获取巩固练习已布置的班级ID失败: %v", err)
		response.Err(ctx, response.ERR_POSTGRESQL)
		return
	}

	// 设置 AssignedClassIDs
	for i, item := range res {
		if item.Practice.ID != 0 {
			res[i].Practice.AssignedClassIDs = resourceID2ClassIDMap[utils.I64ToStr(item.Practice.ID)]
		}
	}

	response.Success(ctx, res)
}

// GetQuestionSetDetailByID 通过题集ID获取题集详情
func (c *TaskController) GetQuestionSetDetailByID(ctx *gin.Context) {
	questionSetID := utils.Atoi64(ctx.Query("questionSetId"))
	if questionSetID == 0 {
		response.ParamError(ctx)
		return
	}

	// 根据题集 ID 从题库获取题集详情
	result, err := c.questionAPI.GetQuestionSetByID(ctx, questionSetID)
	if err != nil {
		c.log.Error(ctx, "获取题集详情失败: %v", err)
		response.Err(ctx, response.ERR_GIL_QUESTION)
		return
	}

	// 统计题型数量
	questionTypeMap := make(map[int64]struct{})
	for _, group := range result.QuestionGroupStableInfoList {
		for _, question := range group.QuestionInfoList {
			questionTypeMap[question.QuestionInfo.QuestionType] = struct{}{}
		}
	}

	res := api.GetQuestionSetDetailResponse{
		QuestionSetStableInfo: *result,
		QuestionTypeCount:     int64(len(questionTypeMap)),
	}

	response.Success(ctx, res)
}

// GetQuestionEnums 获取查询题目支持的下拉框枚举值
func (c *TaskController) GetQuestionEnums(ctx *gin.Context) {
	// 从题库获取查询题目的下拉框枚举值
	result, err := c.questionAPI.GetQuestionEnums(ctx)
	if err != nil {
		c.log.Error(ctx, "获取查询题目下拉框枚举值失败: %v", err)
		response.Err(ctx, response.ERR_GIL_QUESTION)
		return
	}

	response.Success(ctx, result)
}

// GetQuestionList 获取题目列表
func (c *TaskController) GetQuestionList(ctx *gin.Context) {
	// 验证请求参数
	var reqBody api.GetQuestionListRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		c.log.Error(ctx, "绑定请求参数失败: %v", err)
		response.ParamError(ctx)
		return
	}
	if err := reqBody.Validate(); err != nil {
		c.log.Error(ctx, "验证请求参数失败: %v", err)
		response.ParamError(ctx, *err)
		return
	}

	// 检查教师是否具备学科的权限
	if !c.teacherMiddleware.HasTaskCreationSubjectPermission(ctx, reqBody.Subject) {
		response.Forbidden(ctx)
		return
	}

	// 中间件获取老师的学段
	reqBody.Phase = c.teacherMiddleware.ExtractTeacherPhase(ctx) // 学段

	// 实现题目列表获取逻辑
	result, err := c.questionAPI.GetQuestionList(ctx, &reqBody)
	if err != nil {
		c.log.Error(ctx, "获取题目列表失败: %v", err)
		response.Err(ctx, response.ERR_GIL_QUESTION)
		return
	}

	response.Success(ctx, result)
}

// GetQuestionListByIDs 通过ID列表查询题目详情
func (c *TaskController) GetQuestionListByIDs(ctx *gin.Context) {
	// 验证请求参数
	var reqBody api.GetQuestionListByIDsRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		c.log.Error(ctx, "绑定请求参数失败: %v", err)
		response.ParamError(ctx)
		return
	}
	res := []*itl.Question{}
	if len(reqBody.QuestionIDs) == 0 {
		response.Success(ctx, res)
		return
	}

	// 从题库获取题目详情
	res, err := c.questionAPI.GetQuestionListByID(ctx, reqBody.QuestionIDs, true)
	if err != nil {
		c.log.Error(ctx, "获取题目详情失败: %v", err)
		response.Err(ctx, response.ERR_GIL_QUESTION)
		return
	}

	response.Success(ctx, res)
}
