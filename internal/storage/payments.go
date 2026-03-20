package storage

import (
	"database/sql"
	"strconv"
	"time"
)

func AddPayment(db *sql.DB, name, subType string, amount int) error {
	_, err := db.Exec(`
		INSERT INTO payments (client_name, type, amount, paid_at)
		VALUES (?, ?, ?, ?)
	`,
		name,
		subType,
		amount,
		time.Now().Format(time.RFC3339),
	)

	return err
}

func GetTotalMoney(db *sql.DB) (int, error) {
	row := db.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM payments`)

	var total int
	err := row.Scan(&total)
	return total, err
}

func GetMoneyByType(db *sql.DB, subType string) (int, error) {
	row := db.QueryRow(
		`SELECT COALESCE(SUM(amount), 0) FROM payments WHERE type = ?`,
		subType,
	)

	var total int
	err := row.Scan(&total)
	return total, err
}

func GetMoneyLastDays(db *sql.DB, days int) (int, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM payments
		WHERE paid_at >= date('now', ?)
	`

	row := db.QueryRow(query, "-"+strconv.Itoa(days)+" days")

	var total int
	err := row.Scan(&total)
	return total, err
}
