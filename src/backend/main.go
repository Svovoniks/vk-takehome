package main

import (
	"backend/config"
	"backend/db"
	"backend/ping"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type dbHandler func(*gin.Context, *event_db.DB)

type updateHandler struct {
	busy          sync.Mutex
	wsConnections []*websocket.Conn
}

func pingSocet(wsc *websocket.Conn, ch chan<- *websocket.Conn) {
	if err := wsc.WriteMessage(websocket.TextMessage, []byte{}); err != nil {
		wsc.Close()
		ch <- nil
		return
	}
	ch <- wsc
}

// sends update to all listeners if there is no update that is already in progress
// removes listeners that have disconnected
func (u *updateHandler) sendUpdate() {
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

func (u *updateHandler) registerConnection(c *gin.Context) {
	u.busy.Lock()
	defer u.busy.Unlock()

	u.wsConnections = append(u.wsConnections, getWebsocket(c))
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func getEvents(c *gin.Context, db *event_db.DB) {
	ls, err := db.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, ls)
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

func getWebsocket(c *gin.Context) *websocket.Conn {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket:\n%v", err)
		return nil
	}

	go func() {
		discardWsInput(conn)
	}()

	return conn
}

func addEvent(c *gin.Context, db *event_db.DB) {
	var events ping.PingEventList
	if err := c.BindJSON(&events); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := db.PutBulk(events); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}

func dbWrap(handler dbHandler, db *event_db.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handler(ctx, db)
	}
}

func updateWrap(handler gin.HandlerFunc, u *updateHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handler(ctx)
		go u.sendUpdate()
	}
}

func main() {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	updHandler := &updateHandler{}

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := event_db.GetDB(cfg)
	if err != nil {
		log.Fatal("Couldn't connect to database:\n", err)
	}

	router.POST("/events", updateWrap(dbWrap(addEvent, db), updHandler))
	router.GET("/events", dbWrap(getEvents, db))
	router.GET("/ws", updHandler.registerConnection)

	if err := router.Run(":4242"); err != nil {
		log.Println("Engine failed:\n", err)
	}

	log.Println("Shutting down...")
}
