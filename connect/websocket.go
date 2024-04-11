package connect

import (
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gochat/config"
	"net/http"
)

func (c *Connect) InitWebsocket() error {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c.serveWs(DefaultServer, w, r)
	})
	err := http.ListenAndServe(config.Conf.Connect.ConnectWebsocket.Bind, nil)
	return err
}

// 将 http 服务升级成 websocket 服务
func (c *Connect) serveWs(server *Server, w http.ResponseWriter, r *http.Request) {
	var upGrader = websocket.Upgrader{
		ReadBufferSize:  server.Options.ReadBufferSize,
		WriteBufferSize: server.Options.WriteBufferSize,
	}
	// cross origin domain support
	upGrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upGrader.Upgrade(w, r, nil)

	if err != nil {
		logrus.Errorf("serverWs err:%s", err.Error())
		return
	}
	var ch *Channel
	// 	default broadcast size eq 512
	ch = NewChannel(server.Options.BroadcastSize)
	ch.conn = conn
	// send data to websocket conn
	go server.writePump(ch, c)
	// get data from websocket conn
	go server.readPump(ch, c)
}
