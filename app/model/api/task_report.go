package api

import (
	"errors"
	"gil_teacher/app/consts"
	dao_task "gil_teacher/app/dao/task"
	"gil_teacher/app/model/itl"
)

/*************************************************************
            		任务报告首页汇总
*************************************************************/
// 任务报告列表
type TaskReportsResponse struct {
	Tasks    []*TaskReport           `json:"tasks"`
	PageInfo *consts.ApiPageResponse `json:"pageInfo"`
}

// 最近布置的任务报告（每种类型一个）
type LatestReportsResponse struct {
	Tasks []*TaskReport `json:"tasks"`
}

// 单个任务报告(包含全部布置对象)
type TaskReport struct {
	CreatorID int64       `json:"creatorId"` // 创建者ID
	TaskID    int64       `json:"taskId"`    // 任务ID
	TaskName  string      `json:"taskName"`  // 任务名称
	TaskType  int64       `json:"taskType"`  // 任务类型
	Subject   int64       `json:"subject"`   // 科目ID
	Resources []*Resource `json:"resources"` // 资源列表
	Reports   []*Report   `json:"reports"`   // 任务对象统计信息
}

// 资源信息
type Resource struct {
	ID   string `json:"id"`   // 资源ID
	Type int64  `json:"type"` // 资源类型
	Name string `json:"name"` // 资源名称
}

// 单个任务布置对象的统计信息
type Report struct {
	AssignID     int64             `json:"assignId"`     // 任务布置ID
	AssignObject *AssignObject     `json:"assignObject"` // 布置对象信息
	StatData     *AssignObjectStat `json:"statData"`     // 布置对象任务的统计信息
}

// 任务对象信息(班级、小组等的名称)
type AssignObject struct {
	ID       int64  `json:"id"`                 // 任务对象ID
	Type     int64  `json:"type"`               // 任务对象类型
	Name     string `json:"name"`               // 任务对象名称
	Students []any  `json:"students,omitempty"` // 任务对象学生列表
}

// 任务中每个布置对象的统计信息
type AssignObjectStat struct {
	// 非资源类任务数据
	CompletionRate           float64 `json:"completionRate,omitempty"`           // 完成率
	CorrectRate              float64 `json:"correctRate,omitempty"`              // 正确率
	NeedAttentionQuestionNum int64   `json:"needAttentionQuestionNum,omitempty"` // 待关注题目数
	// 资源类任务数据
	AverageProgress float64 `json:"averageProgress,omitempty"` // 平均进度
	ClassHours      int64   `json:"classHours,omitempty"`      // 课时数
	// 任务时间
	StartTime int64 `json:"startTime"` // 开始时间
	Deadline  int64 `json:"deadline"`  // 结束时间
}

/*************************************************************
			            任务报告设置
*************************************************************/
// 任务报告设置(查询、设置、更新)
type TaskReportSetting struct {
	ClassID       int64             `json:"classId"`       // 班级ID
	SubjectID     int64             `json:"subject"`       // 学科ID
	ReportSetting *dao_task.Setting `json:"reportSetting"` // 设置
}

func (s *TaskReportSetting) Validate() error {
	if s.ClassID == 0 || s.SubjectID == 0 {
		return errors.New("classId and subjectId is required")
	}
	if s.ReportSetting == nil {
		return errors.New("reportSetting is required")
	}
	return nil
}

/*************************************************************
		            学生单次任务报告汇总
*************************************************************/
// 学生层级
type StudentLevel string

const (
	StudentLevelSPlus StudentLevel = "S+"
	StudentLevelS     StudentLevel = "S"
	StudentLevelA     StudentLevel = "A"
	StudentLevelB     StudentLevel = "B"
	StudentLevelC     StudentLevel = "C"
)

// 报告详情
type TaskReportSummaryDetailResponse struct {
	Students []*Student             `json:"students"` // 学生列表
	Detail   *ReportSummaryDetail   `json:"detail"`   // 报告详情
	PageInfo consts.ApiPageResponse `json:"pageInfo"` // 分页信息
}

type ReportSummaryDetail struct {
	PraiseList      []int64               `json:"praiseList"`      // 值得表扬学生id列表
	AttentionList   []int64               `json:"attentionList"`   // 需要提醒学生id列表
	AvgAccuracy     float64               `json:"avgAccuracy"`     // 平均正确率
	AvgCostTime     int64                 `json:"avgCostTime"`     // 平均用时，秒
	StudentReports  []*TaskStudentReport  `json:"studentReports"`  // 学生维度的任务统计报告
	ResourceReports []*TaskResourceReport `json:"resourceReports"` // 资源维度的任务统计报告
}

