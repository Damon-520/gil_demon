package consts

import (
	"fmt"
	"time"
)

// 缓存默认过期时间 3s
const (
	DefaultExpire = 3
)

const (
	// 会话缓存 key，communication_session:{sessionId} => sessionId
	CommunicationSessionKey = "communication_session:%s"
	// 会话缓存过期时间
	CommunicationSessionExpire = 7 * 24 * 3600 // 7天

	// 会话消息 hashmap，session_message:{sessionId} => {messageId, timestamp}
	SessionMessageKey = "session_message:%s"
	// 会话消息 id 列表过期时间
	SessionMessageExpire = 7 * 24 * 3600 // 7天

	// 用户最后已读消息，session_user_last_read_message:{sessionId}:{userId} => messageId
	UserLastReadMessageKey = "user_last_read_message:%s:%d"
	// 用户最后已读消息过期时间
	UserLastReadMessageExpire = 7 * 24 * 3600 // 7天
)

// 用户最后已读消息缓存键
func GetSessionUserLastReadMessageKey(sessionID string, userID int64) string {
	return fmt.Sprintf(UserLastReadMessageKey, sessionID, userID)
}

// 会话消息 id 缓存键，缓存有序集合数据
//
//	session_message:{sessionId} => {key:messageId, score:timestamp}
func GetSessionMessageKey(sessionID string) string {
	return fmt.Sprintf(SessionMessageKey, sessionID)
}

// 会话缓存键
func GetCommunicationSessionKey(sessionID string) string {
	return fmt.Sprintf(CommunicationSessionKey, sessionID)
}

// 课程表缓存相关常量
const (
	// TeacherScheduleKeyFormat 教师课程表缓存键格式：teacher_course:{schoolID}:{teacherID}
	TeacherScheduleKeyFormat = "teacher_course:%d:%d"
	// TeacherScheduleExpiration 教师课程表缓存过期时间，设置为24小时
	TeacherScheduleExpiration = 24 * 3600 // 24小时

	// TeacherListKey 存储所有需要获取课程表的教师ID和学校ID的集合
	TeacherListKey = "schedule:teacher:list:%d:%d"
	// TeacherListExpiration 教师列表缓存过期时间
	TeacherListExpiration = 30 * 24 * 3600 // 30天

	// TeacherInfoFormat 教师信息的格式化字符串，用于存储教师ID和学校ID
	TeacherInfoFormat = "teacher_info:%d:%d"

	// TeacherScheduleSearchPattern 教师课程表搜索模式，用于在Redis中搜索特定教师的所有课程表
	TeacherScheduleSearchPattern = "teacher_course:*:%d"
)

// 教师课程表缓存键
func GetTeacherScheduleKey(schoolID, teacherID int64) string {
	return fmt.Sprintf(TeacherScheduleKeyFormat, schoolID, teacherID)
}

// 教师列表缓存键
func GetTeacherListKey(schoolID, teacherID int64) string {
	return fmt.Sprintf(TeacherListKey, schoolID, teacherID)
}

// 教师信息缓存键
func GetTeacherInfoKey(schoolID, teacherID int64) string {
	return fmt.Sprintf(TeacherInfoFormat, schoolID, teacherID)
}

// 教师课程表搜索模式
func GetTeacherScheduleSearchPattern(teacherID int64) string {
	return fmt.Sprintf(TeacherScheduleSearchPattern, teacherID)
}

// 任务缓存相关常量
const (
	// TaskLastUsedBizTreeKeyFormat 最近使用过的业务树缓存键格式：task:lt:{t}:{s}
	TaskLastUsedBizTreeKeyFormat = "task:lt:%d:%d"
)

func TaskLastUsedBizTreeKey(teacherID int64, subject int64) string {
	return fmt.Sprintf(TaskLastUsedBizTreeKeyFormat, teacherID, subject)
}

