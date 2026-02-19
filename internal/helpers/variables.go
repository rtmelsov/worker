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

func GetVarNum(ctx *processor.Context, key string) int32 {
	if v, ok := ctx.Task.Variables[key]; ok && v.Value != nil {
		value, ok := v.Value.(int32)
		if !ok {
			return 0
		}
		return value
	}
	return 0
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
