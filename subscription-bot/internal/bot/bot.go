package bot

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"subscription-bot/subscription-bot/internal/storage"
	"sync"
	"time"

	tele "gopkg.in/telebot.v3"
)

type Session struct {
	Step     string
	Name     string
	Type     string
	ClientID int64
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

func (a *App) menu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}

	btnAdd := menu.Text("Добавить")
	btnList := menu.Text("Список")
	btnSoon := menu.Text("Скоро истекают")
	btnDelete := menu.Text("Удалить")
	btnExtend := menu.Text("Продлить")
	btnStats := menu.Text("Статистика")

	menu.Reply(
		menu.Row(btnAdd, btnList),
		menu.Row(btnSoon, btnDelete),
		menu.Row(btnExtend, btnStats),
	)

	return menu
}

func (a *App) typeMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}

	btnMonthly := menu.Text("Месячный")
	btnSingle := menu.Text("Разовый")
	btnBack := menu.Text("Назад")

	menu.Reply(
		menu.Row(btnMonthly, btnSingle),
		menu.Row(btnBack),
	)

	return menu
}

func (a *App) listMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}

	btnAll := menu.Text("Все")
	btnMonthly := menu.Text("Только месячные")
	btnSingle := menu.Text("Только разовые")
	btnBack := menu.Text("Назад")

	menu.Reply(
		menu.Row(btnAll),
		menu.Row(btnMonthly, btnSingle),
		menu.Row(btnBack),
	)

	return menu
}

func (a *App) dateMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}

	btnToday := menu.Text("Сегодня")
	btnBack := menu.Text("Назад")

	menu.Reply(
		menu.Row(btnToday),
		menu.Row(btnBack),
	)

	return menu
}

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
func (a *App) sendDeleteList(c tele.Context) error {

	clients, err := storage.ListClients(a.DB, "")

	if err != nil {
		return c.Send("Ошибка при получении списка")
	}

	if len(clients) == 0 {
		return c.Send("Список пуст", a.menu())
	}

	var b strings.Builder

	b.WriteString("Выберите номер клиента для удаления\n\n")

	for i, cl := range clients {

		b.WriteString(fmt.Sprintf(
			"%d. %s — %s\n",
			i+1,
			cl.Name,
			mapType(cl.Type),
		))
	}

	return c.Send(b.String())
}

func (a *App) backMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	btnBack := menu.Text("Назад")
	menu.Reply(menu.Row(btnBack))
	return menu
}

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
			fmt.Sprintf(
				"Абонемент продлён\n\nКлиент: %s\nНовая дата окончания: %s",
				client.Name,
				newExpire,
			),
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

		return c.Send(
			fmt.Sprintf("Удалить клиента %s?", client.Name),
			a.confirmDeleteMenu())
	case "delete_confirm":
		if text == "Да" {
			err := storage.DeleteClient(a.DB, s.ClientID)
			if err != nil {
				return c.Send("Ошибка удаления", a.menu())
			}

			a.clearSession(c.Chat().ID)

			return c.Send(
				fmt.Sprintf("Клиент %s удалён", s.Name),
				a.menu(),
			)
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

func mapType(t string) string {
	if t == "monthly" {
		return "месячный"
	}
	return "разовый"
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
		b.WriteString(fmt.Sprintf(
			"%d. %s — истекает %s\n",
			i+1,
			cl.Name,
			cl.ExpireDate.Format("02.01.2006"),
		))
	}

	return c.Send(b.String(), a.backMenu())
}
func (a *App) confirmDeleteMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}

	btnYes := menu.Text("Да")
	btnNo := menu.Text("Нет")

	menu.Reply(
		menu.Row(btnYes, btnNo),
	)

	return menu
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
