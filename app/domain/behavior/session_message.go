package behavior

import (
	"context"
	"math"

	"gil_teacher/app/consts"
	clogger "gil_teacher/app/core/logger"
	"gil_teacher/app/dao"
	behaviorDao "gil_teacher/app/dao/behavior"
	"gil_teacher/app/model/dto"

	"github.com/pkg/errors"
)

// 会话消息，包括创建会话、关闭会话、发送消息、接收消息、关闭会话、标记已读、获取未读消息数量、获取未读消息列表
type SessionMessageHandler struct {
	behaviorDAO behaviorDao.BehaviorDAO
	redisClient *dao.ApiRdbClient
	logger      *clogger.ContextLogger
}

func NewSessionMessageHandler(
	behaviorDAO behaviorDao.BehaviorDAO,
	redisClient *dao.ApiRdbClient,
	logger *clogger.ContextLogger,
) *SessionMessageHandler {
	return &SessionMessageHandler{
		behaviorDAO: behaviorDAO,
		redisClient: redisClient,
		logger:      logger,
	}
}

// redis 记录每个会话的消息，有序集合，key 为会话id，value 为消息id，score 为消息时间戳
func (h *SessionMessageHandler) SaveSessionMessage(ctx context.Context, sessionID string, messageID string, timestamp int64) error {
	key := consts.GetSessionMessageKey(sessionID)
	err := h.redisClient.ZAdd(ctx, key, float64(timestamp), messageID, consts.SessionMessageExpire)
	if err != nil {
		h.logger.Error(ctx, "SaveSessionMessage, redis zadd error: %v", err)
		return err
	}
	return nil
}

