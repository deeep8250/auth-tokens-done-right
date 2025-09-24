package main

import (
	redis "authsvc/internal/Redis"
	"authsvc/internal/db"
	"authsvc/internal/keys"
	"authsvc/internal/middleware"
	"authsvc/internal/routes"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	redis.Init()
	limiter := &middleware.RateLimit{
		Redis:     redis.RDB,
		Max:       5,
		Window:    10 * time.Minute,
		BlockTime: 5 * time.Minute,
	}

	if err := godotenv.Load(); err != nil {
		log.Println(err.Error())
	}

	db.DBInit()
	port := "8080"

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	keyManager, err := keys.New() // create ONE key pair at startup
	if err != nil {
		log.Fatal("cannot generate RSA keys:", err)
	}

	routes.Routes(r, keyManager, limiter)

	log.Println("listening on :", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err.Error())
	}
}
