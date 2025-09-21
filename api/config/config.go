package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	WebSockets []WebSocketConfig `json:",optional"`
}

type WebSocketConfig struct {
	Name             string `json:",optional"`      // 连接名称标识
	URL              string `json:",optional"`      // WebSocket URL
	ReconnectDelay   int    `json:",optional"`      // 重连延迟(秒)，默认5秒
	MaxReconnects    int    `json:",optional"`      // 最大重连次数，0表示无限重连
	PingInterval     int    `json:",optional"`      // 心跳间隔(秒)，默认30秒
	HandshakeTimeout int    `json:",optional"`      // 握手超时(秒)，默认10秒
	Headers          map[string]string `json:",optional"` // 请求头
}