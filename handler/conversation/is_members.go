package conversation

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"im/biz/model"
	"im/dal"
	"strconv"
	"sync"
)

func IsConversationMembers(ctx context.Context, conShortIds []int64, userId int64) (map[int64]int32, error) {
	wg := sync.WaitGroup{}
	statusChan := make([]chan int32, len(conShortIds))
	for i, conShortId := range conShortIds {
		wg.Add(1)
		go func(i int, conShortId int64) {
			defer wg.Done()
			_, err := dal.RedisServer.ZScore(ctx, "members:"+strconv.FormatInt(conShortId, 10), strconv.FormatInt(userId, 10))
			if err == nil {
				statusChan[i] <- 1
				return
			} else if err == redis.Nil {
				statusChan[i] <- 0
				return
			} else {
				userInfo, err := model.GetUserInfos(ctx, conShortId, []int64{userId}, false)
				if err != nil {
					logrus.Errorf("[IsMembers] GetUserInfos err. err = %v", err)
					statusChan[i] <- -1
				}
				if len(userInfo) > 0 {
					statusChan[i] <- 1
				} else {
					statusChan[i] <- 0
				}
			}
		}(i, conShortId)
	}
	wg.Wait()
	status := make(map[int64]int32)
	for i, conShortId := range conShortIds {
		status[conShortId] = <-statusChan[i]
	}
	return status, nil
	//获取core？
	//并发redis zscore;err mysql
}
