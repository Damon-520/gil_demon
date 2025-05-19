package task

import (
	"context"
	"fmt"
	"strings"

	"gil_teacher/app/consts"
	"gil_teacher/app/controller/http_server/response"
	dao_task "gil_teacher/app/dao/task"
	"gil_teacher/app/model/api"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/model/itl"
	"gil_teacher/app/utils"
)

type reportRelationData struct {
	// 基础数据
	Task              *dao_task.Task                      // 任务数据
	TaskResource      map[string]*dao_task.TaskResource   // 资源列表 resource_key -> TaskResource
	ResourceQuestions map[string]map[string]*itl.Question // 资源维度的题目列表 resource_key -> question_id -> Question
	// 作答数据需要单独获取
	QuestionAnswerCount    map[string]int64 // 题目维度答题数 question_key -> answer_count
	QuestionIncorrectCount map[string]int64 // 题目维度错题数 question_key -> incorrect_count
	QuestionTotalCostTime  map[string]int64 // 题目维度总用时 question_key -> total_cost_time
	// 学生作答结果 question_key -> *dao_task.TaskStudentDetails
	StudentAnswers map[int64]map[string]*dao_task.TaskStudentDetails
}

// StudentTaskReport 获取指定学生单次任务的答题结果
func (h *TaskReportHandler) StudentTaskReport(ctx context.Context, studentID int64, query *dto.StudentTaskReportQuery, pageInfo *consts.APIReqeustPageInfo) (*api.TaskStudentAnswerDetailResponse, error) {
	taskID := query.TaskID
	taskHandler, err := h.getTaskData(ctx, taskID)
	if err != nil {
		h.log.Error(ctx, "StudentTaskReport error:%v", err)
		return nil, err
	}

	if err := taskHandler.getResources(ctx, taskID); err != nil {
		h.log.Error(ctx, "[StudentTaskReport] getTaskResource failed, taskID:%d, err:%v", taskID, err)
		return nil, err
	}

	if err := taskHandler.getResourceQuestions(ctx, taskID, &query.TaskReportCommonQuery); err != nil {
		h.log.Error(ctx, "[StudentTaskReport] getResourceQuestions failed, taskID:%d, err:%v", taskID, err)
		return nil, err
	}

	questionIDs := taskHandler.resourceQuestion.questionIDs
	// 获取学生作答结果
	studentAnswers, answerCount, incorrectCount, err := h.taskAnswerService.GetTaskStudentAnswers(ctx, studentID, query, pageInfo.ToDBPageInfo())
	if err != nil {
		h.log.Error(ctx, "StudentTaskReport error:%v", err)
		return nil, err
	}

	// 获取题目作答统计数据
	answerStat, err := h.taskAnswerService.GetTaskAnswerCount(ctx, taskID, query.AssignID, questionIDs)
	if err != nil {
		h.log.Error(ctx, "StudentTaskReport error:%v", err)
		return nil, err
	}

	// 处理作答结果
	taskReportRelationData := &reportRelationData{
		Task:                   taskHandler.task,
		StudentAnswers:         map[int64]map[string]*dao_task.TaskStudentDetails{studentID: studentAnswers},
		QuestionAnswerCount:    answerStat.ResourceAnswerCount,
		QuestionIncorrectCount: answerStat.ResourceIncorrectCount,
		QuestionTotalCostTime:  answerStat.ResourceTotalCostTime,
		TaskResource:           taskHandler.taskResources,
		ResourceQuestions:      taskHandler.resourceQuestion.questionMap,
	}
	userAnswers := h.handleStudentAnswers(studentID, query.AllQuestions, taskReportRelationData)
	totalQuestionNum := taskHandler.resourceQuestion.totalQuestionNum
	// 处理返回数据
	reportCommon := h.getReportCommon(&query.TaskReportCommonQuery)
	reportCommon.TotalCount = totalQuestionNum
	reportCommon.PageInfo = consts.ApiPageResponse{
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
		Total:    totalQuestionNum,
	}

	return &api.TaskStudentAnswerDetailResponse{
		TaskAnswerReportCommon: *reportCommon,
		InCorrectCount:         incorrectCount,
		QuestionAnswers:        userAnswers,
		Progress:               utils.F64Div(float64(answerCount), float64(totalQuestionNum), 2), // 答题进度
	}, nil
}

