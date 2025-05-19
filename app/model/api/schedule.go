package api

import (
	"errors"
)

// TODO 等运营平台修改为驼峰形式
// 课程表数据结构
type Schedule struct {
	ScheduleID                 int64  `json:"schedule_id"`
	Grade                      int64  `json:"grade"`
	GradeName                  string `json:"grade_name"`
	ClassID                    int64  `json:"class_id"`
	ClassName                  string `json:"class_name"`
	ScheduleTplPeriodTimeSpan  string `json:"schedule_tpl_period_time_span"`
	ScheduleTplPeriodStartTime string `json:"schedule_tpl_period_start_time"` // 只有时间字段，无日期部分
	ScheduleTplPeriodEndTime   string `json:"schedule_tpl_period_end_time"`   // 只有时间字段，无日期部分
	ClassScheduleCourseID      int64  `json:"class_schedule_course_id"`
	ClassScheduleCourse        string `json:"class_schedule_course"`
	ClassScheduleTeacherID     int64  `json:"class_schedule_teacher_id"`
	TeacherName                string `json:"teacher_name"`
	ClassScheduleStudyType     int64  `json:"class_schedule_study_type"`
	ClassScheduleStudyTypeName string `json:"class_schedule_study_type_name"`
	IsTmp                      int64  `json:"is_tmp"`
	AtDay                      string `json:"at_day"`
	OriginScheduleID           int64  `json:"origin_schedule_id"`
	TmpScheduleID              int64  `json:"tmp_schedule_id"`
	IsInTimeRange              bool   `json:"is_in_time_range"` // 当前时间是否在课程时间范围内（可进入课程）
	IsInClass                  bool   `json:"is_in_class"`      // 当前时间是否在上课中
	ClassroomID                int64  `json:"classroom_id"`     // 教室ID，由schedule_id或tmp_schedule_id生成
}

// 日程请求参数
type FetchScheduleRequest struct {
	TeacherID    int64  `json:"teacherId"`
	SchoolID     int64  `json:"schoolId"`
	SchoolYearID int64  `json:"schoolYearId"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
}

// 解析完整的API响应
type ApiResponse struct {
	Status  int64            `json:"status"`
	Code    int64            `json:"code"`
	Message string           `json:"message"`
	Data    ScheduleResponse `json:"data"`
}

// 日程响应参数
type ScheduleResponse struct {
	Schedule map[string][]Schedule `json:"schedule"` // key: 1-7 表示周一到周日
	Dates    map[string]string     `json:"dates"`    // key: 1-7 表示周一到周日，value: 对应的日期 YYYY-MM-DD
}

// TeacherTeachingStatus 老师当前教学状态
type TeacherTeachingStatus struct {
	IsTeaching bool      `json:"isTeaching"`         // 是否在上课
	Schedule   *Schedule `json:"schedule,omitempty"` // 当前课程信息，如果不在上课则为nil
}

// CheckTeachingStatusRequest 检查教师上课状态的请求参数
type CheckTeachingStatusRequest struct {
	TeacherID    int64  `json:"teacherId"`    // 教师ID
	SchoolID     int64  `json:"schoolId"`     // 学校ID
	SchoolYearID int64  `json:"schoolYearId"` // 学年ID
	StartDate    string `json:"startDate"`    // 开始日期
	EndDate      string `json:"endDate"`      // 结束日期
	Now          int64  `json:"now"`          // 当前时间（UTC秒数）
	CheckTime    string `json:"checkTime"`    // 检查时间(HH:MM:SS)
	Weekday      int64  `json:"weekday"`      // 星期几(1-7)
}

// DirectScheduleRequest 直接获取课程表的请求
type DirectScheduleRequest struct {
	TeacherID int64  `json:"teacherId" binding:"required"` // 教师ID
	SchoolID  int64  `json:"schoolId" binding:"required"`  // 学校ID
	Date      string `json:"date" binding:"required"`      // 日期，格式：YYYY-MM-DD
}

// ScheduleDayResponse 单日课程表响应
type ScheduleDayResponse struct {
	Date     string     `json:"date"`               // 日期
	Schedule []Schedule `json:"schedule,omitempty"` // 课程表数据
}

// TeacherInfo 教师信息
type TeacherInfo struct {
	TeacherID int64 `json:"teacherId"` // 教师ID
	SchoolID  int64 `json:"schoolId"`  // 学校ID
}

// TeacherListResponse 教师列表响应
type TeacherListResponse struct {
	Count       int64         `json:"count"`       // 教师数量
	TeacherList []TeacherInfo `json:"teacherList"` // 教师列表
}

// SaveScheduleToRedisRequest 保存课程表到Redis的请求
type SaveScheduleToRedisRequest struct {
	TeacherID    int64  `json:"teacherId" binding:"required"`    // 教师ID
	SchoolID     int64  `json:"schoolId" binding:"required"`     // 学校ID
	SchoolYearID int64  `json:"schoolYearId" binding:"required"` // 学年ID
	StartDate    string `json:"startDate" binding:"required"`    // 开始日期
	EndDate      string `json:"endDate" binding:"required"`      // 结束日期
	Data         string `json:"data" binding:"required"`         // 课程表数据（JSON字符串）
}

// ClassroomStatusRequest 获取班级名称和上下课状态的请求
type ClassroomStatusRequest struct {
	ClassroomID int64 `form:"classroomId" binding:"required"` // 课堂ID
}

// Validate 验证请求参数
func (r *ClassroomStatusRequest) Validate() error {
	if r.ClassroomID <= 0 {
		return errors.New("课堂ID不能为0")
	}
	return nil
}

// ClassroomStatusResponse 获取班级名称和上下课状态的响应
type ClassroomStatusResponse struct {
	ClassroomID int64         `json:"classroomId"`        // 课堂ID
	GradeName   string        `json:"gradeName"`          // 年级名称
	ClassName   string        `json:"className"`          // 班级名称
	IsInClass   bool          `json:"isInClass"`          // 是否在上课
	Subject     string        `json:"subject,omitempty"`  // 当前课程学科内容
	Subjects    []SubjectInfo `json:"subjects,omitempty"` // 教师所教学科列表
}

// SubjectInfo 学科信息
type SubjectInfo struct {
	SubjectKey  int64  `json:"subjectKey"`  // 学科 1 ~ 9
	SubjectName string `json:"subjectName"` // 学科名称
}
