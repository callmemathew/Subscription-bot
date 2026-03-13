package storage

import "database/sql"

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
		return nil, err
	}
	defer rows.Close()

	return scanClients(rows)
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
		return nil, err
	}
	defer rows.Close()

	return scanClients(rows)
}

func MarkNotified(db *sql.DB, id int64) error {
	_, err := db.Exec(`UPDATE clients SET notified_7 = 1 WHERE id = ?`, id)
	return err
}