// 学生维度的任务统计报告
type TaskStudentReport struct {
	StudentID        int64         `json:"studentId"`        // 学生ID
	StudyScore       int64         `json:"studyScore"`       // 学习分
	Tags             []*StudentTag `json:"tags"`             // 学生标签
	Progress         float64       `json:"progress"`         // 完成进度
	DifficultyDegree float64       `json:"difficultyDegree"` // 答题难度，取用户已做完题目的平均值计算，若缺少难度标签，则取 3（中等难度）
	AccuracyRate     float64       `json:"accuracyRate"`     // 正确率
	IncorrectNum     int64         `json:"incorrectNum"`     // 错题数
	AnswerNum        int64         `json:"answerNum"`        // 答题数
	CostTime         int64         `json:"costTime"`         // 用时数，秒
}

// 资源维度的任务统计报告
type TaskResourceReport struct {
	ResourceID               string  `json:"resourceId"`               // 资源ID
	ResourceType             int64   `json:"resourceType"`             // 资源类型
	ResourceName             string  `json:"resourceName"`             // 资源名称
	CompletionRate           float64 `json:"completionRate"`           // 完成率
	CorrectRate              float64 `json:"correctRate"`              // 正确率
	NeedAttentionQuestionNum int64   `json:"needAttentionQuestionNum"` // 待关注题目数
	NeedAttentionUserNum     int64   `json:"needAttentionUserNum"`     // 待关注学生数
	AverageCostTime          int64   `json:"averageCostTime"`          // 平均用时，秒
}

type StudentTag struct {
	Label string `json:"label"` // 标签内容
	Type  int    `json:"type"`  // 标签类型，1 正向，0 中性，-1 负向
}

/**********************************************************
		        		答题详情
**********************************************************/
// 单次任务布置的答题详情
type TaskAnswerDetailReportResponse struct {
	TaskAnswerReportCommon
	CommonIncorrectCount int64             `json:"commonIncorrectCount"` // 共性错题数
	Students             []*Student        `json:"students"`             // 任务布置学生列表，方便前端展示
	QuestionAnswers      []*QuestionAnswer `json:"questionAnswers"`      // 答题信息
}

// 单个学生单次任务布置的答题详情
type TaskStudentAnswerDetailResponse struct {
	TaskAnswerReportCommon
	InCorrectCount  int64             `json:"inCorrectCount"`  // 错题数
	Progress        float64           `json:"progress"`        // 完成进度
	QuestionAnswers []*QuestionAnswer `json:"questionAnswers"` // 学生作答详情列表
}

// 公共返回
type TaskAnswerReportCommon struct {
	TaskID       int64                  `json:"taskId"`                 // 任务ID
	TaskType     int64                  `json:"taskType,omitempty"`     // 任务类型
	AssignID     int64                  `json:"assignId"`               // 任务ID
	ResourceID   string                 `json:"resourceId,omitempty"`   // 资源ID
	ResourceType int64                  `json:"resourceType,omitempty"` // 资源类型
	TotalCount   int64                  `json:"totalCount"`             // 题目总数
	PageInfo     consts.ApiPageResponse `json:"pageInfo"`               // 分页信息
}

// 每个题目的答题信息
type QuestionAnswer struct {
	QuestionIndex  int64            `json:"-"`                  // 题目序号，从 1 开始，按添加顺序自动赋值
	ResourceID     string           `json:"resourceId"`         // 资源ID
	ResourceType   int64            `json:"resourceType"`       // 资源类型
	AnswerCount    int64            `json:"answerCount"`        // 作答人数
	IncorrectCount int64            `json:"inCorrectCount"`     // 答错人数
	AvgCostTime    int64            `json:"avgCostTime"`        // 平均用时，秒
	Question       *itl.Question    `json:"question,omitempty"` // 题目信息
	Answer         *StudentAnswer   `json:"answer,omitempty"`   // 当前学生的答题信息，查询单个学生答题时返回
	Answers        []*StudentAnswer `json:"answers,omitempty"`  // 每个学生的答题信息，查询所有学生答题时返回
}

// 单个学生的答题详情
type StudentAnswer struct {
	StudentID int64  `json:"studentId"` // 学生ID
	Answer    string `json:"answer"`    // 学生答案
	IsCorrect bool   `json:"isCorrect"` // 是否正确
	CostTime  int64  `json:"costTime"`  // 用时，秒
}

/*************************************************************
		                提问数据
*************************************************************/
// 提问数据
type QuestionDataResponse struct {
	Questions []*StudentQuestion     `json:"questions"` // 题目数据
	PageInfo  consts.ApiPageResponse `json:"pageInfo"`  // 分页信息
}

type StudentQuestion struct {
	KnowledgeID      int64      `json:"knowledgeId"`      // 知识点ID
	QuestionStudents []*Student `json:"questionStudents"` // 提问学生
}

