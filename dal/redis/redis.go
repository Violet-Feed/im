package redis

import "github.com/redis/go-redis/v9"

type RedisService interface {
}

type RedisServiceImpl struct {
	client *redis.Client
}

func NewRedisServiceImpl() RedisServiceImpl {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	return RedisServiceImpl{client: redisClient}
}
