package bot

import (
	"database/sql"
	"sync"
	"time"

	tele "gopkg.in/telebot.v3"
)

type Session struct {
	Step         string
	Name         string
	Type         string
	ClientID     int64
	PurchaseDate time.Time
	Amount       int
}

type App struct {
	Bot         *tele.Bot
	DB          *sql.DB
	AdminChatID int64

	mu       sync.Mutex
	sessions map[int64]*Session
}

func NewApp(bot *tele.Bot, db *sql.DB, adminChatID int64) *App {
	return &App{
		Bot:         bot,
		DB:          db,
		AdminChatID: adminChatID,
		sessions:    make(map[int64]*Session),
	}
}
