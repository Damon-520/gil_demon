package consts

import "slices"

// 文件访问权限枚举
const (
	FILE_SCOPE_PUBLIC  int = 1 // 公开
	FILE_SCOPE_SCHOOL  int = 2 // 校本
	FILE_SCOPE_PRIVATE int = 3 // 私有
)

// 文件访问权限名称常量
const (
	FILE_SCOPE_PUBLIC_NAME  = "公开"
	FILE_SCOPE_SCHOOL_NAME  = "校本"
	FILE_SCOPE_PRIVATE_NAME = "私有"
)

// FileScopeNameMap 文件访问权限名称映射
var FileScopeNameMap = map[int]string{
	FILE_SCOPE_PUBLIC:  FILE_SCOPE_PUBLIC_NAME,
	FILE_SCOPE_SCHOOL:  FILE_SCOPE_SCHOOL_NAME,
	FILE_SCOPE_PRIVATE: FILE_SCOPE_PRIVATE_NAME,
}

// 学段枚举
const (
	PHASE_PRIMARY int64 = 1 // 小学
	PHASE_JUNIOR  int64 = 2 // 初中
	PHASE_HIGH    int64 = 3 // 高中
)

// 学段名称常量
const (
	PHASE_PRIMARY_NAME = "小学"
	PHASE_JUNIOR_NAME  = "初中"
	PHASE_HIGH_NAME    = "高中"
)

// PhaseExists 检查学段是否存在
func PhaseExists(phase int64) bool {
	return slices.Contains([]int64{
		PHASE_PRIMARY,
		PHASE_JUNIOR,
		PHASE_HIGH,
	}, phase)
}

// 职务枚举
const (
	JOB_TYPE_PRINCIPAL       int64 = 1 // 校长
	JOB_TYPE_GRADE_HEAD      int64 = 2 // 年级主任
	JOB_TYPE_SUBJECT_HEAD    int64 = 3 // 学科组长
	JOB_TYPE_SUBJECT_TEACHER int64 = 4 // 学科教师
	JOB_TYPE_CLASS_TEACHER   int64 = 5 // 班主任
)

// 职务名称常量
const (
	JOB_TYPE_PRINCIPAL_NAME       = "校长"
	JOB_TYPE_GRADE_HEAD_NAME      = "年级主任"
	JOB_TYPE_SUBJECT_HEAD_NAME    = "学科组长"
	JOB_TYPE_SUBJECT_TEACHER_NAME = "学科教师"
	JOB_TYPE_CLASS_TEACHER_NAME   = "班主任"
)

// JobTypeNameMap 职务名称映射
var JobTypeNameMap = map[int64]string{
	JOB_TYPE_PRINCIPAL:       JOB_TYPE_PRINCIPAL_NAME,
	JOB_TYPE_GRADE_HEAD:      JOB_TYPE_GRADE_HEAD_NAME,
	JOB_TYPE_SUBJECT_HEAD:    JOB_TYPE_SUBJECT_HEAD_NAME,
	JOB_TYPE_SUBJECT_TEACHER: JOB_TYPE_SUBJECT_TEACHER_NAME,
	JOB_TYPE_CLASS_TEACHER:   JOB_TYPE_CLASS_TEACHER_NAME,
}

// 学科枚举
const (
	SUBJECT_CHINESE   int64 = 1 // 语文
	SUBJECT_MATH      int64 = 2 // 数学
	SUBJECT_ENGLISH   int64 = 3 // 英语
	SUBJECT_PHYSICS   int64 = 4 // 物理
	SUBJECT_CHEMISTRY int64 = 5 // 化学
	SUBJECT_BIOLOGY   int64 = 6 // 生物
	SUBJECT_HISTORY   int64 = 7 // 历史
	SUBJECT_GEOGRAPHY int64 = 8 // 地理
	SUBJECT_MORAL     int64 = 9 // 道德与法治
)

