package authn

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const apiKeyPrefix = "wmcp_"

// GenerateAPIKey returns a plaintext token (shown once) and its sha256 hash (stored).
func GenerateAPIKey() (plain, hash string, err error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	plain = apiKeyPrefix + hex.EncodeToString(raw)
	return plain, HashAPIKey(plain), nil
}

func HashAPIKey(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

func IsAPIKeyFormat(s string) bool {
	return len(s) > len(apiKeyPrefix) && s[:len(apiKeyPrefix)] == apiKeyPrefix
}
