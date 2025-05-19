package consts

// MessageType 消息类型
type MessageType string

const (
	MessageTypeTeacherBehavior MessageType = "teacher_behavior"
	MessageTypeStudentBehavior MessageType = "student_behavior"
	MessageTypeCommunication   MessageType = "communication"
)

// CommunicationUserType 会话用户类型
type CommunicationUserType string

const (
	CommunicationUserTypeAI        CommunicationUserType = "ai"        // ai
	CommunicationUserTypeStudent   CommunicationUserType = "student"   // 学生
	CommunicationUserTypeTeacher   CommunicationUserType = "teacher"   // 老师
	CommunicationUserTypeAssistant CommunicationUserType = "assistant" // 助教
)

// CommunicationSessionType 会话类型
type CommunicationSessionType string

const (
	CommunicationSessionTypeQuestion CommunicationSessionType = "question" // 提问
	CommunicationSessionTypeAnswer   CommunicationSessionType = "answer"   // 答疑
	CommunicationSessionTypeChat     CommunicationSessionType = "chat"     // 聊天
	CommunicationSessionTypeOther    CommunicationSessionType = "other"    // 其他
)

// 行为类型
type BehaviorType string

// 行为类型常量
const (
	BehaviorTypeBrowse               BehaviorType = "browse"                // 浏览
	BehaviorTypeAnswer               BehaviorType = "answer"                // 答题
	BehaviorTypeQuestion             BehaviorType = "question"              // 提问
	BehaviorTypeLearning             BehaviorType = "learning"              // 学习
	BehaviorTypeInteract             BehaviorType = "interact"              // 互动
	BehaviorTypeChat                 BehaviorType = "chat"                  // 沟通
	BehaviorTypeAssignTask           BehaviorType = "assign_task"           // 教师布置任务的行为
	BehaviorTypeClassComment         BehaviorType = "class_comment"         // 教师课堂评价的行为
	BehaviorTypePraise               BehaviorType = "praise"                // 教师表扬学生的行为
	BehaviorTypeAttention            BehaviorType = "attention"             // 教师关注/提醒学生的行为
	BehaviorTypeCommunication        BehaviorType = "communication"         // 会话
	BehaviorTypeOfflineCommunication BehaviorType = "offline_communication" // 教师发起线下沟通的行为

	// 教师对学生作业进行点赞、提醒
	BehaviorTypeTaskPraise    BehaviorType = "task_praise"    // 教师对学生作业进行点赞
	BehaviorTypeTaskAttention BehaviorType = "task_attention" // 教师对学生作业进行提醒
)

// 学习类型常量
const (
	LearningTypeCourse    = "课程学习" // 课程学习
	LearningTypeSelfStudy = "课堂自学" // 课堂自学
)

// 行为分类常量
const (
	BehaviorCategoryPraise    = "值得表扬" // 值得表扬的学生
	BehaviorCategoryAttention = "建议关注" // 建议关注的学生
	BehaviorCategoryHandled   = "已处理"  // 已处理的学生
)

