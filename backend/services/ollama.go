package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"recall/models"
	"strings"
)

func ProcessWithOllama(message models.SlackMessage, context []string) (models.ExtractedEvent, error) {

	contexts := strings.Join(context, "\n")
	fmt.Println("Reply contexts:", contexts)
	prompt := `Reply ONLY with raw JSON, no explanation, no markdown:
	{
	"type": "task|decision|deadline|none",
	"summary": "one line summary",
	"owner": "person if mentioned, else null",
	"deadline": "deadline if mentioned, else null",
	"confidence": "high|low"
	}

	Previous conversation context if any:
	"` + contexts + `"

	Current message to analyze:
	"` + message.Text + `"
	
	Rules:
	- Only extract if there is a CLEAR commitment, decision, or deadline that was explicitly agreed upon
	- Vague questions, suggestions, or casual chat should be type "none"
	- Confidence should be "high" only if the commitment is explicit and unambiguous
	- If someone is just asking a question with no answer, type is "none"
	`

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
	fmt.Println("raw ollama response:", ollamaResp.Response)
	var extracted models.ExtractedEvent
	json.Unmarshal([]byte(ollamaResp.Response), &extracted)
	extracted.Text = message.Text
	extracted.Channel = message.Channel
	extracted.Timestamp = message.Timestamp
	extracted.User = message.User
	if extracted.Confidence == "high" {
		extracted.Status = "confirmed"
	} else {
		extracted.Status = "pending"
	}
	return extracted, nil
}
