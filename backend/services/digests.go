package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"recall/models"
	"sync"
	"time"
)

func GetTodaysDecisions(db *sql.DB, channel string) []models.ExtractedEvent {

	rows, err := db.Query(`
    SELECT id, type, summary, owner, deadline, confidence, channel, timestamp, slack_user, created_at, status
    FROM decisions WHERE DATE(created_at) = CURRENT_DATE and channel = $1
    ORDER BY created_at DESC
`, channel)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var decisions []models.ExtractedEvent
	for rows.Next() {
		var d models.ExtractedEvent
		rows.Scan(&d.ID, &d.Type, &d.Summary, &d.Owner, &d.Deadline, &d.Confidence, &d.Channel, &d.Timestamp, &d.User, &d.CreatedAt, &d.Status)
		decisions = append(decisions, d)
	}
	return decisions
}

func GetActiveChannelsToday(db *sql.DB) []string {
	rows, err := db.Query(`
	SELECT DISTINCT channel FROM decisions WHERE DATE(created_at) = CURRENT_DATE
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var channels []string
	for rows.Next() {
		var channelID string
		rows.Scan(&channelID)
		channels = append(channels, channelID)
	}
	return channels
}

func DigestAlreadyExists(db *sql.DB, channel string) bool {
	today := time.Now().Format("2006-01-02")
	var exists bool
	err := db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM digests
            WHERE date = $1 AND channel = $2
        )
    `, today, channel).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func SaveDigest(db *sql.DB, digest models.Digest) error {
	_, err := db.Exec(`
    INSERT INTO digests (date, channel, summary)
    VALUES ($1, $2, $3)`,
		time.Now().Format("2006-01-02"), digest.Channel, digest.Summary,
	)
	return err
}

func GenerateDigest(decisions []models.ExtractedEvent) string {
	systemPrompt := `You are an assistant that creates a daily digest summary for decisions made today in a Slack channel.

		Write a clean, concise digest in this format:
		- Start with a one line overview of the day
		- List each decision/task/deadline as a bullet point
		- End with a one line closing summary

		Keep it short and readable. Plain text only, no markdown, no JSON.`
	var channel_decisions string
	for _, e := range decisions {
		channel_decisions += fmt.Sprintf("- [id:%d] [%s] %s (owner: %s, deadline: %s)\n", e.ID, e.Type, e.Summary, e.Owner, e.Deadline)
	}

	userMessage := "Decisions of this channel today:\n" + channel_decisions

	reqBody := models.OllamaRequest{
		Model: "qwen3.5:9b",
		Messages: []models.OllamaMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Stream: false,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return ""
	}
	resp, err := http.Post(
		"http://localhost:11434/api/chat",
		"application/json",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var ollamaResp models.OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return ""
	}

	return ollamaResp.Message.Content
}

func RunDailyDigest(db *sql.DB) {
	channels := GetActiveChannelsToday(db)
	fmt.Println("active channels today:", channels)

	var wg sync.WaitGroup
	for _, channel := range channels {
		wg.Add(1)
		go func(ch string) {
			defer wg.Done()
			fmt.Println("processing channel:", ch)

			if DigestAlreadyExists(db, ch) {
				fmt.Println("digest already exists for:", ch)
				return
			}
			decisions := GetTodaysDecisions(db, ch)
			fmt.Println("decisions found:", len(decisions))

			if len(decisions) == 0 {
				return
			}
			summary := GenerateDigest(decisions)
			fmt.Println("summary generated:", summary)

			err := SaveDigest(db, models.Digest{Channel: ch, Summary: summary})
			fmt.Println("save error:", err)
		}(channel)
	}
	wg.Wait()
}
