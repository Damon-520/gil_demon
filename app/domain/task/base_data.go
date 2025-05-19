package task

import (
	"context"
	"errors"
	"slices"
	"sort"
	"strings"

	"gil_teacher/app/consts"
	dao_task "gil_teacher/app/dao/task"
	"gil_teacher/app/model/api"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/model/itl"
	"gil_teacher/app/utils"
)

// 任务布置数据，包括任务基础数据、任务资源数据、布置对象基础数据、布置对象报告数据
type singleTask struct {
	*TaskReportHandler
	*taskData
}

// 任务布置数据列表，包括任务基础数据、任务资源数据、布置对象基础数据、布置对象报告数据
type multiTasks struct {
	*TaskReportHandler
	taskDataMap map[int64]*taskData // taskID -> taskData 任务数据
}

type taskData struct {
	task             *dao_task.Task                    // 任务基础信息
	taskResources    map[string]*dao_task.TaskResource // resource_key -> TaskResource 任务资源列表
	resourceQuestion *resourceQuestion                 // 资源题目数据
	assignDataMap    map[int64]*assignData             // assignID -> assignData 布置对象数据
}

type resourceQuestion struct {
	questionMap      map[string]map[string]*itl.Question // resource_key -> question_id -> question
	questionIDs      []string                            // 按布置顺序排列的题目ID列表（题目、题集中的题目）
	totalQuestionNum int64
}

// 单个任务布置的基础数据
type assignData struct {
	assign          *dao_task.TaskAssign                // taskAssign 任务布置信息
	report          *dao_task.TaskReport                // taskReport 任务布置报告信息
	questionAnswers map[string]*dao_task.QuestionAnswer // question_key -> QuestionAnswer 布置粒度的题目答题数据
	classInfo       *itl.ClassInfo                      // 班级信息，当布置对象为班级时存在
}

// 获取单个任务的原始数据
func (h *TaskReportHandler) getTaskData(ctx context.Context, taskId int64) (*singleTask, error) {
	task, err := h.taskService.GetTaskByID(ctx, taskId)
	if err != nil {
		h.log.Error(ctx, "[getTaskData] GetTasksByIDs error:%v", err)
		return nil, err
	}

	return &singleTask{
		TaskReportHandler: h,
		taskData: &taskData{
			task: task,
		},
	}, nil
}

// 获取多个任务的原始数据
func (h *TaskReportHandler) getTasksData(ctx context.Context, taskIds []int64, taskAssigns []*dao_task.TaskAssign) (map[int64]*taskData, error) {
	if len(taskIds) == 0 {
		return nil, errors.New("taskIds is empty")
	}

	tasks, err := h.taskService.GetTasksByIDs(ctx, taskIds)
	if err != nil {
		h.log.Error(ctx, "[getTaskData] GetTasksByIDs error:%v", err)
		return nil, err
	}

	taskAssignDataMap := make(map[int64]*taskData)
	for _, task := range tasks {
		taskAssignDataMap[task.TaskID] = &taskData{
			task: task,
		}
	}

	for _, taskAssign := range taskAssigns {
		if taskAssignDataMap[taskAssign.TaskID] == nil {
			h.log.Warn(ctx, "[getTasksData] taskData is nil, taskID:%d", taskAssign.TaskID)
			continue
		}
		if taskAssignDataMap[taskAssign.TaskID].assignDataMap == nil {
			taskAssignDataMap[taskAssign.TaskID].assignDataMap = make(map[int64]*assignData)
		}
		taskAssignDataMap[taskAssign.TaskID].assignDataMap[taskAssign.AssignID] = &assignData{
			assign: taskAssign,
		}
	}

	return taskAssignDataMap, nil
}

// 作业题目排序，不同任务类型排序方式不同
// 按 QuestionAnswer.QuestionIndex 顺序返回
// 返回：
//  1. 排序后的题目列表
//  2. 排序后的题目ID列表
func (h *TaskReportHandler) sortQuestions(taskData *taskData, questionAnswers []*api.QuestionAnswer, query *dto.TaskReportCommonQuery) ([]*api.QuestionAnswer, []string) {
	switch taskData.task.TaskType {
	case consts.TASK_TYPE_COURSE: // 课程任务，按题库中的顺序返回，注意questions的数据不一定连续，需要重新从 1 开始编号
		sort.Slice(questionAnswers, func(i, j int) bool {
			return questionAnswers[i].Question.QuestionContentFormat.QuestionOrder <
				questionAnswers[j].Question.QuestionContentFormat.QuestionOrder
		})
		for i, qa := range questionAnswers {
			qa.QuestionIndex = int64(i + 1)
		}
	case consts.TASK_TYPE_HOMEWORK: // 作业任务，按题目添加顺序返回
		taskResources := make([]*dao_task.TaskResource, 0)
		for _, resource := range taskData.taskResources {
			taskResources = append(taskResources, resource)
		}
		sort.Slice(taskResources, func(i, j int) bool {
			return taskResources[i].ID < taskResources[j].ID
		})
		qaMap := make(map[string]*api.QuestionAnswer)
		for _, qa := range questionAnswers {
			qaMap[qa.Question.QuestionId] = qa
		}
		idx := 1
		for _, resource := range taskResources {
			if qa, ok := qaMap[resource.ResourceID]; ok {
				qa.QuestionIndex = int64(idx)
				idx++
			}
		}
	default: // 其他任务类型，不排序
	}

	// 如果 query 中的 sortKey 不为空，则按 sortKey 排序，排序方式根据 sortType(默认 asc) 决定，否则按题目序号排序
	isSorted := false
	if query != nil && query.SortKey != "" {
		switch query.SortKey {
		case "answerCount":
			sort.Slice(questionAnswers, func(i, j int) bool {
				if query.SortType == consts.SortTypeDesc {
					return questionAnswers[i].AnswerCount > questionAnswers[j].AnswerCount
				} else {
					return questionAnswers[i].AnswerCount < questionAnswers[j].AnswerCount
				}
			})
			isSorted = true
		case "incorrectCount":
			sort.Slice(questionAnswers, func(i, j int) bool {
				if query.SortType == consts.SortTypeDesc {
					return questionAnswers[i].IncorrectCount > questionAnswers[j].IncorrectCount
				} else {
					return questionAnswers[i].IncorrectCount < questionAnswers[j].IncorrectCount
				}
			})
			isSorted = true
		}
	}

	// 如果没有执行自定义排序，则按题目序号排序
	if !isSorted {
		sort.Slice(questionAnswers, func(i, j int) bool {
			return questionAnswers[i].QuestionIndex < questionAnswers[j].QuestionIndex
		})
	}

	questionIDs := make([]string, 0)
	for _, qa := range questionAnswers {
		questionIDs = append(questionIDs, qa.Question.QuestionId)
	}

	return questionAnswers, questionIDs
}

