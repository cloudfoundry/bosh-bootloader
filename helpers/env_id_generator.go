package helpers

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"time"
)

type EnvIDGenerator struct {
	reader io.Reader
}

func NewEnvIDGenerator(reader io.Reader) EnvIDGenerator {
	return EnvIDGenerator{
		reader: reader,
	}
}

func (e EnvIDGenerator) Generate() (string, error) {
	lake, err := e.randomLake()
	if err != nil {
		return "", err
	}
	timestamp := time.Now().UTC().Format("2006-01-02T15:04Z")

	return fmt.Sprintf("bbl-env-%s-%s", lake, timestamp), nil
}

func (e EnvIDGenerator) randomLake() (string, error) {
	lakes := []string{
		"caspian",
		"superior",
		"victoria",
		"huron",
		"michigan",
		"tanganyika",
		"baikal",
		"great-bear",
		"malawi",
		"erie",
		"winnipeg",
		"ontario",
		"ladoga",
		"balkhash",
		"vostok",
		"onega",
		"titicaca",
		"nicaragua",
		"athabasca",
		"taymyr",
		"turkana",
		"reindeer",
		"issyk-kul",
		"urmia",
		"vanern",
		"albert",
		"mweru",
		"nettilling",
		"sarygamysh",
		"nipigon",
		"manitoba",
		"great-salt",
		"qinghai",
		"saimaa",
		"khanka",
	}

	lakeIdx, err := rand.Int(e.reader, big.NewInt(int64(len(lakes))))
	if err != nil {
		return "", err
	}
	return lakes[lakeIdx.Int64()], nil
}
