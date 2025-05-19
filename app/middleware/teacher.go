package middleware

import (
	"errors"
	"gil_teacher/app/consts"
	"gil_teacher/app/controller/http_server/response"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/model/itl"
	"gil_teacher/app/service/gil_internal/admin_service"
	"slices"

	"github.com/gin-gonic/gin"
)

// TeacherMiddleware 教师中间件，包含日志记录器
type TeacherMiddleware struct {
	log            *logger.ContextLogger
	ucenterService *admin_service.UcenterClient
}

// NewTeacherMiddleware 创建教师中间件
func NewTeacherMiddleware(log *logger.ContextLogger, ucenterService *admin_service.UcenterClient) *TeacherMiddleware {
	return &TeacherMiddleware{
		log:            log,
		ucenterService: ucenterService,
	}
}

// WithTeacherContext 中间件用于获取教师信息并存储到context中
func (tm *TeacherMiddleware) WithTeacherContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取Authorization 和 schoolID
		token := c.GetHeader("Authorization")
		schoolID := c.GetHeader(consts.UcenterCustomHeaderOrganizationID)
		if token == "" || schoolID == "" {
			tm.log.Warn(c, "未提供Authorization token 或 schoolID")
			response.Unauthorized(c)
			c.Abort()
			return
		}

		// 调用服务获取教师信息
		result, err := tm.ucenterService.GetTeacherDetail(c, token, schoolID)
		if err != nil {
			// 区分未授权错误和其它错误
			if err == &response.ERR_UNAUTHORIZED {
				response.Unauthorized(c)
			} else {
				response.Err(c, response.ERR_GIL_ADMIN)
			}
			c.Abort()
			return
		}

		tm.log.Debug(c, "成功获取教师信息: %d", result.UserID)
		tm.setTeacherDetailToContext(c, result)

		c.Next()
	}
}

/*
老师的职务信息 teacherJobInfos 示例如下：

	{
	  "teacherJobInfos": [
	    {
	      "jobType": { "jobType": 1, "name": "校长" },  	// 职务类型
	      "jobInfos": null,  							   // 年级、班级信息
	      "jobSubject": { "jobSubject": 0, "name": "" }    // 学科信息
	    },
	    {
	      "jobType": { "jobType": 2, "name": "年级主任" },
	      "jobInfos": [
	        { "jobGrade": 11, "name": "高一", "jobClass": null },
	        { "jobGrade": 12, "name": "高二", "jobClass": null }
	      ],
	      "jobSubject": { "jobSubject": 0, "name": "" }
	    },
	    {
	      "jobType": { "jobType": 3, "name": "学科组长" },
	      "jobInfos": [
	        { "jobGrade": 11, "name": "高一", "jobClass": null },
	        { "jobGrade": 12, "name": "高二", "jobClass": null },
	        { "jobGrade": 13, "name": "高三", "jobClass": null }
	      ],
	      "jobSubject": { "jobSubject": 2, "name": "数学" }
	    },
	    {
	      "jobType": { "jobType": 4, "name": "学科教师" },
	      "jobInfos": [
	        {
	          "jobGrade": 11,
	          "name": "高一",
	          "jobClass": [
	            { "jobClass": 33, "name": "高一 1 班" },
	            { "jobClass": 34, "name": "高一 2 班" },
	            { "jobClass": 35, "name": "高一 3 班" }
	          ]
	        }
	      ],
	      "jobSubject": { "jobSubject": 2, "name": "数学" }
	    },
	    {
	      "jobType": { "jobType": 5, "name": "班主任" },
	      "jobInfos": [
	        {
	          "jobGrade": 11,
	          "name": "高一",
	          "jobClass": [
	            { "jobClass": 34, "name": "高一 2 班" },
	            { "jobClass": 33, "name": "高一 1 班" },
	            { "jobClass": 35, "name": "高一 3 班" }
	          ]
	        }
	      ],
	      "jobSubject": { "jobSubject": 0, "name": "" }
	    }
	  ]
	}
*/
func (tm *TeacherMiddleware) setTeacherDetailToContext(c *gin.Context, detail *itl.TeacherDetailData) {
	// 将教师全部信息存入context
	c.Set(consts.CtxTeacherDetailKey, detail)

	// 提取教师的学段、学科存入context
	phase := int64(0)
	subjects := []int64{}
	subjectMap := make(map[int64]struct{})
	for _, school := range detail.SchoolInfos {
		if school.SchoolID == detail.CurrentSchoolID {
			phase = school.SchoolEduLevel
		}
	}

	for _, jobInfo := range detail.TeacherJobInfos {
		// 校长之类的职务学科信息返回了 0 ，这里直接去掉
		if jobInfo.JobSubject.JobSubject != 0 {
			if _, ok := subjectMap[jobInfo.JobSubject.JobSubject]; !ok {
				subjectMap[jobInfo.JobSubject.JobSubject] = struct{}{}
				subjects = append(subjects, jobInfo.JobSubject.JobSubject)
			}
		}
	}

	c.Set(consts.CtxTeacherPhaseKey, phase)
	c.Set(consts.CtxTeacherSubjectsKey, subjects)

	// 提取教师ID和学校ID存入context
	c.Set(consts.CtxTeacherIDKey, detail.UserID)
	c.Set(consts.CtxSchoolIDKey, detail.CurrentSchoolID)
	tm.log.Debug(c, "学段: %d, 学科: %v, 教师ID: %d, 学校ID: %d", phase, subjects, detail.UserID, detail.CurrentSchoolID)
}

