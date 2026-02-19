// Package config
package config

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

// Глобальная переменная для хранения конфигурации
var cfg *Configuration

// Configuration — основная структура конфига
type Configuration struct {
	ServiceName   string
	CamundaClient CamundaConfig
	DBConnection  DatabaseConfig
}

// CamundaConfig — настройки для Camunda
type CamundaConfig struct {
	EndpointURL string
	APIUser     string
	APIPassword string
	Timeout     int // Тайм-аут в секундах
}

// DatabaseConfig — настройки для БД
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Initialize — читает переменные окружения и заполняет структуру
func Initialize(logger *logrus.Entry, serviceName string) {
	logger.Info("Initializing configuration...")

	cfg = &Configuration{
		ServiceName: serviceName,
		CamundaClient: CamundaConfig{
			EndpointURL: getEnv("CAMUNDA_URL", "http://localhost:8080/engine-rest"),
			APIUser:     getEnv("CAMUNDA_USER", ""),     // Оставь пустым, если нет Auth
			APIPassword: getEnv("CAMUNDA_PASSWORD", ""), // Оставь пустым, если нет Auth
			Timeout:     getEnvAsInt("CAMUNDA_TIMEOUT", 30),
		},
		DBConnection: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "camunda"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}

	logger.Infof("Configuration loaded for service: %s", serviceName)
}

// Config — геттер для доступа к конфигурации из других пакетов
func Config() *Configuration {
	if cfg == nil {
		// Защита от дурака, если забыли вызвать Initialize
		panic("Configuration not initialized! Call config.Initialize() first.")
	}
	return cfg
}

// --- Вспомогательные функции ---

// getEnv возвращает значение переменной окружения или дефолтное значение
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvAsInt возвращает int из переменной окружения или дефолтное значение
func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

