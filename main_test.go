// Package main_test Использовать для локального
// теста воркеров
package main

import (
	"testing"
	"time"

	"log"
	"worker/internal/config"
	"worker/internal/database"
	"worker/internal/handlers"
	"worker/internal/helpers"
	"worker/internal/services"

	camundaClient "github.com/citilinkru/camunda-client-go/v3"
	"github.com/citilinkru/camunda-client-go/v3/processor"
	"github.com/joho/godotenv"
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
		EndpointUrl: config.Config().CamundaClient.EndpointURL,
		ApiUser:     config.Config().CamundaClient.APIUser,
		ApiPassword: config.Config().CamundaClient.APIPassword,
		Timeout:     time.Second * time.Duration(config.Config().CamundaClient.Timeout),
	})

	db, err := database.InitDB(entry, config.Config().DBConnection)
	if err != nil {
		t.Fatalf("failed to connect to the database: %v", err)
	}

	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("failed to close database connection: %v", closeErr)
		}
	})

	service := services.NewService(entry, db, config.Config())
	handler := handlers.NewHandler(client, entry, db, service, config.Config())

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
	_, err := env.handler.CollectInitialData(ctx)
	if err != nil {
		log.Print("ColvirGetGraph: ", "Error text\n", err.Error())
	}

	t.Logf("запрос для топика %s и бизнес ключ %s", ctx.Task.TopicName, ctx.Task.BusinessKey)
}
