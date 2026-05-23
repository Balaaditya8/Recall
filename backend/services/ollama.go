package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"recall/models"
)

func ProcessWithOllama(message models.SlackMessage) (models.ExtractedEvent, error) {

	prompt := `Reply ONLY with raw JSON, no explanation, no markdown:
	{
	"type": "task|decision|deadline|none",
	"summary": "one line summary",
	"owner": "person if mentioned, else null",
	"deadline": "deadline if mentioned, else null",
	"confidence": "high|low"
	}
	Message: "` + message.Text + `"`

	reqBody := models.OllamaRequest{
		Model:  "mistral:latest",
		Prompt: prompt,
		Stream: false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return models.ExtractedEvent{}, err
	}

	resp, err := http.Post(
		"http://localhost:11434/api/generate",
		"application/json",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return models.ExtractedEvent{}, err
	}
	defer resp.Body.Close()

	var ollamaResp models.OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return models.ExtractedEvent{}, err
	}

	var extracted models.ExtractedEvent
	json.Unmarshal([]byte(ollamaResp.Response), &extracted)
	extracted.Text = message.Text
	extracted.Channel = message.Channel
	extracted.Timestamp = message.Timestamp
	extracted.User = message.User
	return extracted, nil
}
