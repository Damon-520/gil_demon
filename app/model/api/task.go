package api

import (
	"gil_teacher/app/consts"
	"gil_teacher/app/controller/http_server/response"
	dao_task "gil_teacher/app/dao/task"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/model/itl"
	"gil_teacher/app/utils"
)

// SubjectTaskType 学科任务类型
type SubjectTaskType struct {
	SubjectKey  int64   `json:"subjectKey"`  // 学科 Key
	SubjectName string  `json:"subjectName"` // 学科名称
	TaskTypes   []int64 `json:"taskTypes"`   // 任务类型列表
}

// GetTaskTypeResponse 获取任务类型响应
type GetTaskTypeResponse struct {
	SubjectTaskTypes []SubjectTaskType `json:"subjectTaskTypes"` // 学科任务类型列表
}

// TaskResource 任务关联资源
type TaskResource struct {
	ResourceID    string `json:"resourceId" binding:"required"`
	ResourceType  int64  `json:"resourceType" binding:"required"`
	ResourceExtra string `json:"resourceExtra"` // 资源额外信息
}

// StudentGroups 任务关联学生群组
type StudentGroup struct {
	GroupType  int64   `json:"groupType" binding:"required"` // 群组类型，1 自定义学生，2 班级，3 群组
	GroupID    int64   `json:"groupId" binding:"required"`   // 群组ID，自定义时为0，班级时为班级ID，群组时为群组ID
	StudentIDs []int64 `json:"studentIds,omitempty"`         // 类型为自定义学生时，前端传递学生ID列表，类型为班级或群组时后端自动填充
	StartTime  int64   `json:"startTime" binding:"required"` // 开始时间
	Deadline   int64   `json:"deadline" binding:"required"`  // 截止时间
}

// GetTaskDetailByIDResponse 根据ID查询任务详情响应
type GetTaskDetailByIDResponse struct {
	*dao_task.Task
	Resources     []*dao_task.TaskResource `json:"resources"`     // 任务关联资源列表
	StudentGroups []*dao_task.TaskAssign   `json:"studentGroups"` // 任务关联学生群组列表
}

// GetStudentTaskListRequest 查询学生任务列表请求
type GetStudentTaskListRequest struct {
	*dto.StudentTaskListQuery
}

func (r *GetStudentTaskListRequest) Validate() *response.Response {
	if r.StudentID <= 0 {
		return &response.ERR_INVALID_STUDENT
	}
	if !consts.SubjectExists(r.Subject) {
		return &response.ERR_SUBJECT
	}
	if r.TaskType != 0 && !consts.TaskTypeExists(r.TaskType) {
		return &response.ERR_INVALID_TASK_TYPE
	}
	if r.StartTimeFrom != 0 && !utils.IsValidUnixTimestamp(r.StartTimeFrom) {
		return &response.ERR_INVALID_TIME
	}
	if r.StartTimeTo != 0 && !utils.IsValidUnixTimestamp(r.StartTimeTo) {
		return &response.ERR_INVALID_TIME
	}
	if r.DeadlineFrom != 0 && !utils.IsValidUnixTimestamp(r.DeadlineFrom) {
		return &response.ERR_INVALID_TIME
	}
	if r.DeadlineTo != 0 && !utils.IsValidUnixTimestamp(r.DeadlineTo) {
		return &response.ERR_INVALID_TIME
	}

	r.ValidSortKeys = []string{"task_id", "start_time", "deadline"}
	if err := r.APIReqeustPageInfo.Check(); err != nil {
		return err
	}
	return nil
}

// GetStudentTaskListResponse 查询学生任务列表响应
type GetStudentTaskListResponse struct {
	List []struct {
		*dao_task.TaskAndTaskAssign
		Resources []*dao_task.TaskResource `json:"resources"` // 任务资源
	} `json:"list"` // 任务列表
	*consts.ApiPageResponse // 分页信息
}

// CreateTaskRequestBody 创建任务请求体
type CreateTaskRequestBody struct {
	SchoolID       int64          `json:"-"`
	Phase          int64          `json:"-"`
	Subject        int64          `json:"subject" binding:"required"`
	TaskType       int64          `json:"taskType" binding:"required"`
	TaskName       string         `json:"taskName" binding:"required"`
	TeacherComment string         `json:"teacherComment,omitempty"` // 老师留言
	TaskExtraInfo  string         `json:"taskExtraInfo,omitempty"`  // 任务额外信息
	BizTreeID      int64          `json:"bizTreeId,omitempty"`      // 业务树ID，创建课程任务时使用
	CreatorID      int64          `json:"-"`
	UpdaterID      int64          `json:"-"`
	Resources      []TaskResource `json:"resources" binding:"required"`
	StudentGroups  []StudentGroup `json:"studentGroups" binding:"required"`
}

