package helpers

import (
	"crypto/rand"
	"fmt"
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

func (s StringGenerator) Generate(prefix string, length int) (string, error) {
	var alphaNumericRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	randomString := make([]rune, length)
	for i := range randomString {
		charIndex, err := rand.Int(s.Reader, big.NewInt(int64(len(alphaNumericRunes))))
		if err != nil {
			return "", err
		}
		randomString[i] = alphaNumericRunes[charIndex.Int64()]
	}
	return fmt.Sprintf("%s%s", prefix, string(randomString)), nil
}
