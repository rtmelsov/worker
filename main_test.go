// Package main_test Использовать для локального
// теста воркеров
package main

import (
	"testing"
	"time"

	"log"

	"worker/internal/config"
	"worker/internal/handlers"
	"worker/internal/helpers"
	"worker/internal/services"

	camundaClient "github.com/citilinkru/camunda-client-go/v3"
	"github.com/citilinkru/camunda-client-go/v3/processor"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type localEnv struct {
	handler *handlers.Handler
}

// Функция для подключения к базе данных
// Получения переменных окружения
// инициализации camunda clientOptions
func bootstrapLocalEnv(t *testing.T) *localEnv {
	t.Helper()
	if err := godotenv.Load(); err != nil {
		t.Logf(".env file not found, falling back to environment variables: %v", err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	entry := logger.WithFields(logrus.Fields{})

	config.Initialize(entry, "bpm-workers")

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

	service := services.NewService(entry, kafkaWriter, config.Config())
	handler := handlers.NewHandler(client, entry, service, config.Config())

	return &localEnv{handler: handler}
}

func newBaseContext() *processor.Context {
	return &processor.Context{
		Task: &camundaClient.ResLockedExternalTask{
			ProcessDefinitionKey: "",
			BusinessKey:          "LOAN-REQUEST-2026-006",
			TopicName:            "test",
			WorkerId:             "local-test-worker",
			Variables:            helpers.LoanCamundaVariables,
		},
	}
}

func TestLocalBootstrap(t *testing.T) {
	env := bootstrapLocalEnv(t)
	ctx := newBaseContext()

	if env.handler == nil {
		t.Fatal("handler is not initialised")
	}

	// AmlCheckClient
	_, err := env.handler.LiteProcessRouter(ctx)
	if err != nil {
		log.Print("ColvirGetGraph: ", "Error text\n", err.Error())
	}

	t.Logf("запрос для топика %s и бизнес ключ %s", ctx.Task.TopicName, ctx.Task.BusinessKey)
}
