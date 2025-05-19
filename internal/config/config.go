package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	filePath := homeDir + "/" + configFileName
	return filePath, nil

}

func write(cfg *Config) error {
	configFile, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("error getting config file path: %v", err)
	}

	fi, err := os.Lstat(configFile)
	if err != nil {
		return fmt.Errorf("error finding config file")
	}
	configJSON, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error mashaling config into json: %v", err)
	}
	err = os.WriteFile(configFile, configJSON, fi.Mode().Perm())
	if err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}
	return nil
}

func (c *Config) SetUser(userName string) error {
	c.CurrentUserName = userName
	return write(c)
}

func Read() (Config, error) {
	configFilePath, err := getConfigFilePath()
	var config Config
	if err != nil {
		return config, fmt.Errorf("error getting config file path: %v", err)
	}

	configRaw, err := os.Open(configFilePath)
	if err != nil {
		return config, fmt.Errorf("error reading config file: %v", err)
	}
	defer configRaw.Close()

	decoder := json.NewDecoder(configRaw)
	err = decoder.Decode(&config)
	if err != nil {
		return config, fmt.Errorf("error decoding json config file: %v", err)
	}
	return config, nil
}
