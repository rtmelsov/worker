// Package handlers
package handlers

import (
	"fmt"
	"time"

	camundaClient "github.com/citilinkru/camunda-client-go/v3"
	"github.com/citilinkru/camunda-client-go/v3/processor"
	"github.com/sirupsen/logrus"

	"database/sql"

	"worker/internal/config"
	"worker/internal/services"
)

type Handler struct {
	processor *processor.Processor
	logger    *logrus.Entry
	db        *sql.DB
	services  *services.Service
	config    *config.Configuration
}

func NewHandler(
	client *camundaClient.Client,
	logger *logrus.Entry,
	service *services.Service,
	cfg *config.Configuration,
) *Handler {
	// Создаем процессор (менеджер воркеров)
	// Опции берем из конфига или ставим дефолтные
	proc := processor.NewProcessor(client, &processor.Options{
		WorkerId:                  "go-worker-" + cfg.ServiceName,
		LockDuration:              time.Second * 10, // Время лока задачи
		MaxTasks:                  1,                // Сколько задач за раз брать
		MaxParallelTaskPerHandler: 1,                // Параллелизм внутри одного хендлера
		LongPollingTimeout:        time.Second * 10,
	}, func(err error) {
		logger.Error("Processor error:", err)
	})

	return &Handler{
		config:    cfg,
		processor: proc,
		logger:    logger,
		services:  service,
	}
}

// AddWorker регистрирует обработчик для топика
func (h *Handler) AddWorker(processKey string, topicName string, handler processor.Handler) {
	var camundaHandlerRequest = []*camundaClient.QueryFetchAndLockTopic{{TopicName: topicName}}
	if processKey != "" {
		camundaHandlerRequest = []*camundaClient.QueryFetchAndLockTopic{{TopicName: topicName, ProcessDefinitionKey: &processKey}}
	}
	h.processor.AddHandler(camundaHandlerRequest, handler)
	if processKey != "" {
		h.logger.Printf("Registered worker for process: %s, topic: %s", processKey, topicName)
	} else {
		h.logger.Printf("Registered worker topic: %s", topicName)
	}
}

// WrapHandler - Middleware для логирования, обработки паники и завершения задач
func (h *Handler) WrapHandler(handlerFunc func(ctx *processor.Context) (map[string]camundaClient.Variable, error), autoRetry, updateStatus bool) processor.Handler {
	return func(ctx *processor.Context) error {
		h.logger.Printf("Handling task for proccess: %s, topic: %s", ctx.Task.BusinessKey, ctx.Task.TopicName)

		var err error
		var processVariables map[string]camundaClient.Variable
		defer func() {
			if r := recover(); r != nil {
				// Приводим панику к типу error
				switch v := r.(type) {
				case error:
					err = v
				default:
					err = fmt.Errorf("panic: %v", v)
				}

				h.logger.Errorf("Recovered from panic for proccess: %s, topic: %s, error: %s", ctx.Task.BusinessKey, ctx.Task.TopicName, err)

				// Сообщаем брокеру об ошибке (при панике обычно autoRetry = false)
				if handleFailureErr := h.handleFailure(ctx, err.Error(), false); handleFailureErr != nil {
					h.logger.Errorf("Failed to handleFailure (panic) for proccess: %s, topic: %s, error: %s", ctx.Task.BusinessKey, ctx.Task.TopicName, handleFailureErr)
				}

				return // Прерываем выполнение defer, чтобы не пойти в ветку if err != nil
			}
			if err != nil {
				h.logger.Errorf("Catched error for proccess: %s, topic: %s, error: %s", ctx.Task.BusinessKey, ctx.Task.TopicName, err)
				if handleFailureErr := h.handleFailure(ctx, err.Error(), autoRetry); handleFailureErr != nil {
					h.logger.Errorf("Failed to handleFailure task for proccess: %s, topic: %s, error: %s", ctx.Task.BusinessKey, ctx.Task.TopicName, handleFailureErr)
				}

			} else {
				if completeErr := h.completeTask(ctx, processVariables); completeErr != nil {
					h.logger.Errorf("Failed to complete task for proccess: %s, topic: %s, error: %s", ctx.Task.BusinessKey, ctx.Task.TopicName, completeErr)
				}
				h.logger.Infof("Successfully completed task for proccess: %s, topic: %s", ctx.Task.BusinessKey, ctx.Task.TopicName)
			}
		}()

		processVariables, err = handlerFunc(ctx)
		return err
	}
}

// completeTask отправляет в Camunda сигнал о завершении
func (h *Handler) completeTask(ctx *processor.Context, processVariables map[string]camundaClient.Variable) error {
	// Важно: библиотека processor имеет метод ctx.Complete, который сам отправляет запрос
	err := ctx.Complete(processor.QueryComplete{Variables: &processVariables})
	if err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}
	return nil
}

// handleFailure отправляет в Camunda инцидент или ошибку
func (h *Handler) handleFailure(ctx *processor.Context, errorText string, autoRetry bool) error {
	request := processor.QueryHandleFailure{
		ErrorMessage: &errorText,
	}

	if autoRetry {
		retries := 3    // Можно вынести в конфиг
		timeout := 5000 // мс
		request.Retries = &retries
		request.RetryTimeout = &timeout
	} else {
		retries := 0 // 0 = Incident (красный значок в Cockpit)
		request.Retries = &retries
	}

	return ctx.HandleFailure(request)
}

// CamundaCheckLite - пример реализации бизнес-логики для конкретного воркера
// Эту функцию мы передаем в AddWorker
func (h *Handler) CamundaCheckLite(ctx *processor.Context) (map[string]camundaClient.Variable, error) {
	// Логика
	h.logger.Info("Executing business logic for Status Update...")

	// Пример вызова сервиса
	// err := h.service.DoSomething(ctx.Task.ProcessInstanceId)
	// if err != nil { return nil, err }

	// Возвращаем переменные, если нужно обновить их в процессе
	return map[string]camundaClient.Variable{
		"statusUpdated": {Value: true, Type: "Boolean"},
	}, nil
}
