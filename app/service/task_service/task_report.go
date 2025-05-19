package task_service

import (
	"context"
	"errors"
	"time"

	"gil_teacher/app/consts"
	"gil_teacher/app/controller/http_server/response"
	"gil_teacher/app/core/logger"
	dao_task "gil_teacher/app/dao/task"
	"gil_teacher/app/model/api"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/utils"

	"gorm.io/gorm"
)

type TaskReportService struct {
	taskReportDao         dao_task.TaskReportDAO
	taskStudentsReportDao dao_task.TaskStudentsReportDao
	taskStudentDetailsDao dao_task.TaskStudentDetailsDao
	taskReportSettingDao  dao_task.TaskReportSettingDao
	log                   *logger.ContextLogger
}

func NewTaskStatService(
	taskStatsDao dao_task.TaskReportDAO,
	taskStudentsStatsDao dao_task.TaskStudentsReportDao,
	taskStudentDetailsDao dao_task.TaskStudentDetailsDao,
	taskReportSettingDao dao_task.TaskReportSettingDao,
	log *logger.ContextLogger,
) *TaskReportService {
	return &TaskReportService{
		taskReportDao:         taskStatsDao,
		taskStudentsReportDao: taskStudentsStatsDao,
		taskStudentDetailsDao: taskStudentDetailsDao,
		taskReportSettingDao:  taskReportSettingDao,
		log:                   log,
	}
}

// 查询指定任务指定布置的统计数据
func (s *TaskReportService) GetTaskAssignStats(ctx context.Context, taskID int64, assignID int64) (*dao_task.TaskReport, error) {
	if taskID == 0 || assignID == 0 {
		return nil, nil
	}

	stats, err := s.GetTaskAssignsStats(ctx, taskID, []int64{assignID})
	if err != nil {
		return nil, err
	}

	if len(stats) == 0 {
		return nil, nil
	}

	return stats[0], nil
}

// 获取指定任务指定布置指定学生 id 列表的统计数据
func (s *TaskReportService) GetTaskAssignStudentReports(ctx context.Context, taskID int64, assignID int64, studentIDs []int64, pageInfo *consts.APIReqeustPageInfo) (map[int64]*dao_task.TaskStudentsReport, int64, error) {
	if taskID == 0 || assignID == 0 {
		return nil, 0, nil
	}

	stats, count, err := s.taskStudentsReportDao.FindTaskStudentsReports(ctx, taskID, assignID, studentIDs, pageInfo.ToDBPageInfo())
	if err != nil {
		return nil, 0, err
	}

	statsMap := make(map[int64]*dao_task.TaskStudentsReport)
	for _, stat := range stats {
		statsMap[stat.StudentID] = stat
	}
	return statsMap, count, nil
}

// 获取指定任务的指定布置ID列表的统计数据
//
//	[]*dao_task.TaskReport
func (s *TaskReportService) GetTaskAssignsStats(ctx context.Context, taskID int64, assignIDs []int64) ([]*dao_task.TaskReport, error) {
	if taskID == 0 || len(assignIDs) == 0 {
		return nil, nil
	}

	stats, err := s.taskReportDao.FindByTaskIDAndAssignIDs(ctx, taskID, assignIDs)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return stats, err
}

// 获取多个 任务id + 布置ID列表 的统计数据，按任务维度合并返回. 返回
//
//	map[taskID][assignID]*dao_task.TaskReport
func (s *TaskReportService) GetTaskReportsByTaskAssignIDs(ctx context.Context, taskAssignIdsMap map[int64][]int64) (map[int64][]*dao_task.TaskReport, error) {
	if len(taskAssignIdsMap) == 0 {
		return nil, nil
	}

	reports, err := s.taskReportDao.FindByTaskAssignIDs(ctx, taskAssignIdsMap)
	if err != nil {
		return nil, err
	}

	reportMap := make(map[int64][]*dao_task.TaskReport) // taskID -> []TaskReport
	for taskId, reports := range reports {
		for _, report := range reports {
			if _, ok := reportMap[taskId]; !ok {
				reportMap[taskId] = make([]*dao_task.TaskReport, 0)
			}
			reportMap[taskId] = append(reportMap[taskId], report)
		}
	}
	return reportMap, nil
}

// 获取指定任务全部学生的作答详情
//
//	map[int64]map[string]*dao_task.TaskStudentDetails,  studentID -> question_key -> TaskStudentDetails
func (s *TaskReportService) GetTaskAssignAnswers(ctx context.Context, query *dto.TaskAssignAnswersQuery, pageInfo *consts.APIReqeustPageInfo) (map[int64]map[string]*dao_task.TaskStudentDetails, error) {
	if query.TaskID == 0 || query.AssignID == 0 {
		return nil, errors.New("taskID, assignID is required")
	}
	studentAnswers, err := s.taskStudentDetailsDao.GetTaskAssignAnswerDetails(ctx, query, pageInfo.ToDBPageInfo())
	if err != nil {
		return nil, err
	}

	studentAnswersMap := make(map[int64]map[string]*dao_task.TaskStudentDetails)
	for _, answer := range studentAnswers {
		if _, ok := studentAnswersMap[answer.StudentID]; !ok {
			studentAnswersMap[answer.StudentID] = make(map[string]*dao_task.TaskStudentDetails)
		}
		questionKey := utils.JoinList([]any{answer.ResourceKey, answer.QuestionID}, consts.CombineKey)
		studentAnswersMap[answer.StudentID][questionKey] = answer
	}
	return studentAnswersMap, nil
}

