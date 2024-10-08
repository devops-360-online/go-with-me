package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort    string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	RedisAddress  string
	KafkaBrokers  []string
	MongoURI      string
	MongoDatabase string
	// Add other configurations as needed
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	config := &Config{
		ServerPort:    viper.GetString("SERVER_PORT"),
		DBHost:        viper.GetString("DB_HOST"),
		DBPort:        viper.GetString("DB_PORT"),
		DBUser:        viper.GetString("DB_USER"),
		DBPassword:    viper.GetString("DB_PASSWORD"),
		DBName:        viper.GetString("DB_NAME"),
		RedisAddress:  viper.GetString("REDIS_ADDRESS"),
		KafkaBrokers:  viper.GetStringSlice("KAFKA_BROKERS"),
		MongoURI:      viper.GetString("MONGO_URI"),
		MongoDatabase: viper.GetString("MONGO_DATABASE"),
	}

	return config
}
