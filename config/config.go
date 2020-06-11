package config

import (
	"os"
	"strconv"

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
	database, _ := strconv.Atoi(getenv("TRACKER_DATABASE", "0"))
	maxIdle, _ := strconv.Atoi(getenv("TRACKER_MAXIDLE", "10"))
	maxActive, _ := strconv.Atoi(getenv("TRACKER_MAXACTIVE", "100"))
	RedisDBConfig = RedisConfig{
		Host:      os.Getenv("TRACKER_HOST"),
		Database:  database,
		Password:  os.Getenv("TRACKER_PASSWORD"),
		Port:      os.Getenv("TRACKER_PORT"),
		MaxIdle:   maxIdle,
		MaxActive: maxActive,
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

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
