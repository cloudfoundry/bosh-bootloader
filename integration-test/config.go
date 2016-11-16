package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Config struct {
	AWSAccessKeyID           string
	AWSSecretAccessKey       string
	AWSRegion                string
	GCPServiceAccountKeyPath string
	GCPProjectID             string
	GCPRegion                string
	GCPZone                  string
	StateFileDir             string
}

var tempDir func(string, string) (string, error) = ioutil.TempDir

func LoadAWSConfig() (Config, error) {
	config, err := loadConfigJson()
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

	if config.StateFileDir == "" {
		dir, err := tempDir("", "")
		if err != nil {
			return Config{}, err
		}
		config.StateFileDir = dir
	}

	return config, nil
}

func LoadGCPConfig() (Config, error) {
	config, err := loadConfigJson()
	if err != nil {
		return Config{}, err
	}

	if config.GCPServiceAccountKeyPath == "" {
		return Config{}, errors.New("gcp service account key path is missing")
	}

	if config.GCPProjectID == "" {
		return Config{}, errors.New("project id is missing")
	}

	if config.GCPRegion == "" {
		return Config{}, errors.New("gcp region is missing")
	}

	if config.GCPZone == "" {
		return Config{}, errors.New("gcp zone is missing")
	}

	if config.StateFileDir == "" {
		dir, err := tempDir("", "")
		if err != nil {
			return Config{}, err
		}
		config.StateFileDir = dir
	}

	return config, nil
}

func loadConfigJson() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}

	configFile, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func configPath() (string, error) {
	path := os.Getenv("BIT_CONFIG")
	if path == "" || !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("$BIT_CONFIG %q does not specify an absolute path to test config file", path)
	}

	return path, nil
}