// 学科名称常量
const (
	SUBJECT_CHINESE_NAME   = "语文"
	SUBJECT_MATH_NAME      = "数学"
	SUBJECT_ENGLISH_NAME   = "英语"
	SUBJECT_PHYSICS_NAME   = "物理"
	SUBJECT_CHEMISTRY_NAME = "化学"
	SUBJECT_BIOLOGY_NAME   = "生物"
	SUBJECT_HISTORY_NAME   = "历史"
	SUBJECT_GEOGRAPHY_NAME = "地理"
	SUBJECT_MORAL_NAME     = "道德与法治"
)

// SubjectExists 检查学科是否存在
func SubjectExists(subject int64) bool {
	return slices.Contains([]int64{
		SUBJECT_CHINESE,
		SUBJECT_MATH,
		SUBJECT_ENGLISH,
		SUBJECT_PHYSICS,
		SUBJECT_CHEMISTRY,
		SUBJECT_BIOLOGY,
		SUBJECT_HISTORY,
		SUBJECT_GEOGRAPHY,
		SUBJECT_MORAL,
	}, subject)
}

// 学科及任务类型
type SubjectTaskTypeInfo struct {
	Key       int64   `json:"key"`
	Value     string  `json:"value"`
	TaskTypes []int64 `json:"taskTypes"`
}

// 学段信息结构
type PhaseInfo struct {
	Key      int64                 `json:"key"`
	Value    string                `json:"value"`
	Subjects []SubjectTaskTypeInfo `json:"subjects"`
}

// PhaseNameMap 学段名称映射
var PhaseNameMap = map[int64]string{
	PHASE_PRIMARY: PHASE_PRIMARY_NAME,
	PHASE_JUNIOR:  PHASE_JUNIOR_NAME,
	PHASE_HIGH:    PHASE_HIGH_NAME,
}

// SubjectNameMap 学科名称映射
var SubjectNameMap = map[int64]string{
	SUBJECT_CHINESE:   SUBJECT_CHINESE_NAME,
	SUBJECT_MATH:      SUBJECT_MATH_NAME,
	SUBJECT_ENGLISH:   SUBJECT_ENGLISH_NAME,
	SUBJECT_PHYSICS:   SUBJECT_PHYSICS_NAME,
	SUBJECT_CHEMISTRY: SUBJECT_CHEMISTRY_NAME,
	SUBJECT_BIOLOGY:   SUBJECT_BIOLOGY_NAME,
	SUBJECT_HISTORY:   SUBJECT_HISTORY_NAME,
	SUBJECT_GEOGRAPHY: SUBJECT_GEOGRAPHY_NAME,
	SUBJECT_MORAL:     SUBJECT_MORAL_NAME,
}

// Phase2SubjectMap 学段学科映射
var Phase2SubjectMap = map[int64][]int64{
	PHASE_PRIMARY: {
		SUBJECT_CHINESE,
		SUBJECT_MATH,
		SUBJECT_ENGLISH,
	},
	PHASE_JUNIOR: {
		SUBJECT_CHINESE,
		SUBJECT_MATH,
		SUBJECT_ENGLISH,
		SUBJECT_PHYSICS,
		SUBJECT_CHEMISTRY,
		SUBJECT_BIOLOGY,
		SUBJECT_HISTORY,
		SUBJECT_GEOGRAPHY,
		SUBJECT_MORAL,
	},
	PHASE_HIGH: {
		SUBJECT_CHINESE,
		SUBJECT_MATH,
		SUBJECT_ENGLISH,
		SUBJECT_PHYSICS,
		SUBJECT_CHEMISTRY,
		SUBJECT_BIOLOGY,
		SUBJECT_HISTORY,
		SUBJECT_GEOGRAPHY,
		SUBJECT_MORAL,
	},
}