// GetTeacherDetailFromContext 从context中获取教师详细信息
func (tm *TeacherMiddleware) GetTeacherDetailFromContext(c *gin.Context) (*itl.TeacherDetailData, bool) {
	value, exists := c.Get(consts.CtxTeacherDetailKey)
	if !exists {
		return nil, false
	}

	detail, ok := value.(*itl.TeacherDetailData)
	return detail, ok
}

// ExtractTeacherPhase 提取教师学段
func (tm *TeacherMiddleware) ExtractTeacherPhase(ctx *gin.Context) int64 {
	return ctx.GetInt64(consts.CtxTeacherPhaseKey)
}

// ExtractTeacherSubjects 提取教师学科
func (tm *TeacherMiddleware) ExtractTeacherSubjects(ctx *gin.Context) []int64 {
	subjectsInterface, exists := ctx.Get(consts.CtxTeacherSubjectsKey)
	tm.log.Debug(ctx, "学科信息 exists=%v, value=%v", exists, subjectsInterface)
	subjects, ok := subjectsInterface.([]int64)
	tm.log.Debug(ctx, "学科信息 转换结果 ok=%v, subjects=%v", ok, subjects)
	return subjects
}

// ExtractTeacherID 提取教师ID
func (tm *TeacherMiddleware) ExtractTeacherID(ctx *gin.Context) int64 {
	return ctx.GetInt64(consts.CtxTeacherIDKey)
}

// ExtractSchoolID 提取学校ID
func (tm *TeacherMiddleware) ExtractSchoolID(ctx *gin.Context) int64 {
	return ctx.GetInt64(consts.CtxSchoolIDKey)
}

// ExtractTeacherClassIDs 提取教师任职的班级ID列表
func (tm *TeacherMiddleware) ExtractTeacherClassIDs(ctx *gin.Context) []int64 {
	detail, _ := tm.GetTeacherDetailFromContext(ctx)

	classIDs := []int64{}
	for _, jobInfo := range detail.TeacherJobInfos {
		for _, job := range jobInfo.JobInfos {
			for _, class := range job.Classes {
				classIDs = append(classIDs, class.ID)
			}
		}
	}
	return classIDs
}

// 提取教师班级 ID => 班级名称
func (tm *TeacherMiddleware) ExtractTeacherClassInfo(ctx *gin.Context) map[int64]*itl.Class {
	detail, _ := tm.GetTeacherDetailFromContext(ctx)

	classIDToNameMap := make(map[int64]*itl.Class)
	for _, jobInfo := range detail.TeacherJobInfos {
		for _, job := range jobInfo.JobInfos {
			for _, class := range job.Classes {
				classIDToNameMap[class.ID] = &class
			}
		}
	}
	return classIDToNameMap
}

// 获取教师的 ID、学校 ID
func (tm *TeacherMiddleware) GetTeacherIDInfo(c *gin.Context) (int64, int64, error) {
	detail, ok := tm.GetTeacherDetailFromContext(c)
	if !ok {
		return 0, 0, errors.New("teacher detail not found")
	}

	return detail.UserID, detail.CurrentSchoolID, nil
}

// 检查教师有科目的权限
// 班主任有班级的全科权限，其他教师只有指定科目的权限
func (tm *TeacherMiddleware) CheckSubjectPermission(ctx *gin.Context, classID int64, subjectID int64) bool {
	detail, ok := tm.GetTeacherDetailFromContext(ctx)
	if !ok {
		return false
	}

	foundClass := false
	foundSubject := false
	for _, jobInfo := range detail.TeacherJobInfos {
		// 班级
		for _, job := range jobInfo.JobInfos {
			if foundClass {
				break
			}

			for _, class := range job.Classes {
				if class.ID == classID {
					foundClass = true
					break
				}
			}
		}

		// 班主任
		if jobInfo.JobType.JobType == consts.JOB_TYPE_CLASS_TEACHER {
			foundSubject = true
			break
		}

		// 学科教师
		if jobInfo.JobType.JobType == consts.JOB_TYPE_SUBJECT_TEACHER &&
			jobInfo.JobSubject.JobSubject == subjectID {
			foundSubject = true
			break
		}
	}

	if !foundClass || !foundSubject {
		return false
	}

	return true
}

