package behavior

import (
	"context"
	"time"

	"gil_teacher/app/consts"
	clogger "gil_teacher/app/core/logger"
	"gil_teacher/app/dao"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/utils/idtools"
)

type TeacherBehaviorDao struct {
	db     *dao.ClickHouseRWClient
	logger *clogger.ContextLogger
}

func newTeacherBehaviorDao(db *dao.ClickHouseRWClient, logger *clogger.ContextLogger) *TeacherBehaviorDao {
	return &TeacherBehaviorDao{
		db:     db,
		logger: logger,
	}
}

// TeacherBehavior 教师行为表结构
type TeacherBehavior struct {
	ID                     string    `ch:"id"` // uuid
	SchoolID               uint64    `ch:"school_id"`
	ClassID                uint64    `ch:"class_id"`
	ClassroomID            *uint64   `ch:"classroom_id"`
	TeacherID              uint64    `ch:"teacher_id"`
	BehaviorType           string    `ch:"behavior_type"`
	CommunicationSessionID *string   `ch:"communication_session_id"`
	LastMessageID          *string   `ch:"last_message_id"`
	Context                string    `ch:"context"`
	TaskID                 uint64    `ch:"task_id"`
	AssignID               uint64    `ch:"assign_id"`
	StudentID              uint64    `ch:"student_id"`
	CreateTime             time.Time `ch:"create_time"`
	UpdateTime             time.Time `ch:"update_time"`
}

func (m *TeacherBehavior) TableName() string {
	return "tbl_teacher_behavior_logs"
}

// 给模型数据生成主键 id，方便插入
func (m *TeacherBehavior) GenerateID(ctx context.Context) string {
	if m.ID == "" {
		uuid := idtools.GetUUID()
		m.ID = uuid
	}
	return m.ID
}

func (m *TeacherBehaviorDao) DB(ctx context.Context) *dao.ClickHouseRWClient {
	return m.db.Model(&TeacherBehavior{})
}

// 获取教师在某堂课的全部行为
func (m *TeacherBehaviorDao) GetTeacherCourseBehaviors(ctx context.Context, teacherID uint64, courseID, classroomID uint64, pageInfo *consts.DBPageInfo) ([]*dto.TeacherBehaviorDTO, error) {
	records := make([]*TeacherBehavior, 0)
	err := m.DB(ctx).FindAll(ctx, &records, map[string]any{"teacher_id": teacherID, "course_id": courseID, "classroom_id": classroomID}, pageInfo)
	if err != nil {
		return nil, err
	}

	var behaviors []*dto.TeacherBehaviorDTO
	for _, behavior := range records {
		behaviors = append(behaviors, &dto.TeacherBehaviorDTO{
			SchoolID:               behavior.SchoolID,
			ClassID:                behavior.ClassID,
			ClassroomID:            behavior.ClassroomID,
			TeacherID:              behavior.TeacherID,
			BehaviorType:           consts.BehaviorType(behavior.BehaviorType),
			CommunicationSessionID: behavior.CommunicationSessionID,
			Context:                behavior.Context,
			CreateTime:             behavior.CreateTime,
		})
	}

	return behaviors, nil
}

// 更新教师最新已读消息 id
func (m *TeacherBehaviorDao) UpdateTeacherLastMessageID(ctx context.Context, teacherID uint64, sessionID, messageID string) error {
	return m.DB(ctx).Update(ctx, map[string]any{"last_message_id": messageID}, map[string]any{"teacher_id": teacherID, "communication_session_id": sessionID})
}

func (m *TeacherBehaviorDao) SaveTeacherBehaviors(ctx context.Context, behaviors []*TeacherBehavior) error {
	_, err := m.DB(ctx).BatchInsert(ctx, behaviors)
	return err
}

// 统计指定任务作业下对学生的点赞和提醒次数
func (m *TeacherBehaviorDao) CountStudentTaskPraiseAndAttention(ctx context.Context, taskID, assignID uint64, studentIDs []uint64) ([]dao.CHGroupCountResult, error) {
	return m.DB(ctx).CountGroupBy(
		ctx,
		[]string{"student_id", "behavior_type"},
		map[string]any{
			"task_id":    taskID,
			"assign_id":  assignID,
			"student_id": studentIDs,
			"behavior_type": []consts.BehaviorType{
				consts.BehaviorTypeTaskPraise,
				consts.BehaviorTypeTaskAttention,
			},
		},
	)
}
