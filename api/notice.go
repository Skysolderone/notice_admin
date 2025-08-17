package main

import (
	"fmt"
	"net/http"

	"notice/rpc/expo"

	"notice/api/rsi"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

func main() {
	expo.NewClient()
	// 创建 go-zero REST 服务，集成静态文件服务
	server := rest.MustNewServer(rest.RestConf{
		Host: "0.0.0.0",
		Port: 5555,
	})
	defer server.Stop()

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
		Path:   "/notice_token",
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
			err := expo.GetExpoClient().Send(data)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Message sent successfully"))
			}
		},
	})

	// 启动币安 RSI 任务（BTCUSDT 4h/1d/1w/1M，RSI(14)）
	go rsi.StartBinanceRSI("btcusdt", "4h", 14)
	go rsi.StartBinanceRSI("btcusdt", "1d", 14)
	go rsi.StartBinanceRSI("btcusdt", "1w", 14)
	go rsi.StartBinanceRSI("btcusdt", "1M", 14)

	logx.Info("Server starting on :5555")
	server.Start()
}
