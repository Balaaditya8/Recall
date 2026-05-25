package services

import (
	"database/sql"
	"recall/models"
)

func SaveDecision(db *sql.DB, event models.ExtractedEvent) error {
	_, err := db.Exec(`
    INSERT INTO decisions 
    (type, summary, owner, deadline, confidence, channel, timestamp, slack_user)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		event.Type, event.Summary, event.Owner, event.Deadline,
		event.Confidence, event.Channel, event.Timestamp, event.User,
	)
	return err
}

func GetDecisions(db *sql.DB) ([]models.ExtractedEvent, error) {
	rows, err := db.Query(`
        SELECT type, summary, owner, deadline, confidence, channel, timestamp, slack_user, created_at 
        FROM decisions 
        ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []models.ExtractedEvent
	for rows.Next() {
		var d models.ExtractedEvent
		rows.Scan(&d.Type, &d.Summary, &d.Owner, &d.Deadline, &d.Confidence, &d.Channel, &d.Timestamp, &d.User, &d.CreatedAt)
		decisions = append(decisions, d)
	}
	return decisions, nil
}
