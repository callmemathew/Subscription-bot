package main

import (
	"log"
	"os"
	"strconv"

	"subscription-bot/internal/bot"
	"subscription-bot/internal/storage"

	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("main: .env file not loaded, using system environment")
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("main: BOT_TOKEN is required")
	}

	adminChatIDStr := os.Getenv("ADMIN_CHAT_ID")
	if adminChatIDStr == "" {
		log.Fatal("main: ADMIN_CHAT_ID is required")
	}

	adminChatID, err := strconv.ParseInt(adminChatIDStr, 10, 64)
	if err != nil {
		log.Fatalf("main: invalid ADMIN_CHAT_ID %q: %v", adminChatIDStr, err)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		log.Fatal("main: DB_PATH is required")
	}

	db, err := storage.OpenDB(dbPath)
	if err != nil {
		log.Fatalf("main: failed to open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("main: failed to close database: %v", err)
		}
	}()

	publicURL := os.Getenv("WEBHOOK_URL")
	if publicURL == "" {
		log.Fatal("main: WEBHOOK_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	pref := tele.Settings{
		Token: token,
		Poller: &tele.Webhook{
			Listen: ":" + port,
			Endpoint: &tele.WebhookEndpoint{
				PublicURL: publicURL,
			},
		},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatalf("main: failed to create bot: %v", err)
	}

	app := bot.NewApp(b, db, adminChatID)
	app.Register()
	app.StartNotifications()

	log.Printf("main: bot started in webhook mode | public_url=%s | listen_port=%s", publicURL, port)

	b.Start()
}
