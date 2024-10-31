package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"time"
)

type RedisService interface {
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	BatchSet(ctx context.Context, keys []string, values []string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	MGet(ctx context.Context, keys []string) ([]interface{}, error)
	Del(ctx context.Context, key string) error
	HSet(ctx context.Context, key string, field string, value interface{}) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, field string) error
	HExists(ctx context.Context, key string, field string) (bool, error)
	ZAdd(ctx context.Context, key string, values []redis.Z) error
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZCard(ctx context.Context, key string) (int64, error)
	FlushDB(ctx context.Context) error
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

func (r *RedisServiceImpl) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	_, err := r.client.Set(ctx, key, value, expiration).Result()
	if err != nil {
		logrus.Errorf("[Set] redis set err. err = %v", err)
		return err
	}
	return nil
}

func (r *RedisServiceImpl) BatchSet(ctx context.Context, keys []string, values []string, expiration time.Duration) error {
	pipe := r.client.Pipeline()
	for i := 0; i < len(keys); i++ {
		pipe.Set(ctx, keys[i], values[i], expiration)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		logrus.Errorf("[BatchSet] redis batch set err. err = %v", err)
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

func (r *RedisServiceImpl) MGet(ctx context.Context, keys []string) ([]interface{}, error) {
	res, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		logrus.Errorf("[MGet] redis mget err. err = %v", err)
		return nil, err
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

func (r *RedisServiceImpl) HExists(ctx context.Context, key string, field string) (bool, error) {
	res, err := r.client.HExists(ctx, key, field).Result()
	if err != nil {
		logrus.Errorf("[HExists] redis hexists err. err = %v", err)
		return false, err
	}
	return res, nil
}

func (r *RedisServiceImpl) ZAdd(ctx context.Context, key string, values []redis.Z) error {
	_, err := r.client.ZAdd(ctx, key, values...).Result()
	if err != nil {
		logrus.Errorf("[ZAdd] redis zadd err. err = %v", err)
		return err
	}
	return nil
}

func (r *RedisServiceImpl) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	res, err := r.client.ZRange(ctx, key, start, stop).Result()
	if err != nil {
		logrus.Errorf("[ZRange] redis zrange err. err = %v", err)
		return nil, err
	}
	return res, nil
}

func (r *RedisServiceImpl) ZCard(ctx context.Context, key string) (int64, error) {
	res, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		logrus.Errorf("[ZCard] redis zcard err. err = %v", err)
		return 0, err
	}
	return res, nil
}

func (r *RedisServiceImpl) FlushDB(ctx context.Context) error {
	_, err := r.client.FlushDB(ctx).Result()
	if err != nil {
		logrus.Errorf("[FlushDB] redis flushdb err. err = %v", err)
		return err
	}
	return nil
}
