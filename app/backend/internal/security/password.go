package security

import (
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12 // Современный уровень безопасности для bcrypt

// HashPassword хеширует пароль с использованием bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword проверяет пароль против хеша (time-constant)
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