// 获取任务指定对象(班级/小组)的答题报告
func (h *TaskReportHandler) GetTaskAnswerReport(ctx context.Context,
	query *dto.TaskAssignAnswersQuery,
	pageInfo *consts.APIReqeustPageInfo,
) (*api.TaskAnswerDetailReportResponse, error) {
	taskHandler, err := h.getTaskData(ctx, query.TaskID)
	if err != nil {
		h.log.Error(ctx, "GetTaskAnswerReport error:%v", err)
		return nil, err
	}

	err = taskHandler.getAssignData(ctx, query.TaskID, query.AssignID)
	if err != nil {
		h.log.Error(ctx, "GetTaskAnswerReport error:%v", err)
		return nil, err
	}

	if err := taskHandler.getResourceQuestions(ctx, query.TaskID, &query.TaskReportCommonQuery); err != nil {
		h.log.Error(ctx, "StudentTaskReport error:%v", err)
		return nil, err
	}

	taskReportRelationData, err := h.getTaskReportRelationData(ctx, taskHandler.taskData, query, pageInfo)
	if err != nil {
		return nil, err
	}

	reportCommon := h.getReportCommon(&query.TaskReportCommonQuery)
	totalQuestionNum := taskHandler.resourceQuestion.totalQuestionNum
	reportCommon.TotalCount = totalQuestionNum
	reportCommon.PageInfo = consts.ApiPageResponse{
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
		Total:    totalQuestionNum,
	}

	if err := taskHandler.getAssignClassInfo(ctx, query.TaskID, query.AssignID); err != nil {
		h.log.Error(ctx, "GetTaskAnswerReport error:%v", err)
		return nil, err
	}

	students := taskHandler.assignDataMap[query.AssignID].classInfo.Students
	resp := &api.TaskAnswerDetailReportResponse{
		TaskAnswerReportCommon: *reportCommon,
		CommonIncorrectCount:   0, //TODO 共性错题数
		Students:               h.getStudentList(students),
		QuestionAnswers:        h.getQuestionAnswers(taskReportRelationData, &query.TaskReportCommonQuery),
	}
	return resp, nil
}

// 解析学生单次任务的作答结果, 错题数，答题信息
//
//	[]*api.QuestionAnswer
func (h *TaskReportHandler) handleStudentAnswers(
	studentID int64,
	allQuestions bool, // 是否查询全部题目
	reportRelationData *reportRelationData, // 任务报告关系数据
) []*api.QuestionAnswer {
	filteredQuestionMap := make(map[string]*api.QuestionAnswer)
	for resourceKey, questions := range reportRelationData.ResourceQuestions {
		for questionId, question := range questions {
			questionKey := utils.JoinList([]any{resourceKey, questionId}, consts.CombineKey)
			studentAnswer := reportRelationData.StudentAnswers[studentID][questionKey]
			// 只看错题，则只返回错题信息
			if !allQuestions && studentAnswer != nil && studentAnswer.Correctness {
				continue
			}
			parts := strings.Split(questionKey, consts.CombineKey)
			filteredQuestionMap[questionKey] = &api.QuestionAnswer{
				ResourceID:     parts[0],
				ResourceType:   utils.Atoi64(parts[1]),
				AnswerCount:    reportRelationData.QuestionAnswerCount[questionKey],
				IncorrectCount: reportRelationData.QuestionIncorrectCount[questionKey],
				AvgCostTime:    utils.I64Div(reportRelationData.QuestionTotalCostTime[questionKey], reportRelationData.QuestionAnswerCount[questionKey]),
				Question:       question,
				Answer:         h.getStudentAnswer(studentAnswer),
			}
		}
	}

	// 题目排序
	taskData := &taskData{
		task:          reportRelationData.Task,
		taskResources: reportRelationData.TaskResource,
	}
	filteredQuestions := make([]*api.QuestionAnswer, 0)
	for _, question := range filteredQuestionMap {
		filteredQuestions = append(filteredQuestions, question)
	}
	sortedQuestions, _ := h.sortQuestions(taskData, filteredQuestions, nil)
	questionAnswers := make([]*api.QuestionAnswer, 0)
	for _, sq := range sortedQuestions {
		for questionKey, question := range filteredQuestionMap {
			parts := strings.Split(questionKey, consts.CombineKey)
			if sq.Question.QuestionId == question.Question.QuestionId {
				questionAnswers = append(questionAnswers, &api.QuestionAnswer{
					QuestionIndex:  sq.QuestionIndex,
					Question:       question.Question,
					ResourceID:     parts[0],
					ResourceType:   utils.Atoi64(parts[1]),
					AnswerCount:    reportRelationData.QuestionAnswerCount[questionKey],
					IncorrectCount: reportRelationData.QuestionIncorrectCount[questionKey],
					AvgCostTime:    utils.I64Div(reportRelationData.QuestionTotalCostTime[questionKey], reportRelationData.QuestionAnswerCount[questionKey]),
					Answer:         h.getStudentAnswer(reportRelationData.StudentAnswers[studentID][questionKey]),
				})
				break
			}
		}
	}
	return questionAnswers
}

