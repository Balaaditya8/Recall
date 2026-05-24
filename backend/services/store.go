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
