package tools

import (
	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
	"gochat/config"
	"sync"
)

var KafkaProducerMap = map[string]sarama.SyncProducer{}
var kafkaMapSyncLock sync.Mutex

type KafkaOption struct {
	Address string
}

func setupProducer() (sarama.SyncProducer, error) {
	conf := sarama.NewConfig()
	conf.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{config.Conf.Task.TaskBase.KafkaServerAddress}, conf)
	if err != nil {
		logrus.Errorf("kafka, setupProducer error: %s", err)
		return nil, err
	}
	return producer, nil
}

func GetProducerInstance(kafkaOption KafkaOption) (sarama.SyncProducer, error) {
	address := kafkaOption.Address
	kafkaMapSyncLock.Lock()
	if producer, ok := KafkaProducerMap[address]; ok {
		return producer, nil
	}
	producer, err := setupProducer()
	if err != nil {
		logrus.Panicf("kafka producer %s", err)
		return nil, err
	}
	KafkaProducerMap[address] = producer
	kafkaMapSyncLock.Unlock()
	return KafkaProducerMap[address], nil
}
