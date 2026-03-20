package bot

import tele "gopkg.in/telebot.v3"

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
func (a *App) amountMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}

	btnDefault := menu.Text("Стандартная")
	btnCustom := menu.Text("Своя")
	btnBack := menu.Text("Назад")

	menu.Reply(
		menu.Row(btnDefault, btnCustom),
		menu.Row(btnBack),
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

func (a *App) backMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	btnBack := menu.Text("Назад")
	menu.Reply(menu.Row(btnBack))
	return menu
}

func (a *App) confirmDeleteMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{ResizeKeyboard: true}

	btnYes := menu.Text("Да")
	btnNo := menu.Text("Нет")

	menu.Reply(menu.Row(btnYes, btnNo))
	return menu
}

func mapType(t string) string {
	if t == "monthly" {
		return "месячный"
	}
	return "разовый"
}
