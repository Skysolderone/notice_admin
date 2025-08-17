package rsi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"notice/rpc/expo"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

// binance futures kline event payload
type binanceKline struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	K         struct {
		OpenTime                 int64  `json:"t"`
		CloseTime                int64  `json:"T"`
		Symbol                   string `json:"s"`
		Interval                 string `json:"i"`
		FirstTradeId             int64  `json:"f"`
		LastTradeId              int64  `json:"L"`
		Open                     string `json:"o"`
		Close                    string `json:"c"`
		High                     string `json:"h"`
		Low                      string `json:"l"`
		Volume                   string `json:"v"`
		NumberOfTrades           int64  `json:"n"`
		IsClosed                 bool   `json:"x"`
		QuoteAssetVolume         string `json:"q"`
		TakerBuyBaseAssetVolume  string `json:"V"`
		TakerBuyQuoteAssetVolume string `json:"Q"`
		Ignore                   string `json:"B"`
	} `json:"k"`
}

// simple string to float parser (Binance numbers are decimal strings)
func toFloat(s string) float64 { f, _ := strconv.ParseFloat(s, 64); return f }

// RSI calculator using Wilder's smoothing
type rsiCalc struct {
	period   int
	initDone bool
	count    int
	prev     float64
	avgGain  float64
	avgLoss  float64
	sumGain  float64
	sumLoss  float64
}

func newRSI(period int) *rsiCalc { return &rsiCalc{period: period} }

func (r *rsiCalc) add(close float64) (rsi float64, ok bool) {
	if r.count == 0 {
		r.prev = close
		r.count = 1
		return 0, false
	}
	change := close - r.prev
	r.prev = close
	gain := math.Max(change, 0)
	loss := math.Max(-change, 0)

	if !r.initDone {
		r.sumGain += gain
		r.sumLoss += loss
		r.count++
		if r.count > r.period {
			r.avgGain = r.sumGain / float64(r.period)
			r.avgLoss = r.sumLoss / float64(r.period)
			r.initDone = true
			return r.compute(), true
		}
		return 0, false
	}

	r.avgGain = (r.avgGain*float64(r.period-1) + gain) / float64(r.period)
	r.avgLoss = (r.avgLoss*float64(r.period-1) + loss) / float64(r.period)
	return r.compute(), true
}

func (r *rsiCalc) compute() float64 {
	if r.avgLoss == 0 {
		if r.avgGain == 0 {
			return 50
		}
		return 100
	}
	rs := r.avgGain / r.avgLoss
	return 100 - 100/(1+rs)
}

// ---- REST 历史K线拉取（USDT 永续：fapi/v1/klines） ----
type histCandle struct {
	close     float64
	closeTime int64
}

func fetchHistoricalKlines(symbol, interval string, limit int) ([]histCandle, error) {
	// Binance Futures REST (USDT-M): https://fapi.binance.com/fapi/v1/klines
	// Response: [[openTime, open, high, low, close, volume, closeTime, ...], ...]
	sym := strings.ToUpper(symbol)
	if limit <= 0 || limit > 1500 {
		limit = 1000
	}
	url := fmt.Sprintf("https://fapi.binance.com/fapi/v1/klines?symbol=%s&interval=%s&limit=%d", sym, interval, limit)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance klines status=%d body=%s", resp.StatusCode, string(b))
	}
	var raw [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	result := make([]histCandle, 0, len(raw))
	for _, it := range raw {
		if len(it) < 7 {
			continue
		}
		// close at index 4, closeTime at index 6
		closeStr, _ := it[4].(string)
		closeTime, _ := it[6].(float64)
		result = append(result, histCandle{
			close:     toFloat(closeStr),
			closeTime: int64(closeTime),
		})
	}
	return result, nil
}

