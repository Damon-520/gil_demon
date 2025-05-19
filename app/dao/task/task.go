package dao_task

import (
	"context"
	"errors"

	"gil_teacher/app/consts"
	"gil_teacher/app/model/dto"

	"gorm.io/gorm"
)

// Task 任务表
type Task struct {
	TaskID         int64  `gorm:"column:task_id;type:bigserial;primaryKey" json:"taskId"`                                                 // 任务ID，主键
	SchoolID       int64  `gorm:"column:school_id;type:bigint;not null" json:"schoolId"`                                                  // 任务所属学校ID
	Phase          int64  `gorm:"column:phase;type:bigint" json:"phase"`                                                                  // 任务所属学段枚举值
	Subject        int64  `gorm:"column:subject;type:bigint" json:"subject"`                                                              // 任务关联学科枚举值
	TaskType       int64  `gorm:"column:task_type;type:bigint" json:"taskType"`                                                           // 任务类型
	TaskSubType    int64  `gorm:"column:task_sub_type;type:bigint;default:0" json:"taskSubType"`                                          // 任务子类型
	TaskName       string `gorm:"column:task_name;type:varchar(32)" json:"taskName"`                                                      // 任务名称
	TeacherComment string `gorm:"column:teacher_comment;type:varchar(256);default:''" json:"teacherComment"`                              // 老师留言
	TaskExtraInfo  string `gorm:"column:task_extra_info;type:text" json:"taskExtraInfo"`                                                  // 任务额外信息
	Deleted        int64  `gorm:"column:deleted;type:bigint" json:"-"`                                                                    // 任务是否删除标识
	CreatorID      int64  `gorm:"column:creator_id;type:bigint" json:"-"`                                                                 // 任务创建者ID
	UpdaterID      int64  `gorm:"column:updater_id;type:bigint" json:"-"`                                                                 // 任务更新者ID
	CreateTime     int64  `gorm:"column:create_time;type:bigint;default:EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT" json:"createTime"` // 任务创建时间（UTC秒数）
	UpdateTime     int64  `gorm:"column:update_time;type:bigint;default:EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT" json:"updateTime"` // 任务信息更新时间（UTC秒数）
}

// TableName 指定表名
func (Task) TableName() string {
	return "tbl_task"
}

// ----------------------------------------------
// ----------------------------------------------
// taskDAO 任务数据访问实现
type taskDAO struct {
	db *gorm.DB
}

// NewTaskDAO 创建任务数据访问实例
func NewTaskDAO(db *gorm.DB) TaskDAO {
	return &taskDAO{db: db}
}

// GetDB 获取数据库连接
func (d *taskDAO) GetDB() *gorm.DB {
	return d.db
}

