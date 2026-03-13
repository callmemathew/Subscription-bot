package storage

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	tele "gopkg.in/telebot.v3"
)

type Client struct {
	ID           int64
	Name         string
	Type         string
	PurchaseDate time.Time
	ExpireDate   *time.Time
	Notified7    bool
}
type Stats struct {
	Total        int
	Monthly      int
	Single       int
	ExpiringSoon int
	Expired      int
}

func OpenDB(path string) (*sql.DB, error) {
	if path == "" {
		return nil, errors.New("empty db path")
	}

	cleanPath := filepath.Clean(path)

	if _, err := os.Stat(cleanPath); err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("database file not found: " + cleanPath)
		}
		return nil, err
	}

	db, err := sql.Open("sqlite3", cleanPath)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS clients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		purchase_date TEXT NOT NULL,
		expire_date TEXT,
		notified_7 INTEGER NOT NULL DEFAULT 0
	);`

	_, err = db.Exec(schema)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
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

func ExpiringSoon(db *sql.DB) ([]Client, error) {
	rows, err := db.Query(`
		SELECT id, name,type, purchase_date, expire_date, notified_7
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
		SELECT id, name,  type, purchase_date, expire_date, notified_7
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
func dateMenu() *tele.ReplyMarkup {

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}

	btnToday := menu.Text("Сегодня")
	btnBack := menu.Text("Назад")

	menu.Reply(
		menu.Row(btnToday),
		menu.Row(btnBack),
	)

	return menu
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

	return s, nil
}
