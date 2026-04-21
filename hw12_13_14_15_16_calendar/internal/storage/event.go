package storage

import "time"

// Event представляет событие в календаре.
type Event struct {
	ID           string        // Уникальный идентификатор события (UUID)
	Title        string        // Короткий текст заголовка
	StartTime    time.Time     // Дата и время события
	Duration     time.Duration // Длительность события, в сек
	Description  string        // Длинный текст описания (опционально)
	UserID       int64         // ID пользователя, владельца события
	NotifyBefore time.Duration // За сколько времени высылать уведомление, в сек (опционально)
}
