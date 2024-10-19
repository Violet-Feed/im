package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"time"
)

type RedisService interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	HSet(ctx context.Context, key string, field string, value interface{}) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, field string) error
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

func (r *RedisServiceImpl) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	_, err := r.client.Set(ctx, key, value, expiration).Result()
	if err != nil {
		logrus.Errorf("[Set] redis set err. err = %v", err)
		return err
	}
	return nil
}

func (r *RedisServiceImpl) Get(ctx context.Context, key string) (string, error) {
	res, err := r.client.Get(ctx, key).Result()
	if err != nil {
		logrus.Errorf("[Get] redis get err. err = %v", err)
		return "", err
	}
	return res, nil
}

func (r *RedisServiceImpl) Del(ctx context.Context, key string) error {
	_, err := r.client.Del(ctx, key).Result()
	if err != nil {
		logrus.Errorf("[Del] redis del err. err = %v", err)
		return err
	}
	return nil
}

func (r *RedisServiceImpl) HSet(ctx context.Context, key string, field string, value interface{}) error {
	_, err := r.client.HSet(ctx, key, field, value).Result()
	if err != nil {
		logrus.Errorf("[HSet] redis hset err. err = %v", err)
		return err
	}
	return nil
}

func (r *RedisServiceImpl) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	res, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		logrus.Errorf("[HGetAll] redis hgetall err. err = %v", err)
		return nil, err
	}
	return res, nil
}

func (r *RedisServiceImpl) HDel(ctx context.Context, key string, field string) error {
	_, err := r.client.HDel(ctx, key, field).Result()
	if err != nil {
		logrus.Errorf("[HDel] redis hdel err. err = %v", err)
		return err
	}
	return nil
}
