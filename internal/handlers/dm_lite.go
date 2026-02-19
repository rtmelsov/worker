package handlers

import (
	"fmt"
	"worker/internal/helpers"

	camundaClient "github.com/citilinkru/camunda-client-go/v3"
	"github.com/citilinkru/camunda-client-go/v3/processor"
)

// LiteProcessRouter - маршрутизатор для топика "lite-process"
func (h *Handler) LiteProcessRouter(ctx *processor.Context) (map[string]camundaClient.Variable, error) {
	defKeyVar := helpers.GetVar(ctx, "startDefinitionKey")

	defKey := fmt.Sprintf("%v", defKeyVar)

	// Для остальных (пока заглушка, чтобы не падало)
	h.logger.Infof("Skipping logic for definition key: %s", defKey)
	return nil, nil
}
