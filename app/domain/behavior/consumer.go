package behavior

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"gil_teacher/app/conf"
	"gil_teacher/app/consts"
	"gil_teacher/app/core/kafka"
	clogger "gil_teacher/app/core/logger"
	"gil_teacher/app/model/dto"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

// 用户行为数据消费处理
type BehaviorConsumer struct {
	kafkaConf *conf.Kafka
	logger    *clogger.ContextLogger
}

func NewBehaviorConsumer(kafkaConf *conf.Kafka, logger *clogger.ContextLogger) *BehaviorConsumer {
	return &BehaviorConsumer{
		kafkaConf: kafkaConf,
		logger:    logger,
	}
}

// 从 topic 批量消费消息，并调用 handler 处理
func (c *BehaviorConsumer) Consume(ctx context.Context, handler *BehaviorHandler) {
	// 添加日志确认实际使用的配置
	c.logger.Info(ctx, "Kafka 配置信息: broker=%s, group=%s, topics=%v",
		c.kafkaConf.Brokers,
		consts.KafkaGroupBehavior,
		consts.KafkaTopicBehaviors)

	consumerGroupHandlerImpl := &kafka.ConsumerGroupHandlerImpl{
		Group:       consts.KafkaGroupBehavior,
		Topics:      consts.KafkaTopicBehaviors,
		BatchSize:   c.kafkaConf.Consumer.BatchSize,
		BatchTime:   c.kafkaConf.Consumer.BatchTime * time.Second,
		SessionTime: c.kafkaConf.Consumer.SessionTime * time.Second,
		ProcMsgList: handler.HandleMessage,
		Log:         c.logger,
	}
	kafka.ConsumeKafkaMsgInSession(ctx, c.kafkaConf, consumerGroupHandlerImpl)
}

// HandleMessage 实现 kafka.MessageHandler 接口
// 不能退出，作为长连接消费
func (h *BehaviorHandler) HandleMessage(msgs []*sarama.ConsumerMessage) error {
	// 有消息才处理
	if len(msgs) == 0 {
		return nil
	}

	ctx := context.Background()
	// 1. 添加处理开始日志
	startTime := time.Now()
	h.logger.Debug(ctx, "开始处理消息批次，数量: %d", len(msgs))

	// 目前 3 种消息都在一个队列中，等后续拆分
	// 2. 对消息进行分类预处理
	messagesByType := make(map[string][]json.RawMessage)
	for _, msg := range msgs {
		behaviorMessage, err := DecodeBehaviorMessage(msg.Value)
		if err != nil {
			h.logger.Error(ctx, "解析消息失败, error:%v, msg:%+v", err, msg)
			continue
		}

		switch behaviorMessage.Type {
		case consts.MessageTypeTeacherBehavior:
			messagesByType[consts.KafkaTopicTeacherBehavior] = append(messagesByType[consts.KafkaTopicTeacherBehavior], behaviorMessage.Content)
		//case consts.MessageTypeStudentBehavior:
		//	messagesByType[consts.KafkaTopicStudentBehavior] = append(messagesByType[consts.KafkaTopicStudentBehavior], behaviorMessage.Content)
		case consts.MessageTypeCommunication:
			messagesByType[consts.KafkaTopicCommunication] = append(messagesByType[consts.KafkaTopicCommunication], behaviorMessage.Content)
		}
	}

	// 3. 并发处理不同类型的消息
	errChan := make(chan error, 3)
	var wg sync.WaitGroup
	for topic, messages := range messagesByType {
		wg.Add(1)
		go func(topic string, msgs []json.RawMessage) {
			defer wg.Done()
			var err error
			switch topic {
			case consts.KafkaTopicTeacherBehavior:
				err = h.processTeacherBehaviors(ctx, msgs)
				if err != nil {
					h.logger.Error(ctx, "处理教师行为失败, error:%v, topic:%s, messages:%+v", err, topic, msgs)
					errChan <- errors.Wrap(err, "处理教师行为失败")
				}
			//case consts.KafkaTopicStudentBehavior:
			//	err = h.processStudentBehaviors(ctx, msgs)
			//	if err != nil {
			//		h.logger.Error(ctx, "处理学生行为失败, error:%v, topic:%s, messages:%+v", err, topic, msgs)
			//		errChan <- errors.Wrap(err, "处理学生行为失败")
			//	}
			case consts.KafkaTopicCommunication:
				err = h.processCommunications(ctx, msgs)
				if err != nil {
					h.logger.Error(ctx, "处理沟通记录失败, error:%v, topic:%s, messages:%+v", err, topic, msgs)
					errChan <- errors.Wrap(err, "处理沟通记录失败")
				}
			default:
				h.logger.Error(ctx, "未知消息类型, topic:%s, messages:%+v", topic, messages)
			}
		}(topic, messages)
	}

	// 等待所有处理完成
	wg.Wait()
	close(errChan)

	// 收集错误
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	// 有消费才打印日志
	if len(msgs) > 0 {
		h.logger.Debug(ctx, "消息批次处理完成，耗时: %v, 成功数: %d, 错误数: %d",
			time.Since(startTime), len(msgs)-len(errs), len(errs))
	}
	return nil
}

