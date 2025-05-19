package behavior

import (
	"encoding/json"
	"time"

	"gil_teacher/app/consts"
)

// BehaviorMessageQueue 行为消息
type BehaviorMessageQueue struct {
	Type      consts.MessageType `json:"type"`
	Content   json.RawMessage    `json:"content"`
	Timestamp time.Time          `json:"timestamp"`
	Version   string             `json:"version"`
}

// 实现 kafka.Message 接口
// 获取BehaviorMessageQueue类型的函数
func (m *BehaviorMessageQueue) GetType() string {
	// 返回m.Type的字符串表示
	return string(m.Type)
}

func (m *BehaviorMessageQueue) GetContent() []byte {
	return m.Content
}

func (m *BehaviorMessageQueue) Encode() []byte {
	msgStr, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return msgStr
}

func DecodeBehaviorMessage(data []byte) (*BehaviorMessageQueue, error) {
	var msg BehaviorMessageQueue
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
