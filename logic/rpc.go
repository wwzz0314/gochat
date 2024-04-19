package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gochat/config"
	"gochat/logic/dao"
	"gochat/proto"
	"gochat/tools"
	"strconv"
	"time"
)

type RpcLogic struct {
}

func (rpc *RpcLogic) Register(ctx context.Context, args *proto.RegisterRequest, reply *proto.RegisterReply) (err error) {
	reply.Code = config.FailReplyCode
	u := new(dao.User)
	uData := u.CheckHaveUserName(args.Name)
	if uData.Id > 0 {
		return errors.New("this user name already have, please login")
	}
	u.UserName = args.Name
	u.Password = args.Password
	userId, err := u.Add()
	if err != nil {
		logrus.Infof("register err: %s", err.Error())
		return err
	}
	if userId == 0 {
		return errors.New("register userId empty")
	}
	reply.Code = config.SuccessReplyCode
	return
}

func (rpc *RpcLogic) Login(ctx context.Context, args *proto.LoginRequest, reply *proto.LoginResponse) (err error) {
	reply.Code = config.FailReplyCode
	u := new(dao.User)
	userName := args.Name
	passWord := args.Password
	data := u.CheckHaveUserName(userName)
	if data.Id == 0 || passWord != data.Password {
		return errors.New("no this user or password error!")
	}
	loginSessionId := tools.GetSessionIdByUserId(data.Id)
	randToken := tools.GetRandomToken(32)
	sessionId := tools.CreateSessionId(randToken)
	userData := make(map[string]interface{})
	userData["userId"] = data.Id
	userData["userName"] = data.UserName
	token, _ := RedisClient.Get(ctx, loginSessionId).Result()
	if token != "" {
		oldSession := tools.CreateSessionId(token)
		err := RedisSessClient.Del(ctx, oldSession).Err()
		if err != nil {
			return errors.New("logout user fail! token is: " + token)
		}
	}
	RedisSessClient.Do(ctx, "MULTI")
	RedisSessClient.HMSet(ctx, sessionId, userData)
	RedisSessClient.Expire(ctx, sessionId, 86400*time.Second)
	RedisSessClient.Set(ctx, loginSessionId, randToken, 86400*time.Second)
	err = RedisSessClient.Do(ctx, "EXEC").Err()
	if err != nil {
		logrus.Infof("register set redis token fail")
		return err
	}
	reply.Code = config.SuccessReplyCode
	reply.AuthToken = randToken
	return
}

func (rpc *RpcLogic) GetUserInfoByUserId(ctx context.Context, args *proto.GetUserInfoRequest, reply *proto.GetUserInfoResponse) (err error) {
	reply.Code = config.FailReplyCode
	userId := args.UserId
	u := new(dao.User)
	userName := u.GetUserNameByUserId(userId)
	reply.UserId = userId
	reply.UserName = userName
	reply.Code = config.SuccessReplyCode
	return
}

func (rpc *RpcLogic) CheckAuth(ctx context.Context, args *proto.CheckAuthRequest, reply *proto.CheckAuthResponse) (err error) {
	reply.Code = config.FailReplyCode
	authToken := args.AuthToken
	sessionName := tools.GetSessionName(authToken)
	var userDataMap = map[string]string{}
	userDataMap, err = RedisSessClient.HGetAll(ctx, sessionName).Result()
	if err != nil {
		logrus.Infof("check auth fail!, authToken is :%s", authToken)
		return err
	}
	if len(userDataMap) == 0 {
		logrus.Infof("no this user session, authToken is: %s", authToken)
		return
	}
	intUserId, _ := strconv.Atoi(userDataMap["userId"])
	reply.UserId = intUserId
	userName, _ := userDataMap["userName"]
	reply.Code = config.SuccessReplyCode
	reply.UserName = userName
	return
}