func (c *CreateTaskRequestBody) Validate() *response.Response {
	if c.TaskName == "" {
		return &response.ERR_EMPTY_TASK_NAME
	}
	if _, ok := consts.TaskTypeNameMap[c.TaskType]; !ok {
		return &response.ERR_INVALID_TASK_TYPE
	}
	if len(c.Resources) == 0 {
		return &response.ERR_EMPTY_RESOURCE
	}
	for _, resource := range c.Resources {
		if _, ok := consts.ResourceTypeNameMap[resource.ResourceType]; !ok {
			return &response.ERR_INVALID_RESOURCE_TYPE
		}
	}
	if len(c.StudentGroups) == 0 {
		return &response.ERR_EMPTY_STUDENT
	}
	// 课程任务必须传递业务树ID
	if c.TaskType == consts.TASK_TYPE_COURSE && c.BizTreeID <= 0 {
		return &response.ERR_BIZ_TREE
	}
	// 当次任务不能重复派发到同一个班级
	classIDs := make(map[int64]struct{})
	for _, studentGroup := range c.StudentGroups {
		if !utils.IsValidUnixTimestamp(studentGroup.StartTime, studentGroup.Deadline) {
			return &response.ERR_INVALID_TIME
		}
		if studentGroup.StartTime >= studentGroup.Deadline {
			return &response.ERR_INVALID_TIME
		}
		// 目前只开放班级
		if studentGroup.GroupType != consts.TASK_GROUP_TYPE_CLASS {
			return &response.ERR_EMPTY_STUDENT
		}
		if studentGroup.GroupID <= 0 {
			return &response.ERR_EMPTY_STUDENT
		}
		// 班级ID不能重复
		if _, ok := classIDs[studentGroup.GroupID]; ok {
			return &response.ERR_DUP_CLASS
		}
		classIDs[studentGroup.GroupID] = struct{}{}
	}

	return nil
}

// DeleteTaskRequestBody 删除任务请求体
type DeleteTaskRequestBody struct {
	TaskIDs []int64 `json:"taskIds" binding:"required"`
}

// UpdateTaskRequestBody 更新任务请求体
type UpdateTaskRequestBody struct {
	TaskID         int64  `json:"taskId" binding:"required"`
	TaskName       string `json:"taskName,omitempty"`
	TeacherComment string `json:"teacherComment,omitempty"`
}

func (u *UpdateTaskRequestBody) Validate() *response.Response {
	if u.TaskName == "" && u.TeacherComment == "" {
		return &response.ERR_EMPTY_TASK_NAME_COMMENT
	}
	return nil
}

// UpdateTaskAssignRequestBody 更新任务分配请求体
type UpdateTaskAssignRequestBody struct {
	AssignID  int64 `json:"assignId" binding:"required"`
	StartTime int64 `json:"startTime"`
	Deadline  int64 `json:"deadline"`
}

// DeleteTaskAssignRequestBody 删除任务分配请求体
type DeleteTaskAssignRequestBody struct {
	TaskID    int64   `json:"taskId" binding:"required"`
	AssignIDs []int64 `json:"assignIds" binding:"required"`
}

func (d *DeleteTaskAssignRequestBody) Validate() *response.Response {
	if d.TaskID <= 0 {
		return &response.ERR_INVALID_TASK
	}
	if len(d.AssignIDs) == 0 {
		return &response.ERR_INVALID_ASSIGN
	}
	for _, assignID := range d.AssignIDs {
		if assignID <= 0 {
			return &response.ERR_INVALID_ASSIGN
		}
	}
	return nil
}

// GetQuestionListRequest 查询题目列表请求
type GetQuestionListRequest struct {
	Phase             int64   `json:"-"`
	Subject           int64   `json:"subject" binding:"required"`
	BizTreeNodeIds    []int64 `json:"bizTreeNodeIds,omitempty"`
	Keyword           string  `json:"keyword,omitempty"`
	QuestionType      []int64 `json:"questionType,omitempty"`
	QuestionDifficult []int64 `json:"questionDifficult,omitempty"`
	QuestionYears     []int64 `json:"questionYears,omitempty"`
	Page              int64   `json:"page,omitempty"`
	PageSize          int64   `json:"pageSize,omitempty"`
	Sort              string  `json:"sort,omitempty"` // 题库支持：createTime 最新题目，useCount 最多使用
}

func (r *GetQuestionListRequest) Validate() *response.Response {
	apiPageInfo := &consts.APIReqeustPageInfo{
		Page:          r.Page,
		PageSize:      r.PageSize,
		SortBy:        r.Sort,
		ValidSortKeys: []string{"createTime", "useCount"},
	}
	if err := apiPageInfo.Check(); err != nil {
		return err
	}
	return nil
}

// GetQuestionListByIDsRequest 通过ID列表查询题目详情请求
type GetQuestionListByIDsRequest struct {
	QuestionIDs []string `json:"questionIds" binding:"required"`
}

// GetAICourseAndPracticeListResponse 获取业务树节点对应的 AI 课和巩固练习列表响应
type GetAICourseAndPracticeListResponse []AICourseAndPracticeInfo

// AICourseAndPracticeInfo 业务树节点对应的 AI 课和巩固练习信息
type AICourseAndPracticeInfo struct {
	BizTreeNodeID   int64    `json:"bizTreeNodeId"`   // 业务树节点ID
	BizTreeNodeName string   `json:"bizTreeNodeName"` // 业务树节点名称
	IsRecommend     bool     `json:"isRecommend"`     // 是否推荐，暂无
	AICourse        struct{} `json:"aiCourse"`        // AI 课，暂无
	Practice        Practice `json:"practice"`        // 巩固练习基本信息
}

// QuestionSet 巩固练习基本信息
type Practice struct {
	ID               int64   `json:"questionSetId,omitempty"`    // 题集ID，巩固练习是题集的一种
	EstimatedTime    int64   `json:"estimatedTime"`              // 预估时间，单位：分钟，暂无
	AssignedClassIDs []int64 `json:"assignedClassIds,omitempty"` // 巩固练习已经布置过的班级ID列表
}

// GetQuestionSetDetailResponse 题集详情响应
type GetQuestionSetDetailResponse struct {
	QuestionTypeCount int64 `json:"questionTypeCount"` // 题型数量
	EstimatedTime     int64 `json:"estimatedTime"`     // 预估时间，单位：分钟，暂无
	itl.QuestionSetStableInfo
}
