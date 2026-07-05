package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Claims struct {
	Subject  int64  `json:"sub"`
	Username string `json:"username,omitempty"`
	Expires  int64  `json:"exp"`
	IssuedAt int64  `json:"iat"`
}

func SignJWT(secret string, userID int64, username string, ttl time.Duration) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", fmt.Errorf("jwt secret is required")
	}
	now := time.Now()
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	claims := Claims{Subject: userID, Username: username, IssuedAt: now.Unix(), Expires: now.Add(ttl).Unix()}
	headerRaw, _ := json.Marshal(header)
	claimsRaw, _ := json.Marshal(claims)
	unsigned := base64.RawURLEncoding.EncodeToString(headerRaw) + "." + base64.RawURLEncoding.EncodeToString(claimsRaw)
	signature := sign(unsigned, secret)
	return unsigned + "." + signature, nil
}

func VerifyJWT(secret, token string) (*Claims, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid jwt")
	}
	unsigned := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(sign(unsigned, secret))) {
		return nil, fmt.Errorf("invalid jwt signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims Claims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, err
	}
	if claims.Subject <= 0 {
		return nil, fmt.Errorf("invalid jwt subject")
	}
	if claims.Expires > 0 && time.Now().Unix() > claims.Expires {
		return nil, fmt.Errorf("jwt expired")
	}
	return &claims, nil
}

func sign(unsigned, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func SubjectString(userID int64) string {
	return strconv.FormatInt(userID, 10)
}
