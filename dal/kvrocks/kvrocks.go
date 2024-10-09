package kvrocks

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type KvrocksService interface {
	Set(ctx context.Context, key string, value string) error
	Get(ctx context.Context, key string) (string, error)
}

type KvrocksServiceImpl struct {
	client *redis.Client
}

func NewKvrocksServiceImpl() KvrocksServiceImpl {
	kvrocksClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6666",
		Password: "",
		DB:       0,
	})
	return KvrocksServiceImpl{client: kvrocksClient}
}

func (k KvrocksServiceImpl) Set(ctx context.Context, key string, value string) error {
	return k.client.Set(ctx, key, value, 0).Err()
}

func (k KvrocksServiceImpl) Get(ctx context.Context, key string) (string, error) {
	return k.client.Get(ctx, key).Result()
}
