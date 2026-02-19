package services

import (
	"fmt"
)

func (s *Service) GetPublicIDFromDossier(crmID string) (string, error) {
	// Эмуляция запроса к https://backend.homebank.kz/dossier-dosmod-usersystems/dossier/
	fmt.Printf("--- Calling DossierDosmodUsersystems for CRMId: %s ---\n", crmID)

	// Имитируем успешный ответ
	if crmID == "" {
		return "", nil
	}

	// Возвращаем тестовый userId
	return "USER-PUB-123456789", nil
}
