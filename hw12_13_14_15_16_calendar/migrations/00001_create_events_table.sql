-- +goose Up
-- Создание таблицы событий календаря
CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    duration BIGINT NOT NULL, -- Длительность в секундах
    description TEXT,
    user_id BIGINT NOT NULL,
    notify_before BIGINT -- Время уведомления в секундах (за сколько до события)
);

-- Индекс для быстрого поиска событий пользователя
CREATE INDEX idx_events_user_id ON events(user_id);

-- Индекс для быстрого поиска событий по времени
CREATE INDEX idx_events_start_time ON events(start_time);

-- Составной индекс для проверки занятости времени
CREATE INDEX idx_events_user_time ON events(user_id, start_time);

-- +goose Down
DROP TABLE IF EXISTS events;
