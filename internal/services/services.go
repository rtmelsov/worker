// Package services
package services

import (
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"

	"worker/internal/config"
	// "worker/internal/models" // Раскомментируйте, когда появятся модели
)

// Service содержит все зависимости для бизнес-логики
type Service struct {
	logger      *logrus.Entry
	kafkaWriter *kafka.Writer
	config      *config.Configuration
}

// NewService создает новый экземпляр сервиса
func NewService(logger *logrus.Entry, kafka *kafka.Writer, cfg *config.Configuration) *Service {
	return &Service{
		logger:      logger,
		kafkaWriter: kafka,
		config:      cfg,
	}
}