// TeacherHasClassPermission 检查教师是否具备班级权限，如果具备所有班级权限，则返回 true，否则返回 false
func (tm *TeacherMiddleware) TeacherHasClassPermission(ctx *gin.Context, classIDs ...int64) bool {
	teacherClassIDs := tm.ExtractTeacherClassIDs(ctx)

	for _, classID := range classIDs {
		if !slices.Contains(teacherClassIDs, classID) {
			return false
		}
	}
	return true
}

// GetTaskCreationSubjects 创建任务相关，返回学科教师和班主任拥有的学科
// 只有学科教师和班主任有权限创建任务，学科教师限制了具体的学科，班主任则具备全部学科的权限
func (tm *TeacherMiddleware) GetTaskCreationSubjects(ctx *gin.Context) []int64 {
	detail, _ := tm.GetTeacherDetailFromContext(ctx)
	phase := tm.ExtractTeacherPhase(ctx)

	subjects := []int64{}
	for _, jobInfo := range detail.TeacherJobInfos {
		// 班主任，具备全部学科的权限，直接返回全部学科
		if jobInfo.JobType.JobType == consts.JOB_TYPE_CLASS_TEACHER {
			return consts.Phase2SubjectMap[phase]
		}

		// 学科教师，具备指定学科的权限，返回指定学科
		if jobInfo.JobType.JobType == consts.JOB_TYPE_SUBJECT_TEACHER {
			subjects = append(subjects, jobInfo.JobSubject.JobSubject)
		}
	}
	return subjects
}

// GetSubjectTeacherSubjects 只返回教师为学科老师时具备的学科列表
func (tm *TeacherMiddleware) GetSubjectTeacherSubjects(ctx *gin.Context) []int64 {
	detail, _ := tm.GetTeacherDetailFromContext(ctx)
	subjects := []int64{}
	for _, jobInfo := range detail.TeacherJobInfos {
		if jobInfo.JobType.JobType == consts.JOB_TYPE_SUBJECT_TEACHER {
			subjects = append(subjects, jobInfo.JobSubject.JobSubject)
		}
	}
	return subjects
}

// HasTaskCreationSubjectPermission 创建任务相关，检查教师是否具备学科权限
// 只有学科教师和班主任有权限创建任务，学科教师限制了具体的学科，班主任则具备全部学科的权限
func (tm *TeacherMiddleware) HasTaskCreationSubjectPermission(ctx *gin.Context, subjectID int64) bool {
	subjects := tm.GetTaskCreationSubjects(ctx)
	return slices.Contains(subjects, subjectID)
}

// GetTaskReportSubjects 获取教师查看作业报告时具备的学科列表
func (tm *TeacherMiddleware) GetTaskReportSubjects(ctx *gin.Context) []int64 {
	// 一个用户可以具备多个角色，取每个角色的学科并集
	// 校长：具备该校学段下的全部学科的权限
	// 年级主任：具备该校学段下的全部学科的权限
	// 学科组长：具备该校学段下的指定学科的权限
	// 学科教师：具备该校学段下的指定学科的权限
	// 班主任：具备该校学段下的全部学科的权限
	detail, _ := tm.GetTeacherDetailFromContext(ctx)
	phase := tm.ExtractTeacherPhase(ctx)

	// 使用 map 去重
	subjectsMap := make(map[int64]struct{})

	// 遍历所有职务信息
	for _, jobInfo := range detail.TeacherJobInfos {
		// 校长、年级主任、班主任：具备全部学科的权限
		if jobInfo.JobType.JobType == consts.JOB_TYPE_PRINCIPAL ||
			jobInfo.JobType.JobType == consts.JOB_TYPE_GRADE_HEAD ||
			jobInfo.JobType.JobType == consts.JOB_TYPE_CLASS_TEACHER {
			// 返回该校学段下的全部学科
			return consts.Phase2SubjectMap[phase]
		}

		// 学科组长、学科教师：具备指定学科的权限
		if jobInfo.JobType.JobType == consts.JOB_TYPE_SUBJECT_HEAD ||
			jobInfo.JobType.JobType == consts.JOB_TYPE_SUBJECT_TEACHER {
			// 添加指定学科
			if jobInfo.JobSubject.JobSubject != 0 {
				subjectsMap[jobInfo.JobSubject.JobSubject] = struct{}{}
			}
		}
	}

	// 将 map 转换为 slice，并保持顺序
	subjectsSlice := make([]int64, 0, len(subjectsMap))
	for subject := range subjectsMap {
		subjectsSlice = append(subjectsSlice, subject)
	}
	slices.Sort(subjectsSlice)
	return subjectsSlice
}

