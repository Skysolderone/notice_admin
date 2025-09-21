package margin_push

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"notice/api/expo"

	"github.com/adshao/go-binance/v2/futures"
)

type Stats struct {
	mu sync.RWMutex
	// 按时间段统计的清算数据
	hourlyStats map[string]*PeriodStats // key: "2024-01-01-15" (年-月-日-时)
	dailyStats  map[string]*PeriodStats // key: "2024-01-01" (年-月-日)
	// 程序启动时间
	startTime time.Time
}

type PeriodStats struct {
	Count      int64     `json:"count"`
	Quantity   float64   `json:"quantity"`
	Value      float64   `json:"value"`
	LongCount  int64     `json:"long_count"`  // 多单数量
	ShortCount int64     `json:"short_count"` // 空单数量
	LongValue  float64   `json:"long_value"`  // 多单价值
	ShortValue float64   `json:"short_value"` // 空单价值
	StartTime  time.Time `json:"start_time"`
}

func (s *Stats) AddLiquidation(event *futures.WsLiquidationOrderEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()

	// 生成时间键
	hourKey := now.Format("2006-01-02-15") // 年-月-日-时
	dayKey := now.Format("2006-01-02")     // 年-月-日

	// 确保map已初始化
	if s.hourlyStats == nil {
		s.hourlyStats = make(map[string]*PeriodStats)
	}
	if s.dailyStats == nil {
		s.dailyStats = make(map[string]*PeriodStats)
	}

	// 初始化小时统计
	if s.hourlyStats[hourKey] == nil {
		s.hourlyStats[hourKey] = &PeriodStats{
			StartTime: now.Truncate(time.Hour),
		}
	}

	// 初始化日统计
	if s.dailyStats[dayKey] == nil {
		s.dailyStats[dayKey] = &PeriodStats{
			StartTime: now.Truncate(24 * time.Hour),
		}
	}

	// 解析数量和价值
	quantity, _ := strconv.ParseFloat(event.LiquidationOrder.OrigQuantity, 64)
	price, _ := strconv.ParseFloat(event.LiquidationOrder.Price, 64)
	value := quantity * price

	// 判断多单还是空单
	isLong := event.LiquidationOrder.Side == "BUY"

	// 更新统计数据
	s.hourlyStats[hourKey].Count++
	s.hourlyStats[hourKey].Quantity += quantity
	s.hourlyStats[hourKey].Value += value

	s.dailyStats[dayKey].Count++
	s.dailyStats[dayKey].Quantity += quantity
	s.dailyStats[dayKey].Value += value

	// 更新多空统计
	if isLong {
		s.hourlyStats[hourKey].LongCount++
		s.hourlyStats[hourKey].LongValue += value
		s.dailyStats[dayKey].LongCount++
		s.dailyStats[dayKey].LongValue += value
	} else {
		s.hourlyStats[hourKey].ShortCount++
		s.hourlyStats[hourKey].ShortValue += value
		s.dailyStats[dayKey].ShortCount++
		s.dailyStats[dayKey].ShortValue += value
	}
}

// 获取当前小时统计
func (s *Stats) GetCurrentHourStats() (int64, float64, float64, int64, int64, float64, float64, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now().UTC()
	hourKey := now.Format("2006-01-02-15")

	if stats, exists := s.hourlyStats[hourKey]; exists {
		return stats.Count, stats.Quantity, stats.Value, stats.LongCount, stats.ShortCount, stats.LongValue, stats.ShortValue, hourKey
	}
	return 0, 0, 0, 0, 0, 0, 0, hourKey
}

// 获取指定小时数的统计总和
func (s *Stats) GetPeriodStats(hours int) (int64, float64, float64, int64, int64, float64, float64, []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now().UTC()
	var totalCount, totalLongCount, totalShortCount int64
	var totalQuantity, totalValue, totalLongValue, totalShortValue float64
	var periods []string

	for i := 0; i < hours; i++ {
		targetTime := now.Add(-time.Duration(i) * time.Hour)
		hourKey := targetTime.Format("2006-01-02-15")

		if stats, exists := s.hourlyStats[hourKey]; exists {
			totalCount += stats.Count
			totalQuantity += stats.Quantity
			totalValue += stats.Value
			totalLongCount += stats.LongCount
			totalShortCount += stats.ShortCount
			totalLongValue += stats.LongValue
			totalShortValue += stats.ShortValue
			periods = append(periods, hourKey)
		}
	}

	return totalCount, totalQuantity, totalValue, totalLongCount, totalShortCount, totalLongValue, totalShortValue, periods
}

