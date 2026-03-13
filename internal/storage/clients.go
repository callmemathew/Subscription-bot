package storage

import (
	"database/sql"
	"time"
)

type Client struct {
	ID           int64
	Name         string
	Type         string
	PurchaseDate time.Time
	ExpireDate   *time.Time
	Notified7    bool
}

func AddClient(db *sql.DB, name, subType string, purchaseDate time.Time) error {
	var expire any = nil

	if subType == "monthly" {
		expire = purchaseDate.AddDate(0, 0, 30).Format(time.RFC3339)
	}

	_, err := db.Exec(`
		INSERT INTO clients (name, type, purchase_date, expire_date, notified_7)
		VALUES (?, ?, ?, ?, 0)
	`, name, subType, purchaseDate.Format(time.RFC3339), expire)

	return err
}

func ListClients(db *sql.DB, filter string) ([]Client, error) {
	query := `
	SELECT id, name, type, purchase_date, expire_date, notified_7
	FROM clients
	`
	var args []any

	if filter != "" {
		query += " WHERE type = ?"
		args = append(args, filter)
	}

	query += " ORDER BY id DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanClients(rows)
}

func DeleteClient(db *sql.DB, id int64) error {
	_, err := db.Exec(`
		DELETE FROM clients
		WHERE id = ?
	`, id)

	return err
}

func ExtendClientFromToday(db *sql.DB, id int64) error {
	now := time.Now()
	expire := now.AddDate(0, 0, 30)

	_, err := db.Exec(`
		UPDATE clients
		SET purchase_date = ?, expire_date = ?, notified_7 = 0
		WHERE id = ?
	`,
		now.Format(time.RFC3339),
		expire.Format(time.RFC3339),
		id,
	)

	return err
}

func scanClients(rows *sql.Rows) ([]Client, error) {
	var clients []Client

	for rows.Next() {
		var c Client
		var purchaseStr string
		var expireStr sql.NullString
		var notified int

		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Type,
			&purchaseStr,
			&expireStr,
			&notified,
		)
		if err != nil {
			return nil, err
		}

		c.PurchaseDate, err = time.Parse(time.RFC3339, purchaseStr)
		if err != nil {
			return nil, err
		}

		if expireStr.Valid {
			t, err := time.Parse(time.RFC3339, expireStr.String)
			if err != nil {
				return nil, err
			}
			c.ExpireDate = &t
		}

		c.Notified7 = notified == 1
		clients = append(clients, c)
	}

	return clients, rows.Err()
}
