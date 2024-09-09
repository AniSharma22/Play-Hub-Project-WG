package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func GetHashedPassword(password []byte) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword verifies if the given password matches the stored hash.
func VerifyPassword(password []byte, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), password)
	return err == nil
}
