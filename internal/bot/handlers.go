package bot

import (
	"fmt"
	"strconv"
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

		err = storage.AddPayment(a.DB, client.Name, "monthly", 50)
		if err != nil {
			return c.Send("Ошибка при сохранении оплаты: "+err.Error(), a.menu())
		}

		newExpire := time.Now().AddDate(0, 0, 30)

		a.clearSession(c.Chat().ID)
		return c.Send(formatExtended(client.Name, newExpire, 50), a.menu())

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

		return c.Send(
			fmt.Sprintf("Удалить клиента %s?", client.Name),
			a.confirmDeleteMenu(),
		)

	case "delete_confirm":
		if text == "Да" {
			err := storage.DeleteClient(a.DB, s.ClientID)
			if err != nil {
				return c.Send("Ошибка удаления", a.menu())
			}

			a.clearSession(c.Chat().ID)
			return c.Send(formatDeleted(s.Name), a.menu())
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
			return c.Send(
				"Введите дату в формате ДД.ММ.ГГГГ\nНапример: 10.03.2026",
				a.dateMenu(),
			)
		}

		s.PurchaseDate = purchaseDate
		s.Step = "amount"

		return c.Send("Выберите сумму оплаты", a.amountMenu())

	case "amount":
		if text == "Стандартная" {
			if s.Type == "monthly" {
				s.Amount = 50
			} else {
				s.Amount = 15
			}

			return a.saveClient(c, s)
		}

		if text == "Своя" {
			s.Step = "custom_amount"
			return c.Send("Введите сумму в леях", a.backMenu())
		}

		return c.Send("Выберите вариант", a.amountMenu())

	case "custom_amount":
		amount, err := strconv.Atoi(text)
		if err != nil || amount <= 0 {
			return c.Send("Введите корректную сумму (например 20)", a.backMenu())
		}

		s.Amount = amount
		return a.saveClient(c, s)

	default:
		return c.Send("Главное меню", a.menu())
	}
}

func (a *App) sendDeleteList(c tele.Context) error {
	clients, err := storage.ListClients(a.DB, "")
	if err != nil {
		return c.Send("Ошибка при получении списка", a.menu())
	}

	if len(clients) == 0 {
		return c.Send("Список пуст", a.menu())
	}

	var text string
	text += "🗑 ВЫБЕРИ КЛИЕНТА ДЛЯ УДАЛЕНИЯ\n\n"

	for i, cl := range clients {
		text += fmt.Sprintf("%d. %s — %s\n", i+1, cl.Name, mapType(cl.Type))
	}

	return c.Send(text, a.backMenu())
}

func (a *App) sendExtendList(c tele.Context) error {
	clients, err := storage.ListClients(a.DB, "monthly")
	if err != nil {
		return c.Send("Ошибка при получении списка", a.menu())
	}

	if len(clients) == 0 {
		return c.Send("Нет месячных абонементов для продления", a.menu())
	}

	var text string
	text += "🔄 ВЫБЕРИ КЛИЕНТА ДЛЯ ПРОДЛЕНИЯ\n\n"

	for i, cl := range clients {
		if cl.ExpireDate == nil {
			continue
		}

		text += fmt.Sprintf(
			"%d. %s — истекает %s\n",
			i+1,
			cl.Name,
			cl.ExpireDate.Format("02.01.2006"),
		)
	}

	return c.Send(text, a.backMenu())
}

func (a *App) saveClient(c tele.Context, s *Session) error {
	err := storage.AddClient(a.DB, s.Name, s.Type, s.PurchaseDate)
	if err != nil {
		return c.Send("Ошибка при сохранении клиента: "+err.Error(), a.menu())
	}

	err = storage.AddPayment(a.DB, s.Name, s.Type, s.Amount)
	if err != nil {
		return c.Send("Ошибка при сохранении оплаты: "+err.Error(), a.menu())
	}

	var expire *time.Time
	if s.Type == "monthly" {
		e := s.PurchaseDate.AddDate(0, 0, 30)
		expire = &e
	}

	msg := formatClientCreated(
		s.Name,
		s.Type,
		s.PurchaseDate,
		s.Amount,
		expire,
	)

	a.clearSession(c.Chat().ID)
	return c.Send(msg, a.menu())
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

	s.PurchaseDate = time.Now()
	s.Step = "amount"

	return c.Send("Выберите сумму оплаты", a.amountMenu())
}

func (a *App) sendList(c tele.Context, filter string) error {
	clients, err := storage.ListClients(a.DB, filter)
	if err != nil {
		return c.Send("Ошибка при получении списка", a.menu())
	}

	if len(clients) == 0 {
		return c.Send("Список пуст", a.menu())
	}

	list := make([]ClientView, 0, len(clients))
	for _, cl := range clients {
		list = append(list, ClientView{
			Name:         cl.Name,
			Type:         cl.Type,
			PurchaseDate: cl.PurchaseDate,
			ExpireDate:   cl.ExpireDate,
		})
	}

	return c.Send(formatClientList(list), a.menu())
}

func (a *App) sendExpiringSoon(c tele.Context) error {
	clients, err := storage.ExpiringSoon(a.DB)
	if err != nil {
		return c.Send("Ошибка при получении данных", a.menu())
	}

	if len(clients) == 0 {
		return c.Send("В ближайшие 7 дней ничего не истекает", a.menu())
	}

	list := make([]ClientView, 0, len(clients))
	for _, cl := range clients {
		list = append(list, ClientView{
			Name:         cl.Name,
			Type:         cl.Type,
			PurchaseDate: cl.PurchaseDate,
			ExpireDate:   cl.ExpireDate,
		})
	}

	return c.Send(formatExpiringSoon(list), a.menu())
}

func (a *App) sendStats(c tele.Context) error {
	stats, err := storage.GetStats(a.DB)
	if err != nil {
		return c.Send("Ошибка при получении статистики", a.menu())
	}

	msg := formatStats(
		stats.Total,
		stats.Monthly,
		stats.Single,
		stats.ExpiringSoon,
		stats.Expired,
		stats.TotalMoney,
		stats.MonthlyMoney,
		stats.SingleMoney,
	)

	return c.Send(msg, a.menu())
}
