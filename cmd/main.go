package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"worker/internal/config"
	"worker/internal/database"
	"worker/internal/handlers"
	"worker/internal/services"

	camundaClient "github.com/citilinkru/camunda-client-go/v3"
	"github.com/joho/godotenv"
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
		EndpointUrl: config.Config().CamundaClient.EndpointURL,
		ApiUser:     config.Config().CamundaClient.APIUser,
		ApiPassword: config.Config().CamundaClient.APIPassword,
		Timeout:     time.Second * time.Duration(config.Config().CamundaClient.Timeout),
	})

	// 3. Подключение к БД
	db, err := database.InitDB(entry, config.Config().DBConnection)
	if err != nil {
		logger.Fatalf("Failed to connect to the database: %v", err)
	}
	// Важно: отложенное закрытие соединения при падении или выходе
	// Но при Graceful Shutdown мы закроем его вручную ниже
	// defer db.Close()

	// 4. Инициализация слоев
	service := services.NewService(entry, db, config.Config())
	handler := handlers.NewHandler(client, entry, db, service, config.Config())

	// 5. Регистрация воркеров
	handler.AddWorker("", "liteProcess", handler.WrapHandler(handler.LiteProcessRouter, false, true))
	handler.AddWorker("", "collectInitialData", handler.WrapHandler(handler.CollectInitialData, false, true))
	handler.AddWorker("", "finishProcess", handler.WrapHandler(handler.FinishProcess, false, true))

	entry.Info("Application started")

	// 6. ПРАВИЛЬНОЕ ОЖИДАНИЕ ЗАВЕРШЕНИЯ (Graceful Shutdown)
	quit := make(chan os.Signal, 1)
	// Слушаем сигналы SIGINT (Ctrl+C) и SIGTERM (Docker stop / Kubernetes terminate)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Блокируем выполнение main, пока не придет сигнал
	<-quit
	entry.Info("Shutting down server...")

	// 7. Очистка ресурсов
	// Здесь можно закрыть соединения с БД, остановить воркеры и т.д.

	// Если у вас есть метод остановки воркеров, вызовите его здесь:
	// handler.Stop()

	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	sqlDB, err := db.Conn(ctx)
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			entry.Errorf("Error closing database connection: %v", err)
		} else {
			entry.Info("Database connection closed")
		}
	}

	entry.Info("Server exited")
}