// 解析任务的答题结果
func (h *TaskReportHandler) getQuestionAnswers(reportRelationData *reportRelationData, query *dto.TaskReportCommonQuery) []*api.QuestionAnswer {
	taskStudentAnswers := reportRelationData.StudentAnswers
	studentQuestionAnswerMap := make(map[string][]*dao_task.TaskStudentDetails) //按题目归类的学生作答结果 question_key -> []TaskStudentDetails
	questionIncorrectCount := make(map[string]int64)                            // question_key -> 错题数
	for _, answers := range taskStudentAnswers {
		for _, answer := range answers {
			questionKey := utils.JoinList([]any{answer.ResourceKey, answer.QuestionID}, consts.CombineKey)
			if _, ok := studentQuestionAnswerMap[questionKey]; !ok {
				studentQuestionAnswerMap[questionKey] = make([]*dao_task.TaskStudentDetails, 0)
			}
			studentQuestionAnswerMap[questionKey] = append(studentQuestionAnswerMap[questionKey], answer)
			if !answer.Correctness {
				questionIncorrectCount[questionKey]++
			}
		}
	}

	resourceQuestions := reportRelationData.ResourceQuestions
	questionAnswerMap := make(map[string]*api.QuestionAnswer)
	for resourceKey, questions := range resourceQuestions {
		questionCostTimeMap := make(map[string][]int64) // 题目作答时长
		parts := strings.Split(resourceKey, consts.CombineKey)
		for questionID, question := range questions {
			questionKey := utils.JoinList([]any{resourceKey, questionID}, consts.CombineKey)
			studentAnswers := studentQuestionAnswerMap[questionKey]
			questionAnswerMap[questionKey] = &api.QuestionAnswer{
				Question:       question,
				Answers:        h.getStudentAnswers(studentAnswers),
				ResourceID:     parts[0],
				ResourceType:   utils.Atoi64(parts[1]),
				AnswerCount:    int64(len(studentAnswers)),
				IncorrectCount: questionIncorrectCount[questionKey],
			}

			for _, answer := range studentAnswers {
				if _, ok := questionCostTimeMap[questionKey]; !ok {
					questionCostTimeMap[questionKey] = make([]int64, 0)
				}
				questionCostTimeMap[questionKey] = append(questionCostTimeMap[questionKey], answer.CostTime)
			}
			// 计算题目平均作答时长
			for questionKey, costTime := range questionCostTimeMap {
				questionAnswerMap[questionKey].AvgCostTime = utils.AvgInt64(costTime)
			}
		}
	}
	answerDetails := make([]*api.QuestionAnswer, 0)
	for _, questionAnswer := range questionAnswerMap {
		answerDetails = append(answerDetails, questionAnswer)
	}
	// 题目排序，或者基于查询条件排序
	taskData := &taskData{
		task:          reportRelationData.Task,
		taskResources: reportRelationData.TaskResource,
	}
	sortedQuestions, _ := h.sortQuestions(taskData, answerDetails, query)
	return sortedQuestions
}

func (h *TaskReportHandler) getReportCommon(query *dto.TaskReportCommonQuery) *api.TaskAnswerReportCommon {
	return &api.TaskAnswerReportCommon{
		TaskID:       query.TaskID,
		AssignID:     query.AssignID,
		ResourceID:   query.ResourceID,
		ResourceType: query.ResourceType,
		// TaskType:     0, //TODO 任务类型
	}
}

func (h *TaskReportHandler) getStudentAnswer(studentAnswer *dao_task.TaskStudentDetails) *api.StudentAnswer {
	if studentAnswer == nil {
		return nil
	}

	return &api.StudentAnswer{
		StudentID: studentAnswer.StudentID,
		Answer:    studentAnswer.AnswerContent,
		IsCorrect: studentAnswer.Correctness,
		CostTime:  studentAnswer.CostTime,
	}
}

func (h *TaskReportHandler) getStudentAnswers(questionAnswer []*dao_task.TaskStudentDetails) []*api.StudentAnswer {
	studentAnswers := make([]*api.StudentAnswer, 0)
	for _, answer := range questionAnswer {
		studentAnswers = append(studentAnswers, h.getStudentAnswer(answer))
	}
	return studentAnswers
}

