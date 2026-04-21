package storage

import "time"

// Notification представляет уведомление о событии.
type Notification struct {
	ID        string    // Уникальный идентификатор уведомления
	EventID   string    // ID события
	Title     string    // Заголовок события
	EventTime time.Time // Дата и время события
	UserID    int64     // Пользователь, которому отправлять
}
