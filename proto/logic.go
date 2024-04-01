package proto

type LoginRequest struct {
	Name     string
	Password string
}

type LoginResponse struct {
	Code      int
	AuthToken string
}

type GetUserInfoRequest struct {
	UserId int
}

type GetUserInfoResponse struct {
	Code     int
	UserId   int
	UserName string
}

type RegisterRequest struct {
	Name     string
	Password string
}

type RegisterReply struct {
	Code      int
	AuthToken string
}

type LogoutRequest struct {
	AuthToken string
}

type LogoutResponse struct {
	Code int
}

type CheckAuthRequest struct {
	AuthToken string
}

type CheckAuthResponse struct {
	Code     int
	UserId   int
	UserName string
}

type ConnectRequest struct {
	AuthToken string `json:"authToken"`
	RoomId    int    `json:"roomId"`
	ServerId  string `json:"serverId"`
}

type ConnectReply struct {
	UserId int
}

type DisConnectRequest struct {
	RoomId int
	UserId int
}

type DisConnectReply struct {
	Has bool
}

type Send struct {
	Code         int    `json:"code,omitempty"`
	Msg          string `json:"msg,omitempty"`
	FromUserId   int    `json:"fromUserId,omitempty"`
	FromUserName string `json:"fromUserName,omitempty"`
	ToUserId     int    `json:"toUserId,omitempty"`
	ToUserName   string `json:"toUserName,omitempty"`
	RoomId       int    `json:"roomId,omitempty"`
	Op           int    `json:"op,omitempty"`
	CreatTime    string `json:"creatTime,omitempty"`
}

type SendTcp struct {
	Code         int    `json:"code,omitempty"`
	Msg          string `json:"msg,omitempty"`
	FromUserId   int    `json:"fromUserId,omitempty"`
	FromUserName string `json:"fromUserName,omitempty"`
	ToUserId     int    `json:"toUserId,omitempty"`
	ToUserName   string `json:"toUserName,omitempty"`
	RoomId       int    `json:"roomId,omitempty"`
	Op           int    `json:"op,omitempty"`
	CreateTime   string `json:"createTime,omitempty"`
	AuthToken    string `json:"authToken,omitempty"`
}
