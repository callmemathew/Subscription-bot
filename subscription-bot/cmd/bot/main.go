package main

import (
	"log"
	"os"
	"strconv"

	"subscription-bot/subscription-bot/internal/bot"
	"subscription-bot/subscription-bot/internal/storage"

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
		dbPath = "data/colonel.db"
	}

	db, err := storage.OpenDB(dbPath)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	publicURL := os.Getenv("WEBHOOK_URL")
	if publicURL == "" {
		log.Fatal("WEBHOOK_URL is required")
	}

	pref := tele.Settings{
		Token: token,
		Poller: &tele.Webhook{
			Listen: ":8080",
			Endpoint: &tele.WebhookEndpoint{
				PublicURL: publicURL,
			},
		},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	app := bot.NewApp(b, db, adminChatID)

	app.Register()
	app.StartNotifications()

	log.Println("bot started with webhook...")
	b.Start()
}
