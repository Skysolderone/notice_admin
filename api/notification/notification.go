package notification

import (
	"notice/api/expo"
	"notice/api/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

// SendNotification 发送通知并保存到存储
func SendNotification(message, source string) error {
	// 保存消息到存储
	err := storage.GetMessageStorage().SaveMessage(message, source)
	if err != nil {
		logx.Errorf("Failed to save message to storage: %v", err)
	}

	// 发送推送通知
	return expo.GetExpoClient().Send(message)
}

// SendNotificationWithTitle 发送带标题的通知并保存到存储
func SendNotificationWithTitle(message, title, source string) error {
	// 保存消息到存储
	err := storage.GetMessageStorage().SaveMessage(message, source)
	if err != nil {
		logx.Errorf("Failed to save message to storage: %v", err)
	}

	// 发送推送通知
	return expo.GetExpoClient().SendWithCustomTitle(message, title)
}

// SendNotificationWithRetry 发送通知并保存到存储（带重试）
func SendNotificationWithRetry(message, source string, maxRetries int) error {
	// 保存消息到存储
	err := storage.GetMessageStorage().SaveMessage(message, source)
	if err != nil {
		logx.Errorf("Failed to save message to storage: %v", err)
	}

	// 发送推送通知
	return expo.GetExpoClient().SendWithRetry(message, maxRetries)
}