// 获取会话的全部消息 id 列表
func (h *SessionMessageHandler) GetAllMessageIDs(ctx context.Context, userID int64, sessionID string) ([]string, error) {
	_, err := h.checkSessionPermission(ctx, userID, sessionID)
	if err != nil {
		h.logger.Error(ctx, "GetAllMessageIDs, check session permission error: %v", err)
		return nil, err
	}

	var messageIDs []string
	key := consts.GetSessionMessageKey(sessionID)
	exists, err := h.redisClient.ZRange(ctx, key, 0, -1, &messageIDs)
	if err != nil {
		h.logger.Error(ctx, "GetAllMessageIDs, redis zrange error: %v", err)
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	return messageIDs, nil
}

// 标记用户消息已读
func (h *SessionMessageHandler) MarkMessageAsRead(ctx context.Context, userID int64, sessionID string, messageID string) error {
	_, err := h.checkSessionPermission(ctx, userID, sessionID)
	if err != nil {
		h.logger.Error(ctx, "MarkMessageAsRead, check session permission error: %v", err)
		return err
	}
	key := consts.GetSessionUserLastReadMessageKey(sessionID, userID)
	err = h.redisClient.Set(ctx, key, messageID, consts.UserLastReadMessageExpire)
	if err != nil {
		h.logger.Error(ctx, "MarkMessageAsRead, redis set error: %v", err)
		return err
	}
	return nil
}

// 获取用户最后已读消息时间戳
func (h *SessionMessageHandler) userSessionLastMessageTimestamp(ctx context.Context, userID int64, sessionID string) (float64, error) {
	_, err := h.checkSessionPermission(ctx, userID, sessionID)
	if err != nil {
		h.logger.Error(ctx, "[GetUserLastReadMessageTimestamp] check session permission error: %v", err)
		return 0, err
	}

	lastMessageID := ""
	lastReadMessageKey := consts.GetSessionUserLastReadMessageKey(sessionID, userID) // 用户最后已读消息 key
	exists, err := h.redisClient.Get(ctx, lastReadMessageKey, &lastMessageID)
	if err != nil {
		h.logger.Error(ctx, "[GetUserLastReadMessageTimestamp] redis get error: %v", err)
		return 0, err
	}

	if !exists {
		return 0, nil
	}

	messageSessionKey := consts.GetSessionMessageKey(sessionID)
	lastReadMessageTimestamp := 0.0
	_, err = h.redisClient.ZScore(ctx, messageSessionKey, lastMessageID, &lastReadMessageTimestamp)
	if err != nil {
		h.logger.Error(ctx, "[GetUserLastReadMessageTimestamp] redis zscore error: %v", err)
		return 0, err
	}

	return lastReadMessageTimestamp, nil
}

// 获取用户指定会话的未读消息数量
func (h *SessionMessageHandler) GetUnreadMessageCount(ctx context.Context, userID int64, sessionID string) (int64, error) {
	_, err := h.checkSessionPermission(ctx, userID, sessionID)
	if err != nil {
		h.logger.Error(ctx, "GetUnreadMessageCount, check session permission error: %v", err)
		return 0, err
	}

	userLastMessageTimestamp, err := h.userSessionLastMessageTimestamp(ctx, userID, sessionID)
	if err != nil {
		h.logger.Error(ctx, "GetUnreadMessageCount, get user last read message timestamp error: %v", err)
		return 0, err
	}

	var count int64
	messageSessionKey := consts.GetSessionMessageKey(sessionID)
	exists, err := h.redisClient.ZCount(ctx, messageSessionKey, userLastMessageTimestamp, math.MaxFloat64, &count)
	if err != nil {
		h.logger.Error(ctx, "GetUnreadMessageCount, redis zcount error: %v", err)
		return 0, err
	}
	if !exists {
		return 0, errors.New("会话不存在")
	}
	return count, nil
}

// 获取用户指定会话的未读消息列表
func (h *SessionMessageHandler) GetUnreadMessageList(ctx context.Context, userID int64, sessionID string) ([]*dto.CommunicationMessageDTO, error) {
	_, err := h.checkSessionPermission(ctx, userID, sessionID)
	if err != nil {
		h.logger.Error(ctx, "[GetUnreadMessageList] check session permission error: %v", err)
		return nil, err
	}

	userLastMessageTimestamp, err := h.userSessionLastMessageTimestamp(ctx, userID, sessionID)
	if err != nil {
		h.logger.Error(ctx, "[GetUnreadMessageList] get user last message timestamp error: %v", err)
		return nil, err
	}
	var messageIDs []string
	sessionKey := consts.GetSessionMessageKey(sessionID)
	exists, err := h.redisClient.ZRangeByScore(ctx, sessionKey, userLastMessageTimestamp, math.MaxFloat64, &messageIDs)
	if err != nil {
		h.logger.Error(ctx, "[GetUnreadMessageList] redis zrangebyscore error: %v", err)
		return nil, err
	}
	if !exists {
		return nil, errors.New("会话不存在")
	}

	messageList, err := h.behaviorDAO.GetCommunicationSessionMessagesByIDs(ctx, sessionID, messageIDs)
	if err != nil {
		h.logger.Error(ctx, "[GetUnreadMessageList] get communication session messages error: %v", err)
		return nil, err
	}

	return messageList, nil
}

// 检查用户对会话的权限
// 1. 会话是否存在
// 2. 用户和会话创建者在同一班级：教师不做限制，因为可能会代课，学生需要限制
func (h *SessionMessageHandler) checkSessionPermission(ctx context.Context, userID int64, sessionID string) (*dto.CommunicationSessionDTO, error) {
	// 先全部绕过检查
	if true {
		return nil, nil
	}

	// 缓存处理
	// sessionKey := consts.GetCommunicationSessionKey(sessionID)

	session, err := h.behaviorDAO.GetCommunicationSession(ctx, sessionID)
	if err != nil {
		h.logger.Error(ctx, "checkSessionPermission, get communication session error: %v", err)
		return nil, err
	}

	if session == nil {
		h.logger.Error(ctx, "checkSessionPermission, session not found")
		return nil, errors.New("session not found")
	}

	// 教师不做限制，学生需要限制
	if session.UserID == uint64(userID) {
		return session, nil
	}

	// 检查用户和会话创建者在同一班级 TODO
	return session, nil
}
