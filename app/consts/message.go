package consts

import (
	"fmt"
)

// 提醒消息模板
const (
	// 基础提醒消息模板
	MsgTplKeepFocused = "请保持专注。已提醒%d次"
	// 带问题描述的提醒消息模板
	MsgTplKeepFocusedWithProblems = "发现你%s，请保持专注。已提醒%d次"

	// 行为描述模板
	BehaviorTplConsecutiveCorrect = "连对%d题，正确率%.1f%%"
	BehaviorTplFocusedLearning    = "专注学习%d分钟"
	BehaviorTplActiveQuestion     = "积极提问：%s"

	// 问题描述模板
	ProblemTplFrequentPageSwitch = "频繁切换页面：%d次"
	ProblemTplOtherContent       = "学习其它内容：%d次"
	ProblemTplPauseOperation     = "停顿操作：%d次"
)

// FormatMessage 格式化消息
// params: 可变参数，按照模板所需参数顺序传入
func FormatMessage(tpl string, params ...interface{}) string {
	return fmt.Sprintf(tpl, params...)
}
