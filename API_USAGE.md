# 消息存储和查询 API 使用文档

## 概述

本文档介绍加密货币通知系统的消息存储和查询 API。系统会自动保存所有发送的通知消息，并提供多种查询方式供客户端应用调用。

## 基本信息

- **服务器地址**: `http://localhost:5555`
- **API前缀**: `/notice`
- **响应格式**: JSON
- **支持方法**: GET/POST
- **认证方式**: 无需认证

## API 接口

### 推送通知相关接口

#### 添加推送令牌
```
POST /notice/notice_token
```

#### 获取令牌统计
```
GET /notice/notice_token/stats
```

#### 发送手动通知
```
POST /notice/notice/query
```

#### Webhook接收
```
POST /notice/webhook
```

### 消息查询相关接口

### 1. 获取消息历史记录

#### 接口地址
```
GET /notice/messages
```

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| limit | int | 否 | 限制返回的消息数量，默认返回所有 | 50 |
| source | string | 否 | 按消息来源过滤 | rsi, liquidation, news, manual, webhook |

#### 请求示例

```bash
# 获取所有消息
curl http://localhost:5555/notice/messages

# 获取最近10条消息
curl http://localhost:5555/notice/messages?limit=10

# 获取RSI相关的消息
curl http://localhost:5555/notice/messages?source=rsi

# 获取清算相关的消息
curl http://localhost:5555/notice/messages?source=liquidation

# 组合查询：获取最近20条RSI消息
curl http://localhost:5555/notice/messages?limit=20&source=rsi
```

#### 响应示例

```json
{
  "success": true,
  "count": 3,
  "data": [
    {
      "id": "1704067200000000000",
      "message": "[RSI] BTCUSDT 2h close=42500.00 RSI(14)=25.50 @ 2024-01-01 12:00:00",
      "source": "rsi",
      "timestamp": "2024-01-01T12:00:00Z"
    },
    {
      "id": "1704067260000000000",
      "message": "大额清算警报：多单清算 $2.5M",
      "source": "liquidation",
      "timestamp": "2024-01-01T12:01:00Z"
    },
    {
      "id": "1704067320000000000",
      "message": "手动发送的测试消息",
      "source": "manual",
      "timestamp": "2024-01-01T12:02:00Z"
    }
  ]
}
```

### 2. 获取消息统计信息

#### 接口地址
```
GET /notice/messages/stats
```

#### 请求参数
无

#### 请求示例

```bash
curl http://localhost:5555/notice/messages/stats
```

#### 响应示例

```json
{
  "success": true,
  "total_count": 156,
  "source_stats": {
    "rsi": 45,
    "liquidation": 67,
    "news": 23,
    "manual": 15,
    "webhook": 6
  }
}
```

### 3. 按时间范围查询消息

#### 接口地址
```
GET /notice/messages/range
```

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| start | string | 是 | 开始时间，RFC3339格式 | 2024-01-01T00:00:00Z |
| end | string | 是 | 结束时间，RFC3339格式 | 2024-01-01T23:59:59Z |

#### 请求示例

```bash
# 查询今天的消息
curl "http://localhost:5555/notice/messages/range?start=2024-01-01T00:00:00Z&end=2024-01-01T23:59:59Z"

# 查询最近1小时的消息
curl "http://localhost:5555/notice/messages/range?start=2024-01-01T10:00:00Z&end=2024-01-01T11:00:00Z"

# 查询特定时间段
curl "http://localhost:5555/notice/messages/range?start=2024-01-01T09:00:00Z&end=2024-01-01T17:00:00Z"
```

#### 响应示例

```json
{
  "success": true,
  "count": 12,
  "start": "2024-01-01T00:00:00Z",
  "end": "2024-01-01T23:59:59Z",
  "data": [
    {
      "id": "1704067200000000000",
      "message": "[RSI] BTCUSDT 2h close=42500.00 RSI(14)=25.50 @ 2024-01-01 12:00:00",
      "source": "rsi",
      "timestamp": "2024-01-01T12:00:00Z"
    }
  ]
}
```

