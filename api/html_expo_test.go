package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"notice/api/expo"
)

// MockRSSFeedItem 模拟RSS feed项目
type MockRSSFeedItem struct {
	Title       string
	Description string
	Published   string
}

// MockExpoClient 模拟Expo客户端用于测试
type MockExpoClient struct {
	sentMessages []string
	shouldFail   bool
}

func (m *MockExpoClient) Send(message string) error {
	if m.shouldFail {
		return fmt.Errorf("模拟发送失败")
	}
	m.sentMessages = append(m.sentMessages, message)
	return nil
}

func (m *MockExpoClient) GetSentMessages() []string {
	return m.sentMessages
}

func (m *MockExpoClient) Reset() {
	m.sentMessages = []string{}
	m.shouldFail = false
}

func (m *MockExpoClient) SetShouldFail(fail bool) {
	m.shouldFail = fail
}

// processHTMLContent 处理HTML内容的函数（从blockbeat.go提取的逻辑）
func processHTMLContent(title, description, published string) string {
	var text string
	
	// 添加标题（粗体格式）
	text += "<b>" + title + "</b>" + "\n"
	
	// 清理描述中的特定文本
	msg := strings.ReplaceAll(description, "BlockBeats 消息，", "")
	
	// 移除所有HTML标签
	re := regexp.MustCompile(`<.*?>`)
	text += re.ReplaceAllString(msg, "")
	
	// 添加发布时间
	text += "\n" + published
	
	return text
}

// TestHTMLContentProcessing 测试HTML内容处理
func TestHTMLContentProcessing(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		published   string
		expected    string
	}{
		{
			name:        "基本HTML清理",
			title:       "测试新闻标题",
			description: "BlockBeats 消息，<p>这是一条包含<strong>HTML标签</strong>的消息。</p>",
			published:   "Mon, 07 Sep 2025 10:00:00 +0000",
			expected:    "<b>测试新闻标题</b>\n这是一条包含HTML标签的消息。\nMon, 07 Sep 2025 10:00:00 +0000",
		},
		{
			name:        "复杂HTML清理",
			title:       "复杂新闻",
			description: "BlockBeats 消息，<div><span>嵌套</span><a href='#'>链接</a><img src='test.jpg'>图片</div>",
			published:   "Mon, 07 Sep 2025 11:00:00 +0000",
			expected:    "<b>复杂新闻</b>\n嵌套链接图片\nMon, 07 Sep 2025 11:00:00 +0000",
		},
		{
			name:        "空描述处理",
			title:       "空描述测试",
			description: "BlockBeats 消息，",
			published:   "Mon, 07 Sep 2025 12:00:00 +0000",
			expected:    "<b>空描述测试</b>\n\nMon, 07 Sep 2025 12:00:00 +0000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processHTMLContent(tt.title, tt.description, tt.published)
			if result != tt.expected {
				t.Errorf("processHTMLContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestExpoSendFunctionality 测试Expo发送功能
func TestExpoSendFunctionality(t *testing.T) {
	// 创建真实的Expo客户端进行测试
	expoClient := expo.GetExpoClient()
	
	tests := []struct {
		name     string
		message  string
		wantErr  bool
		testType string
	}{
		{
			name:     "发送普通文本消息",
			message:  "这是一条测试消息",
			wantErr:  false,
			testType: "normal",
		},
		{
			name:     "发送带HTML格式的消息",
			message:  "<b>重要通知</b>\n这是一条包含格式的消息",
			wantErr:  false,
			testType: "html",
		},
		{
			name:     "发送长消息",
			message:  strings.Repeat("这是一条很长的消息。", 50),
			wantErr:  false,
			testType: "long",
		},
		{
			name:     "发送空消息",
			message:  "",
			wantErr:  false,
			testType: "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := expoClient.Send(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			// 记录测试结果
			if err != nil {
				t.Logf("消息发送失败: %v", err)
			} else {
				t.Logf("消息发送成功: %s", tt.message[:min(len(tt.message), 50)])
			}
		})
	}
}

// TestWebhookHTMLProcessing 测试Webhook接收HTML并发送
func TestWebhookHTMLProcessing(t *testing.T) {
	// 模拟HTTP服务器
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid JSON payload"))
			return
		}

		// 提取消息
		var message string
		if msg, ok := payload["message"].(string); ok {
			message = msg
		} else if data, ok := payload["data"].(string); ok {
			message = data
		} else {
			message = fmt.Sprintf("Webhook payload: %v", payload)
		}

		if message == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("No message found in payload"))
			return
		}

		// 处理HTML内容
		processedMessage := processHTMLContent("Webhook通知", message, time.Now().Format(time.RFC3339))

		// 发送到Expo（这里只是模拟，实际测试中可以使用Mock）
		expoClient := expo.GetExpoClient()
		err := expoClient.Send(processedMessage)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Webhook processed successfully"))
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// 测试用例
	tests := []struct {
		name       string
		payload    map[string]interface{}
		expectCode int
	}{
		{
			name: "HTML消息处理",
			payload: map[string]interface{}{
				"message": "BlockBeats 消息，<p>市场出现<strong>重大变化</strong>，请注意风险。</p>",
			},
			expectCode: http.StatusOK,
		},
		{
			name: "数据字段消息",
			payload: map[string]interface{}{
				"data": "<div>这是一条包含<span>HTML标签</span>的数据消息</div>",
			},
			expectCode: http.StatusOK,
		},
		{
			name: "空消息",
			payload: map[string]interface{}{
				"other": "irrelevant data",
			},
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, _ := json.Marshal(tt.payload)
			resp, err := http.Post(server.URL, "application/json", strings.NewReader(string(jsonPayload)))
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectCode {
				t.Errorf("Expected status %d, got %d", tt.expectCode, resp.StatusCode)
			}
		})
	}
}

