package consts

// 课堂ID相关常量
const (
	// TempClassroomIDBase 临时课堂ID基础值
	// 使用 1 << 62 (即 2^62) 作为基础值
	// 值为: 4,611,686,018,427,387,904
	// 这个值在 int64 正数范围内，并为临时课堂预留了高位的 ID 空间
	TempClassroomIDBase int64 = 1 << 62

	// MaxNormalClassroomID 普通课堂最大ID
	// 小于此值的为普通课堂ID
	MaxNormalClassroomID = TempClassroomIDBase - 1

	DefaultClassroomName = "默认课堂"
	// DefaultTaskName 默认任务名称
	DefaultTaskName = "默认任务"
	// TempClassroomIDPrefix 临时课堂ID前缀
	TempClassroomIDPrefix = "temp_"

	// ClassroomNotFound 课堂不存在时的年级名称
	ClassroomNotFoundGradeName = "不存在的课堂"
	// ClassroomNotFoundClassName 课堂不存在时的班级名称
	ClassroomNotFoundClassName = "不存在的班级"
	// DefaultGradeName 默认年级名称
	DefaultGradeName = "未知年级"
	// DefaultClassName 默认班级名称
	DefaultClassName = "未知班级"
)

// IsTempClassroom 判断是否为临时课堂ID
func IsTempClassroom(classroomID int64) bool {
	return classroomID >= TempClassroomIDBase // >= 也可以，因为 Base 本身不作为普通ID
}

// GetTmpScheduleID 从临时课堂ID中获取临时课表ID
func GetTmpScheduleID(classroomID int64) int64 {
	if !IsTempClassroom(classroomID) {
		return 0 // 或者返回错误，取决于业务逻辑
	}
	return classroomID - TempClassroomIDBase
}

// GenerateTempClassroomID 生成临时课堂ID
func GenerateTempClassroomID(tmpScheduleID int64) int64 {
	// 注意：需要确保 tmpScheduleID 不会过大导致溢出 int64 的正数范围
	// maxTmpScheduleID = MaxInt64 - TempClassroomIDBase
	// 实际应用中可能需要添加检查
	return TempClassroomIDBase + tmpScheduleID
}