// StartBinanceRSI connects to Binance futures kline stream, computes RSI on candle close, and broadcasts
func StartBinanceRSI(symbol, interval string, period int) {
	sym := strings.ToLower(symbol)
	url := fmt.Sprintf("wss://fstream.binance.com/ws/%s@kline_%s", sym, interval)
	backoff := time.Second
	if backoff < time.Second {
		backoff = time.Second
	}

	for {
		// 先拉取历史，完成 RSI 预热
		if hist, err := fetchHistoricalKlines(symbol, interval, 1000); err == nil {
			rsi := newRSI(period)
			var lastVal float64
			var ready bool
			var lastTs int64
			for _, c := range hist {
				lastVal, ready = rsi.add(c.close)
				lastTs = c.closeTime
			}
			if ready {
				ts := time.UnixMilli(lastTs).Format(time.RFC3339)
				// broker.Broadcast(fmt.Sprintf("[RSI] warmup done %s %s RSI(%d)=%.2f @ %s", strings.ToUpper(symbol), interval, period, lastVal, ts))
				expo.GetExpoClient().Send(fmt.Sprintf("[RSI] warmup done %s %s RSI(%d)=%.2f @ %s", strings.ToUpper(symbol), interval, period, lastVal, ts))
			} else {
				// broker.Broadcast(fmt.Sprintf("[RSI] warmup pending %s %s need more candles", strings.ToUpper(symbol), interval))
				expo.GetExpoClient().Send(fmt.Sprintf("[RSI] warmup pending %s %s need more candles", strings.ToUpper(symbol), interval))
			}
		} else {
			// broker.Broadcast(fmt.Sprintf("[RSI] warmup error: %v", err))
			expo.GetExpoClient().Send(fmt.Sprintf("[RSI] warmup error: %v", err))
		}

		dialer := websocket.Dialer{
			HandshakeTimeout:  10 * time.Second,
			EnableCompression: true,
		}
		conn, _, err := dialer.Dial(url, nil)
		if err != nil {
			// broker.Broadcast(fmt.Sprintf("[RSI] connect error: %v", err))
			time.Sleep(backoff)
			backoff = minDuration(backoff*2, 60*time.Second)
			continue
		}

		rsi := newRSI(period)
		// 若 warmup 期间无法获得足够K线，WS 收到的后续 close 会逐步完成初始化
		// broker.Broadcast(fmt.Sprintf("[RSI] connected %s %s period=%d", strings.ToUpper(symbol), interval, period))
		expo.GetExpoClient().Send(fmt.Sprintf("[RSI] connected %s %s period=%d", strings.ToUpper(symbol), interval, period))
		// reset backoff on successful connect
		backoff = time.Second

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				// broker.Broadcast(fmt.Sprintf("[RSI] read error: %v", err))
				_ = conn.Close()
				break
			}

			var ev binanceKline
			if err := json.Unmarshal(data, &ev); err != nil {
				continue
			}
			// only act on candle close to avoid noise
			if !ev.K.IsClosed {
				continue
			}
			closePrice := toFloat(ev.K.Close)
			value, ready := rsi.add(closePrice)
			if !ready {
				continue
			}
			ts := time.UnixMilli(ev.K.CloseTime).Format(time.RFC3339)
			msg := fmt.Sprintf("%s %s close=%.2f RSI(%d)=%.2f @ %s", strings.ToUpper(symbol), interval, closePrice, period, value, ts)
			expo.GetExpoClient().Send(msg)
		}
		// loop to reconnect
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// Client 表示单个 SSE 客户端连接
type Client struct {
	userID string
	send   chan string // 向该客户端发送消息的缓冲队列
}

// Broker 负责管理所有客户端和广播
type Broker struct {
	register   chan *Client
	unregister chan *Client
	broadcast  chan string
	clients    map[*Client]struct{}
}

func NewBroker() *Broker {
	b := &Broker{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan string, 1024),
		clients:    make(map[*Client]struct{}),
	}
	go b.run()
	return b
}

func (b *Broker) run() {
	for {
		select {
		case c := <-b.register:
			b.clients[c] = struct{}{}
		case c := <-b.unregister:
			if _, ok := b.clients[c]; ok {
				delete(b.clients, c)
				close(c.send)
			}
		case msg := <-b.broadcast:
			// 非阻塞广播，慢客户端将被剔除
			for c := range b.clients {
				select {
				case c.send <- msg:
				default:
					// 客户端拥塞，移除
					delete(b.clients, c)
					close(c.send)
				}
			}
		}
	}
}

func (b *Broker) Register(c *Client) {
	b.register <- c
	logx.Infof("client %s registered", c.userID)
}

func (b *Broker) Unregister(c *Client) {
	b.unregister <- c
	logx.Infof("client %s unregistered", c.userID)
}

func (b *Broker) Broadcast(msg string) {
	select {
	case b.broadcast <- msg:
	default:
		// 广播通道已满时丢弃最旧策略不可行，简单丢弃本条以保护系统
	}
}

// SseHandler 使用 Broker 提供 SSE 服务
type SseHandler struct {
	broker *Broker
}

func NewSseHandler() *SseHandler {
	return &SseHandler{broker: NewBroker()}
}

// Serve 处理 SSE 连接
func (h *SseHandler) Serve(w http.ResponseWriter, r *http.Request) {
	logx.Infof("New SSE connection from %s, User-Agent: %s", r.RemoteAddr, r.UserAgent())

	// 设置 SSE 必需的 HTTP 头
	// for versions > v1.8.1, no need to add 3 lines below
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Connection", "keep-alive")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// 为每个客户端创建带缓冲的发送队列

	client := &Client{userID: uuid.New().String(), send: make(chan string, 64)}
	h.broker.Register(client)
	defer h.broker.Unregister(client)

	// 首次连接后，单播返回该连接的 ID
	client.send <- fmt.Sprintf("CONNECTED %s", client.userID)

	// 心跳保持连接活性
	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	for {
		select {
		case msg, ok := <-client.send:
			if !ok {
				logx.Infof("Client %s send channel closed", client.userID)
				return
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", msg); err != nil {
				logx.Infof("Client %s disconnected due to write error: %v", client.userID, err)
				return
			}
			flusher.Flush()
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				logx.Infof("Client %s disconnected due to heartbeat write error: %v", client.userID, err)
				return
			}
			flusher.Flush()
		case <-ctx.Done():
			logx.Infof("Client %s disconnected due to context cancellation", client.userID)
			return
		}
	}
}

// SimulateEvents 模拟周期性事件
func (h *SseHandler) SimulateEvents() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		message := fmt.Sprintf("Server time: %s", time.Now().Format(time.RFC3339))
		h.broker.Broadcast(message)
	}
}
