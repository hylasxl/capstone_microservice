package configs

import "github.com/redis/go-redis/v9"

func ConnectRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "message_redis:16379",
	})
}
