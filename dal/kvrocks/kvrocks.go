package kvrocks

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type KvrocksService interface {
	Set(ctx context.Context, key string, value string) error
	Get(ctx context.Context, key string) (string, error)
	MGet(ctx context.Context, keys []string) ([]string, error)
	Cas(ctx context.Context, key string, oldValue string, newValue string) (int64, error)
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

func (k KvrocksServiceImpl) Cas(ctx context.Context, key string, oldValue string, newValue string) (int64, error) {
	res, err := k.client.Do(ctx, "cas", key, oldValue, newValue).Result()
	if err != nil {
		logrus.Errorf("kvrocks cas err. err = %v", err)
		return res.(int64), err
	}
	logrus.Infof("kvrocks cas result = %v", res)
	return res.(int64), nil
}
