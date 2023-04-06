package telegram

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func SendNewMessage(message string) error {

	url := "https://api.telegram.org/bot5661178899:AAE47vn8dVdEFKwpCLHRTXLSmQFO3OVVJM8/sendMessage?chat_id=5077055661"

	body := map[string]string{
		"text":       message,
		"parse_mode": "HTML",
	}

	json_data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}