// 获取指定任务指定布置指定对象的全部学生
// 数据缓存一天，按班级存储全部学生hashmap
//
//	map[classID]*itl.ClassInfo
func (h *TaskReportHandler) getClassStudents(ctx context.Context, schoolID int64, classIDs []int64) (map[int64]*itl.ClassInfo, error) {
	classMap, err := h.ucenterService.GetClassStudent(ctx, schoolID, classIDs)
	if err != nil {
		h.log.Error(ctx, "[getStudents] GetClassStudent failed, classID:%v, error:%v", classIDs, err)
		return nil, err
	}

	return classMap, nil
}

// 筛选符合要求的题目. 返回：
//  1. 按资源ID和资源类型分类的题目数据 resource_key -> question_id -> question
//  2. 题目ID列表(布置顺序)
//  3. 错误信息
func (h *TaskReportHandler) getQuestions(
	query *dto.TaskReportCommonQuery,
	resourceQuestions map[string][]*itl.Question, // resource_type = 103
	resourcePractices map[int64][]*itl.Question, // resource_type = 102
) (map[string]map[string]*itl.Question, []string, error) {
	// 题目类型和题目关键词筛选
	filteredQuestionMap := make(map[string]map[string]*itl.Question)
	filteredQuestionIDs := make([]string, 0)
	for resourceID, questions := range resourceQuestions {
		filtered := h.filterQuestions(questions, query)
		for _, q := range filtered {
			resourceKey := utils.JoinList([]any{resourceID, consts.RESOURCE_TYPE_QUESTION}, consts.CombineKey)
			if _, ok := filteredQuestionMap[resourceKey]; !ok {
				filteredQuestionMap[resourceKey] = make(map[string]*itl.Question)
			}
			filteredQuestionMap[resourceKey][q.QuestionId] = q
			filteredQuestionIDs = append(filteredQuestionIDs, q.QuestionId)
		}
	}
	for questionSetID, practices := range resourcePractices {
		filtered := h.filterQuestions(practices, query)
		for _, q := range filtered {
			resourceKey := utils.JoinList([]any{questionSetID, consts.RESOURCE_TYPE_PRACTICE}, consts.CombineKey)
			if _, ok := filteredQuestionMap[resourceKey]; !ok {
				filteredQuestionMap[resourceKey] = make(map[string]*itl.Question)
			}
			filteredQuestionMap[resourceKey][q.QuestionId] = q
			filteredQuestionIDs = append(filteredQuestionIDs, q.QuestionId)
		}
	}

	return filteredQuestionMap, filteredQuestionIDs, nil
}

// 筛选符合要求的题目. 返回：
//  1. 按资源ID和资源类型分类的题目数据
//  2. 题目ID列表
//  3. 错误信息
func (h *TaskReportHandler) filterQuestions(questions []*itl.Question, query *dto.TaskReportCommonQuery) []*itl.Question {
	filteredQuestions := make([]*itl.Question, 0)
	for _, q := range questions {
		// 全部，或者指定题目类型：单选、多选、填空
		matchResource := false
		matchKeyword := false
		if query == nil || query.QuestionType == int64(consts.QUESTION_TYPE_ALL) || q.QuestionInfoEntity.QuestionType == query.QuestionType {
			matchResource = true
		}
		// 关键词匹配
		if query == nil || query.Keyword == "" || strings.Contains(q.QuestionContentFormat.QuestionStem, query.Keyword) {
			matchKeyword = true
		}
		if matchResource && matchKeyword {
			filteredQuestions = append(filteredQuestions, q)
		}
	}
	return filteredQuestions
}

func (h *TaskReportHandler) getStudentList(students []*itl.StudentInfo) []*api.Student {
	studentList := make([]*api.Student, 0)
	for _, student := range students {
		studentList = append(studentList, &api.Student{
			StudentID:   student.ID,
			StudentName: student.Name,
			Avatar:      student.Avatar,
		})
	}
	return studentList
}

/*********************************************************
				指定任务的学生维度统计数据
**********************************************************/
// 指定任务的学生维度统计数据查询条件
type assignReportQuery struct {
	TaskID       int64
	AssignID     int64
	ResourceID   string
	ResourceType int64
	StudentIDs   []int64
}

func (h *TaskReportHandler) getStudentIdList(studentMap []*itl.StudentInfo) []int64 {
	studentIDList := make([]int64, 0)
	for _, student := range studentMap {
		studentIDList = append(studentIDList, student.ID)
	}
	return studentIDList
}

/******************************************
				单任务处理
******************************************/

