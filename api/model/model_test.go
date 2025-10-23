package model

import (
	"testing"
	"time"

	"notice/api/config"
	"notice/api/database"
)

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com",
		Port:     5432,
		User:     "wws",
		Password: "Wws5201314",
		DBName:   "notice",
		SSLMode:  "disable",
	}

	err := database.InitDB(cfg)
	if err != nil {
		t.Skipf("跳过测试 - 无法连接数据库: %v", err)
	}

	// 自动迁移所有模型
	err = database.AutoMigrate(
		&Liquidation{},
		&PushToken{},
		&RSISignal{},
		&NewsArticle{},
		&MessageLog{},
		&LiquidationStats{},
	)
	if err != nil {
		t.Fatalf("自动迁移失败: %v", err)
	}
}

// teardownTestDB 清理测试数据库
func teardownTestDB(t *testing.T) {
	db := database.GetDB()
	if db == nil {
		return
	}

	// 删除所有测试表
	db.Migrator().DropTable(
		&Liquidation{},
		&PushToken{},
		&RSISignal{},
		&NewsArticle{},
		&MessageLog{},
		&LiquidationStats{},
	)

	database.CloseDB()
}

// TestLiquidationModel 测试清算记录模型
func TestLiquidationModel(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	db := database.GetDB()

	t.Run("创建清算记录", func(t *testing.T) {
		liq := Liquidation{
			Symbol:       "BTCUSDT",
			Side:         "BUY",
			OrderType:    "LIMIT",
			TimeInForce:  "GTC",
			Quantity:     1.5,
			Price:        50000.0,
			EventTime:    time.Now(),
			ExchangeTime: time.Now(),
		}

		result := db.Create(&liq)
		if result.Error != nil {
			t.Errorf("创建清算记录失败: %v", result.Error)
			return
		}

		// 验证自动计算的字段
		if liq.Value != 1.5*50000.0 {
			t.Errorf("Value = %f, want %f", liq.Value, 1.5*50000.0)
		}

		if !liq.IsLong {
			t.Error("IsLong 应该为 true (BUY)")
		}
	})

	t.Run("查询清算记录", func(t *testing.T) {
		var liq Liquidation
		result := db.Where("symbol = ?", "BTCUSDT").First(&liq)
		if result.Error != nil {
			t.Errorf("查询清算记录失败: %v", result.Error)
			return
		}

		if liq.Symbol != "BTCUSDT" {
			t.Errorf("Symbol = %s, want BTCUSDT", liq.Symbol)
		}
	})

	t.Run("按价格范围查询", func(t *testing.T) {
		var liqs []Liquidation
		result := db.Where("price > ? AND price < ?", 40000.0, 60000.0).Find(&liqs)
		if result.Error != nil {
			t.Errorf("按价格范围查询失败: %v", result.Error)
			return
		}

		if len(liqs) == 0 {
			t.Error("应该找到至少一条记录")
		}
	})
}

// TestPushTokenModel 测试推送令牌模型
func TestPushTokenModel(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	db := database.GetDB()

	t.Run("创建推送令牌", func(t *testing.T) {
		token := PushToken{
			Token:      "ExponentPushToken[xxxxxxxxxxxxxx]",
			DeviceInfo: "iPhone 14 Pro",
			IsActive:   true,
			LastUsed:   time.Now(),
		}

		result := db.Create(&token)
		if result.Error != nil {
			t.Errorf("创建推送令牌失败: %v", result.Error)
		}
	})

	t.Run("唯一性约束", func(t *testing.T) {
		token := PushToken{
			Token:    "ExponentPushToken[xxxxxxxxxxxxxx]",
			IsActive: true,
		}

		result := db.Create(&token)
		if result.Error == nil {
			t.Error("重复的令牌应该失败")
		}
	})

	t.Run("查询活跃令牌", func(t *testing.T) {
		var tokens []PushToken
		result := db.Where("is_active = ?", true).Find(&tokens)
		if result.Error != nil {
			t.Errorf("查询活跃令牌失败: %v", result.Error)
			return
		}

		if len(tokens) == 0 {
			t.Error("应该找到至少一个活跃令牌")
		}
	})
}

