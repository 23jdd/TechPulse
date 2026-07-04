package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func SHA256(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(strings.ToLower(value))))
	return hex.EncodeToString(sum[:])
}
