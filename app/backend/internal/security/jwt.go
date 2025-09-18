package security

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService управляет JWT токенами
type JWTService struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTService создает новый сервис для работы с JWT
func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secret: []byte(secret),
		ttl:    24 * time.Hour, // 24 часа согласно требованиям
	}
}

// Claims содержит данные JWT токена
type Claims struct {
	UserID      string `json:"sub"`
	Login       string `json:"login"`
	DisplayName string `json:"displayName"`
	jwt.RegisteredClaims
}

// GenerateToken создает JWT токен для пользователя
func (j *JWTService) GenerateToken(userID, login, displayName string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(j.ttl)

	claims := &Claims{
		UserID:      userID,
		Login:       login,
		DisplayName: displayName,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, exp, nil
}

// ValidateToken проверяет и парсит JWT токен
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// Константы для контекста
type contextKey string

const UserIDKey contextKey = "userID"

// GetUserIDFromContext извлекает ID пользователя из контекста запроса
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
