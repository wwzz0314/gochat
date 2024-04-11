package task

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gochat/config"
	"gochat/tools"
	"time"
)

var RedisClient *redis.Client

func (task *Task) InitQueueRedisClient() (err error) {
	redisOpt := tools.RedisOption{
		Address:  config.Conf.Common.CommonRedis.RedisAddress,
		Password: config.Conf.Common.CommonRedis.RedisPassword,
		Db:       config.Conf.Common.CommonRedis.Db,
	}
	RedisClient = tools.GetRedisInstance(redisOpt)
	if pong, err := RedisClient.Ping(context.Background()).Result(); err != nil {
		logrus.Infof("RedisClient Ping Result pong: %s, err: %s", pong, err)
	}
	go func() {
		for {
			var result []string
			result, err = RedisClient.BRPop(context.Background(), time.Second*10, config.QueueName).Result()
			if err != nil {
				logrus.Infof("task queue block timeout, no msg err: %s", err.Error())
			}
			if len(result) >= 2 {
				task.Push(result[1])
			}
		}
	}()
	return
}
