package dao_task

import (
	"context"
	"gil_teacher/app/consts"
	"gil_teacher/app/model/dto"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// 合并注入函数
var TaskDAOProvider = wire.NewSet(
	NewTaskDAO,
	NewTaskResourceDAO,
	NewTaskAssignDAO,
	NewTeacherTempSelectionDAO,
	NewTaskReportDAO,
	NewTaskStudentDetailsDao,
	NewTaskStudentsReportDao,
	NewTaskReportSettingDao,
	NewTaskStudentDao,
)

// TaskDAO 任务数据访问接口
type TaskDAO interface {
	// GetDB 获取数据库连接
	GetDB() *gorm.DB
	// CreateTask 创建任务
	CreateTask(ctx context.Context, task *Task) error
	// DeleteTasks 批量删除任务
	DeleteTasks(ctx context.Context, taskIDs []int64, creatorID int64) error
	// UpdateTask 更新任务
	UpdateTask(ctx context.Context, taskID int64, creatorID int64, updates map[string]any) error
	// ListTasks 获取任务列表
	ListTasks(ctx context.Context, conditions map[string]any, page, pageSize int64) ([]*Task, int64, error)
	// GetTaskByIDAndCreatorID 获取指定任务
	GetTaskByIDAndCreatorID(ctx context.Context, taskID int64, creatorID int64) (*Task, error)
	// 通过id 查询任务列表
	GetTasksByIDs(ctx context.Context, taskIDs []int64) ([]*Task, error)
	// GetUserTasks 获取指定用户创建的全部任务列表
	GetUserTasks(ctx context.Context, reqs *dto.TaskAssignListQuery, pageInfo *consts.DBPageInfo) ([]*Task, error)
	// GetLatestTask 获取指定教师最近一次布置的任务（每种任务类型）
	GetLatestTask(ctx context.Context, teacherID, schoolID, subjectID int64) ([]*Task, error)
	// GetStudentTaskList 获取学生的任务列表
	GetStudentTaskList(ctx context.Context, req *dto.StudentTaskListQuery) ([]*TaskAndTaskAssign, int64, error)
}

// TaskResourceDAO 任务资源数据访问接口
type TaskResourceDAO interface {
	// GetDB 获取数据库连接
	GetDB() *gorm.DB
	// Create 创建任务资源关联
	Create(ctx context.Context, resource *TaskResource) error
	// GetByTaskID 获取任务关联的资源列表
	GetByTaskID(ctx context.Context, taskID int64) ([]*TaskResource, error)
	// GetByResourceID 获取资源关联的任务列表
	GetByResourceID(ctx context.Context, resourceID int64) ([]*TaskResource, error)
	// GetTaskResourcesByTaskIDs 获取指定任务列表的资源
	GetTaskResourcesByTaskIDs(ctx context.Context, taskIDs []int64) ([]*TaskResource, error)
	// 获取任务指定资源
	GetTaskResources(ctx context.Context, taskID int64, resourceID string, resourceType int64) ([]*TaskResource, error)
}

// TaskAssignDAO 任务分配数据访问接口
type TaskAssignDAO interface {
	// GetTeacherTasks 获取指定教师指定对象布置（可能为空）的全部任务
	GetTeacherTasks(ctx context.Context, reqs *dto.TaskAssignListQuery, pageInfo *consts.DBPageInfo) (int64, []*TaskAssign, error)
	// GetTaskAssignsByTaskIDs 获取指定任务的全部布置信息
	GetTaskAssignsByTaskIDs(ctx context.Context, taskIDs []int64) ([]*TaskAssign, error)
	// GetTaskAssignInfo 获取指定任务指定布置的统计数据
	GetTaskAssignInfo(ctx context.Context, taskID int64, assignID int64) ([]*TaskAssign, error)
	// GetTaskAssigns 获取指定任务指定布置的统计数据
	GetTaskAssigns(ctx context.Context, taskID int64, assignIds []int64) ([]*TaskAssign, error)
	// UpdateTaskAssign 更新任务分配
	UpdateTaskAssign(ctx context.Context, assginID int64, updates map[string]any) error
	// DeleteTaskAssign 删除任务分配
	DeleteTaskAssign(ctx context.Context, taskID int64, assignIDs []int64) error
	// CountTaskAssignByTaskID 统计指定任务的分配数量
	CountTaskAssignByTaskID(ctx context.Context, taskID int64) (int64, error)
	// GetResourceAssignedClassIDs 获取资源已布置的班级ID列表
	GetResourceAssignedClassIDs(ctx context.Context, req *dto.TaskResourceAssignedClassIDsRequest) ([]ResourceGroupID, error)
}

// TaskStudentDAO 任务学生数据访问接口
type TaskStudentDAO interface {
	// GetTaskAssignStudents 获取任务布置对象的学生ID列表
	// [studentID]int64
	GetTaskAssignStudents(ctx context.Context, taskID int64, assignID int64) ([]int64, error)

	// GetAssignStudents 获取指定布置ID列表的学生ID列表
	//  map[assignID][studentID]int64
	GetAssignStudents(ctx context.Context, assignIDs []int64) (map[int64][]int64, error)
}

// TeacherTempSelectionDAO 教师临时选择数据访问接口
type TeacherTempSelectionDAO interface {
	// CreateSelection 创建教师临时选择
	CreateSelection(ctx context.Context, selection *TeacherTempSelection) error
	// DeleteSelections 删除教师临时选择
	DeleteSelections(ctx context.Context, resourceType int64, resourceIDs []string, teacherID int64) error
	// GetSelectionsByTeacherID 获取教师的临时选择列表
	GetSelectionsByTeacherID(ctx context.Context, teacherID int64) ([]*TeacherTempSelection, error)
}

// TaskReportDAO 任务统计数据访问接口
type TaskReportDAO interface {
	// FindAll 分页查找指定任务的全部统计数据
	FindAll(ctx context.Context, taskID int64, pageInfo *consts.DBPageInfo) ([]*TaskReport, error)
	// FindByTaskIDAndAssignIDs 获取指定任务的指定布置ID列表的统计数据
	FindByTaskIDAndAssignIDs(ctx context.Context, taskID int64, assignIDs []int64) ([]*TaskReport, error)
	// FindByAssignIDs 获取指定布置ID列表的统计数据
	FindByTaskAssignIDs(ctx context.Context, taskAssignIdsMap map[int64][]int64) (map[int64][]*TaskReport, error)
}

type TaskStudentDetailsDao interface {
	// 指定任务指定布置指定学生 id 列表的统计数据. 返回
	//     []*TaskStudentDetails, totalCount, incorrectCount
	GetTaskStudentAnswers(ctx context.Context, studentID int64, query *dto.StudentTaskReportQuery, pageInfo *consts.DBPageInfo) ([]*TaskStudentDetails, int64, int64, error)

	// 指定任务指定布置指定学生 id 列表的统计数据
	GetTaskAnswerCountStat(ctx context.Context, taskID int64, assignID int64, resourceQuestionIDs []string) (*dto.TaskAnswerStat, error)

	// 指定任务指定布置指定资源指定题目的全部作答结果. 返回
	//	[]*TaskStudentDetails
	GetTaskAssignAnswerDetails(ctx context.Context, query *dto.TaskAssignAnswersQuery, pageInfo *consts.DBPageInfo) ([]*TaskStudentDetails, error)

	// 获取指定任务布置下每个题目的正确率
	GetTaskAnswerAccuracy(ctx context.Context, query *dto.TaskReportCommonQuery) (map[string]float64, error)

	// 获取指定题目的正确率数据
	//  map[questionKey]*QuestionAnswer
	GetQuestionAnswers(ctx context.Context, taskID, assignID int64, resourceQuestionIDs map[string][]string) (map[string]*QuestionAnswer, error)

	// 计算指定布置任务每个题的答题时间和答题人数，用于计算每个题目的平均用时
	GetTaskAnswerTime(ctx context.Context, taskID int64, assignID int64, resourceQuestionIDs []string) (map[string]int64, map[string]int64, error)
}

type TaskStudentsReportDao interface {
	// 指定任务指定布置指定学生 id 列表的统计数据
	FindTaskStudentsReports(ctx context.Context, taskID int64, assignID int64, studentIDs []int64, pageInfo *consts.DBPageInfo) ([]*TaskStudentsReport, int64, error)

	// 指定任务指定布置指定学生 id 的统计数据
	FindByTaskIDAndStudentID(ctx context.Context, taskID int64, assignID int64, studentID int64) (*TaskStudentsReport, error)
}

type TaskReportSettingDao interface {
	// 创建任务报告设置
	CreateTaskReportSetting(ctx context.Context, setting *TaskReportSetting) error
	// 更新任务报告设置
	UpdateTaskReportSetting(ctx context.Context, id int64, setting *TaskReportSetting) error
	// 查询指定学校班级指定科目的设置
	GetSettingByClassIDAndSubjectID(ctx context.Context, schoolID, classID, subjectID int64) (*TaskReportSetting, error)
}
