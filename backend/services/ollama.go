package services

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type ExtractedEvent struct {
	Type       string `json:"type"`
	Summary    string `json:"summary"`
	Owner      string `json:"owner"`
	Deadline   string `json:"deadline"`
	Confidence string `json:"confidence"`
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func ProcessWithOllama(message string) (ExtractedEvent, error) {

	prompt := `Reply ONLY with raw JSON, no explanation, no markdown:
	{
	"type": "task|decision|deadline|none",
	"summary": "one line summary",
	"owner": "person if mentioned, else null",
	"deadline": "deadline if mentioned, else null",
	"confidence": "high|low"
	}
	Message: "` + message + `"`

	reqBody := OllamaRequest{
		Model:  "mistral:latest",
		Prompt: prompt,
		Stream: false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return ExtractedEvent{}, err
	}

	resp, err := http.Post(
		"http://localhost:11434/api/generate",
		"application/json",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return ExtractedEvent{}, err
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return ExtractedEvent{}, err
	}

	var extracted ExtractedEvent
	json.Unmarshal([]byte(ollamaResp.Response), &extracted)
	return extracted, nil
}
