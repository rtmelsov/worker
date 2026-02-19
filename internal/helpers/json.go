package helpers

import (
	"encoding/json"
	"net/http"
	"time"
)

// Хелпер для HTTP запросов
func FetchJSON(url string, target any) error {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}
