package config_service

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	Environment string    `json:"ENVIRONMENT,omitempty"`
	Redis       Redis     `json:"REDIS,omitempty"`
	Messenger   Messenger `json:"MESSENGER,omitempty"`
	HostUrl     string    `json:"HOST_URL,omitempty"`
}

type Redis struct {
	HostAddress string `json:"HOST_ADDRESS,omitempty"`
	Port        string `json:"PORT,omitempty"`
	Password    string `json:"PASSWORD,omitempty"`
}

type Messenger struct {
	WebhookVerificationToken string `json:"WEBHOOK_VERIFICATION_TOKEN,omitempty"`
	AccessToken              string `json:"ACCESS_TOKEN,omitempty"`
}

func New() *Config {
	// If on local, grab from local.json
	// If on production, grab from Heroku env variables
	if os.Getenv("ENVIRONMENT") == "PROD" {
		return &Config{
			Environment: os.Getenv("ENVIRONMENT"),
			HostUrl:     os.Getenv("HOST_NAME"),
			Redis: Redis{
				HostAddress: os.Getenv("HOST_ADDRESS"),
				Port:        os.Getenv("PORT"),
				Password:    os.Getenv("PASSWORD"),
			},
			Messenger: Messenger{
				WebhookVerificationToken: os.Getenv("WEBHOOK_VERIFICATION_TOKEN"),
				AccessToken:              os.Getenv("ACCESS_TOKEN"),
			},
		}
	}

	var config Config
	jsonFile, err := os.Open("secrets/local.json")
	if err != nil {
		panic(fmt.Sprintf("failed to start server when loading config: %+v", err))
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		panic(fmt.Sprintf("failed to start server when unmashaling config: %+v", err))
	}

	return &config
}

func (c *Config) GetRedisConfig() Redis {
	return c.Redis
}

func (c *Config) GetMessengerConfig() Messenger {
	return c.Messenger
}

func (c *Config) GetHostUrl() string {
	return c.HostUrl
}
