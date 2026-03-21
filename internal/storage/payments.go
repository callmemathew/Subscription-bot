package storage

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// ==================== ADD PAYMENT ====================

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
	if err != nil {
		return fmt.Errorf("AddPayment: insert failed (name=%s, type=%s, amount=%d): %w",
			name, subType, amount, err)
	}

	return nil
}

// ==================== TOTAL ====================

func GetTotalMoney(db *sql.DB) (int, error) {
	row := db.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM payments`)

	var total int
	err := row.Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("GetTotalMoney: scan failed: %w", err)
	}

	return total, nil
}

// ==================== BY TYPE ====================

func GetMoneyByType(db *sql.DB, subType string) (int, error) {
	row := db.QueryRow(
		`SELECT COALESCE(SUM(amount), 0) FROM payments WHERE type = ?`,
		subType,
	)

	var total int
	err := row.Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("GetMoneyByType: scan failed (type=%s): %w", subType, err)
	}

	return total, nil
}

// ==================== LAST N DAYS ====================

func GetMoneyLastDays(db *sql.DB, days int) (int, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM payments
		WHERE paid_at >= date('now', ?)
	`

	param := "-" + strconv.Itoa(days) + " days"

	row := db.QueryRow(query, param)

	var total int
	err := row.Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("GetMoneyLastDays: scan failed (days=%d): %w", days, err)
	}

	return total, nil
}
