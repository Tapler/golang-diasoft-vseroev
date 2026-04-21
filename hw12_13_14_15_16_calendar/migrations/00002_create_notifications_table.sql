-- +goose Up
-- Создание таблицы уведомлений
CREATE TABLE IF NOT EXISTS notifications (
    id VARCHAR(36) PRIMARY KEY,
    event_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    event_time TIMESTAMP NOT NULL,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска уведомлений пользователя
CREATE INDEX idx_notifications_user_id ON notifications(user_id);

-- Индекс для быстрого поиска уведомлений по событию
CREATE INDEX idx_notifications_event_id ON notifications(event_id);

-- Индекс для быстрого поиска уведомлений по времени события
CREATE INDEX idx_notifications_event_time ON notifications(event_time);

-- +goose Down
DROP TABLE IF EXISTS notifications;
