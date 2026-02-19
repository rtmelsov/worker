package services

// InitialData - структура для обмена данными между сервисом и хендлером
type InitialData struct {
	MRP        int
	PublicID   string
	InGreyList *bool // Используем указатель, чтобы понять, была ли проверка
}

func (s *Service) GetCreditConditionDictionary(iin, channel, repeatProcess string) (*InitialData, error) {
	res := &InitialData{}

	// --- 1. Получение МРП (Всегда) ---
	//
	res.MRP = 4125

	// --- 2. Логика Серого списка ---
	// Аналог условия ${repeatProcess == 1}
	if repeatProcess == "1" {
		// ТУТ БУДЕТ ТВОЙ ЗАПРОС К API
		val := false
		res.InGreyList = &val
	}

	// --- 3. Логика Public ID ---
	// Аналог условия ${channel == "HB"}
	if channel == "HB" {
		// ТУТ БУДЕТ ТВОЙ ЗАПРОС К API
		res.PublicID = "PUB-" + iin
	}

	return res, nil
}
