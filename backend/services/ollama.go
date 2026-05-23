package services

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func ProcessWithOllama(message string) (string, error) {
	reqBody := OllamaRequest{
		Model:  "mistral:latest",
		Prompt: "Extract any decision, task, or important memory from this Slack message:\n\n" + message,
		Stream: false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(
		"http://localhost:11434/api/generate",
		"application/json",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Response, nil
}
