package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"subscription-bot/internal/storage"

	tele "gopkg.in/telebot.v3"
)

func (a *App) handleStep(c tele.Context, text string) error {
	s := a.getSession(c.Chat().ID)
	if s == nil {
		return c.Send("Выберите действие из меню", a.menu())
	}

	switch s.Step {
	case "extend":
		num, err := strconv.Atoi(text)
		if err != nil {
			return c.Send("Введите номер клиента", a.backMenu())
		}

		clients, err := storage.ListClients(a.DB, "monthly")
		if err != nil {
			return c.Send("Ошибка базы", a.menu())
		}

		if num < 1 || num > len(clients) {
			return c.Send("Неверный номер", a.backMenu())
		}

		client := clients[num-1]

		err = storage.ExtendClientFromToday(a.DB, client.ID)
		if err != nil {
			return c.Send("Ошибка продления", a.menu())
		}

		newExpire := time.Now().AddDate(0, 0, 30).Format("02.01.2006")

		a.clearSession(c.Chat().ID)

		return c.Send(
			fmt.Sprintf("Абонемент продлён\n\nКлиент: %s\nНовая дата окончания: %s", client.Name, newExpire),
			a.menu(),
		)

	case "delete":
		num, err := strconv.Atoi(text)
		if err != nil {
			return c.Send("Введите номер клиента", a.backMenu())
		}

		clients, err := storage.ListClients(a.DB, "")
		if err != nil {
			return c.Send("Ошибка базы", a.menu())
		}

		if num < 1 || num > len(clients) {
			return c.Send("Неверный номер", a.backMenu())
		}

		client := clients[num-1]
		s.ClientID = client.ID
		s.Name = client.Name
		s.Step = "delete_confirm"

		return c.Send(fmt.Sprintf("Удалить клиента %s?", client.Name), a.confirmDeleteMenu())

	case "delete_confirm":
		if text == "Да" {
			err := storage.DeleteClient(a.DB, s.ClientID)
			if err != nil {
				return c.Send("Ошибка удаления", a.menu())
			}

			a.clearSession(c.Chat().ID)
			return c.Send(fmt.Sprintf("Клиент %s удалён", s.Name), a.menu())
		}

		if text == "Нет" {
			a.clearSession(c.Chat().ID)
			return c.Send("Удаление отменено", a.menu())
		}

		return c.Send("Нажмите «Да» или «Нет»", a.confirmDeleteMenu())

	case "name":
		s.Name = text
		s.Step = "type"
		return c.Send("Выберите тип абонемента", a.typeMenu())

	case "date":
		purchaseDate, err := time.Parse("02.01.2006", text)
		if err != nil {
			return c.Send("Введите дату в формате ДД.ММ.ГГГГ\nНапример: 10.03.2026", a.dateMenu())
		}

		return a.saveClient(c, s, purchaseDate)

	default:
		return c.Send("Главное меню", a.menu())
	}
}

func (a *App) handleTypeChoice(c tele.Context, text string) error {
	s := a.getSession(c.Chat().ID)
	if s == nil || s.Step != "type" {
		return c.Send("Сначала нажмите «Добавить»", a.menu())
	}

	if text == "Месячный" {
		s.Type = "monthly"
	} else {
		s.Type = "single"
	}

	s.Step = "date"

	return c.Send(
		"Введите дату покупки абонемента в формате ДД.ММ.ГГГГ\n\nНапример: 10.03.2026\n\nИли нажмите кнопку «Сегодня»",
		a.dateMenu(),
	)
}

func (a *App) handleTodayChoice(c tele.Context) error {
	s := a.getSession(c.Chat().ID)
	if s == nil || s.Step != "date" {
		return c.Send("Сначала нажмите «Добавить»", a.menu())
	}

	return a.saveClient(c, s, time.Now())
}