// 获取任务资源数据，没有则从 db 获取
func (h *singleTask) getResources(ctx context.Context, taskId int64) error {
	if h.task == nil {
		return errors.New("任务数据不存在")
	}

	// 已经获取过任务资源数据
	if len(h.taskResources) > 0 {
		return nil
	}

	taskResourceMap, resourceIDs, err := h.taskResourceService.GetTaskResourceByTaskID(ctx, taskId)
	if err != nil {
		h.log.Error(ctx, "[getTaskResource] GetTaskResourcesByTaskIDs error:%v", err)
		return err
	}
	h.taskResources = taskResourceMap
	for _, resource := range taskResourceMap {
		// 资源类型是题目
		switch resource.ResourceType {
		case consts.RESOURCE_TYPE_QUESTION:
			if h.resourceQuestion == nil {
				h.resourceQuestion = &resourceQuestion{
					questionIDs: resourceIDs,
				}
			} else {
				h.resourceQuestion.questionIDs = resourceIDs
			}
		default:
			// TODO 其它资源类型 以后处理
		}
	}
	return nil
}

// 检查有无任务资源题目数据，没有则从题库获取
func (h *singleTask) getResourceQuestions(ctx context.Context, taskId int64, query *dto.TaskReportCommonQuery) error {
	if h.task == nil {
		return errors.New("task is nil")
	}

	// 未获取过任务资源数据，则先获取任务资源数据
	if len(h.taskResources) == 0 {
		if err := h.getResources(ctx, taskId); err != nil {
			h.log.Error(ctx, "[getResourceQuestions] getTaskResource failed, taskID:%d, err:%v", taskId, err)
			return err
		}
	}

	// 已经获取过任务资源题目数据
	if h.resourceQuestion != nil {
		return nil
	}

	if len(h.taskResources) == 0 {
		return errors.New("task has no resources")
	}

	questionIDs := make([]string, 0)
	practiceIDs := make([]int64, 0)
	for _, resource := range h.taskResources {
		switch resource.ResourceType {
		case consts.RESOURCE_TYPE_QUESTION:
			questionIDs = append(questionIDs, resource.ResourceID)
		case consts.RESOURCE_TYPE_PRACTICE:
			practiceIDs = append(practiceIDs, utils.Atoi64(resource.ResourceID))
		default:
			h.log.Error(ctx, "[getResourceQuestions] resource type not supported: %d", resource.ResourceType)
		}
	}

	// 从内容平台查询全部题目信息
	resourceQuestions, resourcePractices, totalQuestionCount, err := h.questionAPI.GetResources(ctx, questionIDs, practiceIDs)
	if err != nil {
		h.log.Error(ctx, "[getResourceQuestions] GetResources error:%v", err)
		return err
	}

	// 筛选符合要求的题目
	filteredQuestionMap, filteredQuestionIDs, err := h.getQuestions(query, resourceQuestions, resourcePractices)
	if err != nil {
		h.log.Error(ctx, "[getTaskResources] getQuestions error:%v", err)
		return err
	}

	h.resourceQuestion = &resourceQuestion{
		questionMap:      filteredQuestionMap,
		questionIDs:      filteredQuestionIDs,
		totalQuestionNum: totalQuestionCount,
	}
	return nil
}

// 获取任务资源报告，如果是作业任务，则合并为一个资源报告 TODO
func (h *singleTask) getResourceReports(ctx context.Context, tarAssignID int64) ([]*api.TaskResourceReport, error) {
	if h.task == nil {
		return nil, errors.New("任务数据不存在")
	}

	// 获取任务资源数据
	if err := h.getResources(ctx, h.task.TaskID); err != nil {
		h.log.Error(ctx, "[getResourceReports] getResources failed, taskID:%d, error:%v", h.task.TaskID, err)
		return nil, err
	}

	// 获取任务布置数据
	if err := h.getAssignData(ctx, h.task.TaskID, tarAssignID); err != nil {
		h.log.Error(ctx, "[getResourceReports] getAssignData failed, taskID:%d, assignID:%d, error:%v", h.task.TaskID, tarAssignID, err)
		return nil, err
	}

	// 获取任务布置报告数据
	if err := h.getAssignReport(ctx, tarAssignID); err != nil {
		h.log.Error(ctx, "[getResourceReports] getAssignReport failed, taskID:%d, assignID:%d, error:%v", h.task.TaskID, tarAssignID, err)
		return nil, err
	}

	assignReport := h.assignDataMap[tarAssignID].report
	resourceReports := make([]*api.TaskResourceReport, 0)
	for resourceKey := range h.taskResources {
		parts := strings.Split(resourceKey, consts.CombineKey)
		resourceID := parts[0]
		resourceType := utils.Atoi64(parts[1])
		tmp := &api.TaskResourceReport{
			ResourceID:   resourceID,
			ResourceType: resourceType,
			ResourceName: consts.GetResourceTypeName(resourceType),
		}
		if assignReport != nil && assignReport.ResourceReportDetail != nil {
			if resourceReport, ok := assignReport.ResourceReportDetail[resourceKey]; ok {
				tmp.CompletionRate = resourceReport.CompletedProgress
				tmp.CorrectRate = resourceReport.AccuracyRate
				tmp.NeedAttentionQuestionNum = resourceReport.NeedAttentionNum
				tmp.NeedAttentionUserNum = resourceReport.NeedAttentionUserNum
				tmp.AverageCostTime = resourceReport.AverageCostTime
			}
		}
		resourceReports = append(resourceReports, tmp)
	}
	return resourceReports, nil
}

