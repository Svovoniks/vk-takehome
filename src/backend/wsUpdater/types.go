package ws_updater

import (
	"log"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type UpdateHandler struct {
	busy          sync.Mutex
	wsConnections []*websocket.Conn
}

func (u *UpdateHandler) SendUpdate() {
	if !u.busy.TryLock() {
		return
	}
	defer u.busy.Unlock()

	onDoneChannel := make(chan *websocket.Conn)

	newCons := make([]*websocket.Conn, 0, len(u.wsConnections))
	for _, wsc := range u.wsConnections {
		go pingSocet(wsc, onDoneChannel)
	}

	for range u.wsConnections {
		if ws := <-onDoneChannel; ws != nil {
			newCons = append(newCons, ws)
		}
	}

	u.wsConnections = newCons
}

func (u *UpdateHandler) RegisterConnection(c *gin.Context) {
	u.busy.Lock()
	defer u.busy.Unlock()

	ws, err := getWebsocket(c)
	if err != nil {
		log.Printf("Couldn't register websocket\n%v\n", err)
		return
	}
	u.wsConnections = append(u.wsConnections, ws)
}
