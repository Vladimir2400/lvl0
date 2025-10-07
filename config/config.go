package config

import (
	"os"
	"strconv"
)

type Config struct {
	Database DatabaseConfig
	Kafka    KafkaConfig
	Server   ServerConfig
	Cache    CacheConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type KafkaConfig struct {
	Brokers   []string
	Topic     string
	GroupID   string
	MinBytes  int
	MaxBytes  int
}

type ServerConfig struct {
	Port string
	Host string
}

type CacheConfig struct {
	MaxSize int
	TTL     int // в секундах
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "127.0.0.1"),
			Port:     getEnv("DB_PORT", "5434"),
			User:     getEnv("DB_USER", "wb_user"),
			Password: getEnv("DB_PASSWORD", "wb_password"),
			DBName:   getEnv("DB_NAME", "wb_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Kafka: KafkaConfig{
			Brokers:  []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			Topic:    getEnv("KAFKA_TOPIC", "orders"),
			GroupID:  getEnv("KAFKA_GROUP_ID", "order-group"),
			MinBytes: getEnvAsInt("KAFKA_MIN_BYTES", 10000),    // 10KB
			MaxBytes: getEnvAsInt("KAFKA_MAX_BYTES", 10000000), // 10MB
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", ""),
		},
		Cache: CacheConfig{
			MaxSize: getEnvAsInt("CACHE_MAX_SIZE", 1000),
			TTL:     getEnvAsInt("CACHE_TTL", 3600), // 1 час
		},
	}
}

func (c *Config) DatabaseDSN() string {
	return "host=" + c.Database.Host +
		" user=" + c.Database.User +
		" password=" + c.Database.Password +
		" dbname=" + c.Database.DBName +
		" port=" + c.Database.Port +
		" sslmode=" + c.Database.SSLMode
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}