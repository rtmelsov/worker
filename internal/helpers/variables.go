// Package helpers
package helpers

import (
	"fmt"

	"github.com/citilinkru/camunda-client-go/v3/processor"
)

func GetVar(ctx *processor.Context, key string) string {
	if v, ok := ctx.Task.Variables[key]; ok && v.Value != nil {
		return fmt.Sprintf("%v", v.Value)
	}
	return ""
}

func GetVarBool(ctx *processor.Context, key string) bool {
	if v, ok := ctx.Task.Variables[key]; ok && v.Value != nil {
		value, ok := v.Value.(bool)
		if !ok {
			return false
		}
		return value
	}
	return false
}
