# CLAUDE.md

此文件为 Claude Code (claude.ai/code) 在此代码库中工作时提供指导。

## 项目概述

这是一个基于 Go 和 go-zero 框架构建的加密货币通知系统。该应用程序监控加密货币市场，跟踪清算数据，计算 RSI 指标，并通过 Expo Push Notifications 向移动设备发送推送通知。

## 构建和开发命令

### 应用程序构建
```bash
# Linux 生产环境构建 - 同时部署到服务器
make build

# 本地开发构建
go build -o notice api/notice.go

# 生产环境优化构建
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o notice api/notice.go
```

### 运行应用程序
```bash
# 使用默认配置运行
./notice

# 使用自定义配置文件运行
./notice -f etc/api.yaml
```

### 开发命令
```bash
# 运行测试
go test ./...

# 运行特定包的测试
go test ./api/expo
go test ./api/websocket

# 运行单个测试函数
go test -run TestExpoClient ./api/expo

# 格式化代码
go fmt ./...

# 检查代码问题
go vet ./...

# 下载依赖
go mod tidy
```

## 架构概述

### 核心组件

1. **主服务 (`api/notice.go`)**:
   - 启动所有服务的入口点
   - 配置端口 5555 的 REST API 服务器
   - 初始化市场数据的 WebSocket 连接
   - 启动监控任务的后台 goroutine

2. **推送通知系统 (`api/expo/`)**:
   - 管理 Expo 推送通知令牌
   - 处理消息发送和重试逻辑
   - 提供令牌验证和管理 API

3. **市场数据监控**:
   - **WebSocket 连接器 (`api/websocket/`)**: 用于实时市场数据的通用 WebSocket 客户端
   - **RSI 监控 (`api/rsi/`)**: 计算 BTC/ETH 多个时间框架的 RSI 指标 (2h, 4h, 1d, 1w)
   - **清算跟踪 (`api/margin_push/`)**: 监控币安期货清算并进行统计分析，金额显示支持万单位格式（≥1万自动转换为w单位）
   - **新闻监控 (`api/listen/`)**: 每 10 秒轮询 TheBlockBeats RSS 源获取加密货币新闻
   - **BlockBeat 数据 (`api/blockbeat/`)**: 加密货币新闻更新的 RSS 源解析器

4. **配置 (`api/config/`)**:
   - 定义应用程序和 WebSocket 配置结构
   - 使用 go-zero 的 RestConf 进行基本服务器设置

5. **消息存储系统 (`api/storage/`)**:
   - 自动保存所有发送的通知消息到本地 JSON 文件
   - 支持消息分类（manual, webhook, rsi, liquidation, news）
   - 提供消息查询、统计和时间范围过滤功能
   - 线程安全的并发访问控制

6. **通知管理 (`api/notification/`)**:
   - 统一的通知发送接口，集成消息存储功能
   - 自动为每条消息添加时间戳和来源标识
   - 支持不同类型的通知发送（普通、带标题、重试）

### 服务架构

```
HTTP API (端口 5555，前缀 /notice)
├── /notice/notice_token (POST) - 添加推送通知令牌
├── /notice/notice_token/stats (GET) - 获取令牌统计信息
├── /notice/notice/query (POST) - 发送手动通知
├── /notice/webhook (POST) - 接收 webhook 通知
├── /notice/sse (GET) - 服务端发送事件端点
├── /notice/test (GET) - SSE 测试页面
├── /notice/messages (GET) - 获取消息历史记录
├── /notice/messages/stats (GET) - 获取消息统计信息
└── /notice/messages/range (GET) - 按时间范围获取消息

后台服务:
├── margin_push.ForceReceive() - 币安期货清算监控（每小时/每日统计）
├── rsi.StartBinanceRSI() - RSI 计算（BTC/ETH 在 2h/4h/1d/1w 时间框架）
├── listen.StartListen() - TheBlockBeats RSS 源轮询（10秒间隔）
└── WebSocket 连接器 - 实时市场数据流（币安 BTC/USDT 1m，Coinbase）
```

### 数据流

1. **市场数据输入**:
   - WebSocket 连接到币安（BTC/USDT 1m K线）和 Coinbase
   - 币安期货 API 获取清算数据
   - TheBlockBeats RSS 源获取加密货币新闻