## 消息来源类型

| 来源类型 | 说明 | 示例消息 |
|----------|------|----------|
| `rsi` | RSI 指标警报 | `[RSI] BTCUSDT 2h close=42500.00 RSI(14)=25.50 @ 2024-01-01 12:00:00` |
| `liquidation` | 清算监控消息 | `📊 1小时清算统计报告\n清算订单数: 15\n总价值: 2.50w USDT` |
| `news` | 新闻推送 | `【BlockBeats】比特币突破新高` |
| `manual` | 手动发送的消息 | `手动测试消息` |
| `webhook` | 通过 webhook 接收的消息 | `外部系统推送的警报` |

## 客户端集成示例

### JavaScript/React Native

```javascript
class MessageAPI {
  constructor(baseURL = 'http://localhost:5555') {
    this.baseURL = baseURL;
    this.apiPrefix = '/notice';
  }

  // 获取消息历史
  async getMessages(limit = 50, source = '') {
    try {
      let url = `${this.baseURL}${this.apiPrefix}/messages?limit=${limit}`;
      if (source) {
        url += `&source=${source}`;
      }

      const response = await fetch(url);
      const data = await response.json();

      if (data.success) {
        return data.data;
      } else {
        throw new Error('获取消息失败');
      }
    } catch (error) {
      console.error('获取消息失败:', error);
      throw error;
    }
  }

  // 获取统计信息
  async getStats() {
    try {
      const response = await fetch(`${this.baseURL}${this.apiPrefix}/messages/stats`);
      const data = await response.json();

      if (data.success) {
        return {
          totalCount: data.total_count,
          sourceStats: data.source_stats
        };
      } else {
        throw new Error('获取统计失败');
      }
    } catch (error) {
      console.error('获取统计失败:', error);
      throw error;
    }
  }

  // 按时间范围查询
  async getMessagesByTimeRange(startTime, endTime) {
    try {
      const url = `${this.baseURL}${this.apiPrefix}/messages/range?start=${startTime}&end=${endTime}`;
      const response = await fetch(url);
      const data = await response.json();

      if (data.success) {
        return data.data;
      } else {
        throw new Error('查询失败');
      }
    } catch (error) {
      console.error('查询失败:', error);
      throw error;
    }
  }

  // 获取今天的消息
  async getTodayMessages() {
    const today = new Date();
    const startOfDay = new Date(today.setHours(0, 0, 0, 0)).toISOString();
    const endOfDay = new Date(today.setHours(23, 59, 59, 999)).toISOString();

    return this.getMessagesByTimeRange(startOfDay, endOfDay);
  }

  // 获取最近N小时的消息
  async getRecentMessages(hours = 1) {
    const end = new Date();
    const start = new Date(end.getTime() - hours * 60 * 60 * 1000);

    return this.getMessagesByTimeRange(start.toISOString(), end.toISOString());
  }
}

// 使用示例
const api = new MessageAPI();

// 获取最近50条消息
const messages = await api.getMessages(50);

// 获取RSI消息
const rsiMessages = await api.getMessages(20, 'rsi');

// 获取统计信息
const stats = await api.getStats();

// 获取今天的消息
const todayMessages = await api.getTodayMessages();

// 获取最近3小时的消息
const recentMessages = await api.getRecentMessages(3);
```

### Python

