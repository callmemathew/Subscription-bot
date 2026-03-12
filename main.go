package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
)

func main() {
	_ = godotenv.Load()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN is required")
	}

	adminChatIDStr := os.Getenv("ADMIN_CHAT_ID")
	if adminChatIDStr == "" {
		log.Fatal("ADMIN_CHAT_ID is required")
	}

	adminChatID, err := strconv.ParseInt(adminChatIDStr, 10, 64)
	if err != nil {
		log.Fatal("invalid ADMIN_CHAT_ID")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "colonel.db"
	}

	db, err := OpenDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	b, err := tele.NewBot(tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	app := NewApp(b, db, adminChatID)
	app.Register()
	app.StartNotifications()

	log.Println("bot started...")
	b.Start()
}
