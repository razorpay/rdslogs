package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/razorpay/rdslogs/constants"
	log "github.com/sirupsen/logrus"
)

//RedisDBConfig ...
var RedisDBConfig RedisConfig

// LoadConfigFromFile ...
func LoadConfigFromFile() {
	err := godotenv.Load()
	if err != nil {
		log.Error("Error loading .env file")
	}
}

//InitilizeRedisConfig ...
func InitilizeRedisConfig() {
	database, _ := strconv.Atoi(getenv(constants.TrackerDatabase, "0"))
	maxIdle, _ := strconv.Atoi(getenv(constants.TrackerMaxIdle, "10"))
	maxActive, _ := strconv.Atoi(getenv(constants.TrackerMaxActive, "100"))
	RedisDBConfig = RedisConfig{
		Host:      os.Getenv(constants.TrackerHost),
		Database:  database,
		Password:  os.Getenv(constants.TrackerPassword),
		Port:      os.Getenv(constants.TrackerPort),
		MaxIdle:   maxIdle,
		MaxActive: maxActive,
	}
}

//InitilizeLogging ...
func InitilizeLogging() {
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
