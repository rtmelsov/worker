package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"worker/internal/config"
	"worker/internal/handlers"
	"worker/internal/services"

	camundaClient "github.com/citilinkru/camunda-client-go/v3"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

func main() {
	// 1. Инициализация логгера
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	entry := logger.WithFields(logrus.Fields{})

	if err := godotenv.Load(); err != nil {
		logger.Printf(".env file not found, skipping local development")
	}

	config.Initialize(entry, "bpm-workers")

	// 2. Инициализация клиента Camunda
	client := camundaClient.NewClient(camundaClient.ClientOptions{
		EndpointUrl:         config.Config().CamundaClient.EndpointURL,
		Timeout:             time.Second * time.Duration(config.Config().CamundaClient.Timeout),
		AuthorizationHeader: config.Config().CamundaClient.CamundaAuthBasic,
	})

	kafkaWriter := &kafka.Writer{
		Addr: kafka.TCP(config.Config().KafkaBrokers...),
		// Balancer определяет, как распределяются сообщения по партициям.
		// LeastBytes - хороший выбор по умолчанию.
		Balancer: &kafka.LeastBytes{},
		// Topic:   "ucp-tracking-group",
	}

	// Обязательно закрываем соединение при завершении работы приложения
	defer func() {
		if err := kafkaWriter.Close(); err != nil {
			logger.Errorf("Ошибка при закрытии Kafka Writer: %v", err)
		}
	}()

	// 4. Инициализация слоев
	service := services.NewService(entry, kafkaWriter, config.Config())
	handler := handlers.NewHandler(client, entry, service, config.Config())

	// 5. Регистрация воркеров
	handler.AddWorker("", "firstParser", handler.WrapHandler(handler.LiteProcessRouter, false, true))

	entry.Info("Application started")

	// 6. ПРАВИЛЬНОЕ ОЖИДАНИЕ ЗАВЕРШЕНИЯ (Graceful Shutdown)
	quit := make(chan os.Signal, 1)
	// Слушаем сигналы SIGINT (Ctrl+C) и SIGTERM (Docker stop / Kubernetes terminate)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Блокируем выполнение main, пока не придет сигнал
	<-quit
	entry.Info("Shutting down server...")

	entry.Info("Server exited")
}
