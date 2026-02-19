// Package config
package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// Глобальная переменная для хранения конфигурации
var cfg *Configuration

// Configuration — основная структура конфига
type Configuration struct {
	ServiceName   string
	KafkaBrokers  []string
	CamundaClient CamundaConfig
}

// CamundaConfig — настройки для Camunda
type CamundaConfig struct {
	EndpointURL      string
	CamundaAuthBasic string
	Timeout          int // Тайм-аут в секундах
}

// Initialize — читает переменные окружения и заполняет структуру
func Initialize(logger *logrus.Entry, serviceName string) {
	logger.Info("Initializing configuration...")

	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		// Если переменная пустая, приложение не сможет работать, поэтому падаем
		logger.Error("Критическая ошибка: KAFKA_BROKERS не задан в окружении")
	}

	// Разбиваем строку по запятой на слайс строк
	kafkaBrokers := strings.Split(brokersEnv, ",")

	cfg = &Configuration{
		ServiceName:  serviceName,
		KafkaBrokers: kafkaBrokers,
		CamundaClient: CamundaConfig{
			EndpointURL:      getEnv("CAMUNDA_URL", "http://localhost:8080/engine-rest"),
			CamundaAuthBasic: getEnv("CAMUNDA_AUTH_BASIC", ""), // Оставь пустым, если нет Auth
			Timeout:          getEnvAsInt("CAMUNDA_TIMEOUT", 30),
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
