package expo

import (
	"fmt"

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
	// Publish message
	response, err := e.client.Publish(
		&expo.PushMessage{
			To:       e.pushToken,
			Body:     message,
			Data:     map[string]string{"withSome": "data"},
			Sound:    "default",
			Title:    "Rsi_signal",
			Priority: expo.DefaultPriority,
		},
	)
	// Check errors
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Validate responses
	if response.ValidateResponse() != nil {
		return fmt.Errorf("failed to send message: %s", response.Message)
	}
	return nil
}
