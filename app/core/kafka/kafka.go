package kafka

// import (
// 	"container/ring"
// 	"context"
// 	"fmt"
// 	"math/rand"
// 	"net"
// 	"strconv"
// 	"strings"
// 	"sync"
// 	"sync/atomic"
// 	"time"

// 	"gil_teacher/app/conf"
// 	"gil_teacher/app/consts"
// 	"gil_teacher/app/core/logger"

// 	"github.com/segmentio/kafka-go"
// )

// // 连接池配置
// const (
// )

// // 连接包装器
// type connWrapper struct {
// 	conn      *kafka.Conn
// 	lastUsed  time.Time
// 	errorCount int32
// }

// // 连接池
// type connectionPool struct {
// 	connections *ring.Ring
// 	size        int
// 	mu          sync.RWMutex
// }

// type ConsumeHandler func(ctx context.Context, key string, body []byte) error

// // ApiKafkaClient kafka 生产消费接口
// type ApiKafkaClient interface {
// 	// Produce 向指定topic生产单条消息
// 	Produce(ctx context.Context, topic string, key string, body string) error
// 	// Consume 消费指定topic的消息
// 	Consume(ctx context.Context, topic string, handler ConsumeHandler) error
// 	// Close 关闭所有连接
// 	Close() error
// }

// // 修改 kafkaClient 结构
// type kafkaClient struct {
// 	ctx      context.Context
// 	timeout  int
// 	config   *conf.Kafka
// 	logger   logger.ContextLogger
// 	mu       sync.RWMutex
// 	pools    map[string]*connectionPool  // topic -> connection pool
// 	readers  map[string]*kafka.Reader
// 	handlers map[string]ConsumeHandler
// }

// // 创建新的连接池
// func newConnectionPool(size int) *connectionPool {
// 	if size <= 0 {
// 		size = consts.KafkaDefaultPoolSize
// 	}
// 	if size > consts.KafkaMaxPoolSize {
// 		size = consts.KafkaMaxPoolSize
// 	}

// 	r := ring.New(size)
// 	return &connectionPool{
// 		connections: r,
// 		size:       size,
// 	}
// }

// // 获取连接
// func (p *connectionPool) getConn() *connWrapper {
// 	p.mu.RLock()
// 	defer p.mu.RUnlock()

// 	if p.connections == nil {
// 		return nil
// 	}

// 	// 遍历环形缓冲区找到可用连接
// 	for range p.size {
// 		if conn := p.connections.Value.(*connWrapper); conn != nil {
// 			if atomic.LoadInt32(&conn.errorCount) < 3 { // 如果错误次数小于3次
// 				conn.lastUsed = time.Now()
// 				return conn
// 			}
// 		}
// 		p.connections = p.connections.Next()
// 	}
// 	return nil
// }

// // 添加连接
// func (p *connectionPool) addConn(conn *kafka.Conn) {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	wrapper := &connWrapper{
// 		conn:      conn,
// 		lastUsed:  time.Now(),
// 		errorCount: 0,
// 	}

// 	// 找到一个空位或替换最旧的连接
// 	current := p.connections
// 	oldest := current
// 	for range p.size {
// 		if current.Value == nil {
// 			current.Value = wrapper
// 			return
// 		}
// 		if c := current.Value.(*connWrapper); c.lastUsed.Before(oldest.Value.(*connWrapper).lastUsed) {
// 			oldest = current
// 		}
// 		current = current.Next()
// 	}

// 	// 关闭最旧的连接
// 	if oldest.Value != nil {
// 		oldest.Value.(*connWrapper).conn.Close()
// 	}
// 	oldest.Value = wrapper
// }

// // 修改 Produce 方法
// func (p *kafkaClient) Produce(ctx context.Context, topic string, key string, body string) error {
// 	if len(body) == 0 {
// 		return fmt.Errorf("empty message body")
// 	}

// 	for range consts.KafkaMaximumRetryCount {
// 		// 获取连接池
// 		pool, exists := p.getOrCreatePool(topic)
// 		if !exists {
// 			time.Sleep(time.Duration(100+rand.Intn(100)) * time.Millisecond)
// 			continue
// 		}

// 		// 获取连接
// 		wrapper := pool.getConn()
// 		if wrapper == nil {
// 			// 尝试创建新连接
// 			if err := p.addNewConnection(topic); err != nil {
// 				p.logger.Error(ctx, "Failed to create new connection for topic %s: %v", topic, err)
// 				time.Sleep(time.Duration(100+rand.Intn(100)) * time.Millisecond)
// 				continue
// 			}
// 			wrapper = pool.getConn()
// 			if wrapper == nil {
// 				continue
// 			}
// 		}

