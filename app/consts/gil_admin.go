package consts

// GilAdminAPI 运营平台 API
type GilAdminAPI struct {
	Method string
	Path   string
}

// 教师服务相关常量
const (
	// TeacherDefaultStatus 教师默认状态
	TeacherDefaultStatus = 1

	// TeacherDefaultType 教师默认类型
	TeacherDefaultType = 1

	// TeacherDefaultIsTest 教师默认测试标记
	TeacherDefaultIsTest = 0

	// TeacherDefaultEmploymentStatus 教师默认在职状态
	TeacherDefaultEmploymentStatus = 1
)

// UcenterCustomHeader 运营平台自定义请求头
const (
	UcenterCustomHeaderUserTypeID      = "userTypeId"
	UcenterCustomHeaderUserTypeIDValue = "2" // 教师端固定值

	UcenterCustomHeaderOrganizationID = "organizationId" // 对应的值为学校ID
)

var (
	// 获取教师详细信息接口
	GetTeacherDetailAPI = GilAdminAPI{
		Method: "GET",
		Path:   "/api/v1/teacher/detail_by_token",
	}

	// 获取学校的学科教材接口
	GetSchoolMaterialAPI = GilAdminAPI{
		Method: "GET",
		Path:   "/internal/api/v1/school/material",
	}

	// 查询班级学生接口
	GetClassStudentAPI = GilAdminAPI{
		Method: "GET",
		Path:   "/internal/api/v1/class/getStudent",
	}

	// 查询课程表接口
	GetInternalClassRoomAPI = GilAdminAPI{
		Method: "GET",
		Path:   "/internal/api/v1/teacher/weekSchedule",
	}

	// 查询年级班级信息接口
	GetGradeClassInfoAPI = GilAdminAPI{
		Method: "GET",
		Path:   "/internal/api/v1/class/classInfo",
	}

	// 通过学生ID查询学生信息接口
	GetStudentInfoByIDAPI = GilAdminAPI{
		Method: "GET",
		Path:   "/internal/api/v1/class/getStudentWihOrg",
	}
)
