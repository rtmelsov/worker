package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"worker/internal/helpers"

	camundaClient "github.com/citilinkru/camunda-client-go/v3"
	"github.com/citilinkru/camunda-client-go/v3/processor"
)

func (h *Handler) FinishProcess(ctx *processor.Context) (map[string]camundaClient.Variable, error) {
	h.logger.Infof("Finishing process %s...", ctx.Task.ProcessInstanceId)

	var procStatus string
	var localStatus string
	var branch string

	tempPkbResponse := helpers.GetVar(ctx, "temp_pkbResponse")
	if tempPkbResponse == "" {
		return nil, fmt.Errorf("ответ от ПКБ отсутствует (temp_pkbResponse пуст)")
	}

	if tempPkbResponse == "-4395" {
		localStatus = "rejectWithSmsPush"
		procStatus = "pkb_cancel_cancel_int_st"
	}

	channel := helpers.GetVar(ctx, "channel")
	fatcaNeeded := helpers.GetVarBool(ctx, "fatcaNeeded")
	OECDNeeded := helpers.GetVarBool(ctx, "OECDNeeded")

	if channel == "HB" && (fatcaNeeded || OECDNeeded) {
		localStatus = "reject"
		if fatcaNeeded {
			procStatus = "fatca_needed_int_st"
		} else {
			procStatus = "oecd_needed_int_st"
		}
	}

	externalProcDecline := helpers.GetVar(ctx, "externalProcDecline")
	if externalProcDecline == "1" {
		procStatus = "external_proc_decline"
		localStatus = "reject"

		deptResp := helpers.GetVar(ctx, "getDepartmentResponse")
		branch = h.extractColvirCode(deptResp)
	}

	// Если ни одно условие не сработало (успешное завершение)
	if procStatus == "" {
		procStatus = "success"
	}

	// 4. ФОРМИРУЕМ PAYLOAD ДЛЯ KAFKA (Объединяем всё, что было в разных скриптах)
	kafkaPayload := map[string]any{
		"procID":           ctx.Task.ProcessInstanceId,
		"businessKey":      ctx.Task.BusinessKey,
		"procStatus":       procStatus,
		"iin":              helpers.GetVar(ctx, "clientIIN"),
		"requestedAmount":  helpers.GetVar(ctx, "creditAmount"),
		"requestedProduct": helpers.GetVar(ctx, "creditProductType"),
		"branch":           branch, // Заполнится только для внешнего отказа
		"endTime":          time.Now().Format(time.RFC3339),
		"initiator":        "HB",
	}

	h.logger.Infof("💾 Сохранение статуса в Kafka [%s]: %v", procStatus, kafkaPayload)

	// h.kafka.Send("ucp-status", kafkaPayload)

	return map[string]camundaClient.Variable{
		"finalStatus": {Value: procStatus, Type: "String"},
		"localStatus": {Value: localStatus, Type: "String"},
	}, nil
}

// Вспомогательная функция для парсинга JSON ответа департамента
func (h *Handler) extractColvirCode(rawJSON string) string {
	if rawJSON == "" {
		return ""
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(rawJSON), &data); err != nil {
		return ""
	}
	// Имитация .prop("colvirCode").value()
	if code, ok := data["colvirCode"].(string); ok {
		return code
	}
	return ""
}
