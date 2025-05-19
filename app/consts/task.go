package consts

import "slices"

// GroupType 群组类型
type GroupType int64

// 连接符
const CombineKey = "#"

// 作业任务报告导出字段
var ExportFields = []string{
	"studentName",                // 学生姓名
	"studyScore",                 // 学习分
	"progress",                   // 完成进度
	"accuracyRate",               // 正确率
	"difficultyDegree",           // 难度
	"incorrectCount/answerCount", // 错题数/答题数
	// "answerTime",                 // 答题用时
}

// 支持的排序字段
var ReportSortKeys = []string{
	"studyScore",   // 学习分
	"progress",     // 完成进度
	"accuracyRate", // 正确率
	"answerCount",  // 答题数
}

// 字段导出中文名
var ExportFieldsCN = map[string]string{
	"studentName":                "学生姓名",
	"studyScore":                 "学习分",
	"progress":                   "完成进度",
	"accuracyRate":               "正确率",
	"difficultyDegree":           "答题难度",
	"incorrectCount/answerCount": "错题/答题",
	// "answerTime":                 "答题用时",
}

// 检查字段是否在导出字段中
func IsExportFields(fields []string) bool {
	for _, f := range fields {
		if !slices.Contains(ExportFields, f) {
			return false
		}
	}
	return true
}

// 提取指定的字段名
func ExtractExportFields(fields []string) []string {
	result := make([]string, 0, len(fields))
	for _, f := range fields {
		result = append(result, ExportFieldsCN[f])
	}
	return result
}
