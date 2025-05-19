package dao_task

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"

	clogger "gil_teacher/app/core/logger"

	"gorm.io/gorm"
)

type taskReportSettingDao struct {
	db     *gorm.DB
	logger *clogger.ContextLogger
}

func NewTaskReportSettingDao(db *gorm.DB, logger *clogger.ContextLogger) TaskReportSettingDao {
	return &taskReportSettingDao{
		db:     db,
		logger: logger,
	}
}

type TaskReportSetting struct {
	ID         int64    `gorm:"column:id;type:bigserial;primaryKey"`
	SchoolID   int64    `gorm:"column:school_id;type:bigint;not null"`
	Subject    int64    `gorm:"column:subject;type:bigint;not null"`
	TeacherID  int64    `gorm:"column:teacher_id;type:bigint;not null"`
	ClassID    int64    `gorm:"column:class_id;type:bigint;not null"`
	Setting    *Setting `gorm:"column:setting;type:jsonb"`
	CreateTime int64    `gorm:"column:create_time;type:bigint"`
	UpdateTime int64    `gorm:"column:update_time;type:bigint"`
}

// 任务报告设置
type Setting struct {
	OneClickPraise    bool `json:"oneClickPraise"`    // 一键点赞
	OneClickAttention bool `json:"oneClickAttention"` // 一键提醒
	// 共性错题
	CommonIncorrectQuestion struct {
		AnswerNum   int64   `json:"answerNum"`   // 答题人数
		CorrectRate float64 `json:"correctRate"` // 正确率
	} `json:"commonIncorrectQuestion"` // 共性错题
	// 学习状态设置，大于设置值提示"有进步"，小于设置值提示"需提醒"
	StudyStatus struct {
		StudyScore       int64   `json:"studyScore"`       // 学习分
		CompletionRate   float64 `json:"completionRate"`   // 完成进度
		CorrectRate      float64 `json:"correctRate"`      // 正确率
		DifficultyDegree float64 `json:"difficultyDegree"` // 答题难度
		IncorrectNum     int64   `json:"incorrectNum"`     // 错题数
		AnswerNum        int64   `json:"answerNum"`        // 答题数
	} `json:"studyStatus"` // 学习状态
}

// Value 实现 driver.Valuer 接口，用于将 Setting 转换为数据库值
func (s Setting) Value() (driver.Value, error) {
	if s == (Setting{}) {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan 实现 sql.Scanner 接口，用于从数据库值转换为 Setting
func (s *Setting) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal Setting value")
	}

	return json.Unmarshal(bytes, s)
}

// Setting字段解析

func (m *TaskReportSetting) TableName() string {
	return "tbl_task_report_setting"
}

func (m *taskReportSettingDao) DB(ctx context.Context) *gorm.DB {
	return m.db.WithContext(ctx).Model(&TaskReportSetting{})
}

func (d *taskReportSettingDao) CreateTaskReportSetting(ctx context.Context, setting *TaskReportSetting) error {
	if setting == nil {
		return errors.New("entity is nil")
	}

	return d.DB(ctx).Create(setting).Error
}

func (d *taskReportSettingDao) UpdateTaskReportSetting(ctx context.Context, id int64, setting *TaskReportSetting) error {
	if setting == nil || id <= 0 {
		return errors.New("entity is nil or id is invalid")
	}

	setting.ID = 0 // 需要置为 0，避免更新 id
	return d.DB(ctx).Where("id = ?", id).Updates(setting).Error
}

// 查询指定学校班级指定科目的设置
func (d *taskReportSettingDao) GetSettingByClassIDAndSubjectID(ctx context.Context, schoolID, classID, subjectID int64) (*TaskReportSetting, error) {
	var setting TaskReportSetting
	err := d.DB(ctx).Where("school_id = ? AND class_id = ? AND subject = ?", schoolID, classID, subjectID).First(&setting).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &setting, nil
}
