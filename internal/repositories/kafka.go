package repositories

import (
    "github.com/IBM/sarama"
	"github.com/devops-360-online/go-with-me/config"
)

func NewKafkaSyncProducer(cfg *config.Config) (sarama.SyncProducer, error) {
    config := sarama.NewConfig()
    config.Producer.Return.Successes = true
    producer, err := sarama.NewSyncProducer(cfg.KafkaBrokers, config)
    if err != nil {
        return nil, err
    }
    return producer, nil
}
