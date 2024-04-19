package task

import (
	"context"
	"github.com/IBM/sarama"
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
		return err
	}
	go func() {
		for {
			var result []string
			result, err = RedisClient.BRPop(context.Background(), time.Second*10, config.QueueName).Result()
			if err != nil {
				logrus.Infof("task queue block timeout, no msg err: %s", err.Error())
			}
			if len(result) >= 2 {
				task.Push([]byte(result[1]))
			}
		}
	}()
	return
}

func (task *Task) InitConsumerGroup(consumer Consumer) (err error) {
	conf := sarama.NewConfig()
	consumerConf := config.Conf.Task.TaskBase
	consumerGroup, err := sarama.NewConsumerGroup([]string{consumerConf.KafkaServerAddress}, consumerConf.ConsumerGroup, conf)
	if err != nil {
		logrus.Errorf("init consumer group err:%s", err)
		return
	}
	go func() {
		for {
			err = consumerGroup.Consume(context.Background(), []string{consumerConf.ConsumerTopic}, &consumer)
			if err != nil {
				logrus.Infof("error of consumer: %v", err)
			}
		}
	}()
	return nil
}
