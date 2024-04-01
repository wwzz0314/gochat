package proto

type RedisMsg struct {
	Op           int               `json:"op,omitempty"`
	ServerId     string            `json:"serverId,omitempty"`
	RoomId       int               `json:"roomId,omitempty"`
	UserId       int               `json:"userId,omitempty"`
	Msg          []byte            `json:"msg,omitempty"`
	Count        int               `json:"count,omitempty"`
	RoomUserInfo map[string]string `json:"roomUserInfo,omitempty"`
}

type RedisRoomInfo struct {
	Op           int               `json:"op,omitempty"`
	RoomId       int               `json:"roomId,omitempty"`
	Count        int               `json:"count,omitempty"`
	RoomUserInfo map[string]string `json:"roomUserInfo,omitempty"`
}

type RedisRoomCountMsg struct {
	Count int `json:"count,omitempty"`
	Op    int `json:"op,omitempty"`
}

type SuccessReply struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
}
