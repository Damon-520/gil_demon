package dao_task

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"gil_teacher/app/consts"
	clogger "gil_teacher/app/core/logger"
)

type taskStudentsReportDao struct {
	db     *gorm.DB
	logger *clogger.ContextLogger
}

func NewTaskStudentsReportDao(db *gorm.DB, logger *clogger.ContextLogger) TaskStudentsReportDao {
	return &taskStudentsReportDao{
		db:     db,
		logger: logger,
	}
}

// ResourceDetailReport 定义任务完成报告结构
type ResourceDetailReport struct {
	StudyScore        int64   `gorm:"study_score" json:"ss"`        // 学习分
	CompletedProgress float64 `gorm:"completed_progress" json:"cp"` // 完成进度
	AccuracyRate      float64 `gorm:"accuracy_rate" json:"ar"`      // 正确率
	AnswerCount       int64   `gorm:"answer_count" json:"ac"`       // 答题数
	IncorrectCount    int64   `gorm:"incorrect_count" json:"ic"`    // 错题数
	CostTime          int64   `gorm:"cost_time" json:"ct"`          // 答题用时，秒
}

// ResourceReportMap 定义任务报告映射类型
type ResourceReportMap map[string]ResourceDetailReport

// TaskStudentsReport 定义任务学生报告结构
type TaskStudentsReport struct {
	ID                   int64             `gorm:"id"`
	TaskID               int64             `gorm:"task_id"`
	AssignID             int64             `gorm:"assign_id"`
	StudentID            int64             `gorm:"student_id"`
	ResourceDetailReport                   // 任务完成报告
	ResourceReport       ResourceReportMap `gorm:"resource_report"` // 分资源统计的报告 map[resource_id#resource_type]TaskCompleteReport
	CreateTime           int64             `gorm:"create_time"`     // 首次统计时间
	UpdateTime           int64             `gorm:"update_time"`     // 最后更新时间
}

// Value 实现 driver.Valuer 接口，用于将 TaskReportMap 转换为数据库值
func (m ResourceReportMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan 实现 sql.Scanner 接口，用于从数据库值转换为 TaskReportMap
func (m *ResourceReportMap) Scan(value any) error {
	if value == nil {
		*m = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil
	}

	if len(data) == 0 {
		*m = make(ResourceReportMap)
		return nil
	}

	return json.Unmarshal(data, m)
}

// GormDataType 实现 gorm 的数据类型接口
func (ResourceReportMap) GormDataType() string {
	return "jsonb"
}

func (m *TaskStudentsReport) TableName() string {
	return "tbl_task_students_report"
}

func (m *taskStudentsReportDao) DB(ctx context.Context) *gorm.DB {
	return m.db.Model(&TaskStudentsReport{}).WithContext(ctx)
}

// 插入任务学生统计数据
func (d *taskStudentsReportDao) Insert(ctx context.Context, report *TaskStudentsReport) error {
	return d.DB(ctx).Create(report).Error
}

// 查询指定任务指定学生的统计数据
func (d *taskStudentsReportDao) FindByTaskIDAndStudentID(ctx context.Context, taskID, assignID, studentID int64) (*TaskStudentsReport, error) {
	var report TaskStudentsReport
	err := d.DB(ctx).Where("task_id = ? and assign_id = ? and student_id = ?", taskID, assignID, studentID).First(&report).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &report, nil
}

// 查询指定任务指定布置指定学生 id 列表的统计数据
func (d *taskStudentsReportDao) FindTaskStudentsReports(ctx context.Context, taskID int64, assignID int64, studentIDs []int64, pageInfo *consts.DBPageInfo) ([]*TaskStudentsReport, int64, error) {
	if taskID == 0 || assignID == 0 {
		return nil, 0, nil
	}

	db := d.DB(ctx).Where("task_id = ? AND assign_id = ?", taskID, assignID)
	if len(studentIDs) > 0 {
		db = db.Where("student_id IN (?)", studentIDs)
	}

	var err error
	var count int64
	// count
	if !pageInfo.All {
		if err = db.Count(&count).Error; err != nil {
			return nil, 0, err
		}
	}

	if pageInfo.SortBy == "" {
		pageInfo.SortBy = "id"
		pageInfo.SortType = consts.SortTypeAsc
	}
	var reports []*TaskStudentsReport
	db = db.Order(fmt.Sprintf("%s %s", pageInfo.SortBy, pageInfo.SortType))
	// 查询全部就不分页
	if pageInfo.All {
		err = db.Find(&reports).Error
	} else {
		err = db.Limit(int(pageInfo.Limit)).Offset(int((pageInfo.Page - 1) * pageInfo.Limit)).Find(&reports).Error
	}
	if err != nil {
		d.logger.Error(ctx, "[FindTaskStudentsReports]查询失败, taskID: %d, assignID: %d, studentIDs: %v, err: %v",
			taskID, assignID, studentIDs, err)
		return nil, 0, err
	}

	return reports, count, nil
}

// 查询指定任务全部学生的统计数据
func (d *taskStudentsReportDao) FindByTaskID(ctx context.Context, taskID int64, pageInfo *consts.DBPageInfo) ([]*TaskStudentsReport, error) {
	var reports []*TaskStudentsReport
	err := d.DB(ctx).Where("task_id = ?", taskID).Find(&reports).Limit(int(pageInfo.Limit)).Offset(int((pageInfo.Page - 1) * pageInfo.Limit)).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}

// 查询指定任务指定布置ID列表的统计数据
func (d *taskStudentsReportDao) FindByTaskIDAndAssignIDs(ctx context.Context, taskID int64, assignIDs []int64, pageInfo *consts.DBPageInfo) ([]*TaskStudentsReport, error) {
	var reports []*TaskStudentsReport
	err := d.DB(ctx).Where("task_id = ? and assign_id in (?)", taskID, assignIDs).Find(&reports).Limit(int(pageInfo.Limit)).Offset(int((pageInfo.Page - 1) * pageInfo.Limit)).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}
