package security

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword создаёт хэш пароля используя bcrypt
func HashPassword(password string) (string, error) {
	// Используем стандартный cost bcrypt (10)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword проверяет соответствие пароля и хэша (тайм-константно)
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}