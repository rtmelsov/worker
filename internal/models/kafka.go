// Package models
package models

// MetricsPayload описывает структуру JSON для Kafka
type MetricsPayload struct {
	ProcID            string      `json:"procID"`
	BusinessKey       string      `json:"businessKey"`
	ProcStatus        string      `json:"procStatus"`
	StartTime         string      `json:"startTime"`
	EndTime           string      `json:"endTime"`
	ProcessReference  interface{} `json:"processReference"`
	Channel           interface{} `json:"channel"`
	Iin               interface{} `json:"iin"`
	Branch            string      `json:"branch"`
	ProcessStartTime  interface{} `json:"processStartTime"`
	HumanTaskDuration interface{} `json:"humanTaskDuration"`
}
