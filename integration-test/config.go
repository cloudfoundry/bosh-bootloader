package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
}

func ConfigPath() (string, error) {
	path := os.Getenv("BIT_CONFIG")
	if path == "" || !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("$BIT_CONFIG %q does not specify an absolute path to test config file", path)
	}

	return path, nil
}

func LoadConfig(configFilePath string) (Config, error) {
	config, err := loadConfigJsonFromPath(configFilePath)
	if err != nil {
		return Config{}, err
	}

	if config.AWSAccessKeyID == "" {
		return Config{}, errors.New("aws access key id is missing")
	}

	if config.AWSSecretAccessKey == "" {
		return Config{}, errors.New("aws secret access key is missing")
	}

	if config.AWSRegion == "" {
		return Config{}, errors.New("aws region is missing")
	}

	return config, nil
}

func loadConfigJsonFromPath(configFilePath string) (Config, error) {
	configFile, err := os.Open(configFilePath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}
