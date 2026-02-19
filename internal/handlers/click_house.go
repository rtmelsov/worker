package handlers

import (
	"encoding/json"
	"fmt"
	"time"
	"worker/internal/helpers"
	"worker/internal/models"

	camundaClient "github.com/citilinkru/camunda-client-go/v3"

	"github.com/citilinkru/camunda-client-go/v3/processor"
)

// ClickHouseMetricsHandler выполняет логику Service Task
func (h *Handler) ClickHouseMetricsHandler(ctx *processor.Context) (map[string]camundaClient.Variable, error) {

	clientDecision := helpers.GetVarBool(ctx, "clientDecision")
	externalProcDecline := helpers.GetVarNum(ctx, "externalProcDecline")

	if !clientDecision || externalProcDecline == 1 {
		if err := h.services.SaveProcStatKFK(ctx); err != nil {
			return nil, err
		}
		return map[string]camundaClient.Variable{
			"localStatus": {Type: "string", Value: "sendToEHD"},
		}, nil
	}

	var branch string
	if deptResp := helpers.GetVar(ctx, "getDepartmentResponse"); deptResp != "" {
		var parsed map[string]interface{}
		// deptResp уже строка, поэтому fmt.Sprint не нужен
		_ = json.Unmarshal([]byte(deptResp), &parsed)
		if code, exists := parsed["colvirCode"]; exists {
			branch = fmt.Sprint(code)
		}
	}

	var startTime string
	if chMetrics := helpers.GetVar(ctx, "clickHouseMetrics"); chMetrics != "" {
		var parsed map[string]interface{}
		_ = json.Unmarshal([]byte(chMetrics), &parsed)
		if et, exists := parsed["endTime"]; exists {
			startTime = fmt.Sprint(et)
		}
	}

	humanTaskDuration := helpers.GetVarNum(ctx, "humantTaskDuration")

	payload := models.MetricsPayload{
		ProcID:            ctx.Task.ProcessInstanceId, // Берем ID процесса из контекста
		BusinessKey:       ctx.Task.BusinessKey,       // Берем бизнес-ключ из контекста
		ProcStatus:        "client_chosen_decision",
		StartTime:         startTime,
		EndTime:           time.Now().Format(time.RFC3339),
		ProcessReference:  helpers.GetVar(ctx, "processReference"),
		Channel:           helpers.GetVar(ctx, "channel"),
		Iin:               helpers.GetVar(ctx, "clientIIN"),
		Branch:            branch,
		ProcessStartTime:  helpers.GetVar(ctx, "startProcessTime"),
		HumanTaskDuration: humanTaskDuration,
	}

	jsBytes, err := h.services.SendToKafka(ctx, &payload)
	if err != nil {
		return nil, fmt.Errorf("ошибка при записи данных в кафку: %w", err)
	}

	return map[string]camundaClient.Variable{
		"clickHouseMetrics":  {Type: "string", Value: string(jsBytes)},
		"humantTaskDuration": {Type: "number", Value: 0},
	}, nil
}
