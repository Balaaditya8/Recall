package services

import (
	"database/sql"
	"recall/handlers"
	"recall/models"
)

func SaveDecision(db *sql.DB, event models.ExtractedEvent) error {
	_, err := db.Exec(`
    INSERT INTO decisions 
    (type, summary, owner, deadline, confidence, channel, timestamp, slack_user, status)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		event.Type, event.Summary, event.Owner, event.Deadline,
		event.Confidence, event.Channel, event.Timestamp, event.User, event.Status,
	)
	return err
}

func GetDecisions(db *sql.DB) ([]models.ExtractedEvent, error) {
	rows, err := db.Query(`
    SELECT id, type, summary, owner, deadline, confidence, channel, timestamp, slack_user, created_at, status
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
		rows.Scan(&d.ID, &d.Type, &d.Summary, &d.Owner, &d.Deadline, &d.Confidence, &d.Channel, &d.Timestamp, &d.User, &d.CreatedAt, &d.Status)
		decisions = append(decisions, d)
	}
	for i := range decisions {
		decisions[i].Channel = handlers.GetChannelName(decisions[i].Channel)
	}
	return decisions, nil
}

func ConfirmDecision(db *sql.DB, id string, updated models.ExtractedEvent) error {
	_, err := db.Exec(`
        UPDATE decisions 
		SET status = 'confirmed', summary = $1, owner = $2, deadline = $3, type = $4
		WHERE id = $5`,
		updated.Summary, updated.Owner, updated.Deadline, updated.Type, id,
	)
	return err
}

func DismissDecision(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM decisions WHERE id = $1`, id)
	return err
}