// 行为触发规则常量
const (
	// 值得表扬规则
	PraiseRuleCorrectAnswersThreshold          = 3    // 连对题目数阈值
	PraiseRuleQuestionCountThreshold           = 1    // 提问次数阈值
	PraiseRuleLearningScoreImprovePercent      = 10.0 // 学习分提升百分比阈值
	PraiseCorrectStreakScoreWeight             = 5    // 连对行为表扬加权分值
	PraiseAccuracyRateThresholdForBonus        = 80   // (表扬加分规则)正确率阈值 - 高于此值则加分
	PraiseAccuracyRateBonusScore               = 10   // (表扬加分规则)正确率奖励分数 - 达到阈值后的加分值
	PraiseAnswerTypeBonusScore                 = 10   // (表扬加分规则)行为类型为答题时的奖励分数
	PraiseQuestionCountScoreWeight             = 10   // (表扬加分规则)提问次数的表扬加权分值
	PraiseQuestionTypeBonusScore               = 15   // (表扬加分规则)行为类型为提问时的奖励分数
	PraiseEarlyLearnCountScoreWeight           = 10   // (表扬加分规则)提前学习次数的表扬加权分值
	PraiseStayDurationThresholdMinutesForBonus = 15   // (表扬加分规则)学习时长奖励阈值(分钟) - 超过此时长开始计算额外加分
	PraiseStayDurationBonusIntervalMinutes     = 5    // (表扬加分规则)学习时长奖励的计分区间(分钟)
	PraiseStayDurationBonusScorePerInterval    = 1    // (表扬加分规则)学习时长每计分区间奖励分数
	PraiseLongStayDurationBaseBonusScore       = 10   // (表扬加分规则)学习时长较长时的基础奖励分数
	PraiseSelfStudyLearningTypeBonusScore      = 15   // (表扬加分规则)学习类型为自学时的奖励分数
	PraiseLearningBehaviorTypeBonusScore       = 10   // (表扬加分规则)行为类型为学习时的奖励分数
	PraiseTypeInitialBaseScore                 = 5    // (表扬加分规则)各类表扬标签的初始基础分数
	PraiseBestTypeMinScoreThreshold            = 5    // (表扬加分规则)判断是否找到最佳表扬类型的最低分数阈值 (通常等于PraiseTypeInitialBaseScore)

	// 表扬检查规则
	PraiseCheckCorrectStreak           = 3    // 连续答对次数阈值
	PraiseCheckMinCorrectAnswers       = 3    // 最小正确答题数
	PraiseCheckCorrectRateThreshold    = 0.8  // 正确率阈值 (80%)
	PraiseCheckMinStayDurationSeconds  = 1800 // 最小学习时长阈值(秒)，30分钟
	PraiseCheckEarlyLearnCountRequired = 1    // 提前学习次数要求
	PraiseCheckQuestionCountRequired   = 1    // 提问次数要求

	// 建议关注规则
	AttentionRuleFrequentPageSwitchTime     = 2   // 频繁切换页面时间阈值(分钟)
	AttentionRuleNotInLearningPageTime      = 2   // 未在学习页面时间阈值(分钟)
	AttentionRuleOtherSubjectTime           = 2   // 学习其他学科时间阈值(分钟)
	AttentionRuleNoOperationTime            = 5   // 单一页面无操作时间阈值(分钟)
	AttentionRuleNotUsingPadTime            = 5   // 未使用pad时间阈值(分钟)
	AttentionRuleLearningScoreDecreasePerct = 20  // 学习分下降百分比阈值
	AttentionRuleVideoPauseTimeSeconds      = 300 // 视频暂停时间阈值(秒)，5分钟
	AttentionWindowSeconds                  = 60  // 关注时间窗口(秒)
	AttentionRuleBaseScoreForLongPause      = 5   // 长时间暂停的基础加分
	AttentionRuleScorePerMinutePaused       = 1   // 每暂停一分钟的加分
	AttentionRulePageSwitchBaseScore        = 10  // 频繁切换页面的基础加分
	AttentionRulePageSwitchScorePerCount    = 2   // 每次切换页面的额外加分
	AttentionRuleOtherContentBaseScore      = 10  // 学习其他内容的基础加分
	AttentionRuleOtherContentScorePerCount  = 2   // 每次学习其他内容的额外加分
	AttentionRuleVideoPauseLearningScore    = 15  // 学习时视频暂停的加分
	AttentionRulePauseCountBaseScore        = 10  // 暂停操作的基础加分
	AttentionRulePauseCountScorePerCount    = 2   // 每次暂停操作的额外加分
)

