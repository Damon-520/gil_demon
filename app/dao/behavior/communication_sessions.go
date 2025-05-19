package behavior

import (
	"context"
	"time"

	"gil_teacher/app/consts"
	clogger "gil_teacher/app/core/logger"
	"gil_teacher/app/dao"
	"gil_teacher/app/utils/idtools"

	"github.com/pkg/errors"
)

type CommunicationSessionDao struct {
	db     *dao.ClickHouseRWClient
	logger *clogger.ContextLogger
}

func newCommunicationSessionDao(db *dao.ClickHouseRWClient, logger *clogger.ContextLogger) *CommunicationSessionDao {
	return &CommunicationSessionDao{
		db:     db,
		logger: logger,
	}
}

// CommunicationSession 沟通会话表结构
type CommunicationSession struct {
	SessionID    string     `ch:"session_id"`
	UserID       uint64     `ch:"user_id"`
	UserType     string     `ch:"user_type"`
	SchoolID     uint64     `ch:"school_id"`
	CourseID     *uint64    `ch:"course_id"`
	ClassroomID  *uint64    `ch:"classroom_id"`
	SessionType  string     `ch:"session_type"`
	TargetID     *string    `ch:"target_id"` // message 表的 message_id
	Closed       bool       `ch:"closed"`
	Participants string     `ch:"participants"`
	StartTime    time.Time  `ch:"start_time"`
	EndTime      *time.Time `ch:"end_time"`
}

func (m *CommunicationSession) TableName() string {
	return "tbl_communication_sessions"
}

// 给模型数据生成主键 id，方便插入
func (m *CommunicationSession) GenerateID(ctx context.Context) string {
	if m.SessionID == "" {
		uuid := idtools.GetUUID()
		m.SessionID = uuid
	}
	return m.SessionID
}

func (m *CommunicationSessionDao) DB(ctx context.Context) *dao.ClickHouseRWClient {
	return m.db.Model(&CommunicationSession{})
}

// 批量保存新会话
func (m *CommunicationSessionDao) SaveCommunicationSessions(ctx context.Context, sessions []*CommunicationSession) error {
	_, err := m.DB(ctx).BatchInsert(ctx, sessions)
	return err
}

// 获取用户在某堂课的全部会话
func (m *CommunicationSessionDao) GetUserCourseSessions(ctx context.Context, userID uint64, courseID, classroomID uint64, pageInfo *consts.DBPageInfo) ([]*CommunicationSession, error) {
	dests := make([]*CommunicationSession, 0)
	err := m.DB(ctx).FindAll(ctx, &dests, map[string]any{"user_id": userID, "course_id": courseID, "classroom_id": classroomID}, pageInfo)
	if err != nil {
		return nil, errors.Wrap(err, "query communication sessions failed")
	}
	return dests, nil
}

// 批量查询会话
func (m *CommunicationSessionDao) GetSessionsByIDs(ctx context.Context, sessionIDs []string) ([]*CommunicationSession, error) {
	if len(sessionIDs) == 0 {
		return nil, nil
	}

	dests := make([]*CommunicationSession, 0)
	err := m.DB(ctx).FindAll(ctx, &dests, map[string]any{"session_id": sessionIDs}, nil)
	if err != nil {
		return nil, errors.Wrap(err, "query communication sessions failed")
	}
	return dests, nil
}

// 查询会话 id 列表是否存在
func (m *CommunicationSessionDao) CheckSessionIDsExist(ctx context.Context, sessionIDs []string) (bool, error) {
	if len(sessionIDs) == 0 {
		return true, nil
	}

	existRecords := make([]*CommunicationSession, 0)
	err := m.DB(ctx).FindAllWithFields(ctx, &existRecords, []string{"session_id"}, map[string]any{"session_id": sessionIDs}, nil)
	if err != nil {
		return false, errors.Wrap(err, "query communication sessions failed")
	}
	return len(existRecords) == len(sessionIDs), nil
}

// 更新会话
func (m *CommunicationSessionDao) UpdateCommunicationSession(ctx context.Context, update map[string]any, where map[string]any) error {
	return m.DB(ctx).Update(ctx, update, where)
}
