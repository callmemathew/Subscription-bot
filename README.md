# CribBot

CribBot is a simple Telegram bot for managing client subscriptions.

Built with **Go** and **SQLite**, it helps track monthly and single subscriptions, view client lists, extend active plans, remove clients, and receive expiration reminders.

## Features

- Add a new client
- Choose subscription type:
  - monthly
  - single
- Set purchase date manually or use **Today**
- View full client list
- Filter clients by subscription type
- Show subscriptions that expire soon
- Delete clients with confirmation
- Extend monthly subscriptions from the current date
- View statistics
- Receive automatic reminders 7 days before expiration

## Tech Stack

- **Go**
- **SQLite**
- **Telebot v3**
- **godotenv**

Add Client
When adding a client, the bot asks for:

client name

subscription type

purchase date

The purchase date can be:

entered manually in DD.MM.YYYY format

selected using the Сегодня button

If the subscription type is monthly, the expiration date is automatically calculated as:

purchase date + 30 days
If the subscription type is single, the expiration date is not set.

Client List
The bot can show:

all clients

only monthly subscriptions

only single subscriptions

Each client card includes:

name

subscription type

purchase date

expiration date

Expiring Soon
Shows all monthly subscriptions that will expire within the next 7 days.

Delete Client
The bot shows a numbered list of clients and asks for confirmation before deletion.

Extend Subscription
Only monthly subscriptions can be extended.

When extended, the bot:

sets purchase date to today

sets expiration date to today + 30 days

resets reminder flag

Statistics
The bot shows:

total clients

monthly subscriptions

single subscriptions

expiring soon

expired

Notifications
The bot checks the database periodically and sends a reminder to the admin when a monthly subscription will expire in 7 days.

Example:

Reminder

Client: Ivan
Expires: 20.03.2026
Environment Variables
Create a .env file in the project root.

Example:

BOT_TOKEN=your_telegram_bot_token
ADMIN_CHAT_ID=your_telegram_chat_id
Run Locally
Clone the repository:

git clone git clone https://github.com/callmemathew/Subscription-bot.git
cd Subscription-bot
Install dependencies:

go mod tidy
Run the bot:

go run ./cmd/bot