func (h *BehaviorHandler) processTeacherBehaviors(ctx context.Context, content []json.RawMessage) error {
	behaviors := make([]*dto.TeacherBehaviorDTO, 0, len(content))
	for _, c := range content {
		var behavior dto.TeacherBehaviorDTO
		if err := json.Unmarshal(c, &behavior); err != nil {
			return errors.Wrap(err, "解析教师行为数据失败")
		}
		h.logger.Debug(ctx, "解析教师行为数据成功, behavior:%+v", behavior)

		if err := h.validateTeacherBehavior(&behavior); err != nil {
			h.logger.Error(ctx, "教师行为数据验证失败, error:%+v, behavior:%+v", err, behavior)
			continue
		}

		behaviors = append(behaviors, &behavior)
	}

	return h.behaviorDAO.SaveTeacherBehavior(ctx, behaviors)
}

func (h *BehaviorHandler) processStudentBehaviors(ctx context.Context, content []json.RawMessage) error {
	behaviors := make([]*dto.StudentBehaviorDTO, 0, len(content))
	for _, c := range content {
		var behavior dto.StudentBehaviorDTO
		if err := json.Unmarshal(c, &behavior); err != nil {
			return errors.Wrap(err, "解析学生行为数据失败")
		}

		if err := h.validateStudentBehavior(&behavior); err != nil {
			h.logger.Error(ctx, "学生行为数据验证失败, error:%v, behavior:%+v", err, behavior)
			continue
		}
		behaviors = append(behaviors, &behavior)
	}

	return h.behaviorDAO.SaveStudentBehavior(ctx, behaviors)
}

func (h *BehaviorHandler) processCommunications(ctx context.Context, content []json.RawMessage) error {
	communications := make([]*dto.CommunicationMessageDTO, 0, len(content))
	for _, c := range content {
		var communication dto.CommunicationMessageDTO
		if err := json.Unmarshal(c, &communication); err != nil {
			return errors.Wrap(err, "解析沟通记录失败")
		}
		if err := h.validateCommunication(&communication); err != nil {
			h.logger.Error(ctx, "沟通记录数据验证失败, error:%v, communication:%+v", err, communication)
			continue
		}

		// 查看会话是否关闭，关闭 5 min后不允许再提交
		session, err := h.GetCommunicationSession(ctx, communication.SessionID)
		if err != nil {
			h.logger.Error(ctx, "查询会话失败, error:%v, communication:%+v", err, communication)
			continue
		}
		if session.Closed && time.Since(*session.EndTime) > 5*time.Minute {
			h.logger.Error(ctx, "会话已关闭，不能再提交消息, communication:%+v", communication)
			continue
		}

		communications = append(communications, &communication)
	}

	return h.behaviorDAO.SaveCommunication(ctx, nil, communications)
}