2. **数据处理**:
   - 使用历史 K线 数据计算多时间框架的 RSI
   - 清算统计分析（多空分析和价值跟踪）
   - 新闻源解析和重复检测
3. **通知触发**:
   - RSI 极值（超卖/超买条件）
   - 大额清算事件统计摘要
   - 新的加密货币新闻文章
4. **客户端推送**: 移动应用通过 Expo SDK 接收推送通知

## 关键模式和约定

### 包结构
- 每个主要功能在 `api/` 下都有自己的包
- 每个包通常包含一个主实现文件和可选的测试文件
- 配置在 `api/config/` 中集中管理
- 共享模型在 `api/model/` 中

### 错误处理
- 使用 go-zero 的内置日志记录 (`logx`)
- HTTP 处理器返回适当的状态码和错误消息
- 后台服务记录错误并继续运行

### 并发处理
- 大量使用 goroutine 进行后台监控任务
- WebSocket 连接器独立处理重连和消息处理
- 在 WebSocket 连接器中使用 sync.RWMutex 进行线程安全操作

### 配置
- 主配置在 `etc/api.yaml`（服务器、日志、示例 WebSocket 配置）
- **重要**: WebSocket 连接在 `api/notice.go` 中硬编码（不使用配置文件）
- go-zero 框架处理大部分服务器配置

## 开发说明

### 测试
- 存在基本测试文件 (`api/expo/expo_test.go`, `api/html_expo_test.go`)
- 使用 `go test ./...` 运行测试
- 示例测试演示 Expo 通知功能

### 依赖关系
- **go-zero**: Web 框架和配置管理
- **gorilla/websocket**: WebSocket 客户端功能
- **oliveroneill/exponent-server-sdk-golang**: Expo 推送通知
- **adshao/go-binance**: 币安 API 集成
- **mmcdole/gofeed**: RSS/源解析功能

### 部署
- Makefile 为 Linux 构建并通过 SSH 部署到远程服务器
- 日志写入 `./logs` 目录，支持轮转
- 服务运行在端口 5555，具有基本健康监控

### WebSocket 连接
- 两个硬编码连接：币安（BTC/USDT 1m K线）和 Coinbase
- 可配置延迟和重试限制的自动重连
- 每个连接可自定义消息处理器
- 通用连接器支持多个同时连接，具有独立的生命周期管理

### 监控系统
- **清算监控**: 跟踪币安期货和 PeriodStats（每小时/每日聚合），金额自动格式化为万单位显示
- **RSI 警报**: 多个货币对和时间框架的基于阈值的通知
- **新闻监控**: 基于时间戳的重复检测的 RSS 源轮询
- **统计**: 使用互斥锁保护的并发访问的线程安全数据收集

### 金额显示格式
- **万单位转换**: 当 USDT 金额 ≥ 10,000 时，自动转换为万单位显示
  - 示例：25000 USDT → 2.50w USDT
  - 示例：150000 USDT → 15.00w USDT
  - 示例：5000 USDT → 5000.00 USDT（小于1万保持原格式）

### 消息存储和查询 API

#### 获取消息历史记录
```bash
# 获取所有消息
GET /notice/messages

# 获取最近 50 条消息
GET /notice/messages?limit=50

# 按来源过滤消息
GET /notice/messages?source=rsi
GET /notice/messages?source=liquidation
GET /notice/messages?source=news
GET /notice/messages?source=manual
GET /notice/messages?source=webhook
```

#### 获取消息统计信息
```bash
# 获取消息总数和按来源分类的统计
GET /notice/messages/stats
```

#### 按时间范围查询消息
```bash
# 查询指定时间范围内的消息（RFC3339 格式）
GET /notice/messages/range?start=2024-01-01T00:00:00Z&end=2024-01-01T23:59:59Z
```

#### 响应格式示例
```json
{
  "success": true,
  "count": 10,
  "data": [
    {
      "id": "1704067200000000000",
      "message": "BTC RSI 超卖信号",
      "source": "rsi",
      "timestamp": "2024-01-01T00:00:00Z"
    }
  ]
}
```