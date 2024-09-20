package server

import (
	"context"
	"github.com/alibaba/sentinel-golang/core/config"
	"strconv"
	"time"

	"github.com/alibaba/sentinel-golang/cluster/client"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

func init() {
	RegisterServiceBuilder("redis", &RedisServiceBuilder{})
}

const (
	InvalidTokenCount   = -1
	redisResourcePrefix = "sentinel-go-##-"
)

type expireKey struct {
	tokenKey   string
	intervalMs time.Duration
}

func (e expireKey) String() string {
	return e.tokenKey + "/" + strconv.Itoa(int(e.intervalMs.Milliseconds()))
}

type RedisServiceBuilder struct {
}

func (builder *RedisServiceBuilder) Builder() client.TokenService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisTokenServiceAddr(),
		Password: config.RedisTokenServicePassword(),
		DB:       config.RedisTokenServiceDB(),
		PoolSize: config.RedisTokenServicePoolSize(),
	})
	return NewTokenService(rdb)
}

type RedisTokenService struct {
	redisCli *redis.Client
	// sized
	expireChan chan expireKey

	ctx context.Context
}

func NewTokenService(cli *redis.Client) client.TokenService {
	ret := &RedisTokenService{
		redisCli:   cli,
		expireChan: make(chan expireKey, 10),
		ctx:        context.Background(),
	}
	go util.RunWithRecover(ret.expireLoop)
	return ret
}

func (r *RedisTokenService) expireLoop() {
	cli := r.redisCli
	for {
		// TODO rethink here
		var expireK *expireKey
		select {
		case key := <-r.expireChan:
			ctx, cancel := context.WithTimeout(context.Background(), cli.Options().ReadTimeout)
			// expire key
			succ, err := r.redisCli.PExpire(ctx, key.tokenKey, key.intervalMs).Result()
			cancel()
			if err != nil {
				expireK = &expireKey{
					tokenKey:   key.tokenKey,
					intervalMs: key.intervalMs,
				}
				logging.Error(err, "Fail to expire token key", "key", key.String())
				break
			} else if !succ {
				expireK = &expireKey{
					tokenKey:   key.tokenKey,
					intervalMs: key.intervalMs,
				}
				logging.Warn("Expire token key failed", "key", key.String())
				break
			}
			expireK = nil
		case <-r.ctx.Done():
			return
		}
		if expireK != nil {
			r.expireChan <- *expireK
		}
	}
}

func (r *RedisTokenService) AcquireFlowToken(resource string, tokenCount uint32, statIntervalInMs uint32) (curCount int64, err error) {
	// 1. general checking logic
	if len(resource) == 0 {
		return InvalidTokenCount, errors.New("empty resource")
	}
	if tokenCount == 0 {
		return InvalidTokenCount, errors.New("token count is zero")
	}
	tokenKey := r.buildResourceKey(resource, statIntervalInMs)

	redisCli := r.redisCli
	ctx, cancel := context.WithTimeout(context.Background(), redisCli.Options().ReadTimeout)
	defer cancel()
	currentVal, err := redisCli.IncrBy(ctx, tokenKey, int64(tokenCount)).Result()
	if err == nil {
		// only one instance meets this condition
		if currentVal == int64(tokenCount) {
			r.expireChan <- expireKey{
				tokenKey:   tokenKey,
				intervalMs: time.Duration(statIntervalInMs) * time.Millisecond,
			}
		}
		return currentVal, nil
	}
	return InvalidTokenCount, err
}

func (r *RedisTokenService) buildResourceKey(res string, statIntervalInMs uint32) string {
	nowMs := util.CurrentTimeMillis()
	startTimeMs := nowMs - (nowMs % uint64(statIntervalInMs))
	return redisResourcePrefix + res + ":" + strconv.Itoa(int(startTimeMs))
}