func (rpc *RpcLogic) Logout(ctx context.Context, args *proto.LogoutRequest, reply *proto.LogoutResponse) (err error) {
	reply.Code = config.FailReplyCode
	authToken := args.AuthToken
	sessionName := tools.GetSessionName(authToken)

	var userDataMap = map[string]string{}
	userDataMap, err = RedisSessClient.HGetAll(ctx, sessionName).Result()
	if err != nil {
		logrus.Infof("check auth fail!, authToken is: %s", authToken)
		return err
	}
	if len(userDataMap) == 0 {
		logrus.Infof("no this user session, authToekn is: %s", authToken)
		return
	}
	intUserId, _ := strconv.Atoi(userDataMap["userId"])
	sessIdMap := tools.GetSessionIdByUserId(intUserId)
	err = RedisSessClient.Del(ctx, sessIdMap).Err()
	if err != nil {
		logrus.Infof("logout del sess map error:%s", err.Error())
		return err
	}
	logic := new(Logic)
	severIdKey := logic.getUserKey(fmt.Sprintf("%d", intUserId))
	err = RedisSessClient.Del(ctx, severIdKey).Err()
	if err != nil {
		logrus.Infof("logout del server id error: %s", err.Error())
		return err
	}
	err = RedisSessClient.Del(ctx, sessionName).Err()
	if err != nil {
		logrus.Infof("logout error: %s", err.Error())
		return err
	}
	reply.Code = config.SuccessReplyCode
	return
}

