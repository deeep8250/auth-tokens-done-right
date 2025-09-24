package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type RateLimit struct {
	Redis     *redis.Client
	Max       int
	Window    time.Duration
	BlockTime time.Duration
}

func (rl *RateLimit) RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx := context.Background()

		ip := c.ClientIP()
		ip = strings.TrimSpace(ip)

		// check if ip is blocked or not
		blockKey := "block" + ip
		blocked, _ := rl.Redis.Exists(ctx, blockKey).Result()
		if blocked > 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many wrong atteps at the same time , plase try again later",
			})
			c.Abort()
			return
		}

		// request counter key
		countKey := "rate" + ip
		count, err := rl.Redis.Incr(ctx, countKey).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		if count == 1 {
			rl.Redis.Expire(ctx, countKey, rl.Window)
		}

		if count > int64(rl.Max) {
			rl.Redis.Set(ctx, blockKey, "1", rl.BlockTime)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, you are blocked temporarily",
			})
			c.Abort()
			return
		}
	}
}
