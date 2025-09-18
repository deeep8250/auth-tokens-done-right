package middleware

import (
	"authsvc/internal/keys"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// JWT Authentication middleware
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is missing",
			})
			c.Abort()
			return
		}

		// Expect format "Bearer <token>"
		tokenString := strings.Split(authHeader, " ")[1]
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Bearer token is missing",
			})
			c.Abort()
			return
		}

		// Retrieve the public key from the Manager (assuming keys.New() gives us the key pair)
		keyManager, err := keys.New()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not load public key",
			})
			c.Abort()
			return
		}

		// Parse the JWT token and verify it using the public key
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			// Validate the token's signing method (RS256)
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.NewValidationError("Invalid signing method", jwt.ValidationErrorSignatureInvalid)
			}
			// Return the public key for verification
			return keyManager.Public(), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Successfully validated the JWT, now extract claims
		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid claims in token",
			})
			c.Abort()
			return
		}

		// Extract email or any other information from the claims and set it in context
		c.Set("email", claims.Subject)

		// Continue to the next handler
		c.Next()
	}
}
