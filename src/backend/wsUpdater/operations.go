package ws_updater

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func pingSocet(wsc *websocket.Conn, ch chan<- *websocket.Conn) {
	if err := wsc.WriteMessage(websocket.TextMessage, []byte{}); err != nil {
		wsc.Close()
		ch <- nil
		return
	}
	ch <- wsc
}

func discardWsInput(conn *websocket.Conn) {
	for {
		if _, _, err := conn.NextReader(); err != nil {
			fmt.Println("Closing")
			conn.Close()
			break
		}
	}
}

func getWebsocket(c *gin.Context) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return nil, err
	}

	go func() {
		discardWsInput(conn)
	}()

	return conn, nil
}
