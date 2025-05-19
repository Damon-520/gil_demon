package behavior

import (
	"context"
	"encoding/json"
	"time"

	"gil_teacher/app/consts"
	"gil_teacher/app/core/kafka"
	clogger "gil_teacher/app/core/logger"
	"gil_teacher/app/model/api"
	"gil_teacher/app/model/dto"
	"gil_teacher/app/utils"
	"gil_teacher/app/utils/idtools"

	"github.com/pkg/errors"
)

// BehaviorProducer 行为服务
type BehaviorProducer struct {
	handler     *BehaviorHandler
	kafkaClient *kafka.KafkaProducerClient
	logger      *clogger.ContextLogger
}

// NewBehaviorProducer 创建行为服务
func NewBehaviorProducer(
	handler *BehaviorHandler,
	kafkaClient *kafka.KafkaProducerClient,
	logger *clogger.ContextLogger) *BehaviorProducer {
	return &BehaviorProducer{
		handler:     handler,
		kafkaClient: kafkaClient,
		logger:      logger,
	}
}

// RecordTeacherBehavior 记录教师行为
func (h *BehaviorProducer) RecordTeacherBehavior(ctx context.Context, req *api.TeacherBehaviorRequest) error {
	// 关联会话的，需要检查会话是否存在
	if req.CommunicationSessionID != "" {
		session, err := h.handler.behaviorDAO.GetCommunicationSession(ctx, req.CommunicationSessionID)
		if err != nil {
			return errors.Wrap(err, "查询会话失败")
		}
		if session == nil {
			return errors.New("会话不存在")
		}
	}

	behavior := &dto.TeacherBehaviorDTO{
		SchoolID:               req.SchoolID,
		ClassID:                req.ClassID,
		ClassroomID:            utils.Ptr(req.ClassroomID),
		TeacherID:              req.TeacherID,
		BehaviorType:           consts.BehaviorType(req.BehaviorType),
		CommunicationSessionID: utils.Ptr(req.CommunicationSessionID),
		Context:                req.Context,
		CreateTime:             time.Now(),
	}
	return h.SendTeacherBehavior(ctx, behavior)
}

// RecordStudentBehavior 记录学生行为
func (h *BehaviorProducer) RecordStudentBehavior(ctx context.Context, req *api.StudentBehaviorRequest) error {
	// 关联会话的，需要检查会话是否存在
	if req.CommunicationSessionID != "" {
		session, err := h.handler.behaviorDAO.GetCommunicationSession(ctx, req.CommunicationSessionID)
		if err != nil {
			return errors.Wrap(err, "查询会话失败")
		}
		if session == nil {
			return errors.New("会话不存在")
		}
	}

	behavior := &dto.StudentBehaviorDTO{
		SchoolID:               req.SchoolID,
		ClassID:                req.ClassID,
		ClassroomID:            utils.Ptr(req.ClassroomID),
		StudentID:              req.StudentID,
		BehaviorType:           consts.BehaviorType(req.BehaviorType),
		CommunicationSessionID: utils.Ptr(req.CommunicationSessionID),
		Context:                req.Context,
		CreateTime:             time.Now(),
	}
	return h.SendStudentBehavior(ctx, behavior)
}

// RecordCommunicationMessage 记录会话内容
func (h *BehaviorProducer) RecordCommunicationMessage(ctx context.Context, req *api.SaveMessageRequest) (string, error) {
	// 检查会话是否存在
	exists, err := h.handler.checkCommunicationSession(ctx, req.SessionID)
	if err != nil {
		return "", errors.Wrap(err, "检查会话失败")
	}
	if !exists {
		return "", errors.New("会话不存在")
	}

	// 参数检查 TODO

	message := &dto.CommunicationMessageDTO{
		MessageID:      idtools.GetUUID(),
		SessionID:      req.SessionID,
		UserID:         req.UserID,
		UserType:       req.UserType,
		MessageContent: req.MessageContent,
		MessageType:    req.MessageType,
		AnswerTo:       req.AnswerTo,
		CreatedAt:      time.Now(),
	}
	// 会话写 mq，避免高峰期数据库压力
	return message.MessageID, h.SendCommunicationMessage(ctx, message)
}

// SendTeacherBehavior 发送教师行为
func (s *BehaviorProducer) SendTeacherBehavior(ctx context.Context, behavior *dto.TeacherBehaviorDTO) error {
	content, err := json.Marshal(behavior)
	if err != nil {
		s.logger.Error(ctx, "序列化教师行为失败, error:%v, behavior:%v", err, behavior)
		return errors.Wrap(err, "序列化教师行为失败")
	}

	msg := &BehaviorMessageQueue{
		Type:      consts.MessageTypeTeacherBehavior,
		Content:   content,
		Timestamp: time.Now(),
		Version:   "1.0",
	}

	s.logger.Info(ctx, "发送教师行为, behavior:%+v", behavior)
	return s.kafkaClient.ProduceMsgToKafka(ctx, consts.KafkaTopicTeacherBehavior, string(msg.Encode()))
}

// SendStudentBehavior 发送学生行为
func (s *BehaviorProducer) SendStudentBehavior(ctx context.Context, behavior *dto.StudentBehaviorDTO) error {
	content, err := json.Marshal(behavior)
	if err != nil {
		s.logger.Error(ctx, "序列化学生行为失败, error:%v, behavior:%+v", err, behavior)
		return errors.Wrap(err, "序列化学生行为失败")
	}

	msg := &BehaviorMessageQueue{
		Type:      consts.MessageTypeStudentBehavior,
		Content:   content,
		Timestamp: time.Now(),
		Version:   "1.0",
	}

	return s.kafkaClient.ProduceMsgToKafka(ctx, consts.KafkaTopicTeacherBehavior, string(msg.Encode()))
}

// SendCommunicationMessage 发送沟通会话
func (s *BehaviorProducer) SendCommunicationMessage(ctx context.Context, message *dto.CommunicationMessageDTO) error {
	content, err := json.Marshal(message)
	if err != nil {
		s.logger.Error(ctx, "序列化沟通会话失败, error:%v, message:%+v", err, message)
		return errors.Wrap(err, "序列化沟通会话失败")
	}

	msg := &BehaviorMessageQueue{
		Type:      consts.MessageTypeCommunication,
		Content:   content,
		Timestamp: time.Now(),
		Version:   "1.0",
	}

	return s.kafkaClient.ProduceMsgToKafka(ctx, consts.KafkaTopicTeacherBehavior, string(msg.Encode()))
}
