package task

import (
	"context"
	"fmt"
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

// GetTeacherTaskReportList 获取教师作业报告列表
func (h *TaskReportHandler) GetTeacherTasksReportList(ctx context.Context, reqs *dto.TaskAssignListQuery, pageInfo *consts.APIReqeustPageInfo) (*api.TaskReportsResponse, error) {
	// 先查询教师的所有任务布置信息
	total, taskAssigns, err := h.taskAssignService.GetTeacherTasks(ctx, reqs, pageInfo.ToDBPageInfo())
	if err != nil {
		h.log.Error(ctx, "GetTeacherTaskReportList error:%v", err)
		return nil, err
	}

	resp := &api.TaskReportsResponse{
		PageInfo: &consts.ApiPageResponse{
			Page:     pageInfo.Page,
			PageSize: pageInfo.PageSize,
			Total:    total,
		},
	}

	if total == 0 || len(taskAssigns) == 0 {
		return resp, nil
	}

	// 任务名映射
	taskIds := make([]int64, 0)
	for _, taskAssign := range taskAssigns {
		taskIds = append(taskIds, taskAssign.TaskID)
	}
	// 去重
	taskIds = slices.Compact(taskIds)

	// 获取任务相关的原始数据
	taskDataMap, err := h.getTasksData(ctx, taskIds, taskAssigns)
	if err != nil {
		h.log.Error(ctx, "GetTeacherTaskReportList error:%v", err)
		return nil, err
	}

	taskReports, err := h.handleTaskReports(ctx, taskDataMap, reqs.ClassInfo)
	if err != nil {
		h.log.Error(ctx, "GetTeacherTaskReportList error:%v", err)
		return nil, err
	}
	resp.Tasks = taskReports
	return resp, nil
}

// GetLatestTeacherTaskReport 获取教师最近一次布置的任务报告
func (h *TaskReportHandler) GetTeacherLatestTaskReport(ctx context.Context, query *dto.TeacherLatestTaskReportsQuery) (*api.LatestReportsResponse, error) {
	// 查询教师最近一次布置的任务
	tasks, err := h.taskService.GetLatestTask(ctx, query.TeacherID, query.SchoolID, query.Subject)
	if err != nil {
		h.log.Error(ctx, "GetLatestTeacherTaskReport error:%v", err)
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	taskIds := make([]int64, 0)
	for _, task := range tasks {
		taskIds = append(taskIds, task.TaskID)
	}

	taskDataMap, err := h.getTasksData(ctx, taskIds, nil)
	if err != nil {
		h.log.Error(ctx, "GetLatestTeacherTaskReport error:%v", err)
		return nil, err
	}

	taskReports, err := h.handleTaskReports(ctx, taskDataMap, query.ClassInfo)
	if err != nil {
		h.log.Error(ctx, "GetLatestTeacherTaskReport error:%v", err)
		return nil, err
	}

	return &api.LatestReportsResponse{
		Tasks: taskReports,
	}, nil
}

// 获取指定任务指定布置对象(班级/小组)的汇总报告数据(学生粒度)
func (h *TaskReportHandler) GetTaskReportSummaryDetail(ctx context.Context, query *dto.TaskAssignReportQuery, pageInfo *consts.APIReqeustPageInfo) (*api.TaskReportSummaryDetailResponse, error) {
	taskID := query.TaskID
	taskHandler, err := h.getTaskData(ctx, taskID)
	if err != nil {
		h.log.Error(ctx, "[GetTaskReportSummaryDetail] getTaskData failed, taskID:%d, error:%v", taskID, err)
		return nil, err
	}

	if err := taskHandler.getResources(ctx, taskID); err != nil {
		h.log.Error(ctx, "[GetTaskReportSummaryDetail] getTaskResources failed, taskID:%d, error:%v", taskID, err)
		return nil, err
	}

	if err := taskHandler.getAssignData(ctx, taskID, query.AssignID); err != nil {
		h.log.Error(ctx, "[GetTaskReportSummaryDetail] GetTaskAssignData failed, taskID:%d, assignID:%d, error:%v", taskID, query.AssignID, err)
		return nil, err
	}

	if err := taskHandler.getAssignClassInfo(ctx, taskID, query.AssignID); err != nil {
		h.log.Error(ctx, "[GetTaskReportSummaryDetail] getTaskAssignClassInfo failed, taskID:%d, assignID:%d, error:%v", taskID, query.AssignID, err)
		return nil, err
	}

	if err := taskHandler.getAssignReport(ctx, query.AssignID); err != nil {
		h.log.Error(ctx, "[GetTaskReportSummaryDetail] getTaskAssignReport failed, taskID:%d, assignID:%d, error:%v", taskID, query.AssignID, err)
		return nil, err
	}

	reports, err := taskHandler.getResourceReports(ctx, query.AssignID)
	if err != nil {
		h.log.Error(ctx, "[GetTaskReportSummaryDetail] getTaskResourceReports failed, taskID:%d, assignID:%d, error:%v", taskID, query.AssignID, err)
		return nil, err
	}

	students := taskHandler.assignDataMap[query.AssignID].classInfo.Students
	filterStudents := make([]*itl.StudentInfo, 0)
	// 如果是模糊查询学生
	if query.StudentName != "" {
		for _, student := range students {
			if strings.Contains(student.Name, query.StudentName) {
				query.StudentIDs = append(query.StudentIDs, student.ID)
				filterStudents = append(filterStudents, student)
			}
		}
	} else {
		filterStudents = students
	}

	studentReportQuery := &dto.TaskAssignAnswersQuery{
		TaskReportCommonQuery: dto.TaskReportCommonQuery{
			TaskID:       taskID,
			AssignID:     query.AssignID,
			ResourceID:   query.ResourceID,
			ResourceType: query.ResourceType,
		},
		StudentIDs: query.StudentIDs,
	}
	studentReports, total, err := taskHandler.getStudentReports(ctx, filterStudents, studentReportQuery, pageInfo)
	if err != nil {
		h.log.Error(ctx, "[GetTaskReportSummaryDetail] getTaskStudentReports failed, taskID:%d, assignID:%d, studentIDs:%v, error:%v",
			taskID, query.AssignID, query.StudentIDs, err)
		return nil, err
	}

	if err := taskHandler.getAssignReport(ctx, query.AssignID); err != nil {
		h.log.Error(ctx, "[GetTaskReportSummaryDetail] getTaskResourceReports failed, taskID:%d, assignID:%d, error:%v", taskID, query.AssignID, err)
		return nil, err
	}

	// 计算平均正确率和平均用时
	accuracies := make([]float64, 0)
	costTimes := make([]int64, 0)
	for _, report := range studentReports {
		accuracies = append(accuracies, report.AccuracyRate)
		costTimes = append(costTimes, report.CostTime)
	}

	// 返回报告数据
	return &api.TaskReportSummaryDetailResponse{
		Students: h.getStudentList(students),
		Detail: &api.ReportSummaryDetail{
			PraiseList:      h.getStudentIdList(students), // TODO 表扬列表
			AttentionList:   h.getStudentIdList(students), // TODO 需要关注列表
			AvgAccuracy:     utils.AvgFloat64(accuracies),
			AvgCostTime:     utils.AvgInt64(costTimes),
			StudentReports:  studentReports,
			ResourceReports: reports,
		},
		PageInfo: consts.ApiPageResponse{
			Page:     pageInfo.Page,
			PageSize: pageInfo.PageSize,
			Total:    total,
		},
	}, nil
}

// 导出指定班级的任务报告（csv）
// 文件名：任务名-班级名(-资源名).csv
func (h *TaskReportHandler) ExportTaskReport(ctx context.Context, query *dto.ExportTaskReportQuery) (string, *dto.ExportTaskReportResult, error) {
	taskID := query.TaskID
	taskHandler, err := h.getTaskData(ctx, taskID)
	if err != nil {
		h.log.Error(ctx, "[ExportTaskReport] GetTaskData failed, taskID:%d, error:%v", taskID, err)
		return "", nil, err
	}
	if err := taskHandler.getAssignData(ctx, taskID, query.AssignID); err != nil {
		h.log.Error(ctx, "[ExportTaskReport] GetTaskAssignData failed, taskID:%d, assignID:%d, error:%v", taskID, query.AssignID, err)
		return "", nil, err
	}

	if err := taskHandler.getAssignClassInfo(ctx, taskID, query.AssignID); err != nil {
		h.log.Error(ctx, "[ExportTaskReport] GetTaskAssignClassInfo failed, taskID:%d, assignID:%d, error:%v", taskID, query.AssignID, err)
		return "", nil, err
	}

	// 获取题目数据，计算题目难度
	if err := taskHandler.getResourceQuestions(ctx, taskID, nil); err != nil {
		h.log.Error(ctx, "[ExportTaskReport] GetTaskResourceQuestions failed, taskID:%d, assignID:%d, error:%v", taskID, query.AssignID, err)
		return "", nil, err
	}

	students := taskHandler.assignDataMap[query.AssignID].classInfo.Students
	assignReportQuery := &dto.TaskAssignAnswersQuery{
		TaskReportCommonQuery: dto.TaskReportCommonQuery{
			TaskID:       taskID,
			AssignID:     query.AssignID,
			ResourceID:   query.ResourceID,
			ResourceType: query.ResourceType,
			SortKey:      query.SortBy,
			SortType:     query.SortType,
		},
	}
	studentReports, _, err := taskHandler.getStudentReports(ctx, students, assignReportQuery, consts.AllDataPageInfo())
	if err != nil {
		h.log.Error(ctx, "[ExportTaskReport] GetTaskStudentReports failed, taskID:%d, assignID:%d, error:%v", taskID, query.AssignID, err)
		return "", nil, err
	}

	result := &dto.ExportTaskReportResult{
		Meta: consts.ExtractExportFields(query.Fields),
		Data: make([][]string, 0, len(studentReports)),
	}

	studentMap := make(map[int64]*itl.StudentInfo)
	for _, student := range students {
		studentMap[student.ID] = student
	}

	// 解析每个学生的报告数据，需要对齐字段顺序
	for _, studentReport := range studentReports {
		row := make([]string, 0)
		for _, field := range query.Fields {
			switch field {
			case "studentName":
				row = append(row, studentMap[studentReport.StudentID].Name)
			case "studyScore":
				row = append(row, utils.I64ToStr(studentReport.StudyScore))
			case "progress":
				row = append(row, utils.F64ToString(studentReport.AccuracyRate, 2, "-"))
			case "accuracyRate":
				row = append(row, utils.F64ToString(studentReport.AccuracyRate, 2, "-"))
			case "difficultyDegree":
				row = append(row, utils.F64ToString(studentReport.AccuracyRate, 1, "-"))
			case "incorrectCount/answerCount":
				row = append(row, fmt.Sprintf("'%d/%d", studentReport.IncorrectNum, studentReport.AnswerNum))
			case "answerTime":
				row = append(row, utils.I64ToStr(studentReport.CostTime))
			}
		}
		result.Data = append(result.Data, row)
	}
	exportFileName := fmt.Sprintf("%s-%s", taskHandler.task.TaskName, taskHandler.assignDataMap[query.AssignID].classInfo.ClassName)
	if query.ResourceID != "" {
		resourceName := consts.GetResourceTypeName(query.ResourceType)
		if resourceName != "" {
			exportFileName = fmt.Sprintf("%s-%s", exportFileName, resourceName)
		}
	} else {
		taskTypeName := consts.GetTaskTypeName(taskHandler.task.TaskType)
		if taskTypeName != "" {
			exportFileName = fmt.Sprintf("%s-%s", exportFileName, taskTypeName)
		}
	}
	return exportFileName, result, nil
}

// 获取任务布置对象每个题目的答题正确率统计，即答题面板统计数据，统计每个题目的答题正确率
func (h *TaskReportHandler) GetTaskAnswerPanel(ctx context.Context, query *dto.TaskReportCommonQuery) (*api.QuestionPanelResponse, error) {
	taskID := query.TaskID
	// 获取任务资源列表
	taskHandler, err := h.getTaskData(ctx, taskID)
	if err != nil {
		h.log.Error(ctx, "[GetTaskAnswerPanel] GetTaskData failed, taskID:%d, error:%v", query.TaskID, err)
		return nil, err
	}

	if err := taskHandler.getResources(ctx, taskID); err != nil {
		h.log.Error(ctx, "[GetTaskAnswerPanel] GetTaskResource failed, taskID:%d, error:%v", query.TaskID, err)
		return nil, err
	}

	if err := taskHandler.getResourceQuestions(ctx, taskID, query); err != nil {
		h.log.Error(ctx, "[GetTaskAnswerPanel] GetTaskResourceQuestions failed, taskID:%d, error:%v", query.TaskID, err)
		return nil, err
	}

	// 提取全部问题id（按resource_key分组）, map[resource_key] -> []question_id
	questionMap := taskHandler.resourceQuestion.questionMap
	questionIdMap := make(map[string][]string)
	for resourceKey, questions := range questionMap {
		for questionId := range questions {
			questionIdMap[resourceKey] = append(questionIdMap[resourceKey], questionId)
		}
	}

	// 布置的答题正确率
	if err := taskHandler.getAssignAnswerAccuracy(ctx, taskID, query.AssignID, questionIdMap); err != nil {
		h.log.Error(ctx, "[GetTaskAnswerPanel] GetTaskAnswerPanel failed, taskID:%d, assignID:%d, error:%v", taskID, query.AssignID, err)
		return nil, err
	}

	questionPanels := make([]*api.QuestionPanel, 0)
	questionIndex := int64(1)
	for resourceKey, questions := range questionMap {
		parts := strings.Split(resourceKey, consts.CombineKey) // resource_id#resource_type
		for questionId := range questions {
			questionKey := utils.JoinList([]any{resourceKey, questionId}, consts.CombineKey)
			tmp := &api.QuestionPanel{
				ResourceID:    parts[0],
				ResourceType:  utils.Atoi64(parts[1]),
				QuestionID:    questionId,
				QuestionIndex: questionIndex,
			}
			if answer, ok := taskHandler.assignDataMap[query.AssignID].questionAnswers[questionKey]; ok {
				tmp.CorrectRate = answer.Accuracy
				tmp.AnswerCount = answer.AnswerCount
				tmp.IncorrectCount = answer.IncorrectCount
			}
			questionPanels = append(questionPanels, tmp)
			questionIndex++
		}
	}

	// 按请求条件进行排序
	isSorted := false
	switch query.SortKey {
	case "answerCount":
		sort.Slice(questionPanels, func(i, j int) bool {
			if query.SortType == consts.SortTypeDesc {
				return questionPanels[i].AnswerCount > questionPanels[j].AnswerCount
			} else {
				return questionPanels[i].AnswerCount < questionPanels[j].AnswerCount
			}
		})
		isSorted = true
	case "incorrectCount":
		sort.Slice(questionPanels, func(i, j int) bool {
			if query.SortType == consts.SortTypeDesc {
				return questionPanels[i].IncorrectCount > questionPanels[j].IncorrectCount
			} else {
				return questionPanels[i].IncorrectCount < questionPanels[j].IncorrectCount
			}
		})
		isSorted = true
	}

	if !isSorted {
		sort.Slice(questionPanels, func(i, j int) bool {
			return questionPanels[i].QuestionIndex < questionPanels[j].QuestionIndex
		})
	}

	return &api.QuestionPanelResponse{
		TaskID:   query.TaskID,
		AssignID: query.AssignID,
		Panel:    questionPanels,
	}, nil
}

// handleTaskReport 处理单个任务的所有布置报告
func (h *TaskReportHandler) handleTaskReport(ctx context.Context, taskData *taskData, classInfo map[int64]*itl.Class) (*api.TaskReport, error) {
	// 按布置ID分组统计数据
	statMap := make(map[int64]*dao_task.TaskReport) // assignID -> taskReport
	taskHandler := &singleTask{
		TaskReportHandler: h,
		taskData:          taskData,
	}
	if err := taskHandler.getAssignReport(ctx, 0); err != nil {
		return nil, err
	}
	for assignID, assignData := range taskData.assignDataMap {
		statMap[assignID] = assignData.report
	}

	// 处理每个布置的报告
	reports := make([]*api.Report, 0)
	for assignID, assignData := range taskData.assignDataMap {
		tmp := &api.Report{
			AssignID: assignID,
			AssignObject: &api.AssignObject{
				ID:   assignData.assign.GroupID,   // 班级/群组ID，0 为临时群组ID
				Type: assignData.assign.GroupType, // 群组类型，1 为班级，2 为临时群组
			},
		}
		if class, ok := classInfo[assignData.assign.GroupID]; ok {
			tmp.AssignObject.Name = class.Name
		}

		tmp.StatData = &api.AssignObjectStat{
			StartTime: assignData.assign.StartTime,
			Deadline:  assignData.assign.Deadline,
		}
		if stat, ok := statMap[assignID]; ok && stat != nil {
			reportDetail := stat.ReportDetail
			tmp.StatData.CompletionRate = float64(reportDetail.CompletedProgress)
			tmp.StatData.CorrectRate = float64(reportDetail.AccuracyRate)
			tmp.StatData.NeedAttentionQuestionNum = reportDetail.NeedAttentionNum
			tmp.StatData.AverageProgress = float64(reportDetail.AverageProgress)
			tmp.StatData.ClassHours = reportDetail.ClassHours
		}
		reports = append(reports, tmp)
	}

	resources := make([]*api.Resource, 0)
	for _, resource := range taskData.taskResources {
		resources = append(resources, &api.Resource{
			ID:   resource.ResourceID,
			Type: resource.ResourceType,
			Name: consts.ResourceTypeNameMap[resource.ResourceType],
		})
	}

	return &api.TaskReport{
		CreatorID: taskData.task.CreatorID,
		TaskID:    taskData.task.TaskID,
		TaskName:  taskData.task.TaskName,
		TaskType:  taskData.task.TaskType,
		Subject:   taskData.task.Subject,
		Resources: resources,
		Reports:   reports,
	}, nil
}

// handleTaskReport 处理全部任务的所有布置报告
func (h *TaskReportHandler) handleTaskReports(ctx context.Context, taskDataMap map[int64]*taskData, classInfo map[int64]*itl.Class) ([]*api.TaskReport, error) {
	taskHandler := &multiTasks{
		TaskReportHandler: h,
		taskDataMap:       taskDataMap,
	}

	// 按布置ID分组统计数据
	if err := taskHandler.getAssignReport(ctx, 0, 0); err != nil {
		return nil, err
	}

	taskReports := make([]*api.TaskReport, 0)
	// 处理每个布置的报告
	for taskId, taskData := range taskHandler.taskDataMap {
		reports := make([]*api.Report, 0)
		for assignID, assignData := range taskData.assignDataMap {
			tmp := &api.Report{
				AssignID: assignID,
				AssignObject: &api.AssignObject{
					ID:   assignData.assign.GroupID,   // 班级/群组ID，0 为临时群组ID
					Type: assignData.assign.GroupType, // 群组类型，1 为班级，2 为临时群组
				},
			}
			if class, ok := classInfo[assignData.assign.GroupID]; ok {
				tmp.AssignObject.Name = class.Name
			}

			tmp.StatData = &api.AssignObjectStat{
				StartTime: assignData.assign.StartTime,
				Deadline:  assignData.assign.Deadline,
			}
			if assignData.report != nil {
				reportDetail := assignData.report.ReportDetail
				tmp.StatData.CompletionRate = float64(reportDetail.CompletedProgress)
				tmp.StatData.CorrectRate = float64(reportDetail.AccuracyRate)
				tmp.StatData.NeedAttentionQuestionNum = reportDetail.NeedAttentionNum
				tmp.StatData.AverageProgress = float64(reportDetail.AverageProgress)
				tmp.StatData.ClassHours = reportDetail.ClassHours
			}
			reports = append(reports, tmp)
		}

		resources := make([]*api.Resource, 0)
		for _, resource := range taskData.taskResources {
			resources = append(resources, &api.Resource{
				ID:   resource.ResourceID,
				Type: resource.ResourceType,
				Name: consts.ResourceTypeNameMap[resource.ResourceType],
			})
		}

		taskReports = append(taskReports, &api.TaskReport{
			CreatorID: taskData.task.CreatorID,
			TaskID:    taskId,
			TaskName:  taskData.task.TaskName,
			TaskType:  taskData.task.TaskType,
			Subject:   taskData.task.Subject,
			Resources: resources,
			Reports:   reports,
		})
	}

	// 按布置时间倒序，一样则按 id 倒序	
	sort.Slice(taskReports, func(i, j int) bool {
		if taskReports[i].Reports[0].StatData.StartTime == taskReports[j].Reports[0].StatData.StartTime {
			return taskReports[i].TaskID > taskReports[j].TaskID
		}
		return taskReports[i].Reports[0].StatData.StartTime > taskReports[j].Reports[0].StatData.StartTime
	})

	return taskReports, nil
}

// 获取学生标签 TODO
func (h *TaskReportHandler) getStudentTags(report *dao_task.TaskStudentsReport) []*api.StudentTag {
	tags := make([]*api.StudentTag, 0)
	tags = append(tags, &api.StudentTag{
		Label: "有进步",
		Type:  consts.STUDENT_TAG_TYPE_POSITIVE,
	})
	tags = append(tags, &api.StudentTag{
		Label: "需关注",
		Type:  consts.STUDENT_TAG_TYPE_NEGATIVE,
	})
	tags = append(tags, &api.StudentTag{
		Label: "已提醒",
		Type:  consts.STUDENT_TAG_TYPE_NEUTRAL,
	})
	tags = append(tags, &api.StudentTag{
		Label: "已表扬",
		Type:  consts.STUDENT_TAG_TYPE_NEUTRAL,
	})
	return tags
}
