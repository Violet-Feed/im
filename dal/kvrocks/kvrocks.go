package kvrocks

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"runtime"
	"time"
)

type KvrocksService interface {
	Set(ctx context.Context, key string, value string) error
	MSet(ctx context.Context, values map[string]string) error
	Get(ctx context.Context, key string) (string, error)
	MGet(ctx context.Context, keys []string) ([]string, error)
	SetNX(ctx context.Context, key string, value string) (bool, error)
	Cas(ctx context.Context, key string, oldValue string, newValue string) (int64, error)
	RPush(ctx context.Context, key string, values []string) (int64, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	LLen(ctx context.Context, key string) (int64, error)
	ZAdd(ctx context.Context, key string, members []redis.Z) (int64, error)
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error)
	ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error)
	ZRevRangByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) ([]redis.Z, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
}

type KvrocksServiceImpl struct {
	client *redis.Client
}

func NewKvrocksServiceImpl() KvrocksServiceImpl {
	if runtime.GOOS == "windows" {
		kvrocksClient := redis.NewClient(&redis.Options{
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       1,
		})
		return KvrocksServiceImpl{client: kvrocksClient}
	}
	kvrocksClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6666",
		Password: "",
		DB:       0,
	})
	return KvrocksServiceImpl{client: kvrocksClient}
}

func (k *KvrocksServiceImpl) Set(ctx context.Context, key string, value string) error {
	_, err := k.client.Set(ctx, key, value, 0).Result()
	if err != nil {
		logrus.Errorf("kvrocks set err. err = %v", err)
		return err
	}
	return nil
}

func (k *KvrocksServiceImpl) MSet(ctx context.Context, values map[string]string) error {
	_, err := k.client.MSet(ctx, values).Result()
	if err != nil {
		logrus.Errorf("[MSet] redis mset err. err = %v", err)
		return err
	}
	return nil
}

func (k *KvrocksServiceImpl) Get(ctx context.Context, key string) (string, error) {
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

func (k *KvrocksServiceImpl) MGet(ctx context.Context, keys []string) ([]string, error) {
	resInters, err := k.client.MGet(ctx, keys...).Result()
	if err != nil {
		logrus.Errorf("kvrocks mget err. err = %v", err)
		return nil, err
	}
	var res []string
	for _, resInter := range resInters {
		if resInter == nil {
			res = append(res, "")
		} else if resStr, ok := resInter.(string); ok {
			res = append(res, resStr)
		} else {
			logrus.Errorf("kvrocks mget assert err. err = %v", err)
			return nil, err
		}
	}
	return res, nil
}

func (k *KvrocksServiceImpl) SetNX(ctx context.Context, key string, value string) (bool, error) {
	res, err := k.client.SetNX(ctx, key, value, 0).Result()
	if err != nil {
		logrus.Errorf("kvrocks set nx err. err = %v", err)
		return false, err
	}
	return res, nil
}

func (k *KvrocksServiceImpl) Cas(ctx context.Context, key string, oldValue string, newValue string) (int64, error) {
	if runtime.GOOS == "windows" {
		locked, err := k.client.SetNX(ctx, "lock:"+key, "1", 1*time.Second).Result()
		defer k.client.Del(ctx, "lock:"+key)
		if err != nil {
			logrus.Errorf("kvrocks cas lock err. err = %v", err)
			return 0, err
		}
		if !locked {
			return 0, nil
		}
		_, err = k.client.Set(ctx, key, newValue, 0).Result()
		if err != nil {
			logrus.Errorf("kvrocks cas set err. err = %v", err)
			return 0, err
		}
		return 1, nil
	}
	res, err := k.client.Do(ctx, "cas", key, oldValue, newValue).Result()
	if err != nil {
		logrus.Errorf("kvrocks cas err. err = %v", err)
		return 0, err
	}
	return res.(int64), nil
}

func (k *KvrocksServiceImpl) RPush(ctx context.Context, key string, values []string) (int64, error) {
	res, err := k.client.RPush(ctx, key, values).Result()
	if err != nil {
		logrus.Errorf("kvrocks rpush err. err = %v", err)
		return 0, err
	}
	return res, nil
}

func (k *KvrocksServiceImpl) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	res, err := k.client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		logrus.Errorf("kvrocks lrange err. err = %v", err)
		return nil, err
	}
	return res, nil
}

func (k *KvrocksServiceImpl) LLen(ctx context.Context, key string) (int64, error) {
	res, err := k.client.LLen(ctx, key).Result()
	if err != nil {
		logrus.Errorf("kvrocks llen err. err = %v", err)
		return 0, err
	}
	return res, nil
}

func (k *KvrocksServiceImpl) ZAdd(ctx context.Context, key string, members []redis.Z) (int64, error) {
	res, err := k.client.ZAdd(ctx, key, members...).Result()
	if err != nil {
		logrus.Errorf("kvrocks zadd err. err = %v", err)
		return 0, err
	}
	return res, nil
}

func (k *KvrocksServiceImpl) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	res, err := k.client.ZRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		logrus.Errorf("kvrocks zrange err. err = %v", err)
		return nil, err
	}
	return res, nil
}

func (k *KvrocksServiceImpl) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) {
	res, err := k.client.ZRemRangeByRank(ctx, key, start, stop).Result()
	if err != nil {
		logrus.Errorf("kvrocks zremrangebyrank err. err = %v", err)
	}
	return res, err
}

func (k *KvrocksServiceImpl) ZRevRangByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
	res, err := k.client.ZRevRangeByScoreWithScores(ctx, key, opt).Result()
	if err != nil {
		logrus.Errorf("kvrocks zrevrangbyscorewithscores err. err = %v", err)
		return nil, err
	}
	return res, nil
}

func (k *KvrocksServiceImpl) Expire(ctx context.Context, key string, expiration time.Duration) error {
	_, err := k.client.Expire(ctx, key, expiration).Result()
	if err != nil {
		logrus.Errorf("kvrocks expire err. err = %v", err)
		return err
	}
	return nil
}
