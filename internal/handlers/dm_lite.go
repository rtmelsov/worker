package handlers

import (
	"encoding/json"
	"fmt"
	"worker/internal/helpers"

	camundaClient "github.com/citilinkru/camunda-client-go/v3"
	"github.com/citilinkru/camunda-client-go/v3/processor"
)

// --- Структуры для внешних API ---

// RandomUserResponse 1. Random User (для имитации клиента)
type RandomUserResponse struct {
	Results []struct {
		Gender string `json:"gender"`
		Name   struct {
			First string `json:"first"`
			Last  string `json:"last"`
		} `json:"name"`
		Dob struct {
			Age int `json:"age"`
		} `json:"dob"`
		Location struct {
			City    string `json:"city"`
			Country string `json:"country"`
		} `json:"location"`
	} `json:"results"`
}

// CoinDeskResponse 2. CoinDesk (для генерации Кредитного Лимита на основе BTC)
type CoinDeskResponse struct {
	Bpi struct {
		USD struct {
			RateFloat float64 `json:"rate_float"`
		} `json:"USD"`
	} `json:"bpi"`
}

// CurrencyResponse 3. Frankfurter (Курсы валют)
type CurrencyResponse struct {
	Rates map[string]float64 `json:"rates"`
	Date  string             `json:"date"`
}

// --- Основной метод ---

func (h *Handler) GatherDMLiteData(ctx *processor.Context) (map[string]camundaClient.Variable, error) {
	h.logger.Infof("Starting Real-World Data Collection for process %s", ctx.Task.ProcessInstanceId)

	// =========================================================================
	// ШАГ 1: Получаем данные о "человеке" (RandomUser.me)
	// =========================================================================
	var userResp RandomUserResponse
	if err := helpers.FetchJSON("https://randomuser.me/api/", &userResp); err != nil {
		h.logger.Error("Failed to fetch RandomUser", err)
	}

	userData := userResp.Results[0]
	h.logger.Infof("Generated Client: %s %s, Age: %d, City: %s",
		userData.Name.First, userData.Name.Last, userData.Dob.Age, userData.Location.City)

	// Логика: Если мужчина -> есть военник
	militaryService := userData.Gender == "male"

	// Логика: Доход зависит от возраста (Возраст * 5000)
	income := userData.Dob.Age * 5000

	// =========================================================================
	// ШАГ 2: Считаем Кредитный Лимит по курсу Биткоина (CoinDesk)
	// =========================================================================
	var btcResp CoinDeskResponse
	creditLimit := 0
	if err := helpers.FetchJSON("https://api.coindesk.com/v1/bpi/currentprice.json", &btcResp); err != nil {
		h.logger.Error("Failed to fetch Bitcoin Price", err)
		creditLimit = 500000 // Фоллбек
	} else {
		// Лимит = Цена биткоина * 10 (в тенге)
		creditLimit = int(btcResp.Bpi.USD.RateFloat) * 10
		h.logger.Infof("Bitcoin Price is %.2f USD. Calculated Credit Limit: %d", btcResp.Bpi.USD.RateFloat, creditLimit)
	}

	// =========================================================================
	// ШАГ 3: Получаем курсы валют (Frankfurter)
	// =========================================================================
	var currResp CurrencyResponse
	// Берем курс к Евро, так как API бесплатное и базовое
	helpers.FetchJSON("https://api.frankfurter.app/latest?from=USD&to=EUR,KZT", &currResp)

	// =========================================================================
	// ШАГ 4: Сборка JSON ответов для Camunda
	// =========================================================================

	// 4.1 JSON Кредитного лимита
	creditLimitJSON, _ := json.Marshal(map[string]any{
		"limit":       creditLimit,
		"currency":    "KZT",
		"btc_rate":    btcResp.Bpi.USD.RateFloat,
		"client_type": "generated_vip",
	})

	// 4.2 JSON Курсов валют
	currencyJSON, _ := json.Marshal(currResp)

	// 4.3 JSON Дохода
	incomeJSON, _ := json.Marshal(map[string]any{
		"official_income": income,
		"source":          "random_generator",
		"employer":        userData.Location.City + " Ltd.",
	})

	// =========================================================================
	// ШАГ 5: Возврат переменных
	// =========================================================================
	return map[string]camundaClient.Variable{
		// Бизнес-данные из внешних API
		"militaryService":              {Value: militaryService, Type: "Boolean"},
		"creditLimitJson":              {Value: string(creditLimitJSON), Type: "String"},
		"incomeData":                   {Value: string(incomeJSON), Type: "String"},
		"getCurrencyRatesResponseData": {Value: string(currencyJSON), Type: "String"},

		// Статусы (рандомно на основе возраста)
		"resultBlockingSign": {Value: userData.Dob.Age > 80, Type: "Boolean"}, // Блок если старше 80
		"resultInvalid":      {Value: false, Type: "Boolean"},
		"phoneVerifyStatus":  {Value: "VERIFIED_BY_API", Type: "String"},

		// Заглушки пустые (чтобы процесс не упал)
		"pkbFrodScoringResp": {Value: "{}", Type: "String"},
		"accountDebtsResp":   {Value: "{}", Type: "String"},
		"antifrodSFDResp":    {Value: "{}", Type: "String"},
		"antifrodSFDRespDM":  {Value: "{}", Type: "String"},
	}, nil
}

// LiteProcessRouter - маршрутизатор для топика "lite-process"
func (h *Handler) LiteProcessRouter(ctx *processor.Context) (map[string]camundaClient.Variable, error) {
	// Смотрим на переменную startDefinitionKey из XML
	defKeyVar := helpers.GetVar(ctx, "startDefinitionKey")

	defKey := fmt.Sprintf("%v", defKeyVar)

	if defKey == "DMDataCollection" {
		// Вызываем наш новый крутой метод с API
		return h.GatherDMLiteData(ctx)
	}

	// Для остальных (пока заглушка, чтобы не падало)
	h.logger.Infof("Skipping logic for definition key: %s", defKey)
	return nil, nil
}
