package kafka

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gil_teacher/app/conf"
	"gil_teacher/app/core/logger"

	"github.com/IBM/sarama"
	"github.com/go-kratos/kratos/v2/log"
)

type KafkaProducerClient struct {
	Producer sarama.SyncProducer
	log      *logger.ContextLogger
}

func newKafkaSyncProducer(conf *conf.Data) (sarama.SyncProducer, error) {
	if conf == nil {
		return nil, fmt.Errorf("kafka config is nil")
	}

	kafkaConf := conf.Kafka
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_6_2_0
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follow都确认
	saramaConfig.Producer.Partitioner = sarama.NewRandomPartitioner // 新选出一个partition
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.MaxMessageBytes = 5000000
	//目前阿里云暂时没有开启sasl，主要是考虑到成本问题
	//saramaConfig.Net.SASL.Enable = true
	//saramaConfig.Net.SASL.User = kafkaConf.SaslUsername
	//saramaConfig.Net.SASL.Password = kafkaConf.SaslPassword
	//saramaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext

	//broker 配置类似 your_host:9200
	brokers := strings.Split(kafkaConf.Brokers, ",")
	return sarama.NewSyncProducer(brokers, saramaConfig)
}

func NewKafkaProducerClient(ctx context.Context, conf *conf.Data, logger *logger.ContextLogger) (*KafkaProducerClient, func(), error) {
	if conf == nil {
		return nil, nil, fmt.Errorf("kafka config is nil")
	}

	// 方便兼容wire，暂时用conf
	kafkaConf := conf.Kafka
	producer, err := newKafkaSyncProducer(conf)
	if err != nil {
		logger.Error(ctx, "NewKafkaProducerClient kafkaConfig:%+v err: %v", kafkaConf, err)
		return nil, func() {}, err
	}

	cleanup := func() {
		if producer != nil {
			err = producer.Close()
			log.NewHelper(logger).Infof("closing kafka producer client failed. err:%+v", err)
		}

		log.NewHelper(logger).Infof("closing kafka producer client success")
	}

	return &KafkaProducerClient{
		Producer: producer,
		log:      logger,
	}, cleanup, nil
}

// 先支持单条发，我们的场景也基本都是单条发送，多条消费
func (kafkaProducerClient *KafkaProducerClient) ProduceMsgToKafka(ctx context.Context, topic string, msgVal string) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(msgVal),
		// key暂时不指定，走轮询策略，防止数据倾斜，如果有特定需求，未来新增按key区分partition
		// Key:   sarama.StringEncoder(msgVal),
	}

	partition, offset, err := kafkaProducerClient.Producer.SendMessage(msg)
	if err != nil {
		kafkaProducerClient.log.Error(ctx, "SendToKafka error:%+v", err)
		return err
	} else {
		kafkaProducerClient.log.Debug(ctx, "SendToKafka success: partition=%d, offset=%d, msg:%s", partition, offset, msgVal)
		return nil
	}
}

// ================================ 消费端代码 ====================================

func newKafkaConsumerClient(kafkaConf *conf.Kafka) (sarama.Client, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_6_2_0
	// 设置从最早的消息开始消费
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	//saramaConfig.Net.SASL.Enable = true
	//saramaConfig.Net.SASL.User = kafkaConf.SaslUsername
	//saramaConfig.Net.SASL.Password = kafkaConf.SaslPassword
	//saramaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	saramaConfig.Net.ReadTimeout = 8 * time.Second
	saramaConfig.Net.WriteTimeout = 8 * time.Second
	saramaConfig.Consumer.Group.Session.Timeout = 360 * time.Second
	saramaConfig.Consumer.Group.Rebalance.Timeout = 10 * time.Second
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = false
	saramaConfig.Consumer.Fetch.Default = 5 * 1024 * 1024

	brokers := strings.Split(kafkaConf.Brokers, ",")
	return sarama.NewClient(brokers, saramaConfig)
}

// 消费组配置
type ConsumerGroupHandlerImpl struct {
	Group       string        // consumer group name
	Topics      []string      // topic
	BatchTime   time.Duration // 5s
	BatchSize   int           // 100
	SessionTime time.Duration // 300s
	ProcMsgList func([]*sarama.ConsumerMessage) error
	Log         *logger.ContextLogger
}