// 		conn := wrapper.conn
// 		if err := conn.SetWriteDeadline(time.Now().Add(consts.KafkaMaxWriteDeadline * time.Second)); err != nil {
// 			atomic.AddInt32(&wrapper.errorCount, 1)
// 			continue
// 		}

// 		_, err := conn.WriteMessages(kafka.Message{Key: []byte(key), Value: []byte(body)})
// 		if err == nil {
// 			atomic.StoreInt32(&wrapper.errorCount, 0) // 重置错误计数
// 			return nil
// 		}

// 		atomic.AddInt32(&wrapper.errorCount, 1)
// 		p.logger.Error(ctx, "Failed to produce message, topic: %s, error: %v", topic, err)

// 		// 处理错误情况
// 		if p.handleProduceError(ctx, topic, err) {
// 			continue
// 		}

// 		time.Sleep(time.Duration(100+rand.Intn(100)) * time.Millisecond)
// 	}

// 	return fmt.Errorf("failed to produce message to topic %s after %d retries", topic, consts.KafkaMaximumRetryCount)
// }

// // 获取或创建连接池
// func (p *kafkaClient) getOrCreatePool(topic string) (*connectionPool, bool) {
// 	p.mu.RLock()
// 	pool, exists := p.pools[topic]
// 	p.mu.RUnlock()

// 	if !exists {
// 		p.mu.Lock()
// 		pool = newConnectionPool(consts.KafkaDefaultPoolSize)
// 		p.pools[topic] = pool
// 		p.mu.Unlock()

// 		// 初始化连接
// 		if err := p.addNewConnection(topic); err != nil {
// 			return nil, false
// 		}
// 	}
// 	return pool, true
// }

// // 添加新连接
// func (p *kafkaClient) addNewConnection(topic string) error {
// 	conn, err := kafka.DialLeader(p.ctx, "tcp", p.config.Brokers, topic, 0)
// 	if err != nil {
// 		return err
// 	}

// 	if err := conn.SetWriteDeadline(time.Now().Add(consts.KafkaMaxWriteDeadline * time.Second)); err != nil {
// 		conn.Close()
// 		return err
// 	}

// 	p.mu.RLock()
// 	pool := p.pools[topic]
// 	p.mu.RUnlock()

// 	pool.addConn(conn)
// 	return nil
// }

// // 处理生产者错误
// func (p *kafkaClient) handleProduceError(ctx context.Context, topic string, err error) bool {
// 	if strings.Contains(err.Error(), "Unknown Topic Or Partition") {
// 		if createErr := p.ensureTopicsExist(); createErr != nil {
// 			p.logger.Error(ctx, "Failed to create missing topic: %v", createErr)
// 			return true
// 		}
// 		return true
// 	}

// 	if strings.Contains(err.Error(), "use of closed network connection") ||
// 		strings.Contains(err.Error(), "broken pipe") ||
// 		strings.Contains(err.Error(), "connection reset by peer") {
// 		return true
// 	}

// 	return false
// }

// // 修改初始化方法
// func NewApiKafkaClient(ctx context.Context, conf *conf.Data, logger *logger.ContextLogger) ApiKafkaClient {
// 	client := &kafkaClient{
// 		ctx:      ctx,
// 		timeout:  10,
// 		config:   conf.Kafka,
// 		logger:   *logger,
// 		pools:    make(map[string]*connectionPool),
// 		readers:  make(map[string]*kafka.Reader),
// 		handlers: make(map[string]ConsumeHandler),
// 	}
// 	client.init()
// 	return client
// }

// func (p *kafkaClient) init() {
// 	// 添加 topic 检查和创建逻辑
// 	if err := p.ensureTopicsExist(); err != nil {
// 		p.logger.Error(p.ctx, "Failed to ensure topics exist: %v", err)
// 		return
// 	}

// 	// 原有的初始化逻辑
// 	for _, topicConfig := range p.config.Topics {
// 		// 初始化生产者连接
// 		if err := p.initProducerConn(topicConfig.Name); err != nil {
// 			p.logger.Error(p.ctx, "Failed to init producer connection for topic %s: %v", topicConfig.Name, err)
// 			continue
// 		}

// 		// 初始化消费者reader
// 		reader := p.getKafkaReader(topicConfig.Name, topicConfig.GroupID)
// 		p.mu.Lock()
// 		p.readers[topicConfig.Name] = reader
// 		p.mu.Unlock()
// 	}
// }

// func (p *kafkaClient) ensureTopicsExist() error {
// 	// 连接到任意一个 broker
// 	conn, err := kafka.Dial("tcp", p.config.Brokers)
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to kafka: %v", err)
// 	}
// 	defer conn.Close()

