package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	App  appConfig
	Bili biliConfig
}

type appConfig struct {
	Port string
}

type biliConfig struct {
	Uid    string
	Url    string
	Avatar string
}

var Config = &config{}

func InitConfig() error {
	configBytes, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Read config failed, %v\n", err)
		return err
	}

	err = yaml.Unmarshal(configBytes, &Config)
	if err != nil {
		log.Fatalf("Decode config failed: %v\n", err)
	}

	return nil
}
