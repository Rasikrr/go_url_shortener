package hash

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher interface {
	Hash(string) (string, error)
	CheckPassword(string, string) (bool, error)
}

type SHA1Hasher struct {
	salt int
}

func NewSHA1Hasher(salt int) *SHA1Hasher {
	return &SHA1Hasher{
		salt: salt,
	}
}

func (h *SHA1Hasher) Hash(password string) (string, error) {
	const fn = "lib.hash.Hash"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.salt)
	if err != nil {
		return "", fmt.Errorf("%s : %w", fn, err)
	}
	return string(hash), nil
}

func (h *SHA1Hasher) CheckPassword(hashedPassword string, password string) (bool, error) {
	const fn = "lib.hash.CheckPassword"

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, fmt.Errorf("%s : %w", fn, err)
	}
	return true, nil
}
