package handlers

import (
	"fmt"
	camundaClient "github.com/citilinkru/camunda-client-go/v3"
	"github.com/citilinkru/camunda-client-go/v3/processor"
	"time"
	"worker/internal/helpers"
)

// CollectInitialData - собирает МРП, PublicID и проверяет GreyList за один проход
func (h *Handler) CollectInitialData(ctx *processor.Context) (map[string]camundaClient.Variable, error) {
	h.logger.Info("Начало сбора первичных данных...")

	// Результирующая карта переменных, которую мы вернем в Camunda
	outputVars := make(map[string]camundaClient.Variable)

	repeatProcess := helpers.GetVar(ctx, "repeatProcess")
	if repeatProcess == "" {
		// Исправляем имя переменной в тексте
		return nil, fmt.Errorf("не найдено значение переменной repeatProcess")
	}

	channel := helpers.GetVar(ctx, "channel")
	iin := helpers.GetVar(ctx, "clientIIN")

	// --- 1. ПОЛУЧЕНИЕ МРП (Выполняется всегда) ---
	// В старой схеме это был самый первый шаг.
	data, err := h.service.GetCreditConditionDictionary(iin, channel, repeatProcess)
	if err != nil {
		return nil, fmt.Errorf("service error: %w", err)
	}

	outputVars["mrp"] = camundaClient.Variable{Value: data.MRP, Type: "Integer"}
	h.logger.Infof("МРП установлен: %d", data.MRP)

	// --- 2. ЛОГИКА ШЛЮЗА: СЕРЫЙ СПИСОК ---
	// Аналог ромбика 'X' перед серым списком.
	if repeatProcess == "1" {
		h.logger.Info("Условие repeatProcess == 1 выполнено. Запрашиваем серый список...")
		processID := ctx.Task.ProcessInstanceId
		crmID := helpers.GetVar(ctx, "CRMId")
		isInGreyList, err := h.service.CheckGreyListOutstaffing(iin, crmID, processID)
		if err != nil {
			return nil, fmt.Errorf("service error: %w", err)
		}

		outputVars["isInGreyList"] = camundaClient.Variable{Value: isInGreyList, Type: "Boolean"}
		outputVars["greyListCheckDate"] = camundaClient.Variable{Value: time.Now().Format(time.RFC3339), Type: "String"}
	} else {
		h.logger.Info("Пропуск проверки серого списка.")
	}

	// --- 3. ЛОГИКА ШЛЮЗА: PUBLIC ID ---
	// Аналог ромбика 'X' перед Public ID.
	if channel == "HB" {
		h.logger.Info("Канал HB (Homebank). Получаем Public ID...")

		// Имитируем получение ID из системы
		publicID, err := h.service.GetCreditConditionDictionary(iin, channel, repeatProcess)
		if err != nil {
			return nil, fmt.Errorf("service error: %w", err)
		}

		outputVars["publicId"] = camundaClient.Variable{Value: publicID, Type: "String"}
	} else {
		h.logger.Info("Канал не HB. Пропуск получения Public ID.")
	}

	// Возвращаем все собранные данные одним пакетом
	return outputVars, nil
}
