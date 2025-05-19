package dao_task

import (
	"context"
	"errors"

	"gil_teacher/app/consts"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/model/dto"

	"gorm.io/gorm"
)

type taskAssignDAO struct {
	db  *gorm.DB
	log *logger.ContextLogger
}

func NewTaskAssignDAO(db *gorm.DB, log *logger.ContextLogger) TaskAssignDAO {
	return &taskAssignDAO{
		db:  db,
		log: log,
	}
}

// TaskAssign 任务分配表
type TaskAssign struct {
	AssignID  int64 `gorm:"column:assign_id;type:bigserial;primaryKey" json:"assignId"` // 自增主键ID
	TaskID    int64 `gorm:"column:task_id;type:bigint;not null" json:"taskId"`          // 任务ID
	SchoolID  int64 `gorm:"column:school_id;type:bigint;not null" json:"schoolId"`      // 学校ID
	GroupType int64 `gorm:"column:group_type;type:bigint;not null" json:"groupType"`    // 群组类型，1 自定义学生，2 班级
	GroupID   int64 `gorm:"column:group_id;type:bigint;not null" json:"groupId"`        // 群组ID，自定义时为0，班级时为班级ID
	StartTime int64 `gorm:"column:start_time;type:bigint;not null" json:"startTime"`    // 任务开始时间
	Deadline  int64 `gorm:"column:deadline;type:bigint;not null" json:"deadline"`       // 任务截止时间
}

// TableName 指定表名
func (TaskAssign) TableName() string {
	return "tbl_task_assign"
}

func (d *taskAssignDAO) DB(ctx context.Context) *gorm.DB {
	return d.db.Model(&TaskAssign{}).WithContext(ctx)
}

// 获取指定教师指定对象布置（可能为空）的任务总数和分页列表
func (d *taskAssignDAO) GetTeacherTasks(ctx context.Context, reqs *dto.TaskAssignListQuery, pageInfo *consts.DBPageInfo) (int64, []*TaskAssign, error) {
	// 构建基础查询
	query := d.DB(ctx).
		Joins("INNER JOIN tbl_task s ON tbl_task_assign.task_id = s.task_id").
		Where("s.creator_id = ? AND s.school_id = ? AND s.deleted = 0", reqs.TeacherID, reqs.SchoolID)

	// 群组或班级条件
	switch consts.GroupType(reqs.GroupType) {
	case consts.TASK_GROUP_TYPE_CLASS, consts.TASK_GROUP_TYPE_STUDENT: // 班级或学生群组
		if reqs.GroupID != 0 {
			query = query.Where("tbl_task_assign.group_type = ? AND tbl_task_assign.group_id = ?", reqs.GroupType, reqs.GroupID)
		}
	case consts.TASK_GROUP_TYPE_TEMP: // 临时群组
		query = query.Where("tbl_task_assign.group_type = 1 AND tbl_task_assign.group_id = 0")
	}

	// 科目条件
	if reqs.Subject != 0 {
		query = query.Where("s.subject = ?", reqs.Subject)
	}

	// 任务查询开始时间条件
	if reqs.StartTime != 0 {
		query = query.Where("start_time >= ?", reqs.StartTime)
	}

	// 任务查询结束时间条件
	if reqs.EndTime != 0 {
		query = query.Where("start_time <= ?", reqs.EndTime)
	}

	// 模糊检索条件
	if reqs.Keyword != "" {
		query = query.Where("s.task_name LIKE ?", "%"+reqs.Keyword+"%")
	}

	// 任务类型条件
	if reqs.TaskType != 0 {
		query = query.Where("s.task_type = ?", reqs.TaskType)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	// 分页查询
	var tasks []*TaskAssign
	if err := query.
		Offset(int((pageInfo.Page - 1) * pageInfo.Limit)).
		Limit(int(pageInfo.Limit)).
		Find(&tasks).Error; err != nil {
		return 0, nil, err
	}

	return total, tasks, nil
}

// 获取指定任务的全部布置信息
func (d *taskAssignDAO) GetTaskAssignsByTaskIDs(ctx context.Context, taskIDs []int64) ([]*TaskAssign, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}

	var taskAssigns []*TaskAssign
	if err := d.DB(ctx).Where("task_id IN (?)", taskIDs).Find(&taskAssigns).Error; err != nil {
		return nil, err
	}

	return taskAssigns, nil
}

