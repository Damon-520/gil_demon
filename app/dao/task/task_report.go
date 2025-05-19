package dao_task

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"gil_teacher/app/consts"
	"gil_teacher/app/core/logger"

	"gorm.io/gorm"
)

type taskReportDAO struct {
	db  *gorm.DB
	log *logger.ContextLogger
}

func NewTaskReportDAO(db *gorm.DB, logger *logger.ContextLogger) TaskReportDAO {
	return &taskReportDAO{
		db:  db,
		log: logger,
	}
}

// TaskCompleteStat 定义任务完成情况统计数据
type TaskCompleteStat struct {
	CompletedProgress    float64 `json:"cp,omitempty"` // 完成进度
	AverageProgress      float64 `json:"ap,omitempty"` // 平均进度
	AccuracyRate         float64 `json:"ar,omitempty"` // 正确率
	NeedAttentionNum     int64   `json:"na,omitempty"` // 待关注题目数
	NeedAttentionUserNum int64   `json:"nu,omitempty"` // 需关注学生数
	ClassHours           int64   `json:"ch,omitempty"` // 课时数
	AverageCostTime      int64   `json:"at,omitempty"` // 平均耗时,单位秒
}

// CompleteReportJSON 实现 GORM 的自定义类型
type CompleteReportJSON TaskCompleteStat

// Value 实现 driver.Valuer 接口，用于将 TaskCompleteStat 转换为数据库值
func (d CompleteReportJSON) Value() (driver.Value, error) {
	if d == (CompleteReportJSON{}) {
		return nil, nil
	}
	return json.Marshal(d)
}

// Scan 实现 sql.Scanner 接口，用于从数据库值转换为 TaskCompleteStat
func (d *CompleteReportJSON) Scan(value any) error {
	if value == nil {
		*d = CompleteReportJSON{}
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported type for CompleteReportJSON: %T", value)
	}

	if len(data) == 0 {
		*d = CompleteReportJSON{}
		return nil
	}

	if err := json.Unmarshal(data, d); err != nil {
		return fmt.Errorf("failed to unmarshal CompleteReportJSON: %w", err)
	}

	return nil
}

// GormDataType 实现 gorm 的数据类型接口
func (CompleteReportJSON) GormDataType() string {
	return "jsonb"
}

// ResourceReportJSON 实现 GORM 的自定义类型
type ResourceReportJSON map[string]TaskCompleteStat

// Value 实现 driver.Valuer 接口，用于将 ResourceReportJSON 转换为数据库值
func (d ResourceReportJSON) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	return json.Marshal(d)
}

// Scan 实现 sql.Scanner 接口，用于从数据库值转换为 ResourceReportJSON
func (d *ResourceReportJSON) Scan(value any) error {
	if value == nil {
		*d = ResourceReportJSON{}
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported type for ResourceReportJSON: %T", value)
	}

	if len(data) == 0 {
		*d = ResourceReportJSON{}
		return nil
	}

	if err := json.Unmarshal(data, d); err != nil {
		return fmt.Errorf("failed to unmarshal ResourceReportJSON: %w", err)
	}

	return nil
}

// GormDataType 实现 gorm 的数据类型接口
func (ResourceReportJSON) GormDataType() string {
	return "jsonb"
}

type TaskReport struct {
	ID                   int64              `gorm:"column:id;type:bigserial;primaryKey"`      // 自增主键ID
	TaskID               int64              `gorm:"column:task_id;type:bigint;not null"`      // 任务ID
	AssignID             int64              `gorm:"column:assign_id;type:bigint;not null"`    // 任务布置ID
	ReportDetail         CompleteReportJSON `gorm:"column:report_detail;type:jsonb"`          // 统计数据,json格式
	ResourceReportDetail ResourceReportJSON `gorm:"column:resource_report_detail;type:jsonb"` // 资源维度统计数据,json格式
	CreateTime           int64              `gorm:"column:create_time;type:bigint"`           // 创建时间
	UpdateTime           int64              `gorm:"column:update_time;type:bigint"`           // 更新时间
}

func (m *TaskReport) TableName() string {
	return "tbl_task_report"
}

