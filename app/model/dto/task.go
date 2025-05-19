package dto

import (
	"gil_teacher/app/consts"
	"gil_teacher/app/model/itl"
)

// 教师布置的全部任务列表（非报告）
type TaskAssignListQuery struct {
	TeacherID int64                `json:"teacherId"` // 教师ID
	SchoolID  int64                `json:"schoolId"`  // 学校ID，教师可能在多个学校任教
	Keyword   string               `json:"keyword"`   // 任务关键词，支持模糊查询
	Subject   int64                `json:"subject"`   // 科目ID，为 0 时查询全部科目，仅在班主任查看时有效，其他教师只能查看任教科目
	GroupType int64                `json:"groupType"` // 群组类型，0 为全部，1 为班级，2 为群组（目前只有虚拟群组）
	GroupID   int64                `json:"groupId"`   // 群组ID或班级ID，根据 groupType 确定
	StartTime int64                `json:"startTime"` // 开始时间
	EndTime   int64                `json:"endTime"`   // 结束时间
	TaskType  int64                `json:"taskType"`  // 任务类型，默认查询全部任务类型
	ClassInfo map[int64]*itl.Class `json:"-"`         // 教师班级信息
}

// 教师最近布置的作业报告列表
type TeacherLatestTaskReportsQuery struct {
	TeacherID int64                `json:"teacherId"` // 教师ID
	SchoolID  int64                `json:"schoolId"`  // 学校ID
	Subject   int64                `json:"subject"`   // 科目ID
	ClassInfo map[int64]*itl.Class `json:"classInfo"` // 教师班级信息
}

// 查询指定任务指定布置对象的完整报告数据
type TaskAssignReportQuery struct {
	TaskID       int64   `json:"taskId"`       // 任务ID
	AssignID     int64   `json:"assignId"`     // 布置对象ID
	ResourceID   string  `json:"resourceId"`   // 资源ID
	ResourceType int64   `json:"resourceType"` // 资源类型
	StudentName  string  `json:"studentName"`  // 学生姓名，可以模糊查询
	StudentIDs   []int64 `json:"-"`            // 学生ID，可以精确查询
}

// 查询指定任务指定布置对象的答题信息
type TaskAssignAnswersQuery struct {
	TaskReportCommonQuery
	QuestionIDs []string `json:"questionIds,omitempty"` // 题目ID列表
	StudentIDs  []int64  `json:"studentIds,omitempty"`  // 学生ID列表
}

// 查询指定学生的作业报告
type StudentTaskReportQuery struct {
	TaskReportCommonQuery
	QuestionIDs []string `json:"questionIds,omitempty"` // 题目ID列表
}

// 查询作业报告公共参数
type TaskReportCommonQuery struct {
	TaskID       int64           `json:"taskId"`       // 任务ID
	AssignID     int64           `json:"assignId"`     // 布置对象ID
	ResourceID   string          `json:"resourceId"`   // 资源ID
	ResourceType int64           `json:"resourceType"` // 资源类型
	QuestionType int64           `json:"questionType"` // 题目类型
	Keyword      string          `json:"keyword"`      // 题目关键词
	AllQuestions bool            `json:"allQuestions"` // 是否查询全部题目
	SortKey      string          `json:"sortKey"`      // 排序关键字
	SortType     consts.SortType `json:"sortType"`     // 排序类型 asc(默认) 或 desc
}

// StudentTaskListQuery 查询学生任务列表请求
type StudentTaskListQuery struct {
	StudentID                  int64 `json:"studentId" binding:"required"` // 学生ID，必传参数
	Subject                    int64 `json:"subject" binding:"required"`   // 学科，必传参数
	TaskType                   int64 `json:"taskType,omitempty"`           // 任务类型
	TaskSubType                int64 `json:"taskSubType,omitempty"`        // 任务子类型
	StartTimeFrom              int64 `json:"startTimeFrom,omitempty"`      // 开始时间 >=
	StartTimeTo                int64 `json:"startTimeTo,omitempty"`        // 开始时间 <
	DeadlineFrom               int64 `json:"deadlineFrom,omitempty"`       // 截止时间 >=
	DeadlineTo                 int64 `json:"deadlineTo,omitempty"`         // 截止时间 <
	*consts.APIReqeustPageInfo       // 分页和排序
}

// 查询资源已布置的班级ID列表
type TaskResourceAssignedClassIDsRequest struct {
	TeacherID    int64    `json:"teacherId"`    // 教师ID
	Subject      int64    `json:"subject"`      // 科目ID
	TaskType     int64    `json:"taskType"`     // 任务类型
	ResourceType int64    `json:"resourceType"` // 资源类型
	ResourceIDs  []string `json:"resourceIds"`  // 资源ID列表
	GroupType    int64    `json:"groupType"`    // 群组类型
}

// 查询答题面板统计数据
type AnswerPanelQuery struct {
	TaskID       int64  `json:"taskId"`       // 任务ID
	AssignID     int64  `json:"assignId"`     // 任务布置ID
	ResourceID   string `json:"resourceId"`   // 资源ID
	ResourceType int64  `json:"resourceType"` // 资源类型
}

// 导出作业报告
type ExportTaskReportQuery struct {
	TaskID       int64           `json:"taskId"`       // 任务ID
	AssignID     int64           `json:"assignId"`     // 任务布置ID
	ResourceID   string          `json:"resourceId"`   // 资源ID
	ResourceType int64           `json:"resourceType"` // 资源类型
	SortBy       string          `json:"sortBy"`       // 排序字段
	SortType     consts.SortType `json:"sortType"`     // 排序类型
	Fields       []string        `json:"fields"`       // 导出字段
}

// 作业报告结果格式
type ExportTaskReportResult struct {
	Meta []string   `json:"meta"` // csv 首行
	Data [][]string `json:"data"` // csv 数据
}

// 任务作业统计数据
type TaskAnswerStat struct {
	ResourceAnswerCount    map[string]int64 `json:"resourceAnswerCount"`    // 资源答题数 resource_key -> answer_count
	ResourceIncorrectCount map[string]int64 `json:"resourceIncorrectCount"` // 资源错题数 resource_key -> incorrect_count
	ResourceTotalCostTime  map[string]int64 `json:"resourceTotalCostTime"`  // 资源总用时 resource_key -> total_cost_time
}