// 学生行为相关缓存键
const (
	// ClassBehaviorHandledKey 班级学生行为已处理记录的缓存键：classroomId -> {studentId: handleTime}
	ClassBehaviorHandledKey = "beh:cls:%d:handled"

	// StudentReminderCountKey 学生提醒次数的缓存键：classroomId, studentId -> reminderCount
	StudentReminderCountKey = "beh:cls:%d:stu:%d:rem_cnt"

	// 已处理学生行为的过期时间：24小时
	ClassBehaviorHandledExpire = 24 * 3600 // 24小时

	// 学生提醒次数的过期时间：24小时
	StudentReminderCountExpire = 24 * 3600 // 24小时

	// PraiseRecordKey 表扬记录缓存键：classroomId:studentId -> timestamp
	// 用于控制每节课只能表扬一次
	PraiseRecordKey = "beh:cls:%d:stu:%d:prs"

	// PraiseTypeRecordKey 按类型表扬记录缓存键：classroomId:studentId:type -> timestamp
	// 用于控制每节课每种类型只能表扬一次
	PraiseTypeRecordKey = "beh:cls:%d:stu:%d:prs:%s"

	// PraiseRecordExpire 表扬记录过期时间：24小时
	// 表扬记录24小时后过期，确保每节课都可以重新表扬
	PraiseRecordExpire = 24 * 3600 // 24小时

	// AttentionTimeWindowKey 关注时间窗口缓存键：classroomId:studentId -> timestamp
	// 用于控制关注后1分钟内不再重复提醒
	AttentionTimeWindowKey = "beh:cls:%d:stu:%d:attn"

	// AttentionTimeWindowExpire 关注时间窗口过期时间：1分钟
	// 1分钟后过期，学生若未改善行为，可以再次提醒
	AttentionTimeWindowExpire = 1 * time.Minute

	// ClassBehaviorPraisedKey 课堂中已表扬的学生集合 key
	ClassBehaviorPraisedKey = "beh:cls:%d:prs_set"
	// ClassBehaviorPraisedExpire 课堂中已表扬学生记录过期时间
	ClassBehaviorPraisedExpire = 24 * time.Hour

	// EvaluateRecordKey 评价记录缓存键：classroomId:studentId:evaluateType -> timestamp
	// 用于检查是否已经评价过该学生
	EvaluateRecordKey = "beh:cls:%d:stu:%d:eval:%s"
	// EvaluateRecordExpire 评价记录过期时间：24小时
	EvaluateRecordExpire = 24 * 3600 // 24小时
)

// 学校班级学生数据，每个班级一个 key
const (
	// 学校班级学生数据缓存键格式：class_student:{schoolID}:{classID}:{date}，缓存当天有效，避免学生信息变化导致的缓存不一致
	ClassStudentKeyFormat = "class_student:%d:%d:%s"
	// 班级信息缓存键格式：class_info:{schoolID}:{classID}:{date}，缓存当天有效，避免学生信息变化导致的缓存不一致
	ClassInfoKeyFormat = "class_info:%d:%d:%s"
	// 过期时间 24 小时
	ClassStudentExpire = 24 * 3600 // 24小时
)

// 学校班级学生数据缓存键列表
func ClassStudentKey(schoolID int64, classID int64) string {
	return fmt.Sprintf(ClassStudentKeyFormat, schoolID, classID, time.Now().Format(TimeFormatDate))
}

// 班级信息缓存键列表
func ClassInfoKey(schoolID int64, classID int64) string {
	return fmt.Sprintf(ClassInfoKeyFormat, schoolID, classID, time.Now().Format(TimeFormatDate))
}

func GetClassBehaviorHandledKey(classroomID int64) string {
	return fmt.Sprintf(ClassBehaviorHandledKey, classroomID)
}

func GetStudentReminderCountKey(classroomID int64, studentID int64) string {
	return fmt.Sprintf(StudentReminderCountKey, classroomID, studentID)
}

// GetPraiseRecordKey 获取表扬记录缓存键
func GetPraiseRecordKey(classroomID, studentID int64) string {
	return fmt.Sprintf(PraiseRecordKey, classroomID, studentID)
}

// GetPraiseTypeRecordKey 获取按类型表扬记录缓存键
func GetPraiseTypeRecordKey(classroomID, studentID int64, behaviorType string) string {
	return fmt.Sprintf(PraiseTypeRecordKey, classroomID, studentID, behaviorType)
}

func GetAttentionTimeWindowKey(classroomID int64, studentID int64) string {
	return fmt.Sprintf(AttentionTimeWindowKey, classroomID, studentID)
}

func GetClassBehaviorPraisedKey(classroomID int64) string {
	return fmt.Sprintf(ClassBehaviorPraisedKey, classroomID)
}

// GetEvaluateRecordKey 获取评价记录缓存键
func GetEvaluateRecordKey(classroomID, studentID int64, evaluateType string) string {
	return fmt.Sprintf(EvaluateRecordKey, classroomID, studentID, evaluateType)
}
