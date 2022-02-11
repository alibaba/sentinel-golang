package server

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
)

const testResource = "cluster-flow-control-test-resource"

/*
func TestRedisTokenServer(t *testing.T) {
	// prepare env
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	redisTokenServer := RedisTokenServer{
		redisCli:               rdb,
		rwMux:                  sync.RWMutex{},
		resClusterTokenResetMap: make(map[string]*redisTokenResetSlidingWindow),
	}
	redisTokenServer.resClusterTokenResetMap[testResource] = redisTokenServer.newResourceTokenReset(testResource, 1000)

	start := time.Now()
	for i := 0; i < 100000; i++ {
		_, err := redisTokenServer.AcquireFlowToken(testResource, 1, nil)
		if err != nil {
			t.Log(err)
		}
	}
	end := time.Now()
	fmt.Println("cap: ", end.Sub(start).Nanoseconds())
}
*/

func Benchmark_1(b *testing.B) {
	// prepare env
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
		PoolSize: 100,
	})

	redisTokenServer := RedisTokenService{
		redisCli: rdb,
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := redisTokenServer.AcquireFlowToken(testResource, 1, 1000)
			if err != nil {
				b.Log(err)
			}
		}
	})
}

var ctx = context.Background()

//
//func ExampleClient() {
//	rdb := redis.NewClient(&redis.Options{
//		Addr:     "localhost:6379",
//		Password: "", // no password set
//		DB:       0,  // use default DB
//	})
//	incr1, err1 := rdb.Incr(ctx, "testincr1").Result()
//	fmt.Println(incr1)
//	fmt.Println(err1)
//
//	incr1, err1 = rdb.IncrBy(ctx, "testincr1", 1000).Result()
//	fmt.Println(incr1)
//	fmt.Println(err1)
//
//	err := rdb.Set(ctx, "key1", "value1", 0).Err()
//	if err != nil {
//		panic(err)
//	}
//
//	val, err := rdb.Get(ctx, "key1").Result()
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println("key1", val)
//
//	val2, err := rdb.Get(ctx, "key2").Result()
//	if err == redis.Nil {
//		fmt.Println("key2 does not exist")
//	} else if err != nil {
//		panic(err)
//	} else {
//		fmt.Println("key2", val2)
//	}
//	// Output: key value
//	// key2 does not exist
//}
