package redis

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	RDB *redis.Client
	Ctx = context.Background()
)

func Init() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("GET_ADDR"),
		Password: os.Getenv("GET_PASSWORD"),
		DB:       0,
	})
}

func StoreRefreshToken(userID uint, refreshToken string) error {
	// Set token with 7-day expiration
	key := strconv.FormatUint(uint64(userID), 10)
	return RDB.Set(Ctx, key, refreshToken, 7*24*time.Hour).Err()
}
