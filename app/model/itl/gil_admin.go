package itl

type CommonResponse struct {
	Status       int64  `json:"status"`       // 状态码
	Code         int64  `json:"code"`         // 业务码
	Message      string `json:"message"`      // 响应消息
	ResponseTime int64  `json:"responseTime"` // 响应时间
}

// TeacherDetailResponse API响应结构
type TeacherDetailResponse struct {
	CommonResponse
	Data TeacherDetailData `json:"data"` // 响应数据
}

// TeacherDetailData API响应数据结构
type TeacherDetailData struct {
	UserID                  int64            `json:"userID"`                  // 教师ID
	UserName                string           `json:"userName"`                // 教师姓名
	UserPhone               string           `json:"userPhone"`               // 教师手机号
	CurrentSchoolID         int64            `json:"currentSchoolID"`         // 当前学校ID
	TeacherEmploymentStatus int64            `json:"teacherEmploymentStatus"` // 当前学校的教师在职状态
	UserIsTest              int64            `json:"userIsTest"`              // 当前学校的教师是否测试用户
	TeacherJobInfos         []TeacherJobInfo `json:"teacherJobInfos"`         // 当前学校的教师任职信息、学科、班级
	SchoolInfos             []SchoolInfo     `json:"schoolInfos"`             // 教师所有学校的基本信息
}

// TeacherJobInfo 教师职务信息
type TeacherJobInfo struct {
	JobType    JobType    `json:"jobType"`    // 职务信息
	JobInfos   []JobInfo  `json:"jobInfos"`   // 职务下的年级和班级
	JobSubject JobSubject `json:"jobSubject"` // 工作学科
}

// JobType 职务信息
type JobType struct {
	JobType int64  `json:"jobType"` // 职务名称枚举值
	Name    string `json:"name"`    // 职务名称：1 校长，2 年级主任，3 学科组长，4 学科教师，5 班主任
}

// JobInfo 职务下的年级和班级
type JobInfo struct {
	Grade     int64   `json:"jobGrade"` // 年级枚举值：1 ~ 14
	GradeName string  `json:"name"`     // 年级名称：一年级 ~ 六年级，初一 ~ 初四，高一 ~ 高三复读
	Classes   []Class `json:"jobClass"` // 班级信息
}

// Class 班级信息
type Class struct {
	ID   int64  `json:"jobClass"` // 班级ID
	Name string `json:"name"`     // 班级名称
}

// JobSubject 学科信息
type JobSubject struct {
	JobSubject int64  `json:"jobSubject"` // 学科枚举值：1 ~ 9
	Name       string `json:"name"`       // 学科名称：语文，数学，英语，物理，化学，生物，历史，地理，道德与法治
}

// SchoolInfo 学校信息
type SchoolInfo struct {
	SchoolID        int64  `json:"schoolID"`        // 学校ID
	SchoolName      string `json:"schoolName"`      // 学校名称
	SchoolNumber    string `json:"schoolNumber"`    // 学校编号，例如 SCH44
	SchoolRegionID  int64  `json:"schoolRegionID"`  // 学校地区ID，省市区
	SchoolAddress   string `json:"schoolAddress"`   // 学校地址
	SchoolEduLevel  int64  `json:"schoolEduLevel"`  // 学段：1 小学，2 初中，3 高中
	SchoolEduSystem int64  `json:"schoolEduSystem"` // 学制
	SchoolNature    int64  `json:"schoolNature"`    // 办学性质：1 私立，2 公立，3 其它
	SchoolFeature   int64  `json:"schoolFeature"`   // 办学特色：1 初中+高中，2 小学+初中，3 小学+初中+高中
	SchoolTag       string `json:"schoolTag"`       // 学校标签
	SchoolIsTest    int64  `json:"schoolIsTest"`    // 是否测试学校
	SchoolStatus    int64  `json:"schoolStatus"`    // 学校状态
	Remark          string `json:"remark"`          // 备注
}

// 获取学校学科教材响应
type GetSchoolMaterialResponse struct {
	CommonResponse
	Data []GradeMaterial `json:"data"` // 响应数据
}

// 学科教材
type GradeMaterial struct {
	Phase     int64   `json:"phase"`     // 学段
	Subject   int64   `json:"subject"`   // 学科
	Grade     int64   `json:"grade"`     // 年级
	ClassType int64   `json:"classType"` // 班级类型：不分文理班、文科班、理科班
	Materials []int64 `json:"materials"` // 教材版本
}

// 查询班级学生响应
type GetClassStudentResponse struct {
	CommonResponse
	Data []ClassInfo `json:"data"` // 响应数据
}

// ClassInfo 班级信息
type ClassInfo struct {
	ClassName string         `json:"className"` // 班级名称
	ClassID   int64          `json:"classID"`   // 班级ID
	Students  []*StudentInfo `json:"students"`  // 学生列表
}

// StudentInfo 学生信息
type StudentInfo struct {
	ID     int64  `json:"id"`     // 学生ID
	Name   string `json:"name"`   // 学生姓名
	Avatar string `json:"avatar"` // 学生头像
}

// GetGradeClassInfoResponse 查询年级班级信息响应
type GetGradeClassInfoResponse struct {
	Status  int64        `json:"status"`  // 状态码
	Code    int64        `json:"code"`    // 业务码
	Message string       `json:"message"` // 响应消息
	Data    []GradeClass `json:"data"`    // 响应数据
}

// GradeClass 年级班级信息
type GradeClass struct {
	GradeID   int64           `json:"gradeID"`   // 年级枚举值：1 ~ 14
	GradeName string          `json:"gradeName"` // 年级名称：一年级 ~ 六年级，初一 ~ 初四，高一 ~ 高三复读
	Class     []ClassInfoItem `json:"class"`     // 班级信息列表
}

// ClassInfoItem 班级信息项
type ClassInfoItem struct {
	ClassID   int64  `json:"classID"`   // 班级ID
	ClassName string `json:"className"` // 班级名称
	// IsTest    bool   `json:"test"`      // 是否测试班级 // 教师端暂未使用
}

// 通过学生ID查询学生信息响应
type GetStudentInfoByIDResponse struct {
	CommonResponse
	Data map[string]StudentInfoData `json:"data"`
}

type StudentInfoData struct {
	Student StudentDetail `json:"student"`
	Class   IDAndName     `json:"class"`
	Grade   IDAndName     `json:"grade"`
	School  IDAndName     `json:"school"`
}

type IDAndName struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type StudentDetail struct {
	IDAndName
	Avatar string `json:"photo"`
}
