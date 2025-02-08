package main

import (
	"backend/config"
	"backend/db"
	"backend/ping"
	"backend/wsUpdater"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type dbHandler func(*gin.Context, *event_db.DB)

func getEvents(c *gin.Context, db *event_db.DB) {
	ls, err := db.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, ls)
}

func addEvent(c *gin.Context, db *event_db.DB) {
	var events ping.PingEventList
	if err := c.BindJSON(&events); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := db.PutBulk(events); err != nil {
		log.Println(err)
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

func updateWrap(handler gin.HandlerFunc, u *ws_updater.UpdateHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handler(ctx)
		go u.SendUpdate()
	}
}

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := event_db.GetDB(cfg)
	if err != nil {
		log.Fatal("Couldn't connect to database:\n", err)
	}

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

	updHandler := &ws_updater.UpdateHandler{}

	router.POST("/events", updateWrap(dbWrap(addEvent, db), updHandler))
	router.GET("/events", dbWrap(getEvents, db))
	router.GET("/ws", updHandler.RegisterConnection)

	if err := router.Run(":4242"); err != nil {
		log.Println("Engine failed:\n", err)
	}

	log.Println("Shutting down...")
}
