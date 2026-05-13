package utils

import (
	"crypto/rand"
	"encoding/base64"
)

type RandomUtil struct{}

func NewRandom() *RandomUtil {
	return &RandomUtil{}
}

func (u *RandomUtil) GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (u *RandomUtil) GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}
