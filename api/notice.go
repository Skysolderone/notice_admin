package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"notice/api/config"
	"notice/api/expo"
	"notice/api/rsi"

	"notice/api/websocket"


	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

func main() {
	expo.NewClient()

	configFile := flag.String("f", "etc/api.yaml", "the config file")
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 确保日志目录存在
	if err := os.MkdirAll("./logs", 0o755); err != nil {
		fmt.Printf("Failed to create logs directory: %v\n", err)
	}

	// 直接写死的WebSocket连接配置
	hardcodedWSConfigs := []config.WebSocketConfig{
		{
			Name:             "binance",
			URL:              "wss://fstream.binance.com/ws/btcusdt@kline_1m",
			ReconnectDelay:   5,
			MaxReconnects:    0,
			PingInterval:     30,
			HandshakeTimeout: 10,
		},
		{
			Name:             "coinbase",
			URL:              "wss://ws-feed.exchange.coinbase.com",
			ReconnectDelay:   3,
			MaxReconnects:    0,
			PingInterval:     25,
			HandshakeTimeout: 10,
		},
	}

	// 初始化多个WebSocket连接器
	wsConnectors := make(map[string]*websocket.WebSocketConnector)

	// 使用写死的配置
	for _, wsConfig := range hardcodedWSConfigs {
		connector := websocket.NewWebSocketConnector(wsConfig)

		// 设置消息处理器
		connector.SetMessageHandler(func(messageType int, data []byte) error {
			logx.Infof("WebSocket [%s] received message: %s", wsConfig.Name, string(data))
			// 可以在这里处理接收到的消息
			return nil
		})

		// 设置连接回调
		connector.SetOnConnect(func() {
			logx.Infof("WebSocket [%s] connected successfully", wsConfig.Name)
		})

		// 设置断开回调
		connector.SetOnDisconnect(func(err error) {
			logx.Errorf("WebSocket [%s] disconnected: %v", wsConfig.Name, err)
		})

		// 启动连接器
		connector.Start()
		wsConnectors[wsConfig.Name] = connector
		logx.Infof("WebSocket connector [%s] started for URL: %s", wsConfig.Name, wsConfig.URL)
	}

	// 创建 go-zero REST 服务，集成静态文件服务
	server := rest.MustNewServer(c.RestConf)
	defer func() {
		server.Stop()
		for name, connector := range wsConnectors {
			logx.Infof("Closing WebSocket connector: %s", name)
			connector.Close()
		}
	}()

	// 初始化 SSE 处理
	sseHandler := rsi.NewSseHandler()

	// 注册 SSE 路由
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/sse",
		Handler: sseHandler.Serve,
	}, rest.WithSSE())

	// 添加测试页面路由
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/test",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "test_sse.html")
		},
	})
	server.AddRoute(rest.Route{
		Method: http.MethodPost,

		Path: "/notice_token",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("add token")
			token := r.FormValue("token")
			if token == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Token is required"))
				return
			}

			err := expo.GetExpoClient().AddToken(token)
			if err != nil {
				if err.Error() == "token already exists" {
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte("Token already exists"))
				} else {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(err.Error()))
				}
				return
			}

			count := expo.GetExpoClient().GetTokenCount()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Token added successfully. Total tokens: %d", count)))
			expo.GetExpoClient().Send("成功订阅rsi通知")
		},
	})

	// 添加 token 管理接口
	// server.AddRoute(rest.Route{
	// 	Method: http.MethodDelete,
	// 	Path:   "/notice_token",
	// 	Handler: func(w http.ResponseWriter, r *http.Request) {
	// 		token := r.FormValue("token")
	// 		if token == "" {
	// 			w.WriteHeader(http.StatusBadRequest)
	// 			w.Write([]byte("Token is required"))
	// 			return
	// 		}

	// 		err := expo.GetExpoClient().RemoveToken(token)
	// 		if err != nil {
	// 			w.WriteHeader(http.StatusNotFound)
	// 			w.Write([]byte(err.Error()))
	// 			return
	// 		}

	// 		count := expo.GetExpoClient().GetTokenCount()
	// 		w.WriteHeader(http.StatusOK)
	// 		w.Write([]byte(fmt.Sprintf("Token removed successfully. Total tokens: %d", count)))
	// 	},
	// })

	// 获取 token 统计信息
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/notice_token/stats",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			count := expo.GetExpoClient().GetTokenCount()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"total_tokens": %d,"tokens":%v}`, count, expo.GetExpoClient().GetTokens())))
		},
	})
	server.AddRoute(rest.Route{
		Method: http.MethodPost,
		Path:   "/notice/query",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			// count := expo.GetExpoClient().GetTokenCount()
			// w.Header().Set("Content-Type", "application/json")
			// w.WriteHeader(http.StatusOK)
			// w.Write([]byte(fmt.Sprintf(`{"total_tokens": %d}`, count)))
			fmt.Println("notice/query")
			data := r.FormValue("data")
			fmt.Println(data)
			if data == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("data is nil"))
				return
			}

			// 记录信号日志
			logx.Infof("Signal sent at %s: %s", time.Now().Format("2006-01-02 15:04:05"), data)

			err := expo.GetExpoClient().Send(data)
			if err != nil {
				logx.Errorf("Failed to send signal: %s, error: %v", data, err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			} else {
				logx.Infof("Signal sent successfully: %s", data)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Message sent successfully"))
			}
		},
	})

	// Webhook API endpoint
	server.AddRoute(rest.Route{
		Method: http.MethodPost,
		Path:   "/webhook",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("webhook received")

			// Parse JSON body
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				logx.Errorf("Invalid JSON payload received from webhook: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid JSON payload"))
				return
			}

			// 记录接收到的webhook数据
			logx.Infof("Webhook received at %s: %v", time.Now().Format("2006-01-02 15:04:05"), payload)

			// Extract message from payload
			var message string
			if msg, ok := payload["message"].(string); ok {
				message = msg
			} else if data, ok := payload["data"].(string); ok {
				message = data
			} else {
				// If no specific message field, use the entire payload as string
				message = fmt.Sprintf("Webhook payload: %v", payload)
			}

			if message == "" {
				logx.Errorf("No message found in webhook payload: %v", payload)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("No message found in payload"))
				return
			}

			// 记录准备发送的信号
			logx.Infof("Webhook signal sent at %s: %s", time.Now().Format("2006-01-02 15:04:05"), message)

			// Send notification
			err := expo.GetExpoClient().Send(message)
			if err != nil {
				logx.Errorf("Failed to send webhook signal: %s, error: %v", message, err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			} else {
				logx.Infof("Webhook signal sent successfully: %s", message)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Webhook processed successfully"))
			}
		},
	})

	// 启动币安 RSI 任务（BTCUSDT 1m/5m/15m/1h，RSI(14)）- 修改为更短周期便于测试
	go rsi.StartBinanceRSI("btcusdt", "2h", 14)
	go rsi.StartBinanceRSI("btcusdt", "4h", 14)
	go rsi.StartBinanceRSI("btcusdt", "1d", 14)
	go rsi.StartBinanceRSI("btcusdt", "1w", 14)
	go rsi.StartBinanceRSI("ethusdt", "2h", 14)
	go rsi.StartBinanceRSI("ethusdt", "4h", 14)
	go rsi.StartBinanceRSI("ethusdt", "1d", 14)
	go rsi.StartBinanceRSI("ethusdt", "1w", 14)

	logx.Infof("Server starting on %s:%d", c.Host, c.Port)
	server.Start()
}
