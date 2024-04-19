package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/go-redis/redis/v8"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/server"
	"gochat/config"
	"gochat/proto"
	"gochat/tools"
	"strconv"
	"strings"
	"time"
)

var RedisClient *redis.Client
var RedisSessClient *redis.Client
var Producer sarama.SyncProducer

func (logic *Logic) InitPublishRedisClient() (err error) {
	redisOpt := tools.RedisOption{
		Address:  config.Conf.Common.CommonRedis.RedisAddress,
		Password: config.Conf.Common.CommonRedis.RedisPassword,
		Db:       config.Conf.Common.CommonRedis.Db,
	}
	RedisClient = tools.GetRedisInstance(redisOpt)
	if pong, err := RedisClient.Ping(context.Background()).Result(); err != nil {
		logrus.Infof("RedisCli Ping Result pong: %s, err: %s", pong, err)
	}
	RedisSessClient = RedisClient
	return err
}

func (logic *Logic) InitKafkaProducer() (err error) {
	kafkaOption := tools.KafkaOption{Address: config.Conf.Task.TaskBase.KafkaServerAddress}
	Producer, err = tools.GetProducerInstance(kafkaOption)
	return err
}

func (logic *Logic) InitRpcServer() (err error) {
	var network, addr string
	// a host multi port case 单个机器开放多个端口
	rpcAddressList := strings.Split(config.Conf.Logic.LogicBase.RpcAddress, ",")
	for _, bind := range rpcAddressList {
		if network, addr, err = tools.ParseNetwork(bind); err != nil {
			logrus.Panicf("InitLogicRpc ParseNetwork error: %s", err.Error())
		}
		logrus.Infof("logic start run at-->%s:%s", network, addr)
		go logic.createRpcServer(network, addr)
	}
	return
}

// 这里创建 rpc 服务器
func (logic *Logic) createRpcServer(network string, addr string) {
	s := server.NewServer()
	logic.addRegisterPlugin(s, network, addr)
	// serverId must be unique
	err := s.RegisterName(config.Conf.Common.CommonEtcd.ServerPathLogic, new(RpcLogic), fmt.Sprintf("%s", logic.ServerId))
	if err != nil {
		logrus.Errorf("register error: %s", err.Error())
	}
	s.RegisterOnShutdown(func(s *server.Server) {
		s.UnregisterAll()
	})
	s.Serve(network, addr)
}

// 在这里进行 ETCD 服务注册
func (logic *Logic) addRegisterPlugin(s *server.Server, network string, addr string) {
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: network + "@" + addr,
		EtcdServers:    []string{config.Conf.Common.CommonEtcd.Host},
		BasePath:       config.Conf.Common.CommonEtcd.BasePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		logrus.Fatal(err)
	}
	s.Plugins.Add(r)
}

func (logic *Logic) RedisPublishChannel(serverId string, toUserId int, msg []byte) (err error) {
	redisMsg := proto.RedisMsg{
		Op:       config.OpSingleSend,
		ServerId: serverId,
		UserId:   toUserId,
		Msg:      msg,
	}
	redisMsgStr, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic, RedisPublishChannel err %s", err.Error())
		return err
	}
	redisChannel := config.QueueName
	if err := RedisClient.LPush(context.Background(), redisChannel, redisMsgStr).Err(); err != nil {
		logrus.Errorf("logic, lpush err: %s", err.Error())
		return err
	}
	return
}

func (logic *Logic) PublishChannel(serverId string, toUserId int, msg []byte) (err error) {
	// 这里 redisMsg 就是生产者要生产的消息
	redisMsg := proto.RedisMsg{
		Op:       config.OpSingleSend,
		ServerId: serverId,
		UserId:   toUserId,
		Msg:      msg,
	}
	redisMsgStr, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic, kafka produce message err %s", err.Error())
		return err
	}
	producerMessage := &sarama.ProducerMessage{
		Topic: config.Conf.Task.TaskBase.ConsumerTopic,
		Key:   sarama.StringEncoder(strconv.Itoa(toUserId)),
		Value: sarama.StringEncoder(redisMsgStr),
	}
	_, _, err = Producer.SendMessage(producerMessage)
	return err
}

func (logic *Logic) RedisPublishRoomInfo(roomId int, count int, RoomUserInfo map[string]string, msg []byte) (err error) {
	var redisMsg = &proto.RedisMsg{
		Op:           config.OpRoomSend,
		RoomId:       roomId,
		Count:        count,
		Msg:          msg,
		RoomUserInfo: RoomUserInfo,
	}
	redisMsgByte, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic, RedisPublishRoomInfo redisMsg error: %s", err.Error())
		return
	}
	err = RedisClient.LPush(context.Background(), config.QueueName, redisMsgByte).Err()
	if err != nil {
		logrus.Errorf("logic, RedisPublishRoomInfo redisMsg error: %s", err.Error())
		return
	}
	return
}

func (logic *Logic) RedisPushRoomCount(roomId int, count int) (err error) {
	var redisMsg = &proto.RedisMsg{
		Op:     config.OpRoomCountSend,
		RoomId: roomId,
		Count:  count,
	}
	redisMsgByte, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic, RedisPushRoomCount redisMsg error: %s", err.Error())
		return
	}
	err = RedisClient.LPush(context.Background(), config.QueueName, redisMsgByte).Err()
	if err != nil {
		logrus.Errorf("logic, RedisPushRoomCount redisMsg error: %s", err.Error())
		return
	}
	return
}

func (logic *Logic) RedisPushRoomInfo(roomId int, count int, roomUserInfo map[string]string) (err error) {
	var redisMsg = &proto.RedisMsg{
		Op:           config.OpRoomInfoSend,
		RoomId:       roomId,
		Count:        count,
		RoomUserInfo: roomUserInfo,
	}
	redisMsgByte, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic, RedisPushRoomInfo redisMsg error: %s", err.Error())
		return
	}
	err = RedisClient.LPush(context.Background(), config.QueueName, redisMsgByte).Err()
	if err != nil {
		logrus.Errorf("logic, RedisPushRoomInfo redisMsg error: %s", err.Error())
		return
	}
	return
}

func (logic *Logic) getRoomUserKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisRoomPrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}

func (logic *Logic) getRoomOnlineCountKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisRoomOnlinePrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}

func (logic *Logic) getUserKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisPrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}