// 获取指定任务的答题面板统计数据
func (s *TaskReportService) GetTaskAnswerAccuracy(ctx context.Context, query *dto.TaskReportCommonQuery) (map[string]float64, error) {
	if query.TaskID == 0 || query.AssignID == 0 {
		return nil, nil
	}

	answerAccuracy, err := s.taskStudentDetailsDao.GetTaskAnswerAccuracy(ctx, query)
	if err != nil {
		return nil, err
	}
	return answerAccuracy, nil
}

// 获取指定任务，指定资源题目的答题面板统计数据
func (s *TaskReportService) GetTaskAnswerAccuracyByResource(ctx context.Context, taskId, assginId int64, questionIdMap map[string][]string) (map[string]*dao_task.QuestionAnswer, error) {
	if taskId == 0 || assginId == 0 {
		return nil, nil
	}

	answerAccuracy, err := s.taskStudentDetailsDao.GetQuestionAnswers(ctx, taskId, assginId, questionIdMap)
	if err != nil {
		return nil, err
	}
	return answerAccuracy, nil
}

// 获取指定学校指定班级指定科目的任务报告设置
func (s *TaskReportService) GetTaskReportSetting(ctx context.Context, schoolID, classID, subjectID int64) (*api.TaskReportSetting, error) {
	reportSetting, err := s.taskReportSettingDao.GetSettingByClassIDAndSubjectID(ctx, schoolID, classID, subjectID)
	if err != nil {
		s.log.Error(ctx, "GetTaskReportSetting failed, schoolID:%d, classID:%d, subjectID:%d, error:%v", schoolID, classID, subjectID, err)
		return nil, err
	}

	var setting *dao_task.Setting
	if reportSetting != nil {
		setting = reportSetting.Setting
	}

	apiSetting := &api.TaskReportSetting{
		ClassID:       classID,
		SubjectID:     subjectID,
		ReportSetting: setting,
	}
	return apiSetting, nil
}

// 新增或更新任务报告设置
func (s *TaskReportService) SaveTaskReportSetting(ctx context.Context, schoolID, teacherID int64, setting *api.TaskReportSetting) *response.Response {
	if schoolID == 0 || teacherID == 0 {
		return &response.ERR_INVALID_TASK_REPORT_SETTING
	}

	if setting == nil {
		return &response.ERR_INVALID_TASK_REPORT_SETTING
	}

	reportSetting, err := s.taskReportSettingDao.GetSettingByClassIDAndSubjectID(ctx, schoolID, setting.ClassID, setting.SubjectID)
	if err != nil {
		s.log.Error(ctx, "GetTaskReportSetting failed, schoolID:%d, teacherID:%d, setting:%v, error:%v", schoolID, teacherID, setting, err)
		return &response.ERR_SYSTEM
	}

	// 只有任课教师能修改配置
	if reportSetting != nil && reportSetting.TeacherID != teacherID {
		return &response.ERR_NO_PERMISSION_TO_MODIFY
	}

	if reportSetting == nil {
		reportSetting = &dao_task.TaskReportSetting{
			SchoolID:   schoolID,
			TeacherID:  teacherID,
			ClassID:    setting.ClassID,
			Subject:    setting.SubjectID,
			Setting:    setting.ReportSetting,
			CreateTime: time.Now().Unix(),
		}
		err = s.taskReportSettingDao.CreateTaskReportSetting(ctx, reportSetting)
	} else {
		reportSetting.UpdateTime = time.Now().Unix()
		err = s.taskReportSettingDao.UpdateTaskReportSetting(ctx, reportSetting.ID, reportSetting)
	}

	if err != nil {
		return &response.ERR_SYSTEM
	}
	return nil
}

// GetTaskReportByTaskIDAndStudentID 获取指定任务指定布置指定学生的统计数据
func (s *TaskReportService) GetTaskReportByTaskIDAndStudentID(ctx context.Context, taskID int64, assignID int64, studentID int64) (*dao_task.TaskStudentsReport, error) {
	if taskID == 0 || assignID == 0 || studentID == 0 {
		return nil, nil
	}

	report, err := s.taskStudentsReportDao.FindByTaskIDAndStudentID(ctx, taskID, assignID, studentID)
	if err != nil {
		return nil, err
	}
	return report, nil
}
