package models

// StatusPayload описывает структуру JSON для сохранения статуса
type StatusPayload struct {
	ProcID               string      `json:"procID"`
	BusinessKey          string      `json:"businessKey"`
	ProcStatus           string      `json:"procStatus"`
	StartTime            string      `json:"startTime"`
	EndTime              string      `json:"endTime"`
	ProcessReference     interface{} `json:"processReference"`
	Channel              interface{} `json:"channel"`
	Iin                  interface{} `json:"iin"`
	RequestedAmount      interface{} `json:"requestedAmount"`
	CardAmountApproved   float64     `json:"cardAmountApproved"`
	CreditAmountApproved float64     `json:"creditAmountApproved"`
	RefAmount            float64     `json:"refAmount"`
	RequestedProduct     interface{} `json:"requestedProduct"`
	ApprovedProduct      string      `json:"approvedProduct"`
	Branch               interface{} `json:"branch"`
	Initiator            interface{} `json:"initiator"`
	ProductCode          interface{} `json:"productCode"`
}
