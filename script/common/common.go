package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var webhookURL = "https://open.feishu.cn/open-apis/bot/v2/hook/25277019-541e-47d9-9155-2bf69a4779dd"

// TextMessage 定义飞书文本消息结构体
type TextMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

// SendFeishuWebhook 发送飞书 Webhook 消息的函数
func SendFeishuWebhook(message string) error {
	// 创建文本消息实例
	msg := TextMessage{
		MsgType: "text",
	}
	msg.Content.Text = message

	// 将消息结构体转换为 JSON 字节切片
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// 创建 HTTP POST 请求
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	// 设置请求头，指定内容类型为 JSON
	req.Header.Set("Content-Type", "application/json")

	// 创建 HTTP 客户端
	client := &http.Client{}
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	// 确保响应体在函数结束时关闭
	defer resp.Body.Close()

	// 读取响应体内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