// 检查有无任务布置数据，没有则从 db 获取
func (h *singleTask) getAssignData(ctx context.Context, taskId, assignId int64) error {
	// 已经获取过任务布置数据
	if assignId == 0 {
		if len(h.assignDataMap) > 0 {
			return nil
		}
	} else {
		if _, ok := h.assignDataMap[assignId]; ok {
			return nil
		}
	}

	taskAssigns, err := h.taskAssignService.GetTaskAssignInfo(ctx, taskId, assignId)
	if err != nil {
		h.log.Error(ctx, "[getTaskAssignData] GetTaskAssignInfo failed, taskID:%d, assignID:%d, error:%v", taskId, assignId, err)
		return err
	}

	h.assignDataMap = make(map[int64]*assignData)
	for _, assign := range taskAssigns {
		h.assignDataMap[assign.AssignID] = &assignData{
			assign: assign,
		}
	}
	return nil
}

// 获取任务布置对象的学生信息(如果是按班级布置，则同时获取班级信息)
func (h *singleTask) getAssignClassInfo(ctx context.Context, taskId, tarAssignId int64) error {
	// 依赖布置基础数据
	if err := h.getAssignData(ctx, taskId, tarAssignId); err != nil {
		h.log.Error(ctx, "[getTaskAssignClassInfo] getTaskAssignData failed, taskID:%d, assignID:%d, error:%v", taskId, tarAssignId, err)
		return err
	}

	// 获取不到数据，直接返回
	if len(h.assignDataMap) == 0 || (tarAssignId != 0 && h.assignDataMap[tarAssignId] == nil) {
		h.log.Warn(ctx, "[getTaskAssignClassInfo] getTaskAssignData empty, taskID:%d, assignID:%d", taskId, tarAssignId)
		return nil
	}

	// 无学生信息的布置 id 列表
	noStudentAssignIds := make([]int64, 0)
	for assignId, assignData := range h.assignDataMap {
		if tarAssignId != 0 && assignId != tarAssignId {
			continue
		}

		// 没有班级信息，则需要获取
		if assignData.classInfo == nil {
			noStudentAssignIds = append(noStudentAssignIds, assignId)
		}
	}

	// 如果学生信息都获取到了，则直接返回
	if len(noStudentAssignIds) == 0 {
		return nil
	}

	// 获取任务布置对象的班级信息
	taskClassMap := make(map[int64]*itl.ClassInfo)       // assignID -> ClassInfo
	taskStudentMap := make(map[int64][]*itl.StudentInfo) // assignID -> []StudentInfo
	for assignId, assignData := range h.assignDataMap {
		if tarAssignId != 0 && assignId != tarAssignId {
			continue
		}

		assign := assignData.assign
		switch assign.GroupType {
		case consts.TASK_GROUP_TYPE_CLASS:
			assignClass, err := h.getClassStudents(ctx, assign.SchoolID, []int64{assign.GroupID})
			if err != nil {
				h.log.Error(ctx, "[getTaskAssignClassInfo] getStudents failed, taskID:%d, assignID:%d, error:%v", taskId, assignId, err)
				return err
			}
			if classInfo, ok := assignClass[assign.GroupID]; ok {
				taskClassMap[assignId] = classInfo
			}
		case consts.TASK_GROUP_TYPE_STUDENT, consts.TASK_GROUP_TYPE_TEMP:
			studentIds, err := h.taskAssignService.GetTaskAssignStudents(ctx, assign.TaskID, assign.AssignID)
			if err != nil {
				h.log.Error(ctx, "[getTaskAssignClassInfo] GetTaskAssignStudents failed, taskID:%d, assignID:%d, error:%v", taskId, assignId, err)
				return err
			}
			// TODO 从运营平台获取学生信息
			students := make([]*itl.StudentInfo, 0)
			for _, studentID := range studentIds {
				students = append(students, &itl.StudentInfo{
					ID:     studentID,
					Name:   "", // TODO 临时小组或者学生小组，以后补充
					Avatar: "", // TODO 临时小组或者学生小组，以后补充
				})
			}
			if len(students) > 0 {
				taskStudentMap[assignId] = students
			}
		}
	}

	if len(taskClassMap) == 0 && len(taskStudentMap) == 0 {
		h.log.Warn(ctx, "[getTaskAssignClassInfo] getTaskAssignClassInfo empty, taskID:%d, assignID:%d", taskId, tarAssignId)
		return errors.New("学生信息不存在")
	}

	for assignId, classInfo := range taskClassMap {
		h.assignDataMap[assignId].classInfo = classInfo
	}
	for assignId, studentList := range taskStudentMap {
		h.assignDataMap[assignId].classInfo = &itl.ClassInfo{
			Students: studentList,
		}
	}
	return nil
}

// 获取任务布置对象的报告数据(作业维度，不处理学生维度的数据)
func (h *singleTask) getAssignReport(ctx context.Context, tarAssignId int64) error {
	if h.task == nil {
		return errors.New("任务数据不存在")
	}

	taskId := h.task.TaskID
	// 依赖布置基础数据
	if len(h.assignDataMap) == 0 {
		if err := h.getAssignData(ctx, taskId, tarAssignId); err != nil {
			h.log.Error(ctx, "[getTaskAssignReport] getTaskAssignData failed, taskID:%d, assignID:%d, error:%v", taskId, tarAssignId, err)
			return err
		}
	}

	// 获取不到数据，直接返回
	if len(h.assignDataMap) == 0 || (tarAssignId != 0 && h.assignDataMap[tarAssignId] == nil) {
		h.log.Warn(ctx, "[getTaskAssignReport] getTaskAssignData empty, taskID:%d, assignID:%d", taskId, tarAssignId)
		return errors.New("缺少任务布置数据")
	}

	noReportAssignIds := make([]int64, 0)
	for assignId := range h.assignDataMap {
		if tarAssignId != 0 && assignId != tarAssignId {
			continue
		}
		if h.assignDataMap[assignId].report == nil {
			noReportAssignIds = append(noReportAssignIds, assignId)
		}
	}

	taskAssignStatsMap, err := h.taskReportService.GetTaskAssignsStats(ctx, taskId, noReportAssignIds)
	if err != nil {
		h.log.Error(ctx, "[getTaskAssignReport] GetTaskStatsByAssignIDs error:%v", err)
		return err
	}

	for _, report := range taskAssignStatsMap {
		h.assignDataMap[report.AssignID].report = report
	}
	return nil
}

