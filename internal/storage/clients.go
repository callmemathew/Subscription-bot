package storage

import (
	"database/sql"
	"fmt"
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
	if err != nil {
		return fmt.Errorf(
			"AddClient: insert failed (name=%s, type=%s, purchaseDate=%s): %w",
			name,
			subType,
			purchaseDate.Format(time.RFC3339),
			err,
		)
	}

	return nil
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
		return nil, fmt.Errorf("ListClients: query failed (filter=%q): %w", filter, err)
	}
	defer rows.Close()

	clients, err := scanClients(rows)
	if err != nil {
		return nil, fmt.Errorf("ListClients: scanClients failed (filter=%q): %w", filter, err)
	}

	return clients, nil
}

func DeleteClient(db *sql.DB, id int64) error {
	_, err := db.Exec(`
		DELETE FROM clients
		WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("DeleteClient: delete failed for client_id=%d: %w", id, err)
	}

	return nil
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
	if err != nil {
		return fmt.Errorf(
			"ExtendClientFromToday: update failed for client_id=%d, purchase_date=%s, expire_date=%s: %w",
			id,
			now.Format(time.RFC3339),
			expire.Format(time.RFC3339),
			err,
		)
	}

	return nil
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
			return nil, fmt.Errorf("scanClients: rows.Scan failed: %w", err)
		}

		c.PurchaseDate, err = time.Parse(time.RFC3339, purchaseStr)
		if err != nil {
			return nil, fmt.Errorf("scanClients: purchase date parse failed (%s): %w", purchaseStr, err)
		}

		if expireStr.Valid {
			t, err := time.Parse(time.RFC3339, expireStr.String)
			if err != nil {
				return nil, fmt.Errorf("scanClients: expire date parse failed (%s): %w", expireStr.String, err)
			}
			c.ExpireDate = &t
		}

		c.Notified7 = notified == 1
		clients = append(clients, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scanClients: rows iteration failed: %w", err)
	}

	return clients, nil
}