// 行为描述常量 (不包含教师姓名)
const (
	BehaviorDescDefaultPraiseSimple        = "表现很棒，继续保持！"
	BehaviorDescAnswerCorrectStreakSimple  = "太棒了！你已经连续答对 %d 题了！"
	BehaviorDescFocusedLearningSimple      = "你学习得很专注，继续加油！"
	BehaviorDescActiveQuestioningSimple    = "你提的问题很有价值，思考很深入！"
	BehaviorDescPraiseGeneralCorrectAnswer = "你的答题表现很棒，继续保持！"
	BehaviorDescPraiseEarlyLearnSpecific   = "你已经提前学习了 %d 次，表现很出色！"
	BehaviorDescPraiseEarlyLearnGeneral    = "你认真预习的态度值得表扬！"
	BehaviorDescPraiseQuestionSpecific     = "你的提问非常有价值，思考很深入！"
	BehaviorDescPraiseQuestionGeneral      = "你积极参与思考的态度很好！"

	// 新增关注提醒文案常量
	AttentionDescVideoPausedNeedHelp    = "视频处于暂停状态，需要帮助吗？"
	AttentionDescFrequentPausesFocus    = "操作频繁暂停，请专注学习"
	AttentionDescIrrelevantContentFocus = "正在浏览课程无关内容，请回到学习页面"
	AttentionDescFrequentSwitchFocus    = "频繁切换页面，请保持专注"

	// 建议关注的描述常量
	BehaviorDescLowAccuracyRate       = "本次答题正确率 %.1f%%，需要加强练习哦。"
	BehaviorDescAnsweringDifficulties = "遇到困难了吗？可以向老师请教。"
	BehaviorDescVideoPausedMinutes    = "视频暂停了 %d 分钟，需要帮助吗？"
	BehaviorDescVideoPaused           = "视频暂停了，需要帮助吗？"
	BehaviorDescNeedAttention         = "请保持专注，积极参与课堂学习。"
)

// 行为类型描述常量
const (
	BehaviorDescTypeCorrectStreak = "连续答对"
	BehaviorDescTypeEarlyLearn    = "提前学习"
	BehaviorDescTypeQuestion      = "主动提问"
	BehaviorDescTypePageSwitch    = "频繁切换页面"
	BehaviorDescTypeOtherContent  = "浏览其他内容"
	BehaviorDescTypePause         = "长时间暂停"
)

// 行为处理结果常量
const (
	BehaviorResultPraiseSuccess    = "表扬成功"
	BehaviorResultAttentionSuccess = "提醒成功"
	BehaviorResultAlreadyPraised   = "已表扬过该学生"
	BehaviorResultAttentionWindow  = "已提醒过该学生，请稍后重试"
	BehaviorResultFailed           = "处理失败"
	BehaviorResultEvaluateSuccess  = "发送评价成功"
	BehaviorResultAlreadyEvaluated = "已评价过该学生"
	BehaviorResultMaxPraiseReached = "该学生今日已达到最大表扬次数(%d次)"
	BehaviorResultNotPraiseWorthy  = "该学生当前不满足表扬条件"
	BehaviorResultSystemError      = "系统错误，请联系管理员"
	BehaviorResultAllPraiseFailed  = "表扬失败：学生已达到最大表扬次数或不满足表扬条件"
)

// 行为限制常量
const (
	MaxDailyPraiseCount = 3 // 每日最大表扬次数
)

// 评价学生提示语
const (
	EvaluatePromptTitle   = "学习评价"
	EvaluatePromptDefault = "今天课堂整体表现很好，学习专注度和正确率都有明显提升，相对比xxx提升了xxx，但仍有xxx需要关注一下。"
)

// 线下沟通文案，只在教师端展示，不推送学生端
var OfflineCommunicationContents = []string{
	"今天答题状态不是很好，是不是课程学习有点难，有需要我帮忙的吗？",
	"我理解有些知识点确实不容易掌握，老师以前学习时也遇到过类似问题。你觉得我们可以一起做些什么来改善呢？",
	"老师很在意你的学习状态，也希望能帮到你。如果有什么需要支持的，可以随时告诉我。",
	"最近课堂上，老师注意到你偶尔会走神/没有参与讨论(描述具体现象)。是不是最近遇到了什么困难？",
}