/*************************************************************
		                学生答题详情
*************************************************************/
// 学生答题详情
type StudentAnswerDetailResponse struct {
	StudentID int64                  `json:"studentId"` // 学生ID
	Answers   []*StudentAnswerDetail `json:"answers"`   // 学生答题详情
	PageInfo  consts.ApiPageResponse `json:"pageInfo"`  // 分页信息
}

type StudentAnswerDetail struct {
	QuestionID    int64  `json:"questionId"`    // 题目ID
	Question      string `json:"question"`      // 题目内容
	Answer        string `json:"answer"`        // 答案
	StudentAnswer string `json:"studentAnswer"` // 学生作答
	IsCorrect     bool   `json:"isCorrect"`     // 是否正确
}

/*************************************************************
		        题目面板（单次任务的班级/小组答题正确率汇总）
*************************************************************/
// 题目面板，不分页，返回所有数据
type QuestionPanelResponse struct {
	TaskID   int64            `json:"taskId"`   // 任务ID
	AssignID int64            `json:"assignId"` // 任务布置ID
	Panel    []*QuestionPanel `json:"panel"`    // 题目答题正确率
}

// 单个题目的答题正确率
type QuestionPanel struct {
	ResourceID   string  `json:"resourceId"`   // 资源ID
	ResourceType int64   `json:"resourceType"` // 资源类型
	QuestionID   string  `json:"questionId"`   // 题目ID
	CorrectRate  float64 `json:"correctRate"`  // 答题正确率
	// 下面字段不输出
	QuestionIndex  int64 `json:"-"` // 题目序号，从 1 开始，按添加顺序自动赋值
	AnswerCount    int64 `json:"-"` // 答题数
	IncorrectCount int64 `json:"-"` // 答错数
}

/*************************************************************
		                学生作业报告处理
*************************************************************/
// 学生作业报告处理请求
type StudentReportHandleRequest struct {
	TaskID       int64               `json:"taskId" binding:"required"`       // 任务ID
	AssignID     int64               `json:"assignId" binding:"required"`     // 任务布置ID
	StudentIDs   []int64             `json:"studentIds" binding:"required"`   // 学生ID列表，支持批量操作
	BehaviorType consts.BehaviorType `json:"behaviorType" binding:"required"` // 行为类型：点赞，提醒
	Content      string              `json:"content"`                         // 行为内容
}

func (r *StudentReportHandleRequest) Validate() error {
	if r.TaskID <= 0 || r.AssignID <= 0 {
		return errors.New("taskId and assignId is required")
	}
	if len(r.StudentIDs) == 0 {
		return errors.New("studentIds is required")
	}
	// 行为类型只能是点赞、提醒
	if r.BehaviorType != consts.BehaviorTypeTaskPraise && r.BehaviorType != consts.BehaviorTypeTaskAttention {
		return errors.New("behaviorType is invalid")
	}
	// 提醒时内容不能为空
	if r.BehaviorType == consts.BehaviorTypeTaskAttention && r.Content == "" {
		return errors.New("content is required")
	}

	return nil
}

// 作业报告-学生详情
type GetStudentDetailResponse struct {
	*Student                          // 学生信息
	PraiseCount              int64    `json:"praiseCount"`              // 已鼓励次数
	AttentionCount           int64    `json:"attentionCount"`           // 需要提醒次数
	StudentAccuracyRate      float64  `json:"studentAccuracyRate"`      // 该学生的正确率
	StudentCompletedProgress float64  `json:"studentCompletedProgress"` // 该学生的完成进度
	ClassAccuracyRate        float64  `json:"classAccuracyRate"`        // 班级平均正确率
	ClassCompletedProgress   float64  `json:"classCompletedProgress"`   // 班级平均进度
	AttentionText            string   `json:"attentionText"`            // 建议干预措施文案
	AttentionTextList        []string `json:"attentionTextList"`        // 建议干预措施列表
	PushDefaultText          string   `json:"pushDefaultText"`          // push题型默认文案
}

/*************************************************************
		                公共定义
*************************************************************/
// 学生信息
type Student struct {
	StudentID   int64  `json:"studentId"`   // 学生ID
	StudentName string `json:"studentName"` // 学生姓名
	Avatar      string `json:"avatar"`      // 学生头像
}

/*************************************************************
		    教师查看作业报告时具备的学科和班级列表
*************************************************************/
// 教师查看作业报告时具备的学科和班级列表
type GetSubjectClassListResponse struct {
	Subjects     []Subject        `json:"subjects"`     // 学科列表
	GradeClasses []itl.GradeClass `json:"gradeClasses"` // 年级-班级二级结构
}

type Subject struct {
	SubjectKey  int64  `json:"subjectKey"`  // 学科 1 ~ 9
	SubjectName string `json:"subjectName"` // 学科名称
}
