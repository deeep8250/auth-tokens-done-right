package users

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// cost controls CPU/time. 12–14 is common in prod; start with 12 locally.
const cost = 12

// HashPassword hashes the plaintext using bcrypt.
// The returned string contains salt and cost; store it directly.
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("empty password")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// VerifyPassword compares a plaintext password against a stored bcrypt hash.
func VerifyPassword(plain, stored string) bool {
	return bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain)) == nil
}