// 获取当天从UTC零点开始的统计
func (s *Stats) GetTodayStats() (int64, float64, float64, int64, int64, float64, float64, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now().UTC()
	dayKey := now.Format("2006-01-02")

	if stats, exists := s.dailyStats[dayKey]; exists {
		return stats.Count, stats.Quantity, stats.Value, stats.LongCount, stats.ShortCount, stats.LongValue, stats.ShortValue, dayKey
	}
	return 0, 0, 0, 0, 0, 0, 0, dayKey
}

// 清理旧的小时数据（保留最近48小时）
func (s *Stats) CleanOldHourlyData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	cutoffTime := now.Add(-48 * time.Hour)

	for key, stats := range s.hourlyStats {
		if stats.StartTime.Before(cutoffTime) {
			delete(s.hourlyStats, key)
		}
	}
}

// 清理旧的日数据（保留最近7天）
func (s *Stats) CleanOldDailyData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	cutoffTime := now.Add(-7 * 24 * time.Hour)

	for key, stats := range s.dailyStats {
		if stats.StartTime.Before(cutoffTime) {
			delete(s.dailyStats, key)
		}
	}
}

func NewStats() *Stats {
	return &Stats{
		startTime:   time.Now().UTC(),
		hourlyStats: make(map[string]*PeriodStats),
		dailyStats:  make(map[string]*PeriodStats),
	}
}

var globalStats = NewStats()

// 发送启动通知
func sendStartupNotification() {
	message := fmt.Sprintf("🚀 清算监控系统启动\n"+
		"启动时间: %s\n"+
		"监控状态: 已开始监听币安清算订单\n"+
		"统计周期: 1h/4h/8h/24h\n"+
		"推送功能: 已激活",
		time.Now().UTC().Format("2006-01-02 15:04:05 UTC"))

	// 发送启动通知推送
	go func() {
		// 等待一小段时间确保expo客户端已初始化
		time.Sleep(2 * time.Second)
		
		client := expo.GetExpoClient()
		if client != nil && client.GetTokenCount() > 0 {
			err := client.SendWithCustomTitle(message, "清算监控系统")
			if err != nil {
				log.Printf("发送启动通知推送失败: %v", err)
			} else {
				log.Printf("启动通知推送发送成功")
			}
		} else {
			log.Printf("跳过启动通知推送: 没有注册的推送token")
		}
	}()
}

// 配置参数
type PushConfig struct {
	EnableLiquidationAlert bool    // 是否启用清算告警
	LargeOrderThreshold    float64 // 大单告警阈值(USDT)
	EnableStatsReport      bool    // 是否启用统计报告
	MaxAlertsPerHour       int     // 每小时最大告警数量
}

var pushConfig = &PushConfig{
	EnableLiquidationAlert: true,
	LargeOrderThreshold:    10000, // 1万USDT
	EnableStatsReport:      true,
	MaxAlertsPerHour:       20, // 限制每小时最多20条告警
}

// 告警计数器
type AlertCounter struct {
	mu         sync.RWMutex
	hourlyData map[string]int // key: hour, value: count
}

var alertCounter = &AlertCounter{
	hourlyData: make(map[string]int),
}

// 检查是否可以发送告警
func (ac *AlertCounter) CanSendAlert() bool {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	hourKey := time.Now().UTC().Format("2006-01-02-15")

	if ac.hourlyData[hourKey] >= pushConfig.MaxAlertsPerHour {
		return false
	}

	ac.hourlyData[hourKey]++
	return true
}

// 清理旧的计数数据
func (ac *AlertCounter) CleanOldData() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	cutoffTime := time.Now().UTC().Add(-24 * time.Hour)
	cutoffHour := cutoffTime.Format("2006-01-02-15")

	for key := range ac.hourlyData {
		if key < cutoffHour {
			delete(ac.hourlyData, key)
		}
	}
}