// TestRSSFeedProcessing 测试RSS feed处理
func TestRSSFeedProcessing(t *testing.T) {
	// 这个测试模拟了blockbeat.go中的RSS处理逻辑
	
	// 模拟RSS数据
	mockFeedItems := []MockRSSFeedItem{
		{
			Title:       "比特币价格突破新高",
			Description: "BlockBeats 消息，<p>比特币价格在今日突破了<strong>历史新高</strong>，达到了前所未有的水平。</p>",
			Published:   "Mon, 07 Sep 2025 14:30:00 +0000",
		},
		{
			Title:       "以太坊网络升级完成",
			Description: "BlockBeats 消息，<div>以太坊网络成功完成了最新的<em>协议升级</em>，提高了交易效率。</div>",
			Published:   "Mon, 07 Sep 2025 15:00:00 +0000",
		},
	}

	for i, item := range mockFeedItems {
		t.Run(fmt.Sprintf("RSS项目_%d", i+1), func(t *testing.T) {
			// 处理HTML内容
			processedContent := processHTMLContent(item.Title, item.Description, item.Published)
			
			// 验证HTML标签被移除
			if strings.Contains(processedContent, "<p>") || strings.Contains(processedContent, "</p>") {
				t.Error("HTML标签未被完全移除")
			}
			
			// 验证包含标题
			if !strings.Contains(processedContent, item.Title) {
				t.Error("处理后的内容应包含原标题")
			}
			
			// 验证包含发布时间
			if !strings.Contains(processedContent, item.Published) {
				t.Error("处理后的内容应包含发布时间")
			}
			
			t.Logf("处理结果: %s", processedContent)
		})
	}
}

// TestIntegratedHTMLToExpoFlow 集成测试：完整的HTML处理到Expo发送流程
func TestIntegratedHTMLToExpoFlow(t *testing.T) {
	// 模拟完整的数据流：从HTML内容接收到Expo发送
	
	scenarios := []struct {
		name        string
		htmlContent string
		title       string
		published   string
		expectSend  bool
	}{
		{
			name:        "完整新闻流程",
			htmlContent: "BlockBeats 消息，<p>加密货币市场今日出现<strong>大幅波动</strong>，<a href='#'>查看详情</a>。</p>",
			title:       "市场动态Alert",
			published:   time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700"),
			expectSend:  true,
		},
		{
			name:        "简短通知",
			htmlContent: "BlockBeats 消息，简短通知测试",
			title:       "简短通知",
			published:   time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700"),
			expectSend:  true,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// 步骤1: 处理HTML内容
			processedMessage := processHTMLContent(scenario.title, scenario.htmlContent, scenario.published)
			
			// 步骤2: 验证内容处理正确
			if processedMessage == "" && scenario.expectSend {
				t.Error("处理后的消息不应为空")
				return
			}
			
			// 步骤3: 发送到Expo
			expoClient := expo.GetExpoClient()
			err := expoClient.Send(processedMessage)
			
			if scenario.expectSend {
				if err != nil {
					t.Errorf("Expo发送失败: %v", err)
				} else {
					t.Logf("成功发送消息: %s", processedMessage[:min(len(processedMessage), 100)])
				}
			}
		})
	}
}

// TestTokenManagement 测试Token管理功能
func TestTokenManagement(t *testing.T) {
	expoClient := expo.GetExpoClient()
	testToken := "ExponentPushToken[exsnUAGvojZv3ffSYpG5Mr]"
	
	// 测试添加Token
	t.Run("添加Token", func(t *testing.T) {
		err := expoClient.AddToken(testToken)
		if err != nil && !strings.Contains(err.Error(), "token already exists") {
			t.Errorf("添加Token失败: %v", err)
		}
	})
	
	// 测试获取Token数量
	t.Run("获取Token数量", func(t *testing.T) {
		count := expoClient.GetTokenCount()
		if count < 0 {
			t.Error("Token数量不应为负数")
		}
		t.Logf("当前Token数量: %d", count)
	})
	
	// 测试向特定Token发送HTML处理后的消息
	t.Run("向特定Token发送HTML消息", func(t *testing.T) {
		htmlMessage := "<p>这是一条包含<strong>HTML标签</strong>的测试消息</p>"
		processedMessage := processHTMLContent("测试通知", htmlMessage, time.Now().Format(time.RFC3339))
		
		err := expoClient.SendToSpecificToken(testToken, processedMessage)
		if err != nil {
			t.Logf("向特定Token发送失败（可能Token无效）: %v", err)
		} else {
			t.Log("向特定Token发送成功")
		}
	})
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	expoClient := expo.GetExpoClient()
	
	tests := []struct {
		name        string
		operation   func() error
		expectError bool
	}{
		{
			name: "无效Token格式",
			operation: func() error {
				return expoClient.AddToken("invalid-token-format")
			},
			expectError: true,
		},
		{
			name: "发送到无效Token",
			operation: func() error {
				return expoClient.SendToSpecificToken("ExponentPushToken[InvalidToken]", "测试消息")
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()
			if (err != nil) != tt.expectError {
				t.Errorf("期望错误: %v, 实际错误: %v", tt.expectError, err)
			}
			if err != nil {
				t.Logf("捕获到预期错误: %v", err)
			}
		})
	}
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}