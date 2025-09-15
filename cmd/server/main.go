package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func getenv(p, d string) string {
	if v := os.Getenv(p); v != "" {
		return p
	}
	return d
}

func main() {

	port := getenv("PORT", "8080")

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"ok":      true,
			"service": "authsvc",
		})
	})

	log.Println("listening on : ", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err.Error())
	}

}
