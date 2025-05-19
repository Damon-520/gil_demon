package consts

// 定义*gin.Context键值
const (
	FormHeaderCtxUserIdKey = "hUserId" // 从header获取的userId

	// CtxTeacherDetailKey context中存储教师信息的key
	CtxTeacherDetailKey = "teacher_detail"

	// CtxTeacherPhaseKey context中存储教师学段的key
	CtxTeacherPhaseKey = "teacher_phase"

	// CtxTeacherSubjectsKey context中存储教师学科的key
	CtxTeacherSubjectsKey = "teacher_subjects"

	// CtxTeacherIDKey context中存储教师ID的key
	CtxTeacherIDKey = "teacher_id"

	// CtxSchoolIDKey context中存储学校ID的key
	CtxSchoolIDKey = "school_id"
)
