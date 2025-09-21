package tokens

import (
	redis "authsvc/internal/Redis"
	"crypto/rand"
	"encoding/base64"
)

// GenerateRefreshToken creates a secure random token
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32) // 256-bit token
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func DeleteRefreshToken(userID string) error {
	return redis.RDB.Del(redis.Ctx, userID).Err()
}