// 参考 https://wcng60ba718p.feishu.cn/wiki/R1N4waSckieoockdM2kcVHcVnvg?sheet=5e8ecf
const (
	// 鼓励类文案
	AttentionTextClassContinueCorrect = "%s同学在本次作业中，以 %d 连对题目数打破了本班之前的连对 %d 题历史，值得老师的肯定。"
	AttentionTextSelfImprovement      = "%s同学相比上次作业，个人学习分提升了 %d%%，这是持续努力学习的最好证明，值得老师的肯定。"
	AttentionTextSelfContinueCorrent  = "%s同学在本次作业中，以 %d 连对题目数打破了自己之前保持的连对 %d 题历史，值得老师的肯定。"
	AttentionTextLayerLeading         = "%s同学在本次作业中，以超过同层同学 %d%% 的优异成绩展现了领跑实力，值得老师的表扬。"
	AttentionTextCompleteBeforeTask   = "%s同学在作业发布前就完成了自学，这种主动性值得老师的肯定。"
	AttentionTextAskFrequently        = "%s同学在本次作业中，以提问 %d 次的表现展现了好学不倦，值得老师的鼓励。"

	PushDefaultTextClassContinueCorrect = "你在本课学习中，以连对 %d 题打破了我们班的连对 %d 题历史。"
	PushDefaultTextSelfImprovement      = "你比之前有了非常大的进步，个人学习分提升了 %d%%。"
	PushDefaultTextSelfContinueCorrent  = "你在本课学习中，以连对 %d 题打破了自己之前保持的连对 %d 题历史。"
	PushDefaultTextLayerLeading         = "你以超过同层同学 %d%% 的优异成绩在课程中展现了领跑者实力。"
	PushDefaultTextCompleteBeforeTask   = "你已经提前学习了本次课程。"
	PushDefaultTextAskFrequently        = "你在本次作业中勤于思考和提问。"

	// 关注类文案
	AttentionTextNoProgress          = "%s同学在本次作业中，进度明显低于班级其他同学。"
	AttentionTextOverdueNotCompleted = "%s同学在本次作业中，存在逾期。可能需要沟通了解卡点，并调整作业截止时间，并在学生完成后鼓励他完成了作业。"
	AttentionTextOverdueCompleted    = "%s同学在本次作业中，存在逾期。可能需要沟通了解卡点，沟通后续学习计划和学习方法是否需要调整。"
	AttentionTextHistoryDeviation    = "%s同学在本次作业中，较上次同类任务学习分下降了 %d%%。"
	AttentionTextLayerDeviation      = "%s同学在本次作业中，较本层同学的任务学习分偏离了 %d%%。"
	AttentionTextContentDeviation    = "%s同学在本次作业中，%s课程的学习情况显著低于其他学习内容。"

	PushDefaultTextNoProgress          = "%s课程中，班级同学的进度已经达到%d%%了，你的进度有点落后了：别让知识等你太久哦，点击这里立即开启学习吧。"
	PushDefaultTextOverdueNotCompleted = "你在%s作业遇到了困难：老师这次可以帮你调整作业完成时间，点击这里快去完成吧。"
	PushDefaultTextOverdueCompleted    = "你在%s作业中遇到了困难：不过已经完成了作业，为你点赞！"
	PushDefaultTextHistoryDeviation    = "发现你在%s课程的掌握情况较上期下降了 %d%%：这是你之前擅长的领域，完成错题重练，做一下提升吧。"
	PushDefaultTextLayerDeviation      = "发现你在%s课程的掌握情况较其他同学偏离了 %d%%。相信你的实力不止是这样，来点巩固练习，做一下提升吧。"
	PushDefaultTextContentDeviation    = "发现你在%s课程存在理解偏差。相信你一定可以攻克，来点专项训练，清除薄弱点吧。"

	// 未命中提醒
	AttentionTextDefault   = "%s同学在本次作业中，进度和班级平均进度接近,学习分和同层同学的平均分也接近一致。建议老师多干预，让学生感受到自己被关注。可通过以下方式进行："
	PushDefaultTextDefault = "%s同学，老师看到:%s课程中，你已经完成了%d%%，完成情况还不错～老师为你点赞！相信你的实力，一定可以更上一层楼。"
)

var (
	// 未命中提醒列表
	AttentionTextDefaultList = []string{
		"提升动力：发送激励信息，安排学习伙伴",
		"优化方法：查看学习报告，了解遇到的卡点",
		"定向提升：针对薄弱环节，错题重做和加练",
	}
)
