package etcdv3

import (
	"github.com/alibaba/sentinel-golang/core/config"
	"strings"
)

const (
	EndPoints  = "csp.sentinel.etcd.endpoint"
	User       = "csp.sentinel.etcd.user"
	PassWord   = "csp.sentinel.etcd.password"
	AuthEnable = "csp.sentinel.etcd.auth.enable"
)

func getEndPoint() []string {
	if config.GetConfig(EndPoints) == "" {
		return nil
	}
	endPoint := strings.Split(config.GetConfig(EndPoints), ",")
	return endPoint
}

func getUser() string {
	return config.GetConfig(User)
}

func getPassWord() string {
	return config.GetConfig(PassWord)
}

func isAuthEnable() bool {
	if config.GetConfig(AuthEnable) == "true" {
		return true
	}
	return false
}