// 获取任务布置的答题正确率
func (h *singleTask) getAssignAnswerAccuracy(ctx context.Context, tarTaskId, tarAssignId int64, questionIdMap map[string][]string) error {
	if h.assignDataMap == nil || h.assignDataMap[tarAssignId] == nil {
		if err := h.getAssignData(ctx, tarTaskId, tarAssignId); err != nil {
			h.log.Error(ctx, "[getAssignAnswerAccuracy] getAssignData failed, taskID:%d, assignID:%d, error:%v", tarTaskId, tarAssignId, err)
			return err
		}
	}

	if h.assignDataMap[tarAssignId] == nil {
		return errors.New("任务布置数据不存在")
	}

	taskAssignAnswerAccuracyMap, err := h.taskReportService.GetTaskAnswerAccuracyByResource(ctx, tarTaskId, tarAssignId, questionIdMap)
	if err != nil {
		h.log.Error(ctx, "[getAssignAnswerAccuracy] GetTaskAnswerAccuracyByResource error:%v", err)
		return err
	}

	if len(h.assignDataMap[tarAssignId].questionAnswers) > 0 {
		return nil
	}

	h.assignDataMap[tarAssignId].questionAnswers = make(map[string]*dao_task.QuestionAnswer)
	for questionKey, answerAccuracy := range taskAssignAnswerAccuracyMap {
		h.assignDataMap[tarAssignId].questionAnswers[questionKey] = answerAccuracy
	}
	return nil
}

// 指定任务的学生维度统计数据
func (h *singleTask) getStudentReports(ctx context.Context, students []*itl.StudentInfo, query *dto.TaskAssignAnswersQuery, pageInfo *consts.APIReqeustPageInfo) ([]*api.TaskStudentReport, int64, error) {
	// 获取答题汇总信息
	taskStudentReports, total, err := h.taskReportService.GetTaskAssignStudentReports(ctx, query.TaskID, query.AssignID, query.StudentIDs, pageInfo)
	if err != nil {
		h.log.Error(ctx, "[getTaskStudentReports] GetTaskAssignStudentReports failed, taskID:%d, assignID:%d, studentIDs:%v, error:%v", query.TaskID, query.AssignID, query.StudentIDs, err)
		return nil, 0, err
	}

	// 获取答题详情
	studentAnswerMap, err := h.taskReportService.GetTaskAssignAnswers(ctx, query, pageInfo)
	if err != nil {
		h.log.Error(ctx, "[getTaskStudentReports] GetTaskAssignAnswers failed, taskID:%d, assignID:%d, studentIDs:%v, error:%v", query.TaskID, query.AssignID, query.StudentIDs, err)
		return nil, 0, err
	}

	if err := h.getResourceQuestions(ctx, query.TaskID, &query.TaskReportCommonQuery); err != nil {
		h.log.Error(ctx, "[getTaskStudentReports] getResourceQuestions failed, taskID:%d, assignID:%d, studentIDs:%v, error:%v", query.TaskID, query.AssignID, query.StudentIDs, err)
		return nil, 0, err
	}

	// 提取题目的难度数据
	questionDifficultyMap := make(map[string]int64)
	for _, questions := range h.resourceQuestion.questionMap {
		for questionId, question := range questions {
			questionDifficultyMap[questionId] = question.QuestionDifficult
		}
	}

	// 处理报告数据, 每个学生的数据都需要返回，如果缺少报告数据，则用默认值填充
	studentReports := make([]*api.TaskStudentReport, 0, len(taskStudentReports))
	for _, student := range students {
		report, ok := taskStudentReports[student.ID]
		if !ok {
			report = &dao_task.TaskStudentsReport{
				StudentID: student.ID,
				TaskID:    query.TaskID,
				AssignID:  query.AssignID,
			}
		}
		var reportData dao_task.ResourceDetailReport
		// 如果指定了资源ID，则查询资源报告
		if query.ResourceID != "" {
			resourceReports := report.ResourceReport
			if resourceReports != nil {
				resourceKey := utils.JoinList([]any{query.ResourceID, query.ResourceType}, consts.CombineKey)
				if resouceReport, ok := resourceReports[resourceKey]; ok {
					reportData = resouceReport
				}
			}
		} else {
			reportData = report.ResourceDetailReport
		}

		// 提取学生作答的题目难度信息进行计算
		difficultyDegree := float64(consts.QUESTION_DIFFICULT_MEDIUM) // 默认中等难度
		if studentAnswers, ok := studentAnswerMap[student.ID]; ok {
			totalDifficultyDegree := int64(0)
			for _, answer := range studentAnswers {
				totalDifficultyDegree += questionDifficultyMap[answer.QuestionID]
			}
			difficultyDegree = utils.F64Div(float64(totalDifficultyDegree), float64(len(studentAnswers)), 1)
		}

		studentReports = append(studentReports, &api.TaskStudentReport{
			StudentID:        student.ID,
			StudyScore:       report.StudyScore,
			DifficultyDegree: difficultyDegree,
			Tags:             h.getStudentTags(report),
			Progress:         reportData.CompletedProgress,
			AccuracyRate:     reportData.AccuracyRate,
			IncorrectNum:     reportData.IncorrectCount,
			AnswerNum:        reportData.AnswerCount,
			CostTime:         reportData.CostTime,
		})
	}

	sorted := false
	if query.SortKey != "" && slices.Contains(consts.ReportSortKeys, query.SortKey) {
		switch query.SortKey {
		case "studyScore": // 学习分
			sorted = true
			sort.Slice(studentReports, func(i, j int) bool {
				if query.SortType == consts.SortTypeDesc {
					return studentReports[i].StudyScore > studentReports[j].StudyScore
				}
				return studentReports[i].StudyScore < studentReports[j].StudyScore
			})
		case "progress": // 完成进度
			sorted = true
			sort.Slice(studentReports, func(i, j int) bool {
				if query.SortType == consts.SortTypeDesc {
					return studentReports[i].Progress > studentReports[j].Progress
				}
				return studentReports[i].Progress < studentReports[j].Progress
			})
		case "accuracyRate": // 正确率
			sorted = true
			sort.Slice(studentReports, func(i, j int) bool {
				if query.SortType == consts.SortTypeDesc {
					return studentReports[i].AccuracyRate > studentReports[j].AccuracyRate
				}
				return studentReports[i].AccuracyRate < studentReports[j].AccuracyRate
			})
		case "answerCount": // 答题数
			sorted = true
			sort.Slice(studentReports, func(i, j int) bool {
				if query.SortType == consts.SortTypeDesc {
					return studentReports[i].AnswerNum > studentReports[j].AnswerNum
				}
				return studentReports[i].AnswerNum < studentReports[j].AnswerNum
			})
		}
	}

	if !sorted {
		// 排序，默认按学生 id 排序
		sort.Slice(studentReports, func(i, j int) bool {
			return studentReports[i].StudentID < studentReports[j].StudentID
		})
	}
	return studentReports, total, nil
}

