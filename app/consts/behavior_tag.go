package consts

// BehaviorTagType 学生行为标签类型
type BehaviorTagType string

const (
	// 表现优秀的行为标签类型
	BehaviorTagTypeEarlyLearn    BehaviorTagType = "earlyLearn"    // 提前学习
	BehaviorTagTypeQuestion      BehaviorTagType = "question"      // 提问
	BehaviorTagTypeCorrectStreak BehaviorTagType = "correctStreak" // 连续答对

	// 需要关注的行为标签类型
	BehaviorTagTypePageSwitch   BehaviorTagType = "pageSwitch"   // 频繁切换页面
	BehaviorTagTypeOtherContent BehaviorTagType = "otherContent" // 学习其他内容
	BehaviorTagTypePause        BehaviorTagType = "pause"        // 停顿操作
)

// 行为标签提示文本模板
const (
	// 表现优秀的行为提示文本模板（更简洁友好）
	BehaviorTagTextEarlyLearn    = "提前学习 %d次"
	BehaviorTagTextQuestion      = "主动提问 %d次"
	BehaviorTagTextCorrectStreak = "连续答对 %d题"

	// 需要关注的行为提示文本模板（更明确具体）
	BehaviorTagTextPageSwitch   = "切换页面 %d次"
	BehaviorTagTextOtherContent = "浏览其他内容 %d次"
	BehaviorTagTextPause        = "长时间暂停 %d次"
)
