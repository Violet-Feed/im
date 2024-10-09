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

func NewKvrocksServiceImpl(client *redis.Client) KvrocksServiceImpl {
	return KvrocksServiceImpl{client: client}
}

func (k KvrocksServiceImpl) Set(ctx context.Context, key string, value string) error {
	return k.client.Set(ctx, key, value, 0).Err()
}

func (k KvrocksServiceImpl) Get(ctx context.Context, key string) (string, error) {
	return k.client.Get(ctx, key).Result()
}