// 	// 获取 controller broker
// 	controller, err := conn.Controller()
// 	if err != nil {
// 		return fmt.Errorf("failed to get controller: %v", err)
// 	}

// 	// 连接到 controller
// 	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to controller: %v", err)
// 	}
// 	defer controllerConn.Close()

// 	// 检查并创建每个 topic
// 	for _, topicConfig := range p.config.Topics {
// 		// 创建 topic 配置
// 		topicConfigs := []kafka.TopicConfig{
// 			{
// 				Topic:             topicConfig.Name,
// 				NumPartitions:     3, // 可以通过配置文件设置
// 				ReplicationFactor: 1, // 可以通过配置文件设置
// 			},
// 		}

// 		err := controllerConn.CreateTopics(topicConfigs...)
// 		if err != nil && !strings.Contains(err.Error(), "Topic already exists") {
// 			return fmt.Errorf("failed to create topic %s: %v", topicConfig.Name, err)
// 		}
// 		p.logger.Info(p.ctx, "Ensured topic exists: %s", topicConfig.Name)
// 	}

// 	return nil
// }

// func (p *kafkaClient) initProducerConn(topic string) error {
// 	// 先关闭旧连接
// 	p.mu.Lock()
// 	if oldConn, exists := p.readers[topic]; exists {
// 		_ = oldConn.Close()
// 		delete(p.readers, topic)
// 	}
// 	p.mu.Unlock()

// 	// 创建新的连接池
// 	pool := newConnectionPool(consts.KafkaDefaultPoolSize)
// 	p.mu.Lock()
// 	p.pools[topic] = pool
// 	p.mu.Unlock()

// 	// 添加新连接
// 	if err := p.addNewConnection(topic); err != nil {
// 		return err
// 	}

// 	p.logger.Info(p.ctx, "init kafka producer conn success, topic: %s", topic)
// 	return nil
// }

// func (p *kafkaClient) Consume(ctx context.Context, topic string, handler ConsumeHandler) error {
// 	p.mu.RLock()
// 	reader, exists := p.readers[topic]
// 	p.mu.RUnlock()

// 	if !exists {
// 		return fmt.Errorf("no reader found for topic: %s", topic)
// 	}

// 	p.mu.Lock()
// 	p.handlers[topic] = handler
// 	p.mu.Unlock()

// 	go func() {
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				return
// 			default:
// 				msgCtx, cancel := context.WithTimeout(ctx, time.Duration(p.timeout)*time.Second)
// 				m, err := reader.FetchMessage(msgCtx)
// 				cancel()

// 				if err != nil {
// 					if err != context.DeadlineExceeded && err != context.Canceled {
// 						p.logger.Error(ctx, "Failed to fetch message, topic: %s, error: %+v", topic, err)
// 					}
// 					continue
// 				}

// 				if err := handler(ctx, string(m.Key), m.Value); err != nil {
// 					p.logger.Error(ctx, "Failed to process message, topic: %s, error: %+v", topic, err)
// 					continue
// 				}

// 				if err := reader.CommitMessages(ctx, m); err != nil {
// 					p.logger.Error(ctx, "Failed to commit message, topic: %s, error: %+v", topic, err)
// 				}
// 			}
// 		}
// 	}()

// 	return nil
// }

// // 修改关闭方法
// func (p *kafkaClient) Close() error {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	// 关闭所有连接池中的连接
// 	for topic, pool := range p.pools {
// 		pool.mu.Lock()
// 		current := pool.connections
// 		for range pool.size {
// 			if wrapper := current.Value.(*connWrapper); wrapper != nil {
// 				if err := wrapper.conn.Close(); err != nil {
// 					p.logger.Error(p.ctx, "Failed to close connection for topic %s: %v", topic, err)
// 				}
// 			}
// 			current = current.Next()
// 		}
// 		pool.mu.Unlock()
// 	}

// 	// 关闭所有消费者reader
// 	for topic, reader := range p.readers {
// 		if err := reader.Close(); err != nil {
// 			p.logger.Error(p.ctx, "Failed to close consumer reader, topic: %s, error: %v", topic, err)
// 		}
// 	}

// 	return nil
// }

// func (p *kafkaClient) getKafkaReader(topic, groupID string) *kafka.Reader {
// 	return kafka.NewReader(kafka.ReaderConfig{
// 		Brokers:               []string{p.config.Brokers},
// 		GroupID:               groupID,
// 		Topic:                 topic,
// 		WatchPartitionChanges: true,
// 		MaxBytes:              10e6, // 10MB
// 		CommitInterval:        p.config.Consumer.CommitInterval,
// 	})
// }