/******************************
			多任务数据
******************************/

// 获取任务资源数据，没有则从 db 获取，如果 taskIds 为空，则获取所有任务的资源数据
func (h *multiTasks) getResource(ctx context.Context, taskIds []int64) error {
	if len(h.taskDataMap) == 0 {
		return errors.New("task is nil")
	}

	// 未获取过任务资源数据的 taskID 列表
	noResourceTaskIds := make([]int64, 0)
	for taskID := range h.taskDataMap {
		if len(taskIds) > 0 && !slices.Contains(taskIds, taskID) {
			continue
		}

		if _, ok := h.taskDataMap[taskID]; !ok {
			noResourceTaskIds = append(noResourceTaskIds, taskID)
		}
	}

	if len(noResourceTaskIds) == 0 {
		return nil
	}

	taskResourceMap, taskResourceIDsMap, err := h.taskResourceService.GetTaskResourcesByTaskIDs(ctx, noResourceTaskIds)
	if err != nil {
		h.log.Error(ctx, "[getTaskResource] GetTaskResourcesByTaskIDs error:%v", err)
		return err
	}
	for taskID, resourceMap := range taskResourceMap {
		h.taskDataMap[taskID].taskResources = resourceMap
	}
	for taskID := range taskResourceMap {
		if h.taskDataMap[taskID].resourceQuestion == nil {
			h.taskDataMap[taskID].resourceQuestion = &resourceQuestion{
				questionIDs: taskResourceIDsMap[taskID],
			}
		} else {
			h.taskDataMap[taskID].resourceQuestion.questionIDs = taskResourceIDsMap[taskID]
		}
	}
	return nil
}

// 检查有无任务资源题目数据，没有则从题库获取，如果 taskIds 为空，则获取所有任务的资源题目数据
func (h *multiTasks) getResourceQuestions(ctx context.Context, tarTaskIds []int64, query *dto.TaskReportCommonQuery) error {
	if len(h.taskDataMap) == 0 {
		return errors.New("task is nil")
	}

	noResourceTaskIds := make([]int64, 0)         // 未获取过任务资源数据的 taskID 列表
	noResourceQuestionTaskIds := make([]int64, 0) // 未获取过任务资源题目数据的 taskID 列表

	for taskID := range h.taskDataMap {
		if len(tarTaskIds) > 0 && !slices.Contains(tarTaskIds, taskID) {
			continue
		}

		// 没有任务资源数据
		if h.taskDataMap[taskID].taskResources == nil {
			noResourceTaskIds = append(noResourceTaskIds, taskID)
		}

		// 没有任务资源题目数据
		if h.taskDataMap[taskID].resourceQuestion == nil {
			noResourceQuestionTaskIds = append(noResourceQuestionTaskIds, taskID)
		}
	}

	if len(noResourceTaskIds) == 0 && len(noResourceQuestionTaskIds) == 0 {
		return nil
	}

	// 获取任务资源数据
	if len(noResourceTaskIds) > 0 {
		if err := h.getResource(ctx, noResourceTaskIds); err != nil {
			h.log.Error(ctx, "[getResourceQuestions] getTaskResource failed, taskID:%d, err:%v", noResourceTaskIds, err)
			return err
		}
	}

	// 获取任务资源题目数据
	if len(noResourceQuestionTaskIds) > 0 {
		questionIDs := make([]string, 0)
		practiceIDs := make([]int64, 0)
		for _, taskId := range noResourceQuestionTaskIds {
			taskData := h.taskDataMap[taskId]
			for _, resource := range taskData.taskResources {
				switch resource.ResourceType {
				case consts.RESOURCE_TYPE_QUESTION:
					questionIDs = append(questionIDs, resource.ResourceID)
				case consts.RESOURCE_TYPE_PRACTICE:
					practiceIDs = append(practiceIDs, utils.Atoi64(resource.ResourceID))
				default:
					h.log.Error(ctx, "[getResourceQuestions] resource type not supported: %d", resource.ResourceType)
				}
			}

			// 从内容平台查询全部题目信息
			resourceQuestions, resourcePractices, totalQuestionCount, err := h.questionAPI.GetResources(ctx, questionIDs, practiceIDs)
			if err != nil {
				h.log.Error(ctx, "[getResourceQuestions] GetResources error:%v", err)
				return err
			}

			// 筛选符合要求的题目
			filteredQuestionMap, filteredQuestionIDs, err := h.getQuestions(query, resourceQuestions, resourcePractices)
			if err != nil {
				h.log.Error(ctx, "[getTaskResources] getQuestions error:%v", err)
				return err
			}

			h.taskDataMap[taskId].resourceQuestion = &resourceQuestion{
				questionMap:      filteredQuestionMap,
				questionIDs:      filteredQuestionIDs,
				totalQuestionNum: totalQuestionCount,
			}
		}
	}
	return nil
}

