package kvrocks

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"time"
)

type KvrocksService interface {
	Set(ctx context.Context, key string, value string) error
	Get(ctx context.Context, key string) (string, error)
	MGet(ctx context.Context, keys []string) ([]string, error)
	SetNX(ctx context.Context, key string, value string) (bool, error)
	Cas(ctx context.Context, key string, oldValue string, newValue string) (int64, error)
	RPush(ctx context.Context, key string, values []string) (int64, error)
	SetExpire(ctx context.Context, key string, expiration time.Duration) error
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
	_, err := k.client.Set(ctx, key, value, 0).Result()
	if err != nil {
		logrus.Errorf("kvrocks set err. err = %v", err)
		return err
	}
	return nil
}

func (k KvrocksServiceImpl) Get(ctx context.Context, key string) (string, error) {
	res, err := k.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", err
	}
	if err != nil {
		logrus.Errorf("kvrocks get err. err = %v", err)
		return "", err
	}
	return res, nil
}

func (k KvrocksServiceImpl) MGet(ctx context.Context, keys []string) ([]string, error) {
	resInters, err := k.client.MGet(ctx, keys...).Result()
	if err != nil {
		logrus.Errorf("kvrocks mget err. err = %v", err)
		return nil, err
	}
	var res []string
	for _, resInter := range resInters {
		if resStr, ok := resInter.(string); ok {
			res = append(res, resStr)
		}
	}
	return res, nil
}

func (k KvrocksServiceImpl) SetNX(ctx context.Context, key string, value string) (bool, error) {
	res, err := k.client.SetNX(ctx, key, value, 0).Result()
	if err != nil {
		logrus.Errorf("kvrocks set nx err. err = %v", err)
		return false, err
	}
	return res, nil
}

func (k KvrocksServiceImpl) Cas(ctx context.Context, key string, oldValue string, newValue string) (int64, error) {
	res, err := k.client.Do(ctx, "cas", key, oldValue, newValue).Result()
	if err != nil {
		logrus.Errorf("kvrocks cas err. err = %v", err)
		return res.(int64), err
	}
	logrus.Infof("kvrocks cas result = %v", res)
	return res.(int64), nil
}

func (k KvrocksServiceImpl) RPush(ctx context.Context, key string, values []string) (int64, error) {
	res, err := k.client.RPush(ctx, key, values).Result()
	if err != nil {
		logrus.Errorf("kvrocks rpush err. err = %v", err)
		return 0, err
	}
	return res, nil
}

func (k KvrocksServiceImpl) SetExpire(ctx context.Context, key string, expiration time.Duration) error {
	_, err := k.client.Expire(ctx, key, expiration).Result()
	if err != nil {
		logrus.Errorf("kvrocks set expire err. err = %v", err)
		return err
	}
	return nil
}
