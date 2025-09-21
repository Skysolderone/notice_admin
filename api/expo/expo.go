package expo

import (
	"fmt"
	"log"
	"time"

	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

type Expo struct {
	pushToken []expo.ExponentPushToken
	client    *expo.PushClient
}

var expoClient *Expo

func GetExpoClient() *Expo {
	if expoClient == nil {
		NewClient()
	}
	return expoClient
}

func NewClient() {
	// Create a new Expo SDK client
	client := expo.NewPushClient(nil)
	expoClient = &Expo{
		pushToken: make([]expo.ExponentPushToken, 0),
		client:    client,
	}
}

func (e *Expo) validateToken(token string) (expo.ExponentPushToken, error) {
	validtoken, err := expo.NewExponentPushToken(token)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}
	return validtoken, nil
}

func (e *Expo) AddToken(token string) error {
	validtoken, err := e.validateToken(token)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// 检查是否已存在该 token
	for _, existingToken := range e.pushToken {
		if existingToken == validtoken {
			return fmt.Errorf("token already exists")
		}
	}

	e.pushToken = append(e.pushToken, validtoken)
	return nil
}

// GetTokenCount 返回当前 token 数量
func (e *Expo) GetTokenCount() int {
	return len(e.pushToken)
}

// RemoveToken 移除指定的 token
func (e *Expo) RemoveToken(token string) error {
	validtoken, err := e.validateToken(token)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	for i, existingToken := range e.pushToken {
		if existingToken == validtoken {
			// 移除该 token
			e.pushToken = append(e.pushToken[:i], e.pushToken[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("token not found")
}

// GetTokens 返回所有 token（用于调试）
func (e *Expo) GetTokens() []expo.ExponentPushToken {
	return e.pushToken
}

func (e *Expo) Send(message string) error {
	return e.SendWithCustomTitle(message, "Rsi_signal")
}

func (e *Expo) SendWithCustomTitle(message, title string) error {
	return e.SendWithCustomTitleAndRetry(message, title, 3)
}

func (e *Expo) SendWithRetry(message string, maxRetries int) error {
	return e.SendWithCustomTitleAndRetry(message, "Rsi_signal", maxRetries)
}

func (e *Expo) SendWithCustomTitleAndRetry(message, title string, maxRetries int) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("推送尝试 %d/%d", attempt, maxRetries)

		// Publish message
		response, err := e.client.Publish(
			&expo.PushMessage{
				To:         e.pushToken,
				Body:       message,
				Data:       map[string]string{"withSome": "data"},
				Sound:      "default",
				Title:      title,
				Priority:   expo.HighPriority,
				TTLSeconds: 0,
			},
		)
		// Check network/client errors
		if err != nil {
			lastErr = fmt.Errorf("网络错误 (尝试 %d): %w", attempt, err)
			log.Printf("推送失败: %v", lastErr)

			if attempt < maxRetries {
				waitTime := time.Duration(attempt) * 2 * time.Second
				log.Printf("等待 %v 后重试...", waitTime)
				time.Sleep(waitTime)
				continue
			}
			return lastErr
		}

		// Print detailed response for debugging
		log.Printf("推送响应: Status=%s, ID=%s", response.Status, response.ID)
		if response.Message != "" {
			log.Printf("响应消息: %s", response.Message)
		}

		// Check if push was accepted by Expo
		if response.Status == "ok" {
			log.Printf("推送成功提交到 Expo 服务器")
			return nil
		}

		// Handle specific error cases
		if response.Status == "error" {
			switch response.Details["error"] {
			case "DeviceNotRegistered":
				return fmt.Errorf("设备未注册，Token 可能已失效: %s", response.Message)
			case "MessageTooBig":
				return fmt.Errorf("消息过大: %s", response.Message)
			case "MessageRateExceeded":
				lastErr = fmt.Errorf("消息频率超限 (尝试 %d): %s", attempt, response.Message)
				log.Printf("推送被限流: %v", lastErr)
			default:
				return fmt.Errorf("推送错误: %s (详情: %v)", response.Message, response.Details)
			}
		}

		// For rate limiting, wait and retry
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * 5 * time.Second // 更长的等待时间
			log.Printf("等待 %v 后重试...", waitTime)
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("推送失败，已重试 %d 次: %v", maxRetries, lastErr)
}

// CheckToken 检查推送 Token 是否仍然有效
func (e *Expo) CheckToken(token string) error {
	validtoken, err := e.validateToken(token)
	if err != nil {
		return fmt.Errorf("Token 格式无效: %w", err)
	}

	// 发送一个测试消息来验证 Token
	response, err := e.client.Publish(
		&expo.PushMessage{
			To:       []expo.ExponentPushToken{validtoken},
			Body:     "", // 空消息用于测试
			Data:     map[string]string{"test": "ping"},
			Sound:    "",
			Title:    "",
			Priority: expo.NormalPriority,
		},
	)
	if err != nil {
		return fmt.Errorf("Token 验证请求失败: %w", err)
	}

	switch response.Status {
	case "ok":
		return nil // Token 有效
	case "error":
		if response.Details["error"] == "DeviceNotRegistered" {
			return fmt.Errorf("Token 已失效 - 设备未注册")
		}
		return fmt.Errorf("Token 验证失败: %s", response.Message)
	default:
		return fmt.Errorf("Token 验证返回未知状态: %s", response.Status)
	}
}

// SendToSpecificToken 向特定 Token 发送消息（用于测试）
func (e *Expo) SendToSpecificToken(token, message string) error {
	validtoken, err := e.validateToken(token)
	if err != nil {
		return fmt.Errorf("Token 无效: %w", err)
	}

	response, err := e.client.Publish(
		&expo.PushMessage{
			To:       []expo.ExponentPushToken{validtoken},
			Body:     message,
			Data:     map[string]string{"timestamp": time.Now().Format(time.RFC3339)},
			Sound:    "default",
			Title:    "测试推送",
			Priority: expo.HighPriority,
		},
	)
	if err != nil {
		return fmt.Errorf("推送失败: %w", err)
	}

	log.Printf("单独推送响应: Status=%s, ID=%s", response.Status, response.ID)

	if response.Status != "ok" {
		return fmt.Errorf("推送被拒绝: %s (详情: %v)", response.Message, response.Details)
	}

	return nil
}
