package helpers

import (
	"crypto/rand"
	"io"
	"math/big"
)

type PasswordGenerator struct {
	Reader io.Reader
}

func NewPasswordGenerator(reader io.Reader) PasswordGenerator {
	return PasswordGenerator{
		Reader: reader,
	}
}

const PASSWORD_LENGTH = 15

func (p PasswordGenerator) Generate() (string, error) {
	var alphaNumericRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	password := make([]rune, PASSWORD_LENGTH)
	for i := range password {
		charIndex, err := rand.Int(p.Reader, big.NewInt(int64(len(alphaNumericRunes))))
		if err != nil {
			return "", err
		}
		password[i] = alphaNumericRunes[charIndex.Int64()]
	}
	return string(password), nil
}
