package routes

import (
	"authsvc/internal/handlers"
	"authsvc/internal/keys"
	"authsvc/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Routes(r *gin.Engine, keyManager *keys.Manager) {
	// Public routes
	r.POST("/login", handlers.Login(keyManager))
	r.POST("/signin", handlers.SignIn)

	// Protected routes (JWT middleware with the same keyManager)
	protected := r.Group("/auth")
	protected.Use(middleware.AuthMiddleware(keyManager))
	{
		protected.GET("/profile", handlers.Profile)
	}
}
