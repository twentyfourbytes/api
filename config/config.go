package config

import (
	"encoding/json"
	"os"

	"github.com/google/logger"
)

//Configuration Holds config struct
type Configuration struct {
	Redis struct {
		Port int    `json:"port"`
		IP   string `json:"ip"`
	} `json:"redis"`
	Server string `json:"server"`
}

//Config Parse config.json and return config struct
func Config() (configuration *Configuration) {
	file, err := os.Open("config.json")
	defer file.Close()
	if err != nil {
		logger.Fatal("Error While opening config file", err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		logger.Error(err)
	}
	return configuration
}