// 检查有无任务布置数据，没有则从 db 获取
// 1.获取全部任务的全部布置数据，tarTaskId = 0
// 2.获取指定任务的全部布置数据，tarTaskId != 0, tarAssignId = 0
// 3.获取指定任务指定布置的数据，tarTaskId != 0, tarAssignId != 0
func (h *multiTasks) getAssignData(ctx context.Context, tarTaskId, tarAssignId int64) error {
	if len(h.taskDataMap) == 0 || (tarTaskId != 0 && (h.taskDataMap[tarTaskId] == nil || h.taskDataMap[tarTaskId].task == nil)) {
		h.log.Warn(ctx, "[getTaskAssignData] taskData is nil, taskID:%d, assignID:%d", tarTaskId, tarAssignId)
		return errors.New("taskData is nil")
	}

	// 获取全部任务的全部布置数据
	if tarTaskId == 0 {
		noAssignTaskIds := make([]int64, 0) // 无布置数据的任务 id 列表
		for taskID := range h.taskDataMap {
			if h.taskDataMap[taskID].assignDataMap == nil {
				h.taskDataMap[taskID].assignDataMap = make(map[int64]*assignData)
				noAssignTaskIds = append(noAssignTaskIds, taskID)
			}
		}

		// 获取无布置数据的任务布置数据
		if len(noAssignTaskIds) > 0 {
			taskAssigns, err := h.taskAssignService.GetTaskAssignsByTaskIDs(ctx, noAssignTaskIds)
			if err != nil {
				h.log.Error(ctx, "[getTaskAssignData] GetTaskAssignsByTaskIDs failed, taskID:%d, error:%v", tarTaskId, err)
				return err
			}
			for taskId, assigns := range taskAssigns {
				for assignID, assign := range assigns {
					h.taskDataMap[taskId].assignDataMap[assignID] = &assignData{
						assign: assign,
					}
				}
			}
		}
	} else {
		// 无布置数据的布置 id 列表
		noAssignIds := make([]int64, 0)
		for taskID := range h.taskDataMap {
			if taskID != tarTaskId {
				continue
			}
			for assignID := range h.taskDataMap[taskID].assignDataMap {
				if assignID != tarAssignId {
					continue
				}
				if h.taskDataMap[taskID].assignDataMap[assignID].assign == nil {
					noAssignIds = append(noAssignIds, assignID)
				}
			}
		}

		// 获取无布置数据的任务布置数据
		if len(noAssignIds) > 0 {
			taskAssigns, err := h.taskAssignService.GetTaskAssigns(ctx, tarTaskId, noAssignIds)
			if err != nil {
				h.log.Error(ctx, "[getTaskAssignData] GetTaskAssignInfo failed, taskID:%d, assignID:%d, error:%v", tarTaskId, noAssignIds, err)
				return err
			}
			for _, assign := range taskAssigns {
				h.taskDataMap[tarTaskId].assignDataMap[assign.AssignID] = &assignData{
					assign: assign,
				}
			}
		}
	}
	return nil
}

