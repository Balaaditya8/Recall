package models

type SlackMessage struct {
	Text      string
	Channel   string
	Timestamp string
	User      string
}

type ExtractedEvent struct {
	Type       string `json:"type"`
	Summary    string `json:"summary"`
	Owner      string `json:"owner"`
	Deadline   string `json:"deadline"`
	Confidence string `json:"confidence"`
	Text       string `json:"text"`
	Channel    string `json:"channel"`
	Timestamp  string `json:"timestamp"`
	User       string `json:"user"`
	CreatedAt  string `json:"created_at"`
	Status     string `json:"status"`
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}
