package main

import (
	redis "authsvc/internal/Redis"
	"authsvc/internal/db"
	"authsvc/internal/keys"
	"authsvc/internal/routes"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var keyManager *keys.Manager

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(err.Error())
	}

	redis.Init()
	db.DBInit()
	port := "8080"

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	keyManager, err := keys.New() // create ONE key pair at startup
	if err != nil {
		log.Fatal("cannot generate RSA keys:", err)
	}

	routes.Routes(r, keyManager)

	log.Println("listening on :", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err.Error())
	}
}
