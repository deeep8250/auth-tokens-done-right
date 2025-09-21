package tokens

import (
	"time"

	"authsvc/internal/keys"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateJWT(email string, keyManager *keys.Manager) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
		"iat":   time.Now().Unix(),
		"iss":   "your-app",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Sign with the private key from the shared keyManager
	return token.SignedString(keyManager.Private())
}
