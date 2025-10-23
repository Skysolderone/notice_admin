package model

import (
	"time"

	"gorm.io/gorm"
)

// Liquidation 清算记录模型
type Liquidation struct {
	gorm.Model
	Symbol       string    `gorm:"size:20;not null;index" json:"symbol"`             // 交易对符号
	Side         string    `gorm:"size:10;not null" json:"side"`                     // BUY/SELL
	OrderType    string    `gorm:"size:20" json:"order_type"`                        // 订单类型
	TimeInForce  string    `gorm:"size:20" json:"time_in_force"`                     // 有效期类型
	Quantity     float64   `gorm:"type:decimal(20,8);not null" json:"quantity"`      // 数量
	Price        float64   `gorm:"type:decimal(20,8);not null;index" json:"price"`   // 价格
	Value        float64   `gorm:"type:decimal(20,8);not null;index" json:"value"`   // 价值(数量*价格)
	EventTime    time.Time `gorm:"not null;index" json:"event_time"`                 // 事件时间
	IsLong       bool      `gorm:"not null;index" json:"is_long"`                    // 是否多单
	ExchangeTime time.Time `gorm:"index" json:"exchange_time"`                       // 交易所时间
}

// TableName 指定表名
func (Liquidation) TableName() string {
	return "liquidations"
}

// BeforeCreate GORM 钩子 - 创建前
func (l *Liquidation) BeforeCreate(tx *gorm.DB) error {
	// 计算价值
	l.Value = l.Quantity * l.Price
	// 判断是否多单
	l.IsLong = l.Side == "BUY"
	return nil
}

// PushToken Expo推送令牌模型
type PushToken struct {
	gorm.Model
	Token      string    `gorm:"size:500;not null;uniqueIndex" json:"token"` // Expo推送令牌
	DeviceInfo string    `gorm:"size:500" json:"device_info"`                // 设备信息
	IsActive   bool      `gorm:"default:true;index" json:"is_active"`        // 是否活跃
	LastUsed   time.Time `gorm:"index" json:"last_used"`                     // 最后使用时间
}

// TableName 指定表名
func (PushToken) TableName() string {
	return "push_tokens"
}

// RSISignal RSI信号记录模型
type RSISignal struct {
	gorm.Model
	Symbol     string    `gorm:"size:20;not null;index" json:"symbol"`          // 交易对符号
	Interval   string    `gorm:"size:10;not null;index" json:"interval"`        // 时间周期
	RSIValue   float64   `gorm:"type:decimal(10,2);not null" json:"rsi_value"`  // RSI值
	SignalType string    `gorm:"size:20;not null;index" json:"signal_type"`     // 信号类型: oversold/overbought
	Price      float64   `gorm:"type:decimal(20,8);not null" json:"price"`      // 当时价格
	Volume     float64   `gorm:"type:decimal(20,8)" json:"volume"`              // 成交量
	SignalTime time.Time `gorm:"not null;index" json:"signal_time"`             // 信号时间
	IsSent     bool      `gorm:"default:false;index" json:"is_sent"`            // 是否已发送通知
}

// TableName 指定表名
func (RSISignal) TableName() string {
	return "rsi_signals"
}

// NewsArticle 新闻文章模型
type NewsArticle struct {
	gorm.Model
	Title       string    `gorm:"size:500;not null" json:"title"`             // 标题
	Link        string    `gorm:"size:1000;not null;uniqueIndex" json:"link"` // 链接
	Description string    `gorm:"type:text" json:"description"`               // 描述
	PubDate     time.Time `gorm:"not null;index" json:"pub_date"`             // 发布日期
	Source      string    `gorm:"size:100;index" json:"source"`               // 来源
	IsSent      bool      `gorm:"default:false;index" json:"is_sent"`         // 是否已发送通知
}

// TableName 指定表名
func (NewsArticle) TableName() string {
	return "news_articles"
}

// MessageLog 消息日志模型
type MessageLog struct {
	gorm.Model
	Message    string    `gorm:"type:text;not null" json:"message"`       // 消息内容
	Source     string    `gorm:"size:50;not null;index" json:"source"`    // 来源: manual/webhook/rsi/liquidation/news
	SendStatus string    `gorm:"size:20;not null;index" json:"send_status"` // 发送状态: success/failed/pending
	RetryCount int       `gorm:"default:0" json:"retry_count"`            // 重试次数
	ErrorMsg   string    `gorm:"type:text" json:"error_msg"`              // 错误信息
	SentAt     time.Time `gorm:"index" json:"sent_at"`                    // 发送时间
}

// TableName 指定表名
func (MessageLog) TableName() string {
	return "message_logs"
}

// LiquidationStats 清算统计模型
type LiquidationStats struct {
	gorm.Model
	PeriodType  string    `gorm:"size:20;not null;index" json:"period_type"`     // 统计周期类型: hourly/daily
	PeriodKey   string    `gorm:"size:50;not null;uniqueIndex" json:"period_key"` // 周期键: 2024-01-01-15
	Count       int64     `gorm:"not null" json:"count"`                         // 清算总数
	TotalValue  float64   `gorm:"type:decimal(20,2);not null" json:"total_value"` // 总价值
	LongCount   int64     `gorm:"not null" json:"long_count"`                    // 多单数量
	ShortCount  int64     `gorm:"not null" json:"short_count"`                   // 空单数量
	LongValue   float64   `gorm:"type:decimal(20,2);not null" json:"long_value"`  // 多单价值
	ShortValue  float64   `gorm:"type:decimal(20,2);not null" json:"short_value"` // 空单价值
	StartTime   time.Time `gorm:"not null;index" json:"start_time"`              // 开始时间
	EndTime     time.Time `gorm:"index" json:"end_time"`                         // 结束时间
}

// TableName 指定表名
func (LiquidationStats) TableName() string {
	return "liquidation_stats"
}
