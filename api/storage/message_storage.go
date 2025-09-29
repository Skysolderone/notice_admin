package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// MessageRecord 消息记录结构
type MessageRecord struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`    // webhook, manual, rsi, liquidation, news等
	Timestamp time.Time `json:"timestamp"`
}

// MessageStorage 消息存储管理器
type MessageStorage struct {
	filePath string
	mutex    sync.RWMutex
}

var (
	instance *MessageStorage
	once     sync.Once
)

// GetMessageStorage 获取单例消息存储实例
func GetMessageStorage() *MessageStorage {
	once.Do(func() {
		// 确保存储目录存在
		storageDir := "./storage"
		if err := os.MkdirAll(storageDir, 0755); err != nil {
			logx.Errorf("Failed to create storage directory: %v", err)
		}

		instance = &MessageStorage{
			filePath: filepath.Join(storageDir, "messages.json"),
		}
	})
	return instance
}

// SaveMessage 保存消息到文件
func (ms *MessageStorage) SaveMessage(message, source string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// 生成唯一ID
	id := fmt.Sprintf("%d", time.Now().UnixNano())

	record := MessageRecord{
		ID:        id,
		Message:   message,
		Source:    source,
		Timestamp: time.Now(),
	}

	// 读取现有记录
	messages, err := ms.readMessages()
	if err != nil {
		logx.Errorf("Failed to read existing messages: %v", err)
		messages = []MessageRecord{}
	}

	// 添加新记录
	messages = append(messages, record)

	// 保持最近1000条记录
	if len(messages) > 1000 {
		messages = messages[len(messages)-1000:]
	}

	// 写入文件
	return ms.writeMessages(messages)
}

// GetMessages 获取消息列表
func (ms *MessageStorage) GetMessages(limit int) ([]MessageRecord, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	messages, err := ms.readMessages()
	if err != nil {
		return nil, err
	}

	// 如果指定了limit且小于总数，返回最新的limit条
	if limit > 0 && limit < len(messages) {
		return messages[len(messages)-limit:], nil
	}

	return messages, nil
}

// GetMessagesByTimeRange 根据时间范围获取消息
func (ms *MessageStorage) GetMessagesByTimeRange(start, end time.Time) ([]MessageRecord, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	messages, err := ms.readMessages()
	if err != nil {
		return nil, err
	}

	var filteredMessages []MessageRecord
	for _, msg := range messages {
		if msg.Timestamp.After(start) && msg.Timestamp.Before(end) {
			filteredMessages = append(filteredMessages, msg)
		}
	}

	return filteredMessages, nil
}

// readMessages 从文件读取消息
func (ms *MessageStorage) readMessages() ([]MessageRecord, error) {
	if _, err := os.Stat(ms.filePath); os.IsNotExist(err) {
		return []MessageRecord{}, nil
	}

	data, err := ioutil.ReadFile(ms.filePath)
	if err != nil {
		return nil, err
	}

	var messages []MessageRecord
	if len(data) == 0 {
		return messages, nil
	}

	err = json.Unmarshal(data, &messages)
	if err != nil {
		logx.Errorf("Failed to unmarshal messages: %v", err)
		return []MessageRecord{}, nil
	}

	return messages, nil
}

// writeMessages 写入消息到文件
func (ms *MessageStorage) writeMessages(messages []MessageRecord) error {
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(ms.filePath, data, 0644)
}

// GetMessageCount 获取消息总数
func (ms *MessageStorage) GetMessageCount() (int, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	messages, err := ms.readMessages()
	if err != nil {
		return 0, err
	}

	return len(messages), nil
}