func (rpc *RpcLogic) Push(ctx context.Context, args *proto.Send, reply *proto.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	sendData := args
	var bodyBytes []byte
	bodyBytes, err = json.Marshal(sendData)
	if err != nil {
		logrus.Errorf("logic, push msg fail err: %s", err.Error())
		return
	}
	logic := new(Logic)
	userSidKey := logic.getUserKey(fmt.Sprintf("%d", sendData.ToUserId))
	serverIdStr := RedisSessClient.Get(ctx, userSidKey).Val()
	if err != nil {
		logrus.Errorf("logic, push parse int fail: %s", err.Error())
		return
	}
	//err = logic.RedisPublishChannel(serverIdStr, sendData.ToUserId, bodyBytes)
	err = logic.PublishChannel(serverIdStr, sendData.ToUserId, bodyBytes)
	if err != nil {
		logrus.Errorf("logic, redis publish err: %s", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

func (rpc *RpcLogic) PushRoom(ctx context.Context, args *proto.Send, reply *proto.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	sendData := args
	roomId := sendData.RoomId
	logic := new(Logic)
	roomUserInfo := make(map[string]string)
	roomUserKey := logic.getRoomUserKey(strconv.Itoa(roomId))
	roomUserInfo, err = RedisClient.HGetAll(ctx, roomUserKey).Result()
	if err != nil {
		logrus.Errorf("logic, PushRoom redis hGetAll err: %s", err.Error())
		return
	}
	var bodyBytes []byte
	sendData.RoomId = roomId
	sendData.Msg = args.Msg
	sendData.FromUserId = args.FromUserId
	sendData.FromUserName = args.FromUserName
	sendData.Op = config.OpRoomSend
	sendData.CreatTime = tools.GetNowDateTime()
	bodyBytes, err = json.Marshal(sendData)
	if err != nil {
		logrus.Errorf("logiic, PushRoom Marshal err: %s", err.Error())
		return
	}
	err = logic.RedisPublishRoomInfo(roomId, len(roomUserInfo), roomUserInfo, bodyBytes)
	if err != nil {
		logrus.Errorf("logic,  PushRoom err: %s", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

func (rpc *RpcLogic) Count(ctx context.Context, args *proto.Send, reply *proto.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	roomId := args.RoomId
	logic := new(Logic)
	var count int
	count, err = RedisSessClient.Get(ctx, logic.getRoomOnlineCountKey(fmt.Sprintf("%d", roomId))).Int()
	err = logic.RedisPushRoomCount(roomId, count)
	if err != nil {
		logrus.Errorf("logic, Count err: %s", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

func (rpc *RpcLogic) GetRoomInfo(ctx context.Context, args *proto.Send, reply *proto.SuccessReply) (err error) {
	reply.Code = config.FailReplyCode
	logic := new(Logic)
	roomId := args.RoomId
	roomUserInfo := make(map[string]string)
	roomUserKey := logic.getRoomUserKey(strconv.Itoa(roomId))
	roomUserInfo, err = RedisClient.HGetAll(ctx, roomUserKey).Result()
	if len(roomUserInfo) == 0 {
		return errors.New("getRoomInfo no this user")
	}
	err = logic.RedisPushRoomInfo(roomId, len(roomUserInfo), roomUserInfo)
	if err != nil {
		logrus.Errorf("logic, GerRoomInfo err: %s", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	return
}

func (rpc *RpcLogic) Connect(ctx context.Context, args *proto.ConnectRequest, reply *proto.ConnectReply) (err error) {
	if args == nil {
		logrus.Errorf("logic, connect args empty")
		return
	}
	logic := new(Logic)
	logrus.Infof("logic,authToken is %s", args.AuthToken)
	key := tools.GetSessionName(args.AuthToken)
	userInfo, err := RedisClient.HGetAll(ctx, key).Result()
	if err != nil {
		logrus.Errorf("RedisCli HGetAll key : %s, err: %s", key, err.Error())
		return
	}
	if len(userInfo) == 0 {
		reply.UserId = 0
		return
	}
	reply.UserId, _ = strconv.Atoi(userInfo["userId"])
	roomUserKey := logic.getRoomUserKey(strconv.Itoa(args.RoomId))
	if reply.UserId != 0 {
		userKey := logic.getUserKey(fmt.Sprintf("%d", reply.UserId))
		logrus.Infof("logic redis set userKey: %s, serverId : %s", userKey, args.ServerId)
		validTime := config.RedisBaseValidTime * time.Second
		err = RedisClient.Set(ctx, userKey, args.ServerId, validTime).Err()
		if err != nil {
			logrus.Warnf("logic set err: %s", err)
		}
		if RedisClient.HGet(ctx, roomUserKey, fmt.Sprintf("%d", reply.UserId)).Val() == "" {
			RedisClient.HSet(ctx, roomUserKey, fmt.Sprintf("%d", reply.UserId), userInfo["userName"])
			RedisClient.Incr(ctx, logic.getRoomOnlineCountKey(fmt.Sprintf("%d", args.RoomId)))
		}
	}
	logrus.Infof("logic rpc userId:  %d", reply.UserId)
	return
}

func (rpc *RpcLogic) DisConnect(ctx context.Context, args *proto.DisConnectRequest, reply *proto.DisConnectReply) (err error) {
	logic := new(Logic)
	roomUserKey := logic.getRoomUserKey(strconv.Itoa(args.RoomId))
	if args.RoomId > 0 {
		count, _ := RedisSessClient.Get(ctx, logic.getRoomOnlineCountKey(fmt.Sprintf("%d", args.RoomId))).Int()
		if count > 0 {
			RedisClient.Decr(ctx, logic.getRoomOnlineCountKey(fmt.Sprintf("%d", args.RoomId))).Result()
		}
	}
	if args.UserId != 0 {
		err = RedisClient.HDel(ctx, roomUserKey, fmt.Sprintf("%d", args.UserId)).Err()
		if err != nil {
			logrus.Warnf("HDel getRoomUserKey err: %s", err)
		}
	}
	roomUserInfo, err := RedisClient.HGetAll(ctx, roomUserKey).Result()
	if err != nil {
		logrus.Warnf("RedisCli HGetAll roomUserInfo key: %s, err :%s", roomUserKey, err)
	}
	if err = logic.RedisPublishRoomInfo(args.RoomId, len(roomUserInfo), roomUserInfo, nil); err != nil {
		logrus.Warnf("publish RedisPublishhRoomCount err: %s", err.Error())
		return
	}
	return
}
