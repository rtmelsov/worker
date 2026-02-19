package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citilinkru/camunda-client-go/v3/processor"
	"github.com/segmentio/kafka-go"
	"worker/internal/helpers" // Замените на ваш путь
	"worker/internal/models"
)

func (s *Service) SaveProcStatKFK(ctx *processor.Context) error {
	var cardAmountApproved float64
	var creditAmountApproved float64
	var refAmount float64
	var productCode interface{}

	// 1. Парсим offerSet
	offerSetStr := helpers.GetVar(ctx, "offerSet")
	var offers []map[string]interface{}
	if offerSetStr != "" {
		_ = json.Unmarshal([]byte(offerSetStr), &offers)
	}

	// Вспомогательная функция для безопасного извлечения чисел
	getFloat := func(val interface{}) float64 {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			// Если приходит как строка, но по факту число (опционально можно добавить strconv)
			return 0
		default:
			return 0
		}
	}

	// Вспомогательная функция для приведения кодов к строке (для безопасного сравнения)
	getStr := func(val interface{}) string {
		if val == nil {
			return ""
		}
		return fmt.Sprint(val)
	}

	// 2. Первый цикл (как в JS)
	for _, offer := range offers {
		offerType := getStr(offer["offerType"])
		pCodeStr := getStr(offer["productCode"])
		productCode = offer["productCode"] // сохраняем оригинальное значение как в JS

		if offerType == "1" {
			if pCodeStr == "1" || pCodeStr == "2" {
				cardAmountApproved = getFloat(offer["creditAmount"])
			} else {
				creditAmountApproved = getFloat(offer["creditAmount"])
				refAmount = getFloat(offer["totalRefinanceAmount"])
			}
		}
	}

	// 3. Второй цикл, если суммы всё ещё 0
	if cardAmountApproved == 0 && creditAmountApproved == 0 {
		for _, offer := range offers {
			pCodeStr := getStr(offer["productCode"])
			productCode = offer["productCode"]

			if pCodeStr == "1" || pCodeStr == "2" {
				cardAmountApproved = getFloat(offer["creditAmount"])
			} else {
				creditAmountApproved = getFloat(offer["creditAmount"])
				refAmount = getFloat(offer["totalRefinanceAmount"])
			}
		}
	}

	// 4. Определение approvedProduct
	var approvedProduct string
	if cardAmountApproved > 0 && creditAmountApproved > 0 {
		approvedProduct = "combo"
	} else if cardAmountApproved > 0 {
		approvedProduct = "card"
	} else if creditAmountApproved > 0 {
		approvedProduct = "credit"
	}

	// 5. Определение procStatus
	procStatus := "client_cancel_int_st"
	externalProcDecline := helpers.GetVar(ctx, "externalProcDecline")
	if externalProcDecline == "1" || externalProcDecline == "true" { // На случай разных типов в Camunda
		procStatus = "external_proc_decline"
	}

	// 6. Формирование итогового payload
	payload := models.StatusPayload{
		ProcID:               ctx.Task.ProcessInstanceId,
		BusinessKey:          ctx.Task.BusinessKey,
		ProcStatus:           procStatus,
		StartTime:            helpers.GetVar(ctx, "startProcessTime"),
		EndTime:              time.Now().Format(time.RFC3339),
		ProcessReference:     helpers.GetVar(ctx, "processReference"),
		Channel:              helpers.GetVar(ctx, "channel"),
		Iin:                  helpers.GetVar(ctx, "clientIIN"),
		RequestedAmount:      helpers.GetVar(ctx, "creditAmount"),
		CardAmountApproved:   cardAmountApproved,
		CreditAmountApproved: creditAmountApproved,
		RefAmount:            refAmount,
		RequestedProduct:     helpers.GetVar(ctx, "creditProductType"),
		ApprovedProduct:      approvedProduct,
		Branch:               helpers.GetVar(ctx, "branch"),
		Initiator:            helpers.GetVar(ctx, "initiator"),
		ProductCode:          productCode,
	}

	// 7. Отправка в Kafka
	jsBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Topic: "ucp-status", // Топик из BPMN схемы
		Value: jsBytes,
	}

	err = s.kafkaWriter.WriteMessages(context.Background(), msg)
	if err != nil {
		return err // Вернет ошибку, Camunda сделает retry
	}

	return nil
}