// AllPhaseSubjectTaskType 所有学段学科任务类型
var AllPhaseSubjectTaskType = []PhaseInfo{
	{
		Key:   PHASE_PRIMARY,
		Value: PHASE_PRIMARY_NAME,
		Subjects: []SubjectTaskTypeInfo{
			{
				Key:       SUBJECT_CHINESE,
				Value:     SUBJECT_CHINESE_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_MATH,
				Value:     SUBJECT_MATH_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_ENGLISH,
				Value:     SUBJECT_ENGLISH_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
		},
	},
	{
		Key:   PHASE_JUNIOR,
		Value: PHASE_JUNIOR_NAME,
		Subjects: []SubjectTaskTypeInfo{
			{
				Key:       SUBJECT_CHINESE,
				Value:     SUBJECT_CHINESE_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_MATH,
				Value:     SUBJECT_MATH_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_ENGLISH,
				Value:     SUBJECT_ENGLISH_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_PHYSICS,
				Value:     SUBJECT_PHYSICS_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_CHEMISTRY,
				Value:     SUBJECT_CHEMISTRY_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_BIOLOGY,
				Value:     SUBJECT_BIOLOGY_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_HISTORY,
				Value:     SUBJECT_HISTORY_NAME,
				TaskTypes: []int64{TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_GEOGRAPHY,
				Value:     SUBJECT_GEOGRAPHY_NAME,
				TaskTypes: []int64{TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_MORAL,
				Value:     SUBJECT_MORAL_NAME,
				TaskTypes: []int64{TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
		},
	},
	{
		Key:   PHASE_HIGH,
		Value: PHASE_HIGH_NAME,
		Subjects: []SubjectTaskTypeInfo{
			{
				Key:       SUBJECT_CHINESE,
				Value:     SUBJECT_CHINESE_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_MATH,
				Value:     SUBJECT_MATH_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_ENGLISH,
				Value:     SUBJECT_ENGLISH_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_PHYSICS,
				Value:     SUBJECT_PHYSICS_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_CHEMISTRY,
				Value:     SUBJECT_CHEMISTRY_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_BIOLOGY,
				Value:     SUBJECT_BIOLOGY_NAME,
				TaskTypes: []int64{TASK_TYPE_COURSE, TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_HISTORY,
				Value:     SUBJECT_HISTORY_NAME,
				TaskTypes: []int64{TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_GEOGRAPHY,
				Value:     SUBJECT_GEOGRAPHY_NAME,
				TaskTypes: []int64{TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
			{
				Key:       SUBJECT_MORAL,
				Value:     SUBJECT_MORAL_NAME,
				TaskTypes: []int64{TASK_TYPE_HOMEWORK, TASK_TYPE_TEST, TASK_TYPE_RESOURCE},
			},
		},
	},
}

// ------------------------------------------------------------
// ------------------------------------------------------------
// 任务类型（10-99）
const (
	// 课程任务（10-19）
	TASK_TYPE_COURSE int64 = 10 // 课程

	// 作业任务（20-29）
	TASK_TYPE_HOMEWORK int64 = 20 // 作业
	// TASK_TYPE_HOMEWORK_STANDARD int64 = 21 // 标准作业
	// TASK_TYPE_HOMEWORK_LAYERED  int64 = 22 // 分层作业
	// TASK_TYPE_HOMEWORK_ANSWER   int64 = 23 // 答题卡作业
	// TASK_TYPE_HOMEWORK_WRONG    int64 = 24 // 错题作业

	// 测验任务（30-39）
	TASK_TYPE_TEST int64 = 30 // 测验
	// TASK_TYPE_TEST_STANDARD int64 = 31 // 标准测验
	// TASK_TYPE_TEST_ANSWER   int64 = 32 // 答题卡测验

	// 资源任务（40-49）
	TASK_TYPE_RESOURCE int64 = 40 // 资源
	// TASK_TYPE_RESOURCE_PUBLIC int64 = 41 // 公共资源
	// TASK_TYPE_RESOURCE_SCHOOL   int64 = 42 // 校本资源
	// TASK_TYPE_RESOURCE_PERSONAL int64 = 43 // 我的资源
)

// TaskTypeExists 检查任务类型是否存在
func TaskTypeExists(taskType int64) bool {
	return slices.Contains([]int64{
		TASK_TYPE_COURSE,
		TASK_TYPE_HOMEWORK,
		TASK_TYPE_TEST,
		TASK_TYPE_RESOURCE,
	}, taskType)
}

// 任务类型名称常量
const (
	TASK_TYPE_COURSE_NAME   = "课程"
	TASK_TYPE_HOMEWORK_NAME = "作业"
	// TASK_TYPE_HOMEWORK_STANDARD_NAME = "标准作业"
	// TASK_TYPE_HOMEWORK_LAYERED_NAME  = "分层作业"
	// TASK_TYPE_HOMEWORK_ANSWER_NAME   = "答题卡作业"
	// TASK_TYPE_HOMEWORK_WRONG_NAME    = "错题作业"
	TASK_TYPE_TEST_NAME = "测验"
	// TASK_TYPE_TEST_STANDARD_NAME     = "标准测验"
	// TASK_TYPE_TEST_ANSWER_NAME       = "答题卡测验"
	TASK_TYPE_RESOURCE_NAME = "资源"
	// TASK_TYPE_RESOURCE_PUBLIC_NAME = "公共资源"
	// TASK_TYPE_RESOURCE_SCHOOL_NAME   = "校本资源"
	// TASK_TYPE_RESOURCE_PERSONAL_NAME = "我的资源"
)

// TaskTypeNameMap 任务类型名称映射
var TaskTypeNameMap = map[int64]string{
	TASK_TYPE_COURSE:   TASK_TYPE_COURSE_NAME,
	TASK_TYPE_HOMEWORK: TASK_TYPE_HOMEWORK_NAME,
	TASK_TYPE_TEST:     TASK_TYPE_TEST_NAME,
	TASK_TYPE_RESOURCE: TASK_TYPE_RESOURCE_NAME,
}

// GetTaskTypeName 获取任务类型名称
func GetTaskTypeName(taskType int64) string {
	if name, ok := TaskTypeNameMap[taskType]; ok {
		return name
	}
	return ""
}

// 任务派发的学生群组相关常量
const (
	TASK_GROUP_TYPE_TEMP    = 1 // 临时群组，对应 group_id 为 0
	TASK_GROUP_TYPE_CLASS   = 2 // 班级，对应 group_id 为班级ID
	TASK_GROUP_TYPE_STUDENT = 3 // 学生群组，对应 group_id 为 group 表的 id
	TASK_GROUP_ID_TEMP      = 0 // 临时群组时的 group_id
)

// 素材资源类型
const (
	// 素材资源类型（100-199）内容平台资源>100
	RESOURCE_TYPE_OTHER     int64 = 100 // 其它资源
	RESOURCE_TYPE_AI_COURSE int64 = 101 // AI课，内容平台
	RESOURCE_TYPE_PRACTICE  int64 = 102 // 巩固练习，内容平台
	RESOURCE_TYPE_QUESTION  int64 = 103 // 试题，内容平台
	RESOURCE_TYPE_PAPER     int64 = 104 // 试卷，内容平台
)

// 素材资源类型名称常量
const (
	RESOURCE_TYPE_OTHER_NAME     = "其它资源"
	RESOURCE_TYPE_AI_COURSE_NAME = "AI课"
	RESOURCE_TYPE_PRACTICE_NAME  = "巩固练习"
	RESOURCE_TYPE_QUESTION_NAME  = "试题"
	RESOURCE_TYPE_PAPER_NAME     = "试卷"
)

// ResourceTypeNameMap 素材资源类型名称映射
var ResourceTypeNameMap = map[int64]string{
	RESOURCE_TYPE_OTHER:     RESOURCE_TYPE_OTHER_NAME,
	RESOURCE_TYPE_AI_COURSE: RESOURCE_TYPE_AI_COURSE_NAME,
	RESOURCE_TYPE_PRACTICE:  RESOURCE_TYPE_PRACTICE_NAME,
	RESOURCE_TYPE_QUESTION:  RESOURCE_TYPE_QUESTION_NAME,
	RESOURCE_TYPE_PAPER:     RESOURCE_TYPE_PAPER_NAME,
}

// GetResourceTypeName 获取素材资源类型名称
func GetResourceTypeName(resourceType int64) string {
	if name, ok := ResourceTypeNameMap[resourceType]; ok {
		return name
	}
	return ""
}

// 题目类型
type QuestionType int64

const (
	QUESTION_TYPE_ALL             QuestionType = 0 // 全部题型
	QUESTION_TYPE_SINGLE_CHOICE   QuestionType = 1 // 单选题
	QUESTION_TYPE_MULTIPLE_CHOICE QuestionType = 2 // 多选题
	QUESTION_TYPE_FILL_BLANK      QuestionType = 3 // 填空题
	// QUESTION_TYPE_JUDGE           QuestionType = 4 // 判断题
	// QUESTION_TYPE_ANSWER          QuestionType = 5 // 解答题
)

// QuestionTypeNameMap 题目类型名称映射
var QuestionTypeNameMap = map[QuestionType]string{
	QUESTION_TYPE_SINGLE_CHOICE:   "单选题",
	QUESTION_TYPE_MULTIPLE_CHOICE: "多选题",
	QUESTION_TYPE_FILL_BLANK:      "填空题",
	// QUESTION_TYPE_JUDGE:           "判断题",
	// QUESTION_TYPE_ANSWER:          "解答题",
}

// QuestionTypeExists 检查题目类型是否存在
func QuestionTypeExists(questionType QuestionType) bool {
	return slices.Contains([]QuestionType{
		QUESTION_TYPE_SINGLE_CHOICE,
		QUESTION_TYPE_MULTIPLE_CHOICE,
		QUESTION_TYPE_FILL_BLANK,
		// QUESTION_TYPE_JUDGE,
		// QUESTION_TYPE_ANSWER,
	}, questionType)
}

// 获取题目类型名称
func GetQuestionTypeName(questionType int64) string {
	if name, ok := QuestionTypeNameMap[QuestionType(questionType)]; ok {
		return name
	}
	return ""
}

type QuestionDifficult int64

const (
	QUESTION_DIFFICULT_SIMPLE      QuestionDifficult = 1 // 简单
	QUESTION_DIFFICULT_EASY        QuestionDifficult = 2 // 较易
	QUESTION_DIFFICULT_MEDIUM      QuestionDifficult = 3 // 中等
	QUESTION_DIFFICULT_CHALLENGING QuestionDifficult = 4 // 较难
	QUESTION_DIFFICULT_HARD        QuestionDifficult = 5 // 困难
)

// QuestionDifficultNameMap 问题难度名称映射
var QuestionDifficultNameMap = map[QuestionDifficult]string{
	QUESTION_DIFFICULT_SIMPLE:      "简单",
	QUESTION_DIFFICULT_EASY:        "较易",
	QUESTION_DIFFICULT_MEDIUM:      "中等",
	QUESTION_DIFFICULT_CHALLENGING: "较难",
	QUESTION_DIFFICULT_HARD:        "困难",
}

// 通过难度值获取难度文字
func GetQuestionDifficultName(questionDifficult int64) string {
	if name, ok := QuestionDifficultNameMap[QuestionDifficult(questionDifficult)]; ok {
		return name
	}
	return QuestionDifficultNameMap[QUESTION_DIFFICULT_MEDIUM]
}

// 学生标签类型
const (
	STUDENT_TAG_TYPE_POSITIVE int = 1  // 正向
	STUDENT_TAG_TYPE_NEUTRAL  int = 0  // 中性
	STUDENT_TAG_TYPE_NEGATIVE int = -1 // 负向
)
