package services

import (
	"context"
	"encoding/json"

	"github.com/citilinkru/camunda-client-go/v3/processor"
	"github.com/segmentio/kafka-go" // Добавляем импорт библиотеки Kafka
	"worker/internal/models"
)

// Подсказка: убедитесь, что ваш Service содержит настроенный writer
// type Service struct {
// 	KafkaWriter *kafka.Writer
// }

func (s *Service) SendToKafka(ctx *processor.Context, req *models.MetricsPayload) ([]byte, error) {

	jsBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err // Возвращаем ошибку маршалинга
	}

	// Формируем сообщение для отправки
	msg := kafka.Message{
		// Топик из вашей BPMN схемы
		Topic: "ucp-tracking-group",
		Value: jsBytes,
	}

	// Вызов Kafka: отправляем сообщение
	// Используем context.Background() или контекст из процессора, если он поддерживает отмену
	err = s.kafkaWriter.WriteMessages(context.Background(), msg)
	if err != nil {
		// Если Kafka недоступна, возвращаем ошибку.
		// Camunda-клиент перехватит её и выполнит retry (R3/PT10S из вашей схемы)
		return nil, err
	}

	return jsBytes, nil
}

