package kafka

import (
	"fmt"
	"strings"

	"github.com/IBM/sarama"
	"github.com/nzb3/diploma/resource-service/internal/configurator"
)

// AppConfig holds the complete Kafka configuration from config file
type AppConfig struct {
	Brokers         []string        `yaml:"brokers" mapstructure:"brokers" validate:"required,min=1"`
	ConsumerGroupID string          `yaml:"consumer_group_id" mapstructure:"consumer_group_id" validate:"required"`
	Topics          TopicsConfig    `yaml:"topics" mapstructure:"topics"`
	Producer        ProducerConfig  `yaml:"producer" mapstructure:"producer"`
	Consumer        ConsumerOptions `yaml:"consumer" mapstructure:"consumer"`
}

// TopicsConfig holds Kafka topic names
type TopicsConfig struct {
	Resource string `yaml:"resource" mapstructure:"resource"`
}

// ProducerConfig holds Kafka producer settings
type ProducerConfig struct {
	RequiredAcks    int    `yaml:"required_acks" mapstructure:"required_acks"`
	RetryMax        int    `yaml:"retry_max" mapstructure:"retry_max"`
	CompressionType string `yaml:"compression_type" mapstructure:"compression_type"`
}

// ConsumerOptions holds Kafka consumer settings
type ConsumerOptions struct {
	AutoOffsetReset string `yaml:"auto_offset_reset" mapstructure:"auto_offset_reset"`
}

// NewConfig loads Kafka configuration from config file and environment variables
func NewConfig() (*Config, error) {
	// Parse configuration from "kafka" section
	appConfig, err := configurator.ParseConfig[AppConfig]("kafka")
	if err != nil {
		return nil, fmt.Errorf("failed to parse kafka config: %w", err)
	}

	// Handle special case for comma-separated brokers from environment
	brokers := appConfig.Brokers
	if brokersEnv := configurator.GetString("KAFKA_BROKERS"); brokersEnv != "" {
		brokerList := strings.Split(brokersEnv, ",")
		for i, broker := range brokerList {
			brokerList[i] = strings.TrimSpace(broker)
		}
		brokers = brokerList
	}

	// Provide default values if brokers are not set
	if len(brokers) == 0 {
		// Default to localhost:9092 for development
		brokers = []string{"localhost:9092"}
	}

	// Convert to producer Config struct
	config := &Config{
		Brokers:         brokers,
		RequiredAcks:    sarama.RequiredAcks(appConfig.Producer.RequiredAcks),
		RetryMax:        appConfig.Producer.RetryMax,
		CompressionType: getCompressionCodec(appConfig.Producer.CompressionType),
	}

	return config, nil
}

// NewConsumerConfig creates a ConsumerConfig from the application configuration
func NewConsumerConfig() (*ConsumerConfig, error) {
	// Parse configuration from "kafka" section
	appConfig, err := configurator.ParseConfig[AppConfig]("kafka")
	if err != nil {
		return nil, fmt.Errorf("failed to parse kafka config: %w", err)
	}

	// Handle special case for comma-separated brokers from environment
	brokers := appConfig.Brokers
	if brokersEnv := configurator.GetString("KAFKA_BROKERS"); brokersEnv != "" {
		brokerList := strings.Split(brokersEnv, ",")
		for i, broker := range brokerList {
			brokerList[i] = strings.TrimSpace(broker)
		}
		brokers = brokerList
	}

	// Provide default values if brokers or consumer group ID are not set
	if len(brokers) == 0 {
		// Default to localhost:9092 for development
		brokers = []string{"localhost:9092"}
	}

	groupID := appConfig.ConsumerGroupID
	if groupID == "" {
		// Default consumer group ID
		groupID = "resource-service-consumer"
	}

	// Provide default for auto offset reset if not set
	autoOffsetReset := appConfig.Consumer.AutoOffsetReset
	if autoOffsetReset == "" {
		// Default to earliest
		autoOffsetReset = "earliest"
	}

	// Convert to consumer Config struct
	config := &ConsumerConfig{
		Brokers:         brokers,
		GroupID:         groupID,
		AutoOffsetReset: autoOffsetReset,
	}

	return config, nil
}

// GetConsumerGroupID returns the consumer group ID from app config
func GetConsumerGroupID() (string, error) {
	appConfig, err := configurator.ParseConfig[AppConfig]("kafka")
	if err != nil {
		return "", err
	}

	// Return default if not set
	if appConfig.ConsumerGroupID == "" {
		return "resource-service-consumer", nil
	}

	return appConfig.ConsumerGroupID, nil
}

// GetBrokers returns the brokers list from app config
func GetBrokers() ([]string, error) {
	appConfig, err := configurator.ParseConfig[AppConfig]("kafka")
	if err != nil {
		return nil, err
	}

	// Handle special case for comma-separated brokers from environment
	if brokersEnv := configurator.GetString("KAFKA_BROKERS"); brokersEnv != "" {
		brokers := strings.Split(brokersEnv, ",")
		for i, broker := range brokers {
			brokers[i] = strings.TrimSpace(broker)
		}
		return brokers, nil
	}

	// Return default if not set
	if len(appConfig.Brokers) == 0 {
		return []string{"localhost:9092"}, nil
	}

	return appConfig.Brokers, nil
}

// GetTopicResource returns the resource topic name from app config
func GetTopicResource() (string, error) {
	appConfig, err := configurator.ParseConfig[AppConfig]("kafka")
	if err != nil {
		return "", err
	}

	// Return default if not set
	if appConfig.Topics.Resource == "" {
		return "resource", nil
	}

	return appConfig.Topics.Resource, nil
}

// getCompressionCodec converts string to sarama compression codec
func getCompressionCodec(compressionType string) sarama.CompressionCodec {
	switch strings.ToLower(compressionType) {
	case "snappy":
		return sarama.CompressionSnappy
	case "gzip":
		return sarama.CompressionGZIP
	case "lz4":
		return sarama.CompressionLZ4
	case "zstd":
		return sarama.CompressionZSTD
	default:
		return sarama.CompressionNone
	}
}
