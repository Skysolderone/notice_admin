package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	WebSockets []WebSocketConfig `json:",optional"`
	Database   DatabaseConfig    `json:",optional"`
}

type WebSocketConfig struct {
	Name             string            `json:",optional"` // 连接名称标识
	URL              string            `json:",optional"` // WebSocket URL
	ReconnectDelay   int               `json:",optional"` // 重连延迟(秒)，默认5秒
	MaxReconnects    int               `json:",optional"` // 最大重连次数，0表示无限重连
	PingInterval     int               `json:",optional"` // 心跳间隔(秒)，默认30秒
	HandshakeTimeout int               `json:",optional"` // 握手超时(秒)，默认10秒
	Headers          map[string]string `json:",optional"` // 请求头
}

type DatabaseConfig struct {
	Host            string `json:",optional"` // 数据库主机地址
	Port            int    `json:",optional"` // 数据库端口，默认5432
	User            string `json:",optional"` // 数据库用户名
	Password        string `json:",optional"` // 数据库密码
	DBName          string `json:",optional"` // 数据库名称
	SSLMode         string `json:",optional"` // SSL模式，默认disable
	MaxOpenConns    int    `json:",optional"` // 最大打开连接数，默认10
	MaxIdleConns    int    `json:",optional"` // 最大空闲连接数，默认5
	ConnMaxLifetime int    `json:",optional"` // 连接最大生命周期(秒)，默认3600
}