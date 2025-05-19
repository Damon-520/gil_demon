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

type CommunicationMessageDao struct {
	db     *dao.ClickHouseRWClient
	logger *clogger.ContextLogger
}

func newCommunicationMessageDao(db *dao.ClickHouseRWClient, logger *clogger.ContextLogger) *CommunicationMessageDao {
	return &CommunicationMessageDao{
		db:     db,
		logger: logger,
	}
}

// CommunicationMessage 沟通消息表结构
type CommunicationMessage struct {
	MessageID      string    `ch:"message_id"`      // 消息ID
	SessionID      string    `ch:"session_id"`      // 会话ID
	UserID         uint64    `ch:"user_id"`         // 用户ID
	UserType       string    `ch:"user_type"`       // 用户类型 Enum8('ai' = 3, 'student' = 1, 'teacher' = 2)
	MessageContent string    `ch:"message_content"` // 消息内容
	MessageType    string    `ch:"message_type"`    // 消息类型
	AnswerTo       string    `ch:"answer_to"`       // 回复消息ID，明确回复某条消息时才记录
	CreatedAt      time.Time `ch:"created_at"`      // 创建时间
}

func (m *CommunicationMessage) TableName() string {
	return "tbl_communication_messages"
}

// 给模型数据生成主键 id，方便插入
func (m *CommunicationMessage) GenerateID(ctx context.Context) string {
	if m.MessageID == "" {
		uuid := idtools.GetUUID()
		m.MessageID = uuid
	}
	return m.MessageID
}

func (m *CommunicationMessageDao) DB(ctx context.Context) *dao.ClickHouseRWClient {
	return m.db.Model(&CommunicationMessage{})
}

// 批量保存
func (m *CommunicationMessageDao) SaveCommunicationMessages(ctx context.Context, messages []*CommunicationMessage) error {
	_, err := m.DB(ctx).BatchInsert(ctx, messages)
	return err
}

// 获取某个session 的全部参与用户列表和类型
func (m *CommunicationMessageDao) GetSessionParticipants(ctx context.Context, sessionID string, pageInfo *consts.DBPageInfo) ([]*CommunicationMessage, error) {
	records := make([]*CommunicationMessage, 0)
	err := m.DB(ctx).FindAllWithFields(ctx, &records, []string{"message_id", "session_id", "user_id", "user_type"}, map[string]any{"session_id": sessionID}, pageInfo)
	if err != nil {
		return nil, errors.Wrap(err, "query session participants failed")
	}
	return records, nil
}

// 获取指定会话的全部消息
func (m *CommunicationMessageDao) GetSessionMessages(ctx context.Context, sessionID string, pageInfo *consts.DBPageInfo) ([]*CommunicationMessage, error) {
	records := make([]*CommunicationMessage, 0)
	err := m.DB(ctx).FindAll(ctx, &records, map[string]any{"session_id": sessionID}, pageInfo)
	if err != nil {
		return nil, errors.Wrap(err, "query session messages failed")
	}
	return records, nil
}

// 获取指定课堂的全部消息
func (m *CommunicationMessageDao) GetClassroomMessages(ctx context.Context, classroomID string, pageInfo *consts.DBPageInfo) ([]*CommunicationMessage, error) {
	records := make([]*CommunicationMessage, 0)
	err := m.DB(ctx).FindAll(ctx, &records, map[string]any{"classroom_id": classroomID}, pageInfo)
	if err != nil {
		return nil, errors.Wrap(err, "query classroom messages failed")
	}
	return records, nil
}

// 按消息 id 批量查询数据
func (m *CommunicationMessageDao) GetMessagesByIDs(ctx context.Context, messageIDs []string) ([]*CommunicationMessage, error) {
	if len(messageIDs) == 0 {
		return nil, nil
	}

	// 预分配一定数量的结果对象
	messages := make([]*CommunicationMessage, len(messageIDs))
	err := m.DB(ctx).FindAll(ctx, &messages, map[string]any{"message_id": messageIDs}, nil)
	if err != nil {
		return nil, errors.Wrap(err, "query messages by ids failed")
	}
	return messages, nil
}

// 检查消息 id 是否存在
func (m *CommunicationMessageDao) CheckMessageIDsExist(ctx context.Context, messageIDs []string) (bool, error) {
	if len(messageIDs) == 0 {
		return true, nil
	}

	existRecords := make([]*CommunicationMessage, 0)
	err := m.DB(ctx).FindAllWithFields(ctx, &existRecords, []string{"message_id"}, map[string]any{"message_id": messageIDs}, nil)
	if err != nil {
		return false, errors.Wrap(err, "query messages by ids failed")
	}
	return len(existRecords) == len(messageIDs), nil
}