func logStats(period string) {
	now := time.Now().UTC()

	switch period {
	case "1小时":
		// 显示刚刚结束的小时统计
		count, quantity, value, longCount, shortCount, longValue, shortValue, timeKey := globalStats.GetCurrentHourStats()
		log.Printf("[%s统计] UTC时间: %s, 清算订单数: %d, 总数量: %.4f, 总价值: %.4f",
			period, timeKey, count, quantity, value)
		log.Printf("   多单: %d笔 价值: %.2f USDT, 空单: %d笔 价值: %.2f USDT",
			longCount, longValue, shortCount, shortValue)

		// 发送1小时统计推送通知
		sendStatsReport(period, count, quantity, value, longCount, shortCount, longValue, shortValue, timeKey)

	case "4小时":
		// 显示过去4小时的总统计（UTC固定时间点：0,4,8,12,16,20点）
		count, quantity, value, longCount, shortCount, longValue, shortValue, periods := globalStats.GetPeriodStats(4)
		log.Printf("[%s统计] UTC时间: %s, 过去4小时清算订单数: %d, 总数量: %.4f, 总价值: %.4f",
			period, now.Format("2006-01-02 15:04"), count, quantity, value)
		log.Printf("   多单: %d笔 价值: %.2f USDT, 空单: %d笔 价值: %.2f USDT",
			longCount, longValue, shortCount, shortValue)
		log.Printf("   包含时段: %v", periods)

		// 发送4小时统计推送通知
		sendStatsReport(period, count, quantity, value, longCount, shortCount, longValue, shortValue, now.Format("2006-01-02 15:04"))

	case "8小时":
		// 显示过去8小时的总统计（UTC固定时间点：0,8,16点）
		count, quantity, value, longCount, shortCount, longValue, shortValue, periods := globalStats.GetPeriodStats(8)
		log.Printf("[%s统计] UTC时间: %s, 过去8小时清算订单数: %d, 总数量: %.4f, 总价值: %.4f",
			period, now.Format("2006-01-02 15:04"), count, quantity, value)
		log.Printf("   多单: %d笔 价值: %.2f USDT, 空单: %d笔 价值: %.2f USDT",
			longCount, longValue, shortCount, shortValue)
		log.Printf("   包含时段: %v", periods)

		// 发送8小时统计推送通知
		sendStatsReport(period, count, quantity, value, longCount, shortCount, longValue, shortValue, now.Format("2006-01-02 15:04"))

	case "24小时":
		// 显示当天从UTC零点开始的统计
		count, quantity, value, longCount, shortCount, longValue, shortValue, timeKey := globalStats.GetTodayStats()
		log.Printf("[%s统计] UTC日期: %s, 从零点开始清算订单数: %d, 总数量: %.4f, 总价值: %.4f",
			period, timeKey, count, quantity, value)
		log.Printf("   多单: %d笔 价值: %.2f USDT, 空单: %d笔 价值: %.2f USDT",
			longCount, longValue, shortCount, shortValue)

		// 发送24小时统计推送通知
		sendStatsReport(period, count, quantity, value, longCount, shortCount, longValue, shortValue, timeKey)

		// UTC零点时清理旧数据
		if now.Hour() == 0 && now.Minute() < 5 {
			globalStats.CleanOldHourlyData()
			globalStats.CleanOldDailyData()
			alertCounter.CleanOldData()
			log.Printf("   已清理旧数据")
		}
	}
}

// 发送统计报告推送消息
func sendStatsReport(period string, count int64, _ /*quantity*/, value float64, longCount, shortCount int64, longValue, shortValue float64, timeKey string) {
	// 只有在有清算数据时才发送推送
	if count == 0 {
		return
	}

	// 计算多空比例
	longPercent := float64(0)
	shortPercent := float64(0)
	if count > 0 {
		longPercent = float64(longCount) / float64(count) * 100
		shortPercent = float64(shortCount) / float64(count) * 100
	}

	// 构建统计报告消息
	message := fmt.Sprintf("📊 %s清算统计报告\n"+
		"时间: %s\n"+
		"清算订单数: %d\n"+
		"总价值: %.2f USDT\n"+
		"━━━━━━━━━━━━━━━━\n"+
		"🟢 多单清算: %d笔 (%.1f%%)\n"+
		"    价值: %.2f USDT\n"+
		"🔴 空单清算: %d笔 (%.1f%%)\n"+
		"    价值: %.2f USDT",
		period,
		timeKey,
		count,
		value,
		longCount, longPercent,
		longValue,
		shortCount, shortPercent,
		shortValue)

	// 发送推送通知
	go func() {
		client := expo.GetExpoClient()
		if client != nil && client.GetTokenCount() > 0 {
			err := client.Send(message)
			if err != nil {
				log.Printf("发送统计报告推送失败: %v", err)
			} else {
				log.Printf("统计报告推送发送成功: %s", period)
			}
		}
	}()
}

