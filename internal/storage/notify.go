package storage

import (
	"database/sql"
	"fmt"
)

func ExpiringSoon(db *sql.DB) ([]Client, error) {
	rows, err := db.Query(`
		SELECT id, name, type, purchase_date, expire_date, notified_7
		FROM clients
		WHERE type = 'monthly'
		  AND expire_date IS NOT NULL
		  AND DATE(expire_date) BETWEEN DATE('now') AND DATE('now', '+7 day')
		ORDER BY expire_date ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("ExpiringSoon: query failed: %w", err)
	}
	defer rows.Close()

	clients, err := scanClients(rows)
	if err != nil {
		return nil, fmt.Errorf("ExpiringSoon: scanClients failed: %w", err)
	}

	return clients, nil
}

func ClientsForNotification(db *sql.DB) ([]Client, error) {
	rows, err := db.Query(`
		SELECT id, name, type, purchase_date, expire_date, notified_7
		FROM clients
		WHERE type = 'monthly'
		  AND expire_date IS NOT NULL
		  AND DATE(expire_date) = DATE('now', '+7 day')
		  AND notified_7 = 0
	`)
	if err != nil {
		return nil, fmt.Errorf("ClientsForNotification: query failed: %w", err)
	}
	defer rows.Close()

	clients, err := scanClients(rows)
	if err != nil {
		return nil, fmt.Errorf("ClientsForNotification: scanClients failed: %w", err)
	}

	return clients, nil
}

func MarkNotified(db *sql.DB, id int64) error {
	_, err := db.Exec(`UPDATE clients SET notified_7 = 1 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("MarkNotified: update failed for client_id=%d: %w", id, err)
	}

	return nil
}