func (a *App) saveClient(c tele.Context, s *Session, purchaseDate time.Time) error {
	err := storage.AddClient(a.DB, s.Name, s.Type, purchaseDate)
	if err != nil {
		return c.Send("Ошибка при сохранении клиента: "+err.Error(), a.menu())
	}

	msg := fmt.Sprintf(
		"Клиент добавлен\n\nИмя: %s\nТип: %s\nКупил: %s",
		s.Name,
		mapType(s.Type),
		purchaseDate.Format("02.01.2006"),
	)

	if s.Type == "monthly" {
		expire := purchaseDate.AddDate(0, 0, 30)
		msg += "\nИстекает: " + expire.Format("02.01.2006")
	} else {
		msg += "\nИстекает: —"
	}

	a.clearSession(c.Chat().ID)
	return c.Send(msg, a.menu())
}

func (a *App) sendDeleteList(c tele.Context) error {
	clients, err := storage.ListClients(a.DB, "")
	if err != nil {
		return c.Send("Ошибка при получении списка", a.menu())
	}

	if len(clients) == 0 {
		return c.Send("Список пуст", a.menu())
	}

	var b strings.Builder
	b.WriteString("Выберите номер клиента для удаления\n\n")

	for i, cl := range clients {
		b.WriteString(fmt.Sprintf("%d. %s — %s\n", i+1, cl.Name, mapType(cl.Type)))
	}

	return c.Send(b.String(), a.backMenu())
}

func (a *App) sendExtendList(c tele.Context) error {
	clients, err := storage.ListClients(a.DB, "monthly")
	if err != nil {
		return c.Send("Ошибка при получении списка", a.menu())
	}

	if len(clients) == 0 {
		return c.Send("Нет месячных абонементов для продления", a.menu())
	}

	var b strings.Builder
	b.WriteString("Выберите номер клиента для продления\n\n")

	for i, cl := range clients {
		b.WriteString(fmt.Sprintf("%d. %s — истекает %s\n", i+1, cl.Name, cl.ExpireDate.Format("02.01.2006")))
	}

	return c.Send(b.String(), a.backMenu())
}

func (a *App) sendList(c tele.Context, filter string) error {
	clients, err := storage.ListClients(a.DB, filter)
	if err != nil {
		return c.Send("Ошибка при получении списка", a.menu())
	}

	if len(clients) == 0 {
		return c.Send("Список пуст", a.menu())
	}

	var b strings.Builder
	b.WriteString("Клиенты:\n\n")

	for i, cl := range clients {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, cl.Name))
		b.WriteString(fmt.Sprintf("Тип: %s\n", mapType(cl.Type)))
		b.WriteString(fmt.Sprintf("Купил: %s\n", cl.PurchaseDate.Format("02.01.2006")))

		if cl.ExpireDate != nil {
			b.WriteString(fmt.Sprintf("Истекает: %s\n", cl.ExpireDate.Format("02.01.2006")))
		} else {
			b.WriteString("Истекает: —\n")
		}

		b.WriteString("\n")
	}

	return c.Send(b.String(), a.menu())
}

func (a *App) sendExpiringSoon(c tele.Context) error {
	clients, err := storage.ExpiringSoon(a.DB)
	if err != nil {
		return c.Send("Ошибка при получении данных", a.menu())
	}

	if len(clients) == 0 {
		return c.Send("В ближайшие 7 дней ничего не истекает", a.menu())
	}

	var b strings.Builder
	b.WriteString("Скоро истекают:\n\n")

	for _, cl := range clients {
		if cl.ExpireDate == nil {
			continue
		}

		days := int(cl.ExpireDate.Sub(time.Now()).Hours() / 24)
		if days < 0 {
			days = 0
		}

		b.WriteString(fmt.Sprintf("%s\n", cl.Name))
		b.WriteString(fmt.Sprintf("Истекает: %s\n", cl.ExpireDate.Format("02.01.2006")))
		b.WriteString(fmt.Sprintf("Осталось дней: %d\n\n", days))
	}

	return c.Send(b.String(), a.menu())
}

func (a *App) sendStats(c tele.Context) error {
	stats, err := storage.GetStats(a.DB)
	if err != nil {
		return c.Send("Ошибка при получении статистики", a.menu())
	}

	msg := fmt.Sprintf(
		"Статистика\n\nВсего клиентов: %d\nМесячных: %d\nРазовых: %d\nСкоро истекают: %d\nИстекли: %d",
		stats.Total,
		stats.Monthly,
		stats.Single,
		stats.ExpiringSoon,
		stats.Expired,
	)

	return c.Send(msg, a.menu())
}