// 计算下一个UTC固定时间点
func getNextUTCTime(hour int) time.Time {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.UTC)

	// 如果今天的时间点已过，则计算明天的
	if next.Before(now) || next.Equal(now) {
		next = next.Add(24 * time.Hour)
	}

	return next
}

// 计算下一个整点时间
func getNextHour() time.Time {
	now := time.Now().UTC()
	next := now.Truncate(time.Hour).Add(time.Hour)
	return next
}

func startStatsTimers() {
	log.Printf("启动基于UTC时间的定时器...")

	// 1小时定时器 - 每个整点触发
	go func() {
		for {
			next := getNextHour()
			sleepDuration := time.Until(next)
			log.Printf("1小时定时器将在 %v 后触发 (UTC: %s)", sleepDuration, next.Format("2006-01-02 15:04:05"))

			time.Sleep(sleepDuration)
			logStats("1小时")
		}
	}()

	// 4小时定时器 - UTC 0,4,8,12,16,20点触发
	go func() {
		hours := []int{0, 4, 8, 12, 16, 20}
		for {
			var nextTime time.Time

			// 找到下一个4小时触发点
			for _, h := range hours {
				candidate := getNextUTCTime(h)
				if nextTime.IsZero() || candidate.Before(nextTime) {
					nextTime = candidate
				}
			}

			sleepDuration := time.Until(nextTime)
			log.Printf("4小时定时器将在 %v 后触发 (UTC: %s)", sleepDuration, nextTime.Format("2006-01-02 15:04:05"))

			time.Sleep(sleepDuration)
			logStats("4小时")
		}
	}()

	// 8小时定时器 - UTC 0,8,16点触发
	go func() {
		hours := []int{0, 8, 16}
		for {
			var nextTime time.Time

			// 找到下一个8小时触发点
			for _, h := range hours {
				candidate := getNextUTCTime(h)
				if nextTime.IsZero() || candidate.Before(nextTime) {
					nextTime = candidate
				}
			}

			sleepDuration := time.Until(nextTime)
			log.Printf("8小时定时器将在 %v 后触发 (UTC: %s)", sleepDuration, nextTime.Format("2006-01-02 15:04:05"))

			time.Sleep(sleepDuration)
			logStats("8小时")
		}
	}()

	// 24小时定时器 - UTC 0点触发
	go func() {
		for {
			nextTime := getNextUTCTime(0)
			sleepDuration := time.Until(nextTime)
			log.Printf("24小时定时器将在 %v 后触发 (UTC: %s)", sleepDuration, nextTime.Format("2006-01-02 15:04:05"))

			time.Sleep(sleepDuration)
			logStats("24小时")
		}
	}()
}

func ForceReceive() {
	// 发送启动通知
	sendStartupNotification()
	
	// 启动统计定时器
	startStatsTimers()

	wsLiquidationOrderHandler := func(event *futures.WsLiquidationOrderEvent) {
		// fmt.Println(event)
		// 记录统计数据
		globalStats.AddLiquidation(event)
	}
	errHandler := func(err error) {
		log.Printf("WebSocket错误: %v", err)
	}

	log.Println("开始监听清算订单...")

	doneC, _, err := futures.WsAllLiquidationOrderServe(wsLiquidationOrderHandler, errHandler)
	if err != nil {
		log.Printf("启动WebSocket失败: %v", err)
		return
	}

	// 移除自动停止，让程序持续运行
	// go func() {
	//     time.Sleep(5 * time.Second)
	//     stopC <- struct{}{}
	// }()

	// 程序将持续运行直到手动停止
	<-doneC
}

// func main() {
// 	log.SetFlags(log.LstdFlags | log.Lshortfile)
// 	log.Println("启动清算订单监控程序...")
// 	ForceReceive()
// }
