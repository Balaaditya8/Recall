package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"recall/models"
	"strconv"
	"strings"
	"time"
)

var decisionTools = []models.OllamaTool{
	{
		Type: "function",
		Function: models.OllamaToolFunction{
			Name:        "save_decision",
			Description: "Save a new decision, task, or deadline",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type":       map[string]string{"type": "string"},
					"summary":    map[string]string{"type": "string"},
					"owner":      map[string]string{"type": "string", "description": "the person responsible, extracted from message"},
					"deadline":   map[string]string{"type": "string", "description": "due date in YYYY-MM-DD format, extracted from message"},
					"confidence": map[string]string{"type": "string", "description": "high or low"},
				},
				"required": []string{"type", "summary", "confidence"},
			},
		},
	},
	{
		Type: "function",
		Function: models.OllamaToolFunction{
			Name:        "update_decision",
			Description: "Update an existing decision that is being amended",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":       map[string]string{"type": "string"},
					"summary":  map[string]string{"type": "string"},
					"owner":    map[string]string{"type": "string"},
					"deadline": map[string]string{"type": "string"},
				},
				"required": []string{"id"},
			},
		},
	},
	{
		Type: "function",
		Function: models.OllamaToolFunction{
			Name:        "ignore",
			Description: "Ignore this message, nothing worth saving",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	},
}

func ProcessWithOllama(message models.SlackMessage, context []string, existing []models.ExtractedEvent) (models.ExtractedEvent, error) {

	contextStr := strings.Join(context, "\n")

	// build existing decisions string
	var existingStr string
	for _, e := range existing {
		existingStr += fmt.Sprintf("- [id:%d] [%s] %s (owner: %s, deadline: %s)\n", e.ID, e.Type, e.Summary, e.Owner, e.Deadline)
	}

	systemPrompt := `You are an assistant that monitors Slack conversations and captures decisions, tasks, and deadlines.

		You have three tools:
		- save_decision: use when a NEW clear commitment, decision or deadline is made
		- update_decision: use when an EXISTING decision is being amended or changed
		- ignore: use when nothing worth saving is happening

		Rules:
		- Only save if there is a CLEAR commitment, decision, or deadline explicitly agreed upon
		- Vague questions, suggestions, or casual chat then ignore
		- If someone is just asking a question with no answer then ignore
		- type "task" when someone commits to DO something (e.g. "Jake will finish the API")
		- type "deadline" when a due date is set without a specific action (e.g. "launch is on Friday")
		- type "decision" when a choice is made (e.g. "let's go with React", "we decided on postgres")
		- "let's go with X", "we're going with X", "agreed on X" are always decisions with high confidence
		- Summary must be under 10 words, clear and concise
			- Good: "Andy will finish the API by Friday"
			- Bad: "Virg will own the dashboard bug and will try and complete by coming monday"
		- When owner or deadline changes, always regenerate the summary to reflect the new information
		- Use save_decision with confidence "high" ONLY if ALL of these are true:
		1. The commitment is explicit and unambiguous
		2. It is clear WHAT needs to be done
		3. There is enough context to understand the full commitment
		- Use save_decision with confidence "low" if ANY of these are true:
		1. Owner is null and it is unclear who is responsible
		2. The summary is vague or refers to "this task", "it", "that thing" without explanation
		3. No previous context and the message alone is not self-explanatory
		4. You are unsure about any key detail
		- Current date: ` + time.Now().Format("2006-01-02") + `
		- Use actual dates for deadlines, not relative terms like "Friday" or "next Monday"`

	userMessage := "Existing saved decisions:\n" + existingStr +
		"\n\nPrevious conversation context:\n" + contextStr +
		"\n\n---\nCurrent message to analyze: \"" + message.Text + "\"\n---"

	fmt.Println("existing decisions passed:", existingStr)
	fmt.Println("context passed:", contextStr)
	fmt.Println("usermessage:", userMessage)
	reqBody := models.OllamaRequest{
		Model: "qwen3.5:9b",
		Messages: []models.OllamaMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Tools:  decisionTools,
		Stream: false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return models.ExtractedEvent{}, err
	}

	resp, err := http.Post(
		"http://localhost:11434/api/chat",
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

	fmt.Println("raw ollama response:", ollamaResp)

	// check if model called a tool
	if len(ollamaResp.Message.ToolCalls) == 0 {
		return models.ExtractedEvent{}, nil // no tool called, ignore
	}

	toolCall := ollamaResp.Message.ToolCalls[0]
	fmt.Println("tool called:", toolCall.Function.Name)
	fmt.Println("arguments:", toolCall.Function.Arguments)

	var extracted models.ExtractedEvent
	extracted.Text = message.Text
	extracted.Channel = message.Channel
	extracted.Timestamp = message.Timestamp
	extracted.User = message.User

	switch toolCall.Function.Name {
	case "ignore":
		return models.ExtractedEvent{}, nil

	case "save_decision":
		args := toolCall.Function.Arguments
		extracted.Type = fmt.Sprintf("%v", args["type"])
		extracted.Summary = fmt.Sprintf("%v", args["summary"])
		extracted.Owner = fmt.Sprintf("%v", args["owner"])
		extracted.Deadline = fmt.Sprintf("%v", args["deadline"])
		extracted.Confidence = fmt.Sprintf("%v", args["confidence"])
		if extracted.Confidence == "high" {
			extracted.Status = "confirmed"
		} else {
			extracted.Status = "pending"
		}
		if args["owner"] != nil {
			owner := fmt.Sprintf("%v", args["owner"])
			if owner != "<nil>" && owner != "null" && owner != "unassigned" {
				extracted.Owner = owner
			}
		}
		if args["deadline"] != nil {
			deadline := fmt.Sprintf("%v", args["deadline"])
			if deadline != "<nil>" && deadline != "null" && deadline != "none" {
				extracted.Deadline = deadline
			}
		}

	case "update_decision":
		args := toolCall.Function.Arguments
		idStr := fmt.Sprintf("%v", args["id"])
		id, _ := strconv.Atoi(idStr)
		extracted.ID = id
		if args["summary"] != nil {
			extracted.Summary = fmt.Sprintf("%v", args["summary"])
		}
		if args["owner"] != nil {
			extracted.Owner = fmt.Sprintf("%v", args["owner"])
		}
		if args["deadline"] != nil {
			extracted.Deadline = fmt.Sprintf("%v", args["deadline"])
		}
		extracted.Status = "update"
		if args["owner"] != nil {
			owner := fmt.Sprintf("%v", args["owner"])
			if owner != "<nil>" && owner != "null" {
				extracted.Owner = owner
			}
		}
		if args["deadline"] != nil {
			deadline := fmt.Sprintf("%v", args["deadline"])
			if deadline != "<nil>" && deadline != "null" && deadline != "none" {
				extracted.Deadline = deadline
			}
		}
	}

	return extracted, nil
}