// setup在sarama中，初始化只有一个协程执行一次，不像ConsumeClaim有partition数量的协程多次执行
func (handler *ConsumerGroupHandlerImpl) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (handler *ConsumerGroupHandlerImpl) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (handler *ConsumerGroupHandlerImpl) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := context.Background()
	handler.Log.Info(ctx, "开始消费分区: %d, 起始偏移量: %d, 高水位: %d",
		claim.Partition(), claim.InitialOffset(), claim.HighWaterMarkOffset())
	defer func() {
		if err := recover(); err != nil {
			handler.Log.Error(ctx, "ConsumeClaim get panic. error:%+v", err)
		}
	}()
	handler.Log.Info(ctx, "ConsumeClaim start")

	procTimer := time.NewTimer(handler.BatchTime)
	defer func() { procTimer.Stop() }()

	sessTimer := time.NewTimer(handler.SessionTime)
	defer func() { sessTimer.Stop() }()

	msgList := make([]*sarama.ConsumerMessage, 0)
	procMsg := func() error {
		defer func() {
			msgList = make([]*sarama.ConsumerMessage, 0)
			procTimer.Reset(handler.BatchTime)
			if err := recover(); err != nil {
				handler.Log.Error(ctx, "procMsg get panic. error:%+v", err)
			}
		}()

		if len(msgList) > 0 {
			msg := msgList[len(msgList)-1]
			handler.Log.Info(ctx, "offset:%d partition:%d time:%+v msgList len:%d ",
				msg.Offset, msg.Partition, msg.Timestamp, len(msgList))
		}

		err := handler.ProcMsgList(msgList)
		if err != nil {
			handler.Log.Error(ctx, "procMsg error:%+v", err)
			return err
		}

		for _, m := range msgList {
			session.MarkMessage(m, "")
		}

		session.Commit()

		return err
	}

	for {
		select {
		case msg := <-claim.Messages():
			//当有两个消费实例在启动时，可能一段时间读不到消息，返回为nil
			if msg == nil {
				time.Sleep(1 * time.Second)
				continue
			}

			msgList = append(msgList, msg)
			if len(msgList) >= handler.BatchSize {
				err := procMsg()
				if err != nil {
					handler.Log.Error(ctx, "procMsg error:%+v", err)
					return nil
				}
			}

		case <-procTimer.C:
			err := procMsg()
			if err != nil {
				handler.Log.Error(ctx, "procMsg error:%+v", err)
				return nil
			}
		case <-sessTimer.C:
			handler.Log.Info(ctx, "会话定时器触发，结束消费")
			//消费session生命周期30秒(saramaConfig.Consumer.Group.Session.Timeout)比sessTimer的10秒定时还多20秒
			//确保最后的一次发送有足够的时间处理完毕
			return procMsg()
		case <-session.Context().Done():
			// Should return when `session.Context()` is done.
			// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
			// https://github.com/Shopify/sarama/issues/1192
			// group:to-new-kafka-group topic:pcdn error:kafka: error while consuming pcdn/0: read tcp 222.187.225.58:33018->39.98.176.107:9092: i/o timeout
			//Log.Logger.Info("consumer session.context done")
			return procMsg()
		default: //防止select阻塞
			time.Sleep(1 * time.Second)
		}
	}
}

// 此函数不走wire注入，因为是长连接消费
func ConsumeKafkaMsgInSession(ctx context.Context, kafkaConf *conf.Kafka, handler *ConsumerGroupHandlerImpl) {
	consumerKafkaClient, err := newKafkaConsumerClient(kafkaConf)
	if err != nil {
		handler.Log.Error(ctx, "create consumer kafka client error:%+v", err)
		return
	}
	defer func() { _ = consumerKafkaClient.Close() }()

	consumerGroup, err := sarama.NewConsumerGroupFromClient(handler.Group, consumerKafkaClient)
	if err != nil {
		handler.Log.Error(ctx, "claim consumer group error:%+v", err)
		return
	}
	defer func() { _ = consumerGroup.Close() }()

	handler.Log.Info(ctx, "consumer group:"+handler.Group+" consume topic:"+strings.Join(handler.Topics, ",")+" start!")
	consumerTopics := handler.Topics
	if err := consumerGroup.Consume(ctx, consumerTopics, handler); err != nil {
		handler.Log.Error(ctx, "consume group error:%+v", err)
		return
	}

	select {
	case err = <-consumerGroup.Errors():
		if err != nil {
			if strings.Contains(err.Error(), "The requested offset is outside the range of offsets maintained by the server for the given topic/partition") {
				//此种重置方案并不靠谱，需要手动在kafka集群中用命令单独重置
				//若消费者给出的offset不正常，则需要重置消费.
				//这种情况多见于消费者不消费对应kafka数据，从而导致当前消费offset低于kafka最低的offset(比如kafka集群消息最多保存三天，结果落后了10天)。
				//handler.needResetConsume = true //强制重置开关
				/*
					if handler.needResetConsume {
						for topic, partitions := range session.Claims() {
							for _, partition := range partitions {
								Log.Logger.Info(fmt.Sprintf("consumer reset topic:%s partition:%d offset", topic, partition))
								session.ResetOffset(topic, partition, sarama.OffsetOldest, "")
							}
						}

						handler.needResetConsume = false
						Log.Logger.Info("consumer reset all topic partition offset")
					}

				*/
				//consumerGroupHandler.needResetConsume = true
			}
			handler.Log.Error(ctx, "consumer group partition offset error")

			//这个报警跟kafka的状态有关系，一旦报警了，大概率是kafka集群可能出问题了，有些case程序的offset还可能出现错误涉及到数据修复
			//所有的错误描述均可在sarama搜索到
			//kafka错误码协议地址：https://kafka.apache.org/protocol#protocol_error_codes
		}

	default:
		handler.Log.Info(ctx, "consumer group finish consumed msg")
	}
}
