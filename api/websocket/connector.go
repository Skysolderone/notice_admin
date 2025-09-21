package websocket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"notice/api/config"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

// MessageHandler 定义消息处理函数类型
type MessageHandler func(messageType int, data []byte) error

// WebSocketConnector WebSocket连接器
type WebSocketConnector struct {
	config         config.WebSocketConfig
	conn           *websocket.Conn
	mu             sync.RWMutex
	isConnected    bool
	reconnectCount int
	ctx            context.Context
	cancel         context.CancelFunc
	messageHandler MessageHandler
	onConnect      func()
	onDisconnect   func(err error)
}

// NewWebSocketConnector 创建新的WebSocket连接器
func NewWebSocketConnector(cfg config.WebSocketConfig) *WebSocketConnector {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WebSocketConnector{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}
}

// SetMessageHandler 设置消息处理函数
func (w *WebSocketConnector) SetMessageHandler(handler MessageHandler) {
	w.messageHandler = handler
}

// SetOnConnect 设置连接成功回调
func (w *WebSocketConnector) SetOnConnect(callback func()) {
	w.onConnect = callback
}

// SetOnDisconnect 设置断开连接回调
func (w *WebSocketConnector) SetOnDisconnect(callback func(err error)) {
	w.onDisconnect = callback
}

// Connect 连接到WebSocket服务器
func (w *WebSocketConnector) Connect() error {
	if w.config.URL == "" {
		return fmt.Errorf("websocket URL not configured")
	}

	// 设置默认值
	handshakeTimeout := w.config.HandshakeTimeout
	if handshakeTimeout == 0 {
		handshakeTimeout = 10
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: time.Duration(handshakeTimeout) * time.Second,
	}

	// 构建请求头
	headers := make(map[string][]string)
	for k, v := range w.config.Headers {
		headers[k] = []string{v}
	}

	logx.Infof("Connecting to WebSocket: %s", w.config.URL)
	
	conn, _, err := dialer.Dial(w.config.URL, headers)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", w.config.URL, err)
	}

	w.mu.Lock()
	w.conn = conn
	w.isConnected = true
	w.reconnectCount = 0
	w.mu.Unlock()

	logx.Infof("WebSocket connected successfully to %s", w.config.URL)
	
	if w.onConnect != nil {
		w.onConnect()
	}

	// 启动消息监听和心跳
	go w.readMessages()
	go w.pingLoop()

	return nil
}

// Start 启动连接器（带自动重连）
func (w *WebSocketConnector) Start() {
	go w.connectLoop()
}

// connectLoop 连接循环（处理重连）
func (w *WebSocketConnector) connectLoop() {
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			if !w.IsConnected() {
				if err := w.Connect(); err != nil {
					w.reconnectCount++
					logx.Errorf("WebSocket connection failed (attempt %d): %v", w.reconnectCount, err)
					
					// 检查是否超过最大重连次数
					if w.config.MaxReconnects > 0 && w.reconnectCount >= w.config.MaxReconnects {
						logx.Errorf("Max reconnect attempts (%d) reached, giving up", w.config.MaxReconnects)
						return
					}
					
					// 设置重连延迟默认值
					reconnectDelay := w.config.ReconnectDelay
					if reconnectDelay == 0 {
						reconnectDelay = 5
					}
					
					// 等待重连延迟
					select {
					case <-time.After(time.Duration(reconnectDelay) * time.Second):
					case <-w.ctx.Done():
						return
					}
				} else {
					// 连接成功，重置重连计数
					w.reconnectCount = 0
				}
			}
			
			// 短暂休眠避免过度循环
			time.Sleep(1 * time.Second)
		}
	}
}

// readMessages 读取消息
func (w *WebSocketConnector) readMessages() {
	defer func() {
		w.mu.Lock()
		if w.conn != nil {
			w.conn.Close()
			w.conn = nil
		}
		w.isConnected = false
		w.mu.Unlock()
	}()

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			w.mu.RLock()
			conn := w.conn
			w.mu.RUnlock()
			
			if conn == nil {
				return
			}

			messageType, data, err := conn.ReadMessage()
			if err != nil {
				logx.Errorf("WebSocket read error: %v", err)
				
				w.mu.Lock()
				w.isConnected = false
				w.mu.Unlock()
				
				if w.onDisconnect != nil {
					w.onDisconnect(err)
				}
				
				return
			}

			// 处理接收到的消息
			if w.messageHandler != nil {
				if err := w.messageHandler(messageType, data); err != nil {
					logx.Errorf("Message handler error: %v", err)
				}
			}
		}
	}
}

// pingLoop 心跳循环
func (w *WebSocketConnector) pingLoop() {
	pingInterval := w.config.PingInterval
	if pingInterval == 0 {
		pingInterval = 30
	}
	
	ticker := time.NewTicker(time.Duration(pingInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.mu.RLock()
			conn := w.conn
			connected := w.isConnected
			w.mu.RUnlock()
			
			if !connected || conn == nil {
				return
			}

			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logx.Errorf("WebSocket ping failed: %v", err)
				
				w.mu.Lock()
				w.isConnected = false
				w.mu.Unlock()
				
				if w.onDisconnect != nil {
					w.onDisconnect(err)
				}
				
				return
			}
		}
	}
}

// SendMessage 发送消息
func (w *WebSocketConnector) SendMessage(messageType int, data []byte) error {
	w.mu.RLock()
	conn := w.conn
	connected := w.isConnected
	w.mu.RUnlock()
	
	if !connected || conn == nil {
		return fmt.Errorf("websocket not connected")
	}

	return conn.WriteMessage(messageType, data)
}

// SendText 发送文本消息
func (w *WebSocketConnector) SendText(message string) error {
	return w.SendMessage(websocket.TextMessage, []byte(message))
}

// SendJSON 发送JSON消息
func (w *WebSocketConnector) SendJSON(v interface{}) error {
	w.mu.RLock()
	conn := w.conn
	connected := w.isConnected
	w.mu.RUnlock()
	
	if !connected || conn == nil {
		return fmt.Errorf("websocket not connected")
	}

	return conn.WriteJSON(v)
}

// IsConnected 检查连接状态
func (w *WebSocketConnector) IsConnected() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.isConnected
}

// Close 关闭连接器
func (w *WebSocketConnector) Close() error {
	w.cancel()
	
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if w.conn != nil {
		err := w.conn.Close()
		w.conn = nil
		w.isConnected = false
		return err
	}
	
	return nil
}

// GetReconnectCount 获取重连次数
func (w *WebSocketConnector) GetReconnectCount() int {
	return w.reconnectCount
}