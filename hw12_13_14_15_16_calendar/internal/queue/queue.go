package queue

import "context"

// Producer определяет интерфейс для отправки сообщений в очередь.
type Producer interface {
	// SendMessage отправляет сообщение в указанный топик.
	SendMessage(ctx context.Context, topic string, key, value []byte) error

	// Close закрывает соединение с брокером.
	Close() error
}

// Consumer определяет интерфейс для получения сообщений из очереди.
type Consumer interface {
	// Consume начинает потребление сообщений из указанных топиков.
	Consume(ctx context.Context, topics []string, handler MessageHandler) error

	// Close закрывает соединение с брокером.
	Close() error
}

// MessageHandler обрабатывает полученное сообщение.
type MessageHandler func(ctx context.Context, message []byte) error