```python
import requests
from datetime import datetime, timedelta
from typing import List, Dict, Optional

class MessageAPI:
    def __init__(self, base_url: str = "http://localhost:5555"):
        self.base_url = base_url
        self.api_prefix = "/notice"

    def get_messages(self, limit: Optional[int] = None, source: Optional[str] = None) -> List[Dict]:
        """获取消息历史"""
        params = {}
        if limit:
            params['limit'] = limit
        if source:
            params['source'] = source

        response = requests.get(f"{self.base_url}{self.api_prefix}/messages", params=params)
        response.raise_for_status()

        data = response.json()
        if data.get('success'):
            return data.get('data', [])
        else:
            raise Exception("获取消息失败")

    def get_stats(self) -> Dict:
        """获取统计信息"""
        response = requests.get(f"{self.base_url}{self.api_prefix}/messages/stats")
        response.raise_for_status()

        data = response.json()
        if data.get('success'):
            return {
                'total_count': data.get('total_count'),
                'source_stats': data.get('source_stats')
            }
        else:
            raise Exception("获取统计失败")

    def get_messages_by_time_range(self, start_time: str, end_time: str) -> List[Dict]:
        """按时间范围查询消息"""
        params = {
            'start': start_time,
            'end': end_time
        }

        response = requests.get(f"{self.base_url}{self.api_prefix}/messages/range", params=params)
        response.raise_for_status()

        data = response.json()
        if data.get('success'):
            return data.get('data', [])
        else:
            raise Exception("查询失败")

    def get_today_messages(self) -> List[Dict]:
        """获取今天的消息"""
        today = datetime.now().date()
        start_time = datetime.combine(today, datetime.min.time()).isoformat() + 'Z'
        end_time = datetime.combine(today, datetime.max.time()).isoformat() + 'Z'

        return self.get_messages_by_time_range(start_time, end_time)

    def get_recent_messages(self, hours: int = 1) -> List[Dict]:
        """获取最近N小时的消息"""
        end_time = datetime.now()
        start_time = end_time - timedelta(hours=hours)

        return self.get_messages_by_time_range(
            start_time.isoformat() + 'Z',
            end_time.isoformat() + 'Z'
        )

# 使用示例
api = MessageAPI()

# 获取最近50条消息
messages = api.get_messages(limit=50)

# 获取RSI消息
rsi_messages = api.get_messages(limit=20, source='rsi')

# 获取统计信息
stats = api.get_stats()
print(f"总消息数: {stats['total_count']}")
print(f"各来源统计: {stats['source_stats']}")

# 获取今天的消息
today_messages = api.get_today_messages()

# 获取最近3小时的消息
recent_messages = api.get_recent_messages(hours=3)
```

## 错误处理

### 常见错误码

| HTTP状态码 | 说明 | 解决方法 |
|------------|------|----------|
| 400 | 请求参数错误 | 检查参数格式，特别是时间格式 |
| 500 | 服务器内部错误 | 检查服务器日志，可能是存储文件权限问题 |

### 错误响应示例

```json
{
  "error": "Invalid limit parameter"
}
```

### 时间格式说明

时间参数必须使用 RFC3339 格式，示例：
- `2024-01-01T00:00:00Z` (UTC时间)
- `2024-01-01T08:00:00+08:00` (带时区)

## 性能建议

1. **分页查询**: 使用 `limit` 参数避免一次性获取大量数据
2. **缓存策略**: 客户端可以缓存统计信息，定期更新
3. **时间范围**: 避免查询过长的时间范围，建议单次查询不超过24小时
4. **轮询频率**: 建议轮询间隔不少于10秒

## 数据存储说明

- 消息存储在服务器的 `./storage/messages.json` 文件中
- 系统最多保留 1000 条历史消息
- 超过限制时，会自动删除最旧的消息
- 每条消息都有唯一的 ID（基于纳秒时间戳）

## 金额显示格式

### 万单位转换规则
清算相关消息中的 USDT 金额会自动进行格式化：

- **≥ 10,000 USDT**: 自动转换为万单位显示（保留2位小数）
  - 例：25,000 USDT → `2.50w USDT`
  - 例：150,000 USDT → `15.00w USDT`
  - 例：1,500,000 USDT → `150.00w USDT`

- **< 10,000 USDT**: 保持原始格式显示（保留2位小数）
  - 例：5,000 USDT → `5000.00 USDT`
  - 例：8,888 USDT → `8888.00 USDT`

### 适用范围
万单位转换适用于以下类型的消息：
- 清算统计报告中的总价值
- 多单清算价值
- 空单清算价值
- 日志输出中的价值信息