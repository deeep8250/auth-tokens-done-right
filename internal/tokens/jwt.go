package tokens

import (
	"authsvc/internal/keys"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Create a function that generates the JWT with RS256 (using private key)
func GenerateJWT(email string) (string, error) {
	// Create a new instance of the RSA Manager (this contains the private key)
	keyManager, err := keys.New() // assuming `keys.New()` creates the key pair
	if err != nil {
		log.Println("Error generating RSA keys:", err)
		return "", err
	}

	// JWT claims (you can add other claims as needed)
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 1).Unix(), // Set expiration for 1 hour
		"iat":   time.Now().Unix(),                    // Issued At (optional)
		"iss":   "your-app",                           // Issuer (optional)
	}

	// Create a new JWT token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Sign the token with the private key
	// `keyManager.Private()` returns the private key for signing
	tokenString, err := token.SignedString(keyManager.Private())
	if err != nil {
		log.Println("Error signing the token:", err)
		return "", err
	}

	return tokenString, nil
}
