package storage

import "database/sql"

type Stats struct {
	Total        int
	Monthly      int
	Single       int
	ExpiringSoon int
	Expired      int

	TotalMoney   int
	MonthlyMoney int
	SingleMoney  int
}

func GetStats(db *sql.DB) (Stats, error) {
	var s Stats

	err := db.QueryRow(`SELECT COUNT(*) FROM clients`).Scan(&s.Total)
	if err != nil {
		return s, err
	}

	err = db.QueryRow(`SELECT COUNT(*) FROM clients WHERE type = 'monthly'`).Scan(&s.Monthly)
	if err != nil {
		return s, err
	}

	err = db.QueryRow(`SELECT COUNT(*) FROM clients WHERE type = 'single'`).Scan(&s.Single)
	if err != nil {
		return s, err
	}

	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM clients
		WHERE type = 'monthly'
		  AND expire_date IS NOT NULL
		  AND DATE(expire_date) BETWEEN DATE('now') AND DATE('now', '+7 day')
	`).Scan(&s.ExpiringSoon)
	if err != nil {
		return s, err
	}

	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM clients
		WHERE type = 'monthly'
		  AND expire_date IS NOT NULL
		  AND DATE(expire_date) < DATE('now')
	`).Scan(&s.Expired)
	if err != nil {
		return s, err
	}

	err = db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM payments
	`).Scan(&s.TotalMoney)
	if err != nil {
		return s, err
	}

	err = db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM payments
		WHERE type = 'monthly'
	`).Scan(&s.MonthlyMoney)
	if err != nil {
		return s, err
	}

	err = db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM payments
		WHERE type = 'single'
	`).Scan(&s.SingleMoney)
	if err != nil {
		return s, err
	}

	return s, nil
}