// 获取任务布置报告相关数据
func (h *TaskReportHandler) getTaskReportRelationData(ctx context.Context,
	taskData *taskData,
	query *dto.TaskAssignAnswersQuery,
	pageInfo *consts.APIReqeustPageInfo,
) (*reportRelationData, error) {
	// 获取题目作答统计数据
	answerStat, err := h.taskAnswerService.GetTaskAnswerCount(ctx, query.TaskID, query.AssignID, taskData.resourceQuestion.questionIDs)
	if err != nil {
		h.log.Error(ctx, "StudentTaskReport error:%v", err)
		return nil, err
	}

	taskStudentAnswers, err := h.taskReportService.GetTaskAssignAnswers(ctx, query, pageInfo)
	if err != nil {
		h.log.Error(ctx, "StudentTaskReport error:%v", err)
		return nil, err
	}

	// 处理作答结果
	taskReportRelationData := &reportRelationData{
		Task:                   taskData.task,
		TaskResource:           taskData.taskResources,
		QuestionAnswerCount:    answerStat.ResourceAnswerCount,
		QuestionIncorrectCount: answerStat.ResourceIncorrectCount,
		QuestionTotalCostTime:  answerStat.ResourceTotalCostTime,
		ResourceQuestions:      taskData.resourceQuestion.questionMap,
		StudentAnswers:         taskStudentAnswers,
	}
	return taskReportRelationData, nil
}

// GetStudentDetail 获取学生作业报告详情
func (h *TaskReportHandler) GetStudentDetail(ctx context.Context, schoolID int64, taskID int64, assignID int64, studentID int64) (*api.GetStudentDetailResponse, *response.Response) {
	var res api.GetStudentDetailResponse

	// 获取学生基本信息
	studentInfoMap, err := h.ucenterService.GetStudentInfoByID(ctx, schoolID, []int64{studentID})
	if err != nil {
		h.log.Error(ctx, "GetStudentInfo error:%v", err)
		return nil, &response.ERR_GIL_ADMIN
	}
	studentInfo, ok := studentInfoMap[studentID]
	if !ok {
		h.log.Error(ctx, "GetStudentInfo not found")
		return nil, &response.ERR_INVALID_STUDENT
	}
	res.Student = &api.Student{
		StudentID:   studentID,
		StudentName: studentInfo.Student.Name,
		Avatar:      studentInfo.Student.Avatar,
	}

	// 获取学生点赞和提醒次数
	praiseAndAttention, err := h.behaviorDAO.CountStudentTaskPraiseAndAttention(ctx, uint64(taskID), uint64(assignID), []uint64{uint64(studentID)})
	if err != nil {
		h.log.Error(ctx, "GetStudentDetail error:%v", err)
		return nil, &response.ERR_CLICKHOUSE
	}
	if len(praiseAndAttention) != 0 {
		for _, studentPraiseAndAttention := range praiseAndAttention {
			h.log.Debug(ctx, "GetStudentDetail praiseAndAttention: %v", studentPraiseAndAttention)
			if v, ok := studentPraiseAndAttention.GroupValues["behavior_type"].(string); ok && v == string(consts.BehaviorTypeTaskPraise) {
				res.PraiseCount = studentPraiseAndAttention.Count
			}
			if v, ok := studentPraiseAndAttention.GroupValues["behavior_type"].(string); ok && v == string(consts.BehaviorTypeTaskAttention) {
				res.AttentionCount = studentPraiseAndAttention.Count
			}
		}
	}

	// 获取学生正确率、完成进度
	taskStudentReport, err := h.taskReportService.GetTaskReportByTaskIDAndStudentID(ctx, taskID, assignID, studentID)
	if err != nil {
		h.log.Error(ctx, "GetTaskReport error:%v", err)
		return nil, &response.ERR_POSTGRESQL
	}
	if taskStudentReport != nil {
		res.StudentAccuracyRate = taskStudentReport.AccuracyRate
		res.StudentCompletedProgress = taskStudentReport.CompletedProgress
	}

	// 获取班级平均正确率、完成进度
	taskReport, err := h.taskReportService.GetTaskAssignStats(ctx, taskID, assignID)
	if err != nil {
		h.log.Error(ctx, "GetTaskReport error:%v", err)
		return nil, &response.ERR_POSTGRESQL
	}
	if taskReport != nil {
		res.ClassAccuracyRate = taskReport.ReportDetail.AccuracyRate
		res.ClassCompletedProgress = taskReport.ReportDetail.CompletedProgress
	}

	// TODO: 根据策略计算得到对应的文案，依赖于学生端上报的数据，这里先写死给前端
	res.AttentionText = fmt.Sprintf(consts.AttentionTextDefault, res.StudentName)
	res.AttentionTextList = consts.AttentionTextDefaultList
	res.PushDefaultText = consts.PushDefaultTextAskFrequently

	return &res, nil
}
