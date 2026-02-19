// Package services
package services

import (
	"database/sql"
	"github.com/sirupsen/logrus"

	"worker/internal/config"
	// "worker/internal/models" // Раскомментируйте, когда появятся модели
)

// Service содержит все зависимости для бизнес-логики
type Service struct {
	logger *logrus.Entry
	db     *sql.DB
	config *config.Configuration
}

// NewService создает новый экземпляр сервиса
func NewService(logger *logrus.Entry, db *sql.DB, cfg *config.Configuration) *Service {
	return &Service{
		logger: logger,
		db:     db,
		config: cfg,
	}
}

// UpdateProcessStatus Пример бизнес-метода (заглушка)
// Сюда вы будете добавлять логику, которую вызывает Handler
func (s *Service) UpdateProcessStatus(processInstanceID string, status string) error {
	s.logger.Infof("Updating process %s to status %s", processInstanceID, status)

	// Пример сохранения в БД (если есть модель Process)
	/*
		err := s.db.Model(&models.Process{}).
			Where("instance_id = ?", processInstanceID).
			Update("status", status).Error

		if err != nil {
			s.logger.Errorf("Failed to update status in DB: %v", err)
			return err
		}
	*/

	return nil
}
