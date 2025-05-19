package behavior

import (
	"context"
	"gil_teacher/app/model/api"
	"gil_teacher/app/model/dto"
)

// Repository 行为域接口
type Repository interface {
	// 记录教师行为
	RecordTeacherBehavior(ctx context.Context, req *api.TeacherBehaviorRequest) error
	// 记录学生行为
	RecordStudentBehavior(ctx context.Context, req *api.StudentBehaviorRequest) error
	// 用户创建一个会话，返回会话 id
	OpenCommunicationSession(ctx context.Context, req *api.OpenSessionRequest) (string, error)
	// 记录会话内容
	SaveCommunicationMessage(ctx context.Context, req *api.SaveMessageRequest) error
	// 关闭会话
	CloseCommunicationSession(ctx context.Context, req *api.CloseSessionRequest) error
	// 查询指定会话的全部消息
	GetCommunicationSessionMessages(ctx context.Context, sessionID string) ([]*dto.CommunicationMessageDTO, error)
	// 查询指定课堂的全部消息
	GetClassroomMessages(ctx context.Context, userID string, classroomID string) ([]*dto.CommunicationMessageDTO, error)
	// 标记用户消息已读
	// MarkMessageAsRead(ctx context.Context, userID string, sessionID string) error
	// // 获取用户指定会话的未读消息数量
	// GetUnreadMessageCount(ctx context.Context, userID string, sessionID string) (int64, error)
	// // 获取用户指定会话的未读消息列表
	// GetUnreadMessageList(ctx context.Context, userID string, sessionID string) ([]*dto.CommunicationMessageDTO, error)
	GetClassLatestBehaviors(ctx context.Context, classID uint64) ([]*dto.StudentLatestBehaviorDTO, error)
}

// 定义 session message 接口
type SessionMessageRepository interface {
	// 保存会话消息
	SaveSessionMessage(ctx context.Context, sessionID string, messageID string, timestamp int64) error
	// 获取会话的全部消息 id 列表
	GetAllMessageIDs(ctx context.Context, userID int64, sessionID string) ([]string, error)
	// 标记用户消息已读
	MarkMessageAsRead(ctx context.Context, userID int64, sessionID string, messageID string) error
	// 获取用户指定会话的未读消息数量
	GetUnreadMessageCount(ctx context.Context, userID int64, sessionID string) (int64, error)
	// 获取用户指定会话的未读消息列表
	GetUnreadMessageList(ctx context.Context, userID int64, sessionID string) ([]*dto.CommunicationMessageDTO, error)
}