// DB 返回带有上下文的数据库连接
func (d *taskDAO) DB(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

// CreateTask 创建任务
func (d *taskDAO) CreateTask(ctx context.Context, task *Task) error {
	return d.DB(ctx).Create(task).Error
}

// DeleteTasks 批量软删除任务
func (d *taskDAO) DeleteTasks(ctx context.Context, taskIDs []int64, creatorID int64) error {
	if err := d.DB(ctx).Model(&Task{}).Where("task_id IN ? AND creator_id = ?", taskIDs, creatorID).Update("deleted", 1).Error; err != nil {
		return err
	}
	return nil
}

// UpdateTask 更新任务
func (d *taskDAO) UpdateTask(ctx context.Context, taskID int64, creatorID int64, updates map[string]interface{}) error {
	// 更新任务，确保只能更新自己创建的任务
	if err := d.DB(ctx).Model(&Task{}).
		Where("task_id = ? AND creator_id = ? AND deleted = 0", taskID, creatorID).
		Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

// ListTasks 获取任务列表
func (d *taskDAO) ListTasks(ctx context.Context, conditions map[string]interface{}, page, pageSize int64) ([]*Task, int64, error) {
	var tasks []*Task
	var total int64
	db := d.DB(ctx).Model(&Task{}).Where("deleted = 0")

	// 应用查询条件
	for key, value := range conditions {
		db = db.Where(key+" = ?", value)
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		db = db.Offset(int(offset)).Limit(int(pageSize))
	}

	// 执行查询
	if err := db.Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// GetTaskByIDAndCreatorID 获取指定任务
func (d *taskDAO) GetTaskByIDAndCreatorID(ctx context.Context, taskID int64, creatorID int64) (*Task, error) {
	var task *Task
	if err := d.DB(ctx).Where("task_id = ? AND creator_id = ? AND deleted = 0", taskID, creatorID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return task, nil
}

// GetTasksByIDs 通过id 查询任务列表
func (d *taskDAO) GetTasksByIDs(ctx context.Context, taskIDs []int64) ([]*Task, error) {
	var tasks []*Task
	err := d.DB(ctx).Where("task_id IN ?", taskIDs).Find(&tasks).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return tasks, err
}

// GetUserTasks 获取指定用户创建的全部任务列表
func (d *taskDAO) GetUserTasks(ctx context.Context, reqs *dto.TaskAssignListQuery, pageInfo *consts.DBPageInfo) ([]*Task, error) {
	var tasks []*Task
	pageInfo.Check()
	if err := d.DB(ctx).Where("creator_id = ? AND school_id = ? AND deleted = 0", reqs.TeacherID, reqs.SchoolID).
		Find(&tasks).Limit(int(pageInfo.Limit)).Offset(int((pageInfo.Page - 1) * pageInfo.Limit)).Order("create_time DESC").Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetLatestTask 获取指定教师每种任务类型的最新任务，如果有限定科目 id，则只查询指定科目的任务
func (d *taskDAO) GetLatestTask(ctx context.Context, teacherID, schoolID, subjectID int64) ([]*Task, error) {
	var tasks []*Task
	// 使用子查询获取每种任务类型的最新任务
	subQuery := d.DB(ctx).Model(&Task{}).
		Select("task_type, MAX(create_time) as max_create_time").
		Where("creator_id = ? AND school_id = ? AND deleted = 0", teacherID, schoolID)

	if subjectID > 0 {
		subQuery = subQuery.Where("subject = ?", subjectID)
	}

	subQuery = subQuery.Group("task_type")

	mainQuery := d.DB(ctx).Model(&Task{}).
		Joins("INNER JOIN (?) as latest ON tbl_task.task_type = latest.task_type AND tbl_task.create_time = latest.max_create_time", subQuery).
		Where("tbl_task.creator_id = ? AND tbl_task.school_id = ? AND tbl_task.deleted = 0", teacherID, schoolID)

	if subjectID > 0 {
		mainQuery = mainQuery.Where("tbl_task.subject = ?", subjectID)
	}

	if err := mainQuery.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

type TaskAndTaskAssign struct {
	Task
	StartTime int64 `gorm:"column:start_time" json:"startTime"` // 任务开始时间
	Deadline  int64 `gorm:"column:deadline" json:"deadline"`    // 任务截止时间
}

// GetStudentTaskList 获取指定学生的任务列表
func (d *taskDAO) GetStudentTaskList(ctx context.Context, req *dto.StudentTaskListQuery) ([]*TaskAndTaskAssign, int64, error) {
	var tasks []*TaskAndTaskAssign
	var total int64
	sql := d.DB(ctx).Model(&Task{}).
		Joins("INNER JOIN tbl_task_assign ON tbl_task.task_id = tbl_task_assign.task_id").
		Joins("INNER JOIN tbl_task_student ON tbl_task_assign.assign_id = tbl_task_student.assign_id").
		Where("tbl_task.deleted = 0 AND tbl_task_assign.deleted = 0").
		Where("tbl_task_student.student_id = ?", req.StudentID).
		Where("tbl_task.subject = ?", req.Subject).
		Select(`tbl_task.*, tbl_task_assign.start_time as "start_time", tbl_task_assign.deadline as "deadline"`)

	// tbl_task 可选参数
	if req.TaskType != 0 {
		sql = sql.Where("tbl_task.task_type = ?", req.TaskType)
	}
	if req.TaskSubType != 0 {
		sql = sql.Where("tbl_task.task_sub_type = ?", req.TaskSubType)
	}

	// tbl_task_assign 可选参数
	if req.StartTimeFrom > 0 {
		sql = sql.Where("tbl_task_assign.start_time >= ?", req.StartTimeFrom)
	}
	if req.StartTimeTo > 0 {
		sql = sql.Where("tbl_task_assign.start_time < ?", req.StartTimeTo)
	}
	if req.DeadlineFrom > 0 {
		sql = sql.Where("tbl_task_assign.deadline >= ?", req.DeadlineFrom)
	}
	if req.DeadlineTo > 0 {
		sql = sql.Where("tbl_task_assign.deadline < ?", req.DeadlineTo)
	}

	// 获取总数
	if err := sql.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	// 分页查询
	query := sql.Limit(int(req.PageSize)).Offset(int((req.Page - 1) * req.PageSize))

	// 排序
	switch req.SortBy {
	case "task_id":
		query = query.Order("tbl_task.task_id " + string(req.SortType))
	case "start_time":
		query = query.Order("tbl_task_assign.start_time " + string(req.SortType))
	case "deadline":
		query = query.Order("tbl_task_assign.deadline " + string(req.SortType))
	default:
		query = query.Order("tbl_task.task_id DESC")
	}

	if err := query.Find(&tasks).Error; err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}
