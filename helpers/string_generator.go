package helpers

import (
	"crypto/rand"
	"io"
	"math/big"
)

type StringGenerator struct {
	Reader io.Reader
}

func NewStringGenerator(reader io.Reader) StringGenerator {
	return StringGenerator{
		Reader: reader,
	}
}

func (s StringGenerator) Generate(length int) (string, error) {
	var alphaNumericRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	randomString := make([]rune, length)
	for i := range randomString {
		charIndex, err := rand.Int(s.Reader, big.NewInt(int64(len(alphaNumericRunes))))
		if err != nil {
			return "", err
		}
		randomString[i] = alphaNumericRunes[charIndex.Int64()]
	}
	return string(randomString), nil
}
