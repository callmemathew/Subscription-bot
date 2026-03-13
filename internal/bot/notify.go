package bot

import (
	"fmt"
	"time"

	"subscription-bot/internal/storage"

	tele "gopkg.in/telebot.v3"
)

func (a *App) StartNotifications() {
	ticker := time.NewTicker(1 * time.Hour)

	go func() {
		a.notify()
		for range ticker.C {
			a.notify()
		}
	}()
}

func (a *App) notify() {
	clients, err := storage.ClientsForNotification(a.DB)
	if err != nil {
		return
	}

	for _, cl := range clients {
		if cl.ExpireDate == nil {
			continue
		}

		msg := fmt.Sprintf(
			"Напоминание\n\nУ клиента %s абонемент истекает %s",
			cl.Name,
			cl.ExpireDate.Format("02.01.2006"),
		)

		_, err := a.Bot.Send(&tele.Chat{ID: a.AdminChatID}, msg)
		if err == nil {
			_ = storage.MarkNotified(a.DB, cl.ID)
		}
	}
}

func (a *App) getSession(chatID int64) *Session {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.sessions[chatID]
}

func (a *App) setSession(chatID int64, s *Session) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sessions[chatID] = s
}

func (a *App) clearSession(chatID int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.sessions, chatID)
}