// TestRSISignalModel 测试RSI信号模型
func TestRSISignalModel(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	db := database.GetDB()

	t.Run("创建RSI信号", func(t *testing.T) {
		signal := RSISignal{
			Symbol:     "BTCUSDT",
			Interval:   "4h",
			RSIValue:   25.5,
			SignalType: "oversold",
			Price:      48000.0,
			Volume:     1000000.0,
			SignalTime: time.Now(),
			IsSent:     false,
		}

		result := db.Create(&signal)
		if result.Error != nil {
			t.Errorf("创建RSI信号失败: %v", result.Error)
		}
	})

	t.Run("查询未发送的信号", func(t *testing.T) {
		var signals []RSISignal
		result := db.Where("is_sent = ?", false).Find(&signals)
		if result.Error != nil {
			t.Errorf("查询未发送信号失败: %v", result.Error)
			return
		}

		if len(signals) == 0 {
			t.Error("应该找到至少一个未发送的信号")
		}
	})

	t.Run("按信号类型统计", func(t *testing.T) {
		var count int64
		result := db.Model(&RSISignal{}).Where("signal_type = ?", "oversold").Count(&count)
		if result.Error != nil {
			t.Errorf("统计失败: %v", result.Error)
			return
		}

		if count == 0 {
			t.Error("应该有超卖信号")
		}
	})
}

// TestNewsArticleModel 测试新闻文章模型
func TestNewsArticleModel(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	db := database.GetDB()

	t.Run("创建新闻文章", func(t *testing.T) {
		article := NewsArticle{
			Title:       "比特币突破新高",
			Link:        "https://example.com/article1",
			Description: "比特币价格突破历史新高",
			PubDate:     time.Now(),
			Source:      "TheBlockBeats",
			IsSent:      false,
		}

		result := db.Create(&article)
		if result.Error != nil {
			t.Errorf("创建新闻文章失败: %v", result.Error)
		}
	})

	t.Run("链接唯一性", func(t *testing.T) {
		article := NewsArticle{
			Title:   "重复文章",
			Link:    "https://example.com/article1",
			PubDate: time.Now(),
		}

		result := db.Create(&article)
		if result.Error == nil {
			t.Error("重复的链接应该失败")
		}
	})

	t.Run("查询最新文章", func(t *testing.T) {
		var articles []NewsArticle
		result := db.Order("pub_date DESC").Limit(10).Find(&articles)
		if result.Error != nil {
			t.Errorf("查询最新文章失败: %v", result.Error)
			return
		}

		if len(articles) == 0 {
			t.Error("应该找到文章")
		}
	})
}

// TestMessageLogModel 测试消息日志模型
func TestMessageLogModel(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	db := database.GetDB()

	t.Run("创建消息日志", func(t *testing.T) {
		log := MessageLog{
			Message:    "测试推送消息",
			Source:     "manual",
			SendStatus: "success",
			RetryCount: 0,
			SentAt:     time.Now(),
		}

		result := db.Create(&log)
		if result.Error != nil {
			t.Errorf("创建消息日志失败: %v", result.Error)
		}
	})

	t.Run("查询失败消息", func(t *testing.T) {
		// 先创建一条失败消息
		failedLog := MessageLog{
			Message:    "失败的消息",
			Source:     "webhook",
			SendStatus: "failed",
			RetryCount: 3,
			ErrorMsg:   "网络超时",
			SentAt:     time.Now(),
		}
		db.Create(&failedLog)

		var logs []MessageLog
		result := db.Where("send_status = ?", "failed").Find(&logs)
		if result.Error != nil {
			t.Errorf("查询失败消息失败: %v", result.Error)
			return
		}

		if len(logs) == 0 {
			t.Error("应该找到失败的消息")
		}
	})

	t.Run("按来源统计", func(t *testing.T) {
		type SourceCount struct {
			Source string
			Count  int64
		}

		var results []SourceCount
		result := db.Model(&MessageLog{}).
			Select("source, COUNT(*) as count").
			Group("source").
			Find(&results)

		if result.Error != nil {
			t.Errorf("按来源统计失败: %v", result.Error)
			return
		}

		if len(results) == 0 {
			t.Error("应该有统计结果")
		}
	})
}

