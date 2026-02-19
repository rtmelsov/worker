// Package helpers содержит вспомогательные утилиты для проекта.
package helpers

import (
	camundaClient "github.com/citilinkru/camunda-client-go/v3"
)

var LoanCamundaVariables = map[string]camundaClient.Variable{
	"CRMId":             {Value: "CRM-777-999"},
	"channel":           {Value: "HB"},
	"clientGivenName":   {Value: "Иван"},
	"clientIIN":         {Value: "900101300123"},
	"clientInfoBD":      {Value: "0321321"},
	"clientMiddleName":  {Value: "Иванович"},
	"clientSurname":     {Value: "Иванов"},
	"creditAmount":      {Value: 500000},
	"creditProductType": {Value: "CashLoan"},
	"initiator":         {Value: "ALMATY_MOBILE_APP"},
	// "repeatProcess":      {Value: "1"},
	"startDefinitionKey": {Value: "0321321"},
}
