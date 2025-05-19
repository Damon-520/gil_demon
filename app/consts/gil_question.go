package consts

import "time"

// 题库服务相关常量
// 年份
const (
	GilQuestionYearBeforeCondition int64 = 1000 // 更早
)

// 问题类型
const (
	GilQuestionTypeSingleChoice   int64 = 1 // 单选
	GilQuestionTypeMultipleChoice int64 = 2 // 多选
	GilQuestionTypeFillInTheBlank int64 = 3 // 填空
	GilQuestionTypeDefault        int64 = 999
)

var GilQuestionTypeNameMap = map[int64]string{
	GilQuestionTypeSingleChoice:   "单选题",
	GilQuestionTypeMultipleChoice: "多选题",
	GilQuestionTypeFillInTheBlank: "填空题",
	GilQuestionTypeDefault:        "默认题",
}

// QuestionAPI 题库 API
type QuestionAPI struct {
	Method string
	Path   string
}

// API 路径常量
const (
	// API 版本
	QuestionAPIPrefix = "/api/internal/v1"

	// 场景类型
	QuestionSceneCategoryAIClass  = 1 // AI 课
	QuestionSceneCategoryPractice = 2 // 巩固练习

	// 业务树类型
	QuestionBizTreeTypeAll            = 0 // 全部类型
	QuestionBizTreeTypeChapter        = 1 // 章节类型
	QuestionBizTreeTypeKnowledgePoint = 2 // 知识点类型

	// 题集上架状态
	QuestionSetShelfStatusOffShelf int64 = 0 // 未上架
	QuestionSetShelfStatusOnShelf  int64 = 1 // 已上架

)

// 题库错误码
const (
	QuestionAPICodeQuestionSetNotExist = 300201 // 题集不存在
)

// API 配置
var (
	QuestionAPIDefaultTimeout = 10 * time.Second
)

var (
	// 业务树相关
	QuestionAPIBizTreeList = QuestionAPI{
		Method: "POST",
		Path:   QuestionAPIPrefix + "/base/data/biz/tree/list",
	}
	QuestionAPIBizTreeDetail = QuestionAPI{
		Method: "GET",
		Path:   QuestionAPIPrefix + "/base/data/biz/tree/detail",
	}

	// 题集、巩固练习相关
	QuestionAPIPracticeListByID = QuestionAPI{
		Method: "POST",
		Path:   QuestionAPIPrefix + "/scene/get/question/set/list",
	}
	QuestionAPIGetQuestionSetInfo = QuestionAPI{
		Method: "GET",
		Path:   QuestionAPIPrefix + "/scene/get/stable/question/set/info",
	}

	// 题目相关
	QuestionAPIQuestionEnums = QuestionAPI{
		Method: "GET",
		Path:   QuestionAPIPrefix + "/enums/get/consts",
	}
	QuestionAPIQuestionList = QuestionAPI{
		Method: "POST",
		Path:   QuestionAPIPrefix + "/resource/get/question/list",
	}
	QuestionAPIQuestionDetail = QuestionAPI{
		Method: "GET",
		Path:   QuestionAPIPrefix + "/resource/get/question/info",
	}
	QuestionAPIQuestionListByID = QuestionAPI{
		Method: "POST",
		Path:   QuestionAPIPrefix + "/resource/get/question/info/list",
	}
)
