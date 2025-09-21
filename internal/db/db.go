package db

import (
	"authsvc/internal/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func DBInit() {
	dsn1 := os.Getenv("DB_DSN")
	fmt.Println("DSN : ", dsn1)
	dsn := os.Getenv("DB_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("cannot connect with db")
	}

	db.AutoMigrate(models.User{})
	DB = db

}
