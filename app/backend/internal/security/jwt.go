package security

import (
	"errors"
	"fmt"
	"time"
	
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims представляет JWT claims
type Claims struct {
	UserID      uuid.UUID `json:"sub"`
	Login       string    `json:"login"`
	DisplayName string    `json:"displayName"`
	jwt.RegisteredClaims
}

// JWTManager управляет JWT токенами
type JWTManager struct {
	secret        []byte
	tokenDuration time.Duration
}

// NewJWTManager создаёт новый менеджер JWT
func NewJWTManager(secret string, duration time.Duration) *JWTManager {
	return &JWTManager{
		secret:        []byte(secret),
		tokenDuration: duration,
	}
}

// GenerateToken создаёт новый JWT токен для пользователя
func (m *JWTManager) GenerateToken(userID uuid.UUID, login, displayName string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.tokenDuration)
	
	claims := Claims{
		UserID:      userID,
		Login:       login,
		DisplayName: displayName,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}
	
	return tokenString, expiresAt, nil
}

// ValidateToken проверяет и парсит JWT токен
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	
	// Дополнительная проверка времени жизни
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	
	return claims, nil
}