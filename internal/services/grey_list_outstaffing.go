package services

import (
	"fmt"
)

func (s *Service) CheckGreyListOutstaffing(iin, userID, processID string) (bool, error) {
	// Эмуляция запроса к grey-list-outstaffing/api/Check/GetRecord
	fmt.Printf("--- Calling GreyListOutstaffing for IIN: %s ---\n", iin)

	// Логика заглушки: если ИИН начинается на "666", клиент в сером списке
	if len(iin) > 0 && iin[0] == '6' {
		return true, nil
	}

	return false, nil
}