// 获取任务布置对象的学生信息(如果是按班级布置，则同时获取班级信息)
// 1.获取全部任务的全部布置对象的学生信息，taskId = 0
// 2.获取指定任务的全部布置对象的学生信息，taskId != 0, tarAssignId = 0
// 3.获取指定任务指定布置对象的学生信息，taskId != 0, tarAssignId != 0
func (h *multiTasks) getAssignClassInfo(ctx context.Context, tarTaskId, tarAssignId int64) error {
	// 缺少前置数据，且无法获取到数据，直接返回
	if len(h.taskDataMap) == 0 || (tarTaskId != 0 && (h.taskDataMap[tarTaskId] == nil || h.taskDataMap[tarTaskId].task == nil)) {
		h.log.Warn(ctx, "[getTaskAssignClassInfo] taskData is nil, taskID:%d, assignID:%d", tarTaskId, tarAssignId)
		return errors.New("task data missing")
	}

	// 依赖布置基础数据
	if err := h.getAssignData(ctx, tarTaskId, tarAssignId); err != nil {
		h.log.Error(ctx, "[getTaskAssignClassInfo] getTaskAssignData failed, taskID:%d, assignID:%d, error:%v", tarTaskId, tarAssignId, err)
		return err
	}

	// 无学生信息的任务布置 id 列表
	noStudentAssignIds := make([]int64, 0)
	for taskId, taskData := range h.taskDataMap {
		if tarTaskId != 0 && tarTaskId != taskId {
			continue
		}
		for assignId := range taskData.assignDataMap {
			if tarAssignId != 0 && assignId != tarAssignId {
				continue
			}
			if taskData.assignDataMap[assignId].classInfo == nil {
				noStudentAssignIds = append(noStudentAssignIds, assignId)
			}
		}
	}

	// 如果学生信息都获取到了，则直接返回
	if len(noStudentAssignIds) == 0 {
		return nil
	}

	classIDs := make([]int64, 0)     // 班级id 列表
	tmpAssignIds := make([]int64, 0) // 临时小组id 列表
	for taskId, taskData := range h.taskDataMap {
		for assignId, assignData := range taskData.assignDataMap {
			if tarTaskId != 0 && tarTaskId != taskId {
				continue
			}
			if tarAssignId != 0 && assignId != tarAssignId {
				continue
			}
			switch assignData.assign.GroupType {
			case consts.TASK_GROUP_TYPE_CLASS:
				classIDs = append(classIDs, assignData.assign.GroupID)
			case consts.TASK_GROUP_TYPE_STUDENT, consts.TASK_GROUP_TYPE_TEMP:
				tmpAssignIds = append(tmpAssignIds, assignData.assign.GroupID)
			}
		}
	}

	// 获取任务布置对象的班级信息
	var err error
	classMap := make(map[int64]*itl.ClassInfo)  // classId -> ClassInfo
	assignStudentIds := make(map[int64][]int64) // assignID -> [studentID]int64

	schoolID := h.taskDataMap[tarTaskId].task.SchoolID
	if len(classIDs) > 0 {
		classMap, err = h.getClassStudents(ctx, schoolID, classIDs)
		if err != nil {
			h.log.Error(ctx, "[getTaskAssignClassInfo] getClassStudents failed, classID:%d, error:%v", classIDs, err)
			return err
		}
	}

	tmpAssignIds = slices.Compact(tmpAssignIds) // 去重
	if len(tmpAssignIds) > 0 {
		assignStudentIds, err = h.taskAssignService.GetAssignStudents(ctx, tmpAssignIds)
		if err != nil {
			h.log.Error(ctx, "[getTaskAssignClassInfo] GetAssignStudents failed, assignID:%d, error:%v", tmpAssignIds, err)
			return err
		}
	}

	if len(classMap) > 0 {
		for classID, classInfo := range classMap {
			for _, assignData := range h.taskDataMap[tarTaskId].assignDataMap {
				if assignData.assign.GroupID == classID {
					assignData.classInfo = classInfo
				}
			}
		}
	} else if len(assignStudentIds) > 0 {
		for assignId, studentList := range assignStudentIds {
			tmpStudents := make([]*itl.StudentInfo, 0)
			for _, studentID := range studentList {
				tmpStudents = append(tmpStudents, &itl.StudentInfo{
					ID:     studentID,
					Name:   "", // TODO 临时小组或者学生小组，以后补充
					Avatar: "", // TODO 临时小组或者学生小组，以后补充
				})
			}
			h.taskDataMap[tarTaskId].assignDataMap[assignId].classInfo = &itl.ClassInfo{
				Students: tmpStudents,
			}
		}
	}
	return nil
}

// 获取任务布置对象的统计数据
// 1.获取全部任务的全部布置对象的统计数据，taskId = 0
// 2.获取指定任务的全部布置对象的统计数据，taskId != 0, tarAssignId = 0
// 3.获取指定任务指定布置对象的统计数据，taskId != 0, tarAssignId != 0
func (h *multiTasks) getAssignReport(ctx context.Context, tarTaskId, tarAssignId int64) error {
	// 依赖任务基础数据
	if len(h.taskDataMap) == 0 || (tarTaskId != 0 && (h.taskDataMap[tarTaskId] == nil || h.taskDataMap[tarTaskId].task == nil)) {
		h.log.Warn(ctx, "[getAssignReport] taskData is nil, taskID:%d, assignID:%d", tarTaskId, tarAssignId)
		return errors.New("taskData is nil")
	}

	if err := h.getAssignData(ctx, tarTaskId, tarAssignId); err != nil {
		h.log.Error(ctx, "[getAssignReport] getTaskAssignData failed, taskID:%d, assignID:%d, error:%v", tarTaskId, tarAssignId, err)
		return err
	}

	noReportAssignIds := make(map[int64][]int64) // taskID -> [assignID]int64
	if tarTaskId == 0 {                          // 获取全部布置对象的统计数据
		for taskID, taskData := range h.taskDataMap {
			if noReportAssignIds[taskID] == nil {
				noReportAssignIds[taskID] = make([]int64, 0)
			}
			for assignID := range taskData.assignDataMap {
				if taskData.assignDataMap[assignID].report == nil {
					noReportAssignIds[taskID] = append(noReportAssignIds[taskID], assignID)
				}
			}
		}
	} else { // 获取指定任务的指定布置对象( tarAssignId!=0 时)的统计数据
		if noReportAssignIds[tarTaskId] == nil {
			noReportAssignIds[tarTaskId] = make([]int64, 0)
		}
		for assignID := range h.taskDataMap[tarTaskId].assignDataMap {
			if tarAssignId != 0 && assignID != tarAssignId {
				continue
			}
			if h.taskDataMap[tarTaskId].assignDataMap[assignID].report == nil {
				noReportAssignIds[tarTaskId] = append(noReportAssignIds[tarTaskId], assignID)
			}
		}
	}

	// 如果布置报告数据都获取到了，则直接返回
	if len(noReportAssignIds) == 0 {
		return nil
	}

	taskAssignReportMap, err := h.taskReportService.GetTaskReportsByTaskAssignIDs(ctx, noReportAssignIds)
	if err != nil {
		h.log.Error(ctx, "[getAssignReport] GetTaskReportsByTaskAssignIDs error:%v", err)
		return err
	}

	for taskID, reports := range taskAssignReportMap {
		if h.taskDataMap[taskID] == nil {
			h.log.Warn(ctx, "[getAssignReport] taskData is nil, taskID:%d", taskID)
			continue
		}

		for _, report := range reports {
			if h.taskDataMap[taskID].assignDataMap[report.AssignID] == nil {
				h.log.Warn(ctx, "[getAssignReport] assignData is nil, taskID:%d, assignID:%d", taskID, report.AssignID)
				continue
			}
			h.taskDataMap[taskID].assignDataMap[report.AssignID].report = report
		}
	}

	return nil
}
