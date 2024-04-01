package proto

type Msg struct {
	Ver       int    `json:"ver,omitempty"`       // protocol version
	Operation int    `json:"operation,omitempty"` // operation for request
	SeqId     string `json:"seqId,omitempty"`     // sequence number chosen by client
	Body      []byte `json:"body,omitempty"`      // binary body bytes
}

type PushMsgRequest struct {
	UserId int
	Msg    Msg
}

type PushRoomMsgRequest struct {
	RoomId int
	Msg    Msg
}

type PushRoomCountRequest struct {
	RoomId int
	Count  int
}