// TestLiquidationStatsModel 测试清算统计模型
func TestLiquidationStatsModel(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	db := database.GetDB()

	t.Run("创建统计记录", func(t *testing.T) {
		stats := LiquidationStats{
			PeriodType: "hourly",
			PeriodKey:  "2024-01-01-15",
			Count:      100,
			TotalValue: 5000000.0,
			LongCount:  60,
			ShortCount: 40,
			LongValue:  3000000.0,
			ShortValue: 2000000.0,
			StartTime:  time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
			EndTime:    time.Date(2024, 1, 1, 15, 59, 59, 0, time.UTC),
		}

		result := db.Create(&stats)
		if result.Error != nil {
			t.Errorf("创建统计记录失败: %v", result.Error)
		}
	})

	t.Run("周期键唯一性", func(t *testing.T) {
		stats := LiquidationStats{
			PeriodType: "hourly",
			PeriodKey:  "2024-01-01-15",
			Count:      50,
			TotalValue: 1000000.0,
			StartTime:  time.Now(),
		}

		result := db.Create(&stats)
		if result.Error == nil {
			t.Error("重复的周期键应该失败")
		}
	})

	t.Run("查询每日统计", func(t *testing.T) {
		var stats []LiquidationStats
		result := db.Where("period_type = ?", "daily").
			Order("start_time DESC").
			Limit(7).
			Find(&stats)

		if result.Error != nil {
			t.Errorf("查询每日统计失败: %v", result.Error)
		}
	})

	t.Run("计算多空比", func(t *testing.T) {
		var stat LiquidationStats
		result := db.Where("period_key = ?", "2024-01-01-15").First(&stat)
		if result.Error != nil {
			t.Errorf("查询统计失败: %v", result.Error)
			return
		}

		longPercent := float64(stat.LongCount) / float64(stat.Count) * 100
		shortPercent := float64(stat.ShortCount) / float64(stat.Count) * 100

		if longPercent+shortPercent != 100.0 {
			t.Errorf("多空比计算错误: %.2f%% + %.2f%% != 100%%", longPercent, shortPercent)
		}

		t.Logf("多单占比: %.2f%%, 空单占比: %.2f%%", longPercent, shortPercent)
	})
}

// TestComplexQueries 测试复杂查询
func TestComplexQueries(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	db := database.GetDB()

	// 插入测试数据
	now := time.Now()
	for i := 0; i < 10; i++ {
		db.Create(&Liquidation{
			Symbol:       "BTCUSDT",
			Side:         []string{"BUY", "SELL"}[i%2],
			Quantity:     float64(i + 1),
			Price:        50000.0 + float64(i*1000),
			EventTime:    now.Add(time.Duration(-i) * time.Hour),
			ExchangeTime: now.Add(time.Duration(-i) * time.Hour),
		})
	}

	t.Run("按时间范围查询", func(t *testing.T) {
		var liqs []Liquidation
		result := db.Where("event_time BETWEEN ? AND ?",
			now.Add(-5*time.Hour), now).Find(&liqs)

		if result.Error != nil {
			t.Errorf("按时间范围查询失败: %v", result.Error)
			return
		}

		if len(liqs) != 6 {
			t.Errorf("查询结果数量 = %d, want 6", len(liqs))
		}
	})

	t.Run("聚合查询", func(t *testing.T) {
		type AggResult struct {
			TotalCount int64
			TotalValue float64
			AvgPrice   float64
		}

		var result AggResult
		db.Model(&Liquidation{}).
			Select("COUNT(*) as total_count, SUM(value) as total_value, AVG(price) as avg_price").
			Scan(&result)

		if result.TotalCount != 10 {
			t.Errorf("TotalCount = %d, want 10", result.TotalCount)
		}

		t.Logf("聚合结果 - 总数: %d, 总价值: %.2f, 平均价格: %.2f",
			result.TotalCount, result.TotalValue, result.AvgPrice)
	})

	t.Run("多条件组合查询", func(t *testing.T) {
		var liqs []Liquidation
		result := db.Where("symbol = ? AND side = ? AND price > ?",
			"BTCUSDT", "BUY", 52000.0).Find(&liqs)

		if result.Error != nil {
			t.Errorf("多条件查询失败: %v", result.Error)
			return
		}

		for _, liq := range liqs {
			if liq.Side != "BUY" || liq.Price <= 52000.0 {
				t.Errorf("查询条件不匹配: Side=%s, Price=%.2f", liq.Side, liq.Price)
			}
		}
	})
}

// BenchmarkModelInsert 性能测试 - 批量插入
func BenchmarkModelInsert(b *testing.B) {
	cfg := config.DatabaseConfig{
		Host:     "pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com",
		Port:     5432,
		User:     "wws",
		Password: "Wws5201314",
		DBName:   "notice",
		SSLMode:  "disable",
	}

	err := database.InitDB(cfg)
	if err != nil {
		b.Skipf("跳过测试 - 无法连接数据库: %v", err)
		return
	}
	defer database.CloseDB()

	database.AutoMigrate(&Liquidation{})
	defer database.GetDB().Migrator().DropTable(&Liquidation{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		liq := Liquidation{
			Symbol:    "BTCUSDT",
			Side:      "BUY",
			Quantity:  1.0,
			Price:     50000.0,
			EventTime: time.Now(),
		}
		database.GetDB().Create(&liq)
	}
}
