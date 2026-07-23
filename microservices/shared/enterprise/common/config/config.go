package config

import (
	"os"
	"strconv"
	"time"
)

type AppConfig struct {
	ServiceName      string
	Environment       string
	Port              string
	GRPCPort         string
	LogLevel          string
	ShutdownTimeout   time.Duration
	Postgres          PostgresConfig
	Kafka             KafkaConfig
	Redis             RedisConfig
	OpenSearch        OpenSearchConfig
	MinIO             MinIOConfig
	JWT               JWTConfig
	OpenTelemetry     OTelConfig
}

type PostgresConfig struct {
	Host              string
	Port              int
	User              string
	Password          string
	Database          string
	SSLMode           string
	MaxOpenConns      int
	MaxIdleConns      int
	ConnMaxLifetime    time.Duration
	MigrationPath      string
}

type KafkaConfig struct {
	Brokers           []string
	ClientID          string
	GroupID           string
	AutoOffsetReset   string
	EnableIDempotence bool
	MaxRetries        int
	RetryBackoff      time.Duration
	Topics            KafkaTopics
}

type KafkaTopics struct {
	OrderCreated       string
	OrderCancelled    string
	PaymentSuccess     string
	PaymentFailed      string
	InventoryUpdated   string
	ShipmentCreated    string
	NotificationRequested string
}

type RedisConfig struct {
	Address           string
	Password          string
	DB                int
	DefaultTTL        time.Duration
}

type OpenSearchConfig struct {
	Addresses        []string
	Username         string
	Password         string
	IndexPrefix      string
}

type MinIOConfig struct {
	Endpoint         string
	AccessKey       string
	SecretKey       string
	UseSSL          bool
	BucketName      string
}

type JWTConfig struct {
	Secret           string
	Issuer           string
	Audience         string
	Expiry           time.Duration
	RefreshExpiry    time.Duration
}

type OTelConfig struct {
	ServiceName      string
	CollectorEndpoint string
	TraceRatio       float64
	EnableMetrics    bool
	EnableLogs      bool
}

func LoadConfig() *AppConfig {
	return &AppConfig{
		ServiceName:      getEnv("SERVICE_NAME", "enterprise-service"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		Port:              getEnv("PORT", "8080"),
		GRPCPort:          getEnv("GRPC_PORT", "50051"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		ShutdownTimeout:   30 * time.Second,
		Postgres: PostgresConfig{
			Host:              getEnv("POSTGRES_HOST", "localhost"),
			Port:              getEnvInt("POSTGRES_PORT", 5432),
			User:              getEnv("POSTGRES_USER", "postgres"),
			Password:          getEnv("POSTGRES_PASSWORD", "postgres"),
			Database:          getEnv("POSTGRES_DB", "enterprise_ecommerce"),
			SSLMode:           getEnv("POSTGRES_SSLMODE", "disable"),
			MaxOpenConns:     25,
			MaxIdleConns:     5,
			ConnMaxLifetime:   5 * time.Minute,
			MigrationPath:     "file://migrations",
		},
		Kafka: KafkaConfig{
			Brokers:           getEnvSlice("KAFKA_BROKERS", "localhost:9092"),
			ClientID:          getEnv("KAFKA_CLIENT_ID", "enterprise"),
			GroupID:           getEnv("KAFKA_GROUP_ID", "enterprise-group"),
			AutoOffsetReset:   "earliest",
			EnableIDempotence: true,
			MaxRetries:        3,
			RetryBackoff:      1 * time.Second,
			Topics: KafkaTopics{
				OrderCreated:         "order-created",
				OrderCancelled:      "order-cancelled",
				PaymentSuccess:       "payment-success",
				PaymentFailed:        "payment-failed",
				InventoryUpdated:     "inventory-updated",
				ShipmentCreated:     "shipment-created",
				NotificationRequested: "notification-requested",
			},
		},
		Redis: RedisConfig{
			Address:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password:    getEnv("REDIS_PASSWORD", ""),
			DB:          0,
			DefaultTTL:  15 * time.Minute,
		},
		OpenSearch: OpenSearchConfig{
			Addresses:   getEnvSlice("OPENSEARCH_ADDR", "http://localhost:9200"),
			Username:    getEnv("OPENSEARCH_USERNAME", "admin"),
			Password:    getEnv("OPENSEARCH_PASSWORD", "admin"),
			IndexPrefix: "ecommerce",
		},
		MinIO: MinIOConfig{
			Endpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:     false,
			BucketName: "ecommerce",
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "super-secret-key-change-in-production"),
			Issuer:        "enterprise-ecommerce",
			Audience:      "enterprise-services",
			Expiry:        15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
		OpenTelemetry: OTelConfig{
			ServiceName:      getEnv("OTEL_SERVICE_NAME", "enterprise"),
			CollectorEndpoint: getEnv("OTEL_COLLECTOR_ENDPOINT", "localhost:4317"),
			TraceRatio:       1.0,
			EnableMetrics:    true,
			EnableLogs:      true,
		},
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvSlice(key string, defaultVal string) []string {
	v := os.Getenv(key)
	if v == "" {
		return []string{defaultVal}
	}
	return []string{v}
}