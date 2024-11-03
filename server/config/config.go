package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Listen address is an array of IP addresses and port combinations.
	// Listen address is an array so that this service can listen to many interfaces at once.
	// You can use this value for example: []string{"192.168.1.12:80", "25.49.25.73:80"} to listen to
	// listen to interfaces with IP address of 192.168.1.12 and 25.49.25.73, both on port 80.
	ListenAddress string `config:"LISTEN_ADDRESS"`

	CorsAllowedHeaders []string `config:"CORS_ALLOWED_HEADERS"`
	CorsAllowedMethods []string `config:"CORS_ALLOWED_METHODS"`
	CorsAllowedOrigins []string `config:"CORS_ALLOWED_ORIGINS"`

	AppName string `config:"APP_NAME"`
	AppKey  string `config:"APP_KEY"`
}

func InitConfig(path string) *Config {
	// Todo: add env checker

	if path == "" {
		godotenv.Load(".env")
	} else {
		godotenv.Load(path)
	}

	appName := getEnv("APP_NAME", "")

	return &Config{
		ListenAddress: fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")),
		CorsAllowedHeaders: []string{
			"Connection", "User-Agent", "Referer",
			"Accept", "Accept-Language", "Content-Type",
			"Content-Language", "Content-Disposition", "Origin",
			"Content-Length", "Authorization", "ResponseType",
			"X-Requested-With", "X-Forwarded-For",
		},
		CorsAllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "PUT"},
		CorsAllowedOrigins: []string{"*"},
		AppName:            appName,
		AppKey:             getEnv("APP_KEY", ""),
	}
}

func (c *Config) AsString() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
