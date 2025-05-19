package consts

const (
	KafkaMaxWriteDeadline  = 5
	KafkaMaximumRetryCount = 5
	KafkaDefaultPoolSize   = 3  // 每个 topic 的默认连接数
	KafkaMaxPoolSize       = 10 // 每个 topic 的最大连接数

	KafkaTopicTeacherBehavior = "topic-teacher-behaviors"
	KafkaTopicCommunication   = "topic-communication"
	KafkaGroupBehavior        = "group-teacher-behaviors"
)

var (
	KafkaTopicBehaviors = []string{
		KafkaTopicTeacherBehavior,
	}
)