// 获取指定任务指定布置的统计数据，如果指定了assignID，则返回指定布置的统计数据，否则返回指定任务的全部布置数据
func (d *taskAssignDAO) GetTaskAssignInfo(ctx context.Context, taskID int64, assignID int64) ([]*TaskAssign, error) {
	if taskID == 0 {
		return nil, errors.New("taskID is invalid")
	}

	var taskAssigns []*TaskAssign
	query := d.DB(ctx).Where("task_id = ?", taskID)

	// 如果指定了assignID，则只返回指定布置的数据
	if assignID != 0 {
		query = query.Where("assign_id = ?", assignID)
	}

	err := query.Find(&taskAssigns).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return taskAssigns, err
}

// GetTaskAssigns 获取指定任务指定布置的统计数据
func (d *taskAssignDAO) GetTaskAssigns(ctx context.Context, taskID int64, assignIds []int64) ([]*TaskAssign, error) {
	if taskID == 0 {
		return nil, errors.New("taskID is invalid")
	}

	var taskAssigns []*TaskAssign
	query := d.DB(ctx).Where("task_id = ?", taskID)
	if len(assignIds) > 0 {
		query = query.Where("assign_id IN (?)", assignIds)
	}

	err := query.Find(&taskAssigns).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return taskAssigns, err
}

// UpdateTaskAssign 更新任务分配
func (d *taskAssignDAO) UpdateTaskAssign(ctx context.Context, assginID int64, updates map[string]any) error {
	return d.DB(ctx).Where("assign_id = ? AND deleted = 0", assginID).Updates(updates).Error
}

// DeleteTaskAssign 删除任务分配，软删除
func (d *taskAssignDAO) DeleteTaskAssign(ctx context.Context, taskID int64, assignIDs []int64) error {
	return d.DB(ctx).Where("task_id = ? AND assign_id IN (?)", taskID, assignIDs).Update("deleted", 1).Error
}

// CountTaskAssignByTaskID 统计指定任务的分配数量
func (d *taskAssignDAO) CountTaskAssignByTaskID(ctx context.Context, taskID int64) (int64, error) {
	var count int64
	if err := d.DB(ctx).Where("task_id = ? AND deleted = 0", taskID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ResourceGroupID 资源ID和群组ID，用于获取资源已布置的班级ID列表
type ResourceGroupID struct {
	ResourceID string `gorm:"column:resource_id"`
	GroupID    int64  `gorm:"column:group_id"`
}

// GetResourceAssignedClassIDs 获取资源已布置的班级ID列表
func (d *taskAssignDAO) GetResourceAssignedClassIDs(
	ctx context.Context,
	req *dto.TaskResourceAssignedClassIDsRequest,
) ([]ResourceGroupID, error) {
	var resourceGroupIDs []ResourceGroupID
	if err := d.DB(ctx).
		Table("tbl_task task").
		Joins("INNER JOIN tbl_task_resource resource ON task.task_id = resource.task_id").
		Joins("INNER JOIN tbl_task_assign assign ON task.task_id = assign.task_id").
		Where("task.task_type = ? AND task.subject = ? AND task.creator_id = ? AND task.deleted = 0 AND resource.resource_type = ? AND resource.resource_id IN (?) AND assign.deleted = 0 AND assign.group_type = ?",
			req.TaskType,
			req.Subject,
			req.TeacherID,
			req.ResourceType,
			req.ResourceIDs,
			req.GroupType).
		Select("DISTINCT resource.resource_id", "assign.group_id").
		Find(&resourceGroupIDs).Error; err != nil {
		return nil, err
	}

	return resourceGroupIDs, nil
}
