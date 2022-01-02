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
		fmt.Println("Loading config from *environment variables*...")
		return getConfigFromEnvVariables()
	}

	fmt.Println("Loading config from *JSON*...")
	return getConfigFromJSON("secrets/local.json")
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

func getConfigFromEnvVariables() *Config {
	return &Config{
		Environment: getEnvOrPanic("ENVIRONMENT"),
		HostUrl:     getEnvOrPanic("HOST_NAME"),
		Redis: Redis{
			HostAddress: getEnvOrPanic("HOST_ADDRESS"),
			Port:        getEnvOrPanic("PORT"),
			Password:    getEnvOrPanic("PASSWORD"),
		},
		Messenger: Messenger{
			WebhookVerificationToken: getEnvOrPanic("WEBHOOK_VERIFICATION_TOKEN"),
			AccessToken:              getEnvOrPanic("ACCESS_TOKEN"),
		},
	}
}

// GetEnvOrPanic checks to see if the key exists or else it panics
func getEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("missing Heroku config key: %v", key))
	}

	return value
}

func getConfigFromJSON(filePath string) *Config {
	var config Config
	jsonFile, err := os.Open(filePath)
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
