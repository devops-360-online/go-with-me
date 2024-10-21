package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort               string
	DBHost                   string
	DBPort                   string
	DBUser                   string
	DBPassword               string
	DBName                   string
	RedisAddress             string
	KafkaBrokers             []string
	MongoURI                 string
	MongoDatabase            string
	AwsAccessKeyID           string
	AwsSecretAccessKey       string
	S3BucketNameEvents       string
	S3BucketNameChatEvent    string
	S3Region                 string
	OtelExporterOTLPEndpoint string
	S3Endpoint              string
	// Add other configurations as needed
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	config := &Config{
		ServerPort:               viper.GetString("SERVER_PORT"),
		DBHost:                   viper.GetString("DB_HOST"),
		DBPort:                   viper.GetString("DB_PORT"),
		DBUser:                   viper.GetString("DB_USER"),
		DBPassword:               viper.GetString("DB_PASSWORD"),
		DBName:                   viper.GetString("DB_NAME"),
		RedisAddress:             viper.GetString("REDIS_ADDRESS"),
		KafkaBrokers:             viper.GetStringSlice("KAFKA_BROKERS"),
		MongoURI:                 viper.GetString("MONGO_URI"),
		MongoDatabase:            viper.GetString("MONGO_DATABASE"),
		AwsAccessKeyID:           viper.GetString("AWS_ACCESS_KEY_ID"),
		AwsSecretAccessKey:       viper.GetString("AWS_SECRET_ACCESS_KEY"),
		S3BucketNameEvents:       viper.GetString("S3_BUCKET_EVENTS"),
		S3BucketNameChatEvent:    viper.GetString("S3_BUCKET_CHAT"),
		S3Region:                 viper.GetString("DEFAULT_S3_REGION"),
		S3Endpoint:              viper.GetString("S3_ENDPOINT"),
		OtelExporterOTLPEndpoint: viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT"),
		        // Add other configurations as needed
	}

	return config
}
