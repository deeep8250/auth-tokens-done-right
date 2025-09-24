package routes

import (
	"authsvc/internal/handlers"
	"authsvc/internal/keys"
	"authsvc/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Routes(r *gin.Engine, keyManager *keys.Manager, limiter *middleware.RateLimit) {
	// Public routes
	r.POST("/login", limiter.RateLimiter(), handlers.Login(keyManager))
	r.POST("/signin", handlers.SignIn)

	// Protected routes (JWT middleware with the same keyManager)
	protected := r.Group("/auth")
	protected.Use(middleware.AuthMiddleware(keyManager))
	{
		protected.GET("/profile", handlers.Profile)
		protected.POST("/reset-request", handlers.RequestReset)
		protected.POST("/reset", handlers.VerifyReset)
	}
	r.GET("/.well-known/jwks.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", keyManager.JWKS())
	})

}
