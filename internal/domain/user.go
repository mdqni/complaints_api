package domain

type User struct {
	ID     int    `json:"id"`
	UserID string `json:"user_id/telegram_id"` // Telegram user ID
}
