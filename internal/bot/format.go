package bot

import (
	"fmt"
	"strings"
	"time"
)

type ClientView struct {
	Name         string
	Type         string
	PurchaseDate time.Time
	ExpireDate   *time.Time
}

const sep = "━━━━━━━━━━━━━━━"

// ==================== STATS ====================

func formatStats(
	total, monthly, single int,
	expiringSoon, expired int,
	totalMoney, monthlyMoney, singleMoney int,
) string {

	return fmt.Sprintf(`📊 СТАТИСТИКА

%s

👥 КЛИЕНТЫ
• Всего: %d
• Месячные: %d
• Разовые: %d

%s

📅 СТАТУС
• Скоро истекают: %d
• Истекли: %d

%s

💰 ДОХОД
• Всего: %d лей
• Месячные: %d лей
• Разовые: %d лей

%s`,
		sep,
		total, monthly, single,
		sep,
		expiringSoon, expired,
		sep,
		totalMoney, monthlyMoney, singleMoney,
		sep,
	)
}

// ==================== CLIENT CREATED ====================

func formatClientCreated(name, subType string, purchaseDate time.Time, amount int, expire *time.Time) string {
	expireStr := "—"
	if expire != nil {
		expireStr = expire.Format("02.01.2006")
	}

	return fmt.Sprintf(`✅ КЛИЕНТ ДОБАВЛЕН

%s

👤 ИМЯ
• %s

📦 ТИП
• %s

📅 ДАТА
• %s

💰 ОПЛАТА
• %d лей

⏳ ИСТЕКАЕТ
• %s

%s`,
		sep,
		name,
		mapType(subType),
		purchaseDate.Format("02.01.2006"),
		amount,
		expireStr,
		sep,
	)
}

// ==================== EXTENDED ====================

func formatExtended(name string, newExpire time.Time, amount int) string {
	return fmt.Sprintf(`🔄 АБОНЕМЕНТ ПРОДЛЁН

%s

👤 КЛИЕНТ
• %s

📅 НОВАЯ ДАТА
• %s

💰 ОПЛАТА
• %d лей

%s`,
		sep,
		name,
		newExpire.Format("02.01.2006"),
		amount,
		sep,
	)
}

// ==================== DELETED ====================

func formatDeleted(name string) string {
	return fmt.Sprintf(`🗑 КЛИЕНТ УДАЛЁН

%s

👤 КЛИЕНТ
• %s

%s`,
		sep,
		name,
		sep,
	)
}

// ==================== EXPIRING SOON ====================

func formatExpiringSoon(clients []ClientView) string {
	if len(clients) == 0 {
		return "⏳ Нет клиентов с истекающим сроком"
	}

	var b strings.Builder

	b.WriteString("⏳ СКОРО ИСТЕКАЮТ\n\n")

	for _, cl := range clients {
		if cl.ExpireDate == nil {
			continue
		}

		days := int(cl.ExpireDate.Sub(time.Now()).Hours() / 24)
		if days < 0 {
			days = 0
		}

		b.WriteString(sep + "\n\n")
		b.WriteString(fmt.Sprintf("👤 %s\n", cl.Name))
		b.WriteString(fmt.Sprintf("📅 Истекает: %s\n", cl.ExpireDate.Format("02.01.2006")))
		b.WriteString(fmt.Sprintf("⏳ Осталось: %d дней\n\n", days))
	}

	b.WriteString(sep)

	return b.String()
}

// ==================== CLIENT LIST ====================

func formatClientList(clients []ClientView) string {
	if len(clients) == 0 {
		return "📋 Список пуст"
	}

	var b strings.Builder

	b.WriteString("📋 СПИСОК КЛИЕНТОВ\n\n")

	for i, cl := range clients {
		b.WriteString(sep + "\n\n")

		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, cl.Name))
		b.WriteString(fmt.Sprintf("📦 Тип: %s\n", mapType(cl.Type)))
		b.WriteString(fmt.Sprintf("📅 Купил: %s\n", cl.PurchaseDate.Format("02.01.2006")))

		if cl.ExpireDate != nil {
			b.WriteString(fmt.Sprintf("⏳ Истекает: %s\n", cl.ExpireDate.Format("02.01.2006")))
		} else {
			b.WriteString("⏳ Истекает: —\n")
		}

		b.WriteString("\n")
	}

	return b.String()
}