func (m *taskReportDAO) DB(ctx context.Context) *gorm.DB {
	return m.db.Model(&TaskReport{}).WithContext(ctx)
}

// 查询指定任务的全部统计数据
func (d *taskReportDAO) FindAll(ctx context.Context, taskID int64, pageInfo *consts.DBPageInfo) ([]*TaskReport, error) {
	var reports []*TaskReport
	pageInfo.Check()
	err := d.DB(ctx).Where("task_id = ?", taskID).Find(&reports).Limit(int(pageInfo.Limit)).Offset(int((pageInfo.Page - 1) * pageInfo.Limit)).Error
	if err != nil {
		d.log.Error(ctx, "[FindAll]查询失败, taskID: %d, pageInfo: %v, err: %v", taskID, pageInfo, err)
		return nil, err
	}

	return reports, nil
}

// 查询指定任务指定布置ID的统计数据
func (d *taskReportDAO) FindByTaskIDAndAssignID(ctx context.Context, taskID int64, assignID int64) (*TaskReport, error) {
	var report *TaskReport
	err := d.DB(ctx).Where("task_id = ? and assign_id = ?", taskID, assignID).First(&report).Error
	if err != nil {
		d.log.Error(ctx, "[FindByTaskIDAndAssignID]查询失败, taskID: %d, assignID: %d, err: %v", taskID, assignID, err)
		return nil, err
	}

	return report, nil
}

// 查询指定任务的指定布置ID列表的统计数据
func (d *taskReportDAO) FindByTaskIDAndAssignIDs(ctx context.Context, taskID int64, assignIDs []int64) ([]*TaskReport, error) {
	var reports []*TaskReport
	err := d.DB(ctx).Where("task_id = ? and assign_id in (?)", taskID, assignIDs).Find(&reports).Error
	if err != nil {
		d.log.Error(ctx, "[FindByTaskIDAndAssignIDs]查询失败, taskID: %d, assignIDs: %v, err: %v", taskID, assignIDs, err)
		return nil, err
	}

	return reports, nil
}

// 查询指定布置ID列表的统计数据
//
//	map[taskID][]*TaskReport
func (d *taskReportDAO) FindByTaskAssignIDs(ctx context.Context, taskAssignIdsMap map[int64][]int64) (map[int64][]*TaskReport, error) {
	reportsMap := make(map[int64][]*TaskReport)

	// 构建查询条件
	var conditions []string
	var args []any
	for taskID, assignIDs := range taskAssignIdsMap {
		conditions = append(conditions, "(task_id = ? AND assign_id IN (?))")
		args = append(args, taskID, assignIDs)
	}

	// 执行查询
	var reports []*TaskReport
	err := d.DB(ctx).Where(strings.Join(conditions, " OR "), args...).Find(&reports).Error
	if err != nil {
		d.log.Error(ctx, "[FindByTaskAssignIDs]查询失败, taskAssignIdsMap: %v, err: %v", taskAssignIdsMap, err)
		return nil, err
	}

	// 按 taskID 分组结果
	for _, report := range reports {
		if _, ok := reportsMap[report.TaskID]; !ok {
			reportsMap[report.TaskID] = make([]*TaskReport, 0)
		}
		reportsMap[report.TaskID] = append(reportsMap[report.TaskID], report)
	}

	return reportsMap, nil
}

// 写入任务统计数据
func (d *taskReportDAO) Insert(ctx context.Context, report TaskReport) error {
	err := d.DB(ctx).Create(&report).Scan(&report).Error
	if err != nil {
		d.log.Error(ctx, "[Insert]写入失败, report: %v, err: %v", report, err)
		return err
	}

	return nil
}

// 更新单个任务布置的统计数据
func (d *taskReportDAO) Update(ctx context.Context, taskID int64, assignID int64, detail *TaskCompleteStat) error {
	if detail == nil {
		return nil
	}

	err := d.DB(ctx).Where("task_id = ? and assign_id = ?", taskID, assignID).
		Update("report_detail", CompleteReportJSON(*detail)).Error
	if err != nil {
		d.log.Error(ctx, "[Update]任务统计数据更新失败, taskID: %d, assignID: %d, detail: %v, err: %v", taskID, assignID, detail, err)
		return err
	}

	return nil
}