// GetTaskReportGradeClasses 获取教师具备的年级班级列表
func (tm *TeacherMiddleware) GetTaskReportGradeClasses(ctx *gin.Context) ([]itl.GradeClass, error) {
	// 一个用户可以具备多个角色，取每个角色的年级班级并集
	// 校长：具备全部年级全部班级的权限
	// 年级主任：具备指定年级下的全部班级的权限
	// 学科组长：具备指定年级下的全部班级的权限
	// 学科教师：具备指定年级下的指定班级的权限
	// 班主任：具备指定班级的权限
	detail, _ := tm.GetTeacherDetailFromContext(ctx)

	// 使用 map 去重，key 为年级ID
	gradeClassMap := make(map[int64]itl.GradeClass)

	// 遍历所有职务信息
	for _, jobInfo := range detail.TeacherJobInfos {
		// 校长：具备全部年级全部班级的权限
		if jobInfo.JobType.JobType == consts.JOB_TYPE_PRINCIPAL {
			// 获取当前学校的全部年级班级信息
			gradeClasses, err := tm.ucenterService.GetGradeClassInfo(ctx, detail.CurrentSchoolID)
			if err != nil {
				return nil, err
			}
			// 最高权限，直接返回
			return gradeClasses, nil
		}

		// 年级主任、学科组长：具备指定年级下的全部班级的权限
		if jobInfo.JobType.JobType == consts.JOB_TYPE_GRADE_HEAD ||
			jobInfo.JobType.JobType == consts.JOB_TYPE_SUBJECT_HEAD {
			// 记录年级信息，此时还没有班级信息
			for _, job := range jobInfo.JobInfos {
				gradeClass := itl.GradeClass{
					GradeID:   job.Grade,
					GradeName: job.GradeName,
					Class:     []itl.ClassInfoItem{},
				}
				gradeClassMap[job.Grade] = gradeClass
			}
			continue
		}

		// 学科教师、班主任：具备指定年级下的指定班级的权限
		if jobInfo.JobType.JobType == consts.JOB_TYPE_SUBJECT_TEACHER ||
			jobInfo.JobType.JobType == consts.JOB_TYPE_CLASS_TEACHER {
			for _, job := range jobInfo.JobInfos {
				// 如果已经记录了年级信息，代表有更高的权限，不需要再记录到班级维度
				// 否则记录到具体的班级
				if _, ok := gradeClassMap[job.Grade]; !ok {
					// 由于运营平台定义的字段名称不一致，这里需要转换一下
					classInfoItems := make([]itl.ClassInfoItem, len(job.Classes))
					for i, class := range job.Classes {
						classInfoItems[i] = itl.ClassInfoItem{
							ClassID:   class.ID,
							ClassName: class.Name,
						}
					}

					// 记录年级和班级信息
					gradeClass := itl.GradeClass{
						GradeID:   job.Grade,
						GradeName: job.GradeName,
						Class:     classInfoItems,
					}
					gradeClassMap[job.Grade] = gradeClass
				}
			}
		}
	}

	// 如果 gradeClassMap 只有年级，没有班级，则需要请求运营平台获取年级下的全部班级列表
	gradeIDs := make([]int64, 0, len(gradeClassMap))
	for gradeID, gradeClass := range gradeClassMap {
		if len(gradeClass.Class) == 0 {
			gradeIDs = append(gradeIDs, gradeID)
		}
	}

	// 获取需要查询的年级下的班级信息
	if len(gradeIDs) > 0 {
		gradeClasses, err := tm.ucenterService.GetGradeClassInfo(ctx, detail.CurrentSchoolID, gradeIDs...)
		if err != nil {
			return nil, err
		}

		// 更新 gradeClassMap 中的班级信息
		for _, gradeClass := range gradeClasses {
			gradeClassMap[gradeClass.GradeID] = gradeClass
		}
	}

	// 将 map 转换为 slice，并按年级ID排序
	result := make([]itl.GradeClass, 0, len(gradeClassMap))
	for _, gradeClass := range gradeClassMap {
		result = append(result, gradeClass)
	}
	slices.SortFunc(result, func(a, b itl.GradeClass) int {
		return int(a.GradeID - b.GradeID)
	})

	// 对每个年级的班级按班级ID排序
	for i := range result {
		slices.SortFunc(result[i].Class, func(a, b itl.ClassInfoItem) int {
			return int(a.ClassID - b.ClassID)
		})
	}

	return result, nil
}
