package ec2

import (
	"fmt"
	"io"
)

type UUIDGenerator struct {
	reader io.Reader
}

func NewUUIDGenerator(reader io.Reader) UUIDGenerator {
	return UUIDGenerator{reader}
}

func (g UUIDGenerator) Generate() (string, error) {
	var buf [16]byte

	_, err := g.reader.Read(buf[:])
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:]), nil
}
