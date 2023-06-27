package server

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/cluster/client"
	"strings"
	"sync"
)

type TokenServiceBuilder interface {
	Builder() client.TokenService
}

var (
	tokenServices        = make(map[string]TokenServiceBuilder, 8)
	tokenServiceInstance client.TokenService
	once                 sync.Once
)

func RegisterServiceBuilder(serviceType string, builder TokenServiceBuilder) {
	tokenServices[serviceType] = builder
}

func GetTokenService(serviceType string) (client.TokenService, error) {
	once.Do(func() {
		tokenServiceInstance = tokenServices[strings.ToLower(serviceType)].Builder()
	})
	if tokenServiceInstance == nil {
		return nil, fmt.Errorf("nil TokenService,serviceType=%s", serviceType)
	}
	return tokenServiceInstance, nil
}
