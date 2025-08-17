package svc

import (
	"notice/rpc/expo"

	"github.com/Skysolderone/zero_core/config"
)

type ServiceContext struct {
	Config config.RpcConfig
	Expo   *expo.Expo
}

func NewServiceContext(c config.RpcConfig) *ServiceContext {
	expo := expo.GetExpoClient()
	return &ServiceContext{
		Config: c,
		Expo:   expo,
	}
}
