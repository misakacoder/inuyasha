package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Claims struct {
	*jwt.RegisteredClaims
	Payload map[string]any `json:"payload"`
}

type Manager interface {
	Create(subject string, extension map[string]any) (string, *Claims)
	Parse(tokenString string) (*Claims, error)
}

type manager struct {
	secretKey  string
	expireTime time.Duration
}

func (manager *manager) Create(subject string, payload map[string]any) (string, *Claims) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: &jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(now.Add(manager.expireTime)),
		},
		Payload: payload,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(manager.secretKey))
	return tokenString, claims
}

func (manager *manager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("invalid signing algorithm")
		}
		return []byte(manager.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func NewManager(secretKey string, expireTime time.Duration) Manager {
	return &manager{
		secretKey:  secretKey,
		expireTime: expireTime,
	}
}
