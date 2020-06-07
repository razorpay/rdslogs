package config

import (
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

//RedisDBConfig ...
var RedisDBConfig RedisConfig

//LoadConfig ...
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Error("Error loading .env file")
	}
	initilizeLogging()
}

//InitilizeRedisConfig ...
func InitilizeRedisConfig() {
	RedisDBConfig = RedisConfig{
		Host:     os.Getenv("TRACKER_HOST"),
		Database: os.Getenv("TRACKER_DATABASE"),
		Username: os.Getenv("TRACKER_USERNAME"),
		Password: os.Getenv("TRACKER_PASSWORD"),
		Port:     os.Getenv("TRACKER_PORT"),
	}
}

func initilizeLogging() {
	log.SetFormatter(&log.JSONFormatter{})
	loglevel := os.Getenv("LOG_LEVEL")
	if loglevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else if loglevel == "info" {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	log.Debug("Logging in debug mode.")
}
