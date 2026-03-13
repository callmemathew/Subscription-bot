package bot

import (
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (a *App) Register() {
	a.Bot.Handle("/start", func(c tele.Context) error {
		a.clearSession(c.Chat().ID)
		return c.Send("Главное меню", a.menu())
	})

	a.Bot.Handle(tele.OnText, func(c tele.Context) error {
		chatID := c.Chat().ID
		text := strings.TrimSpace(c.Text())

		switch text {
		case "Статистика":
			a.clearSession(chatID)
			return a.sendStats(c)

		case "Добавить":
			a.setSession(chatID, &Session{Step: "name"})
			return c.Send("Введите имя клиента", a.backMenu())

		case "Продлить":
			a.setSession(chatID, &Session{Step: "extend"})
			return a.sendExtendList(c)

		case "Удалить":
			a.setSession(chatID, &Session{Step: "delete"})
			return a.sendDeleteList(c)

		case "Список":
			a.clearSession(chatID)
			return c.Send("Выберите список", a.listMenu())

		case "Скоро истекают":
			a.clearSession(chatID)
			return a.sendExpiringSoon(c)

		case "Все":
			a.clearSession(chatID)
			return a.sendList(c, "")

		case "Только месячные":
			a.clearSession(chatID)
			return a.sendList(c, "monthly")

		case "Только разовые":
			a.clearSession(chatID)
			return a.sendList(c, "single")

		case "Назад":
			a.clearSession(chatID)
			return c.Send("Главное меню", a.menu())

		case "Месячный", "Разовый":
			return a.handleTypeChoice(c, text)

		case "Сегодня":
			return a.handleTodayChoice(c)
		}

		return a.handleStep(c, text)
	})
}
