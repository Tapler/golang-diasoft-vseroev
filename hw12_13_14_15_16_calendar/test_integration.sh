#!/bin/bash

set -e

echo "=== Тест интеграции Kafka ==="
echo ""

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Проверка что мы в правильной директории
if [ ! -f "./bin/calendar" ]; then
    echo -e "${RED}Ошибка: бинарные файлы не найдены. Выполните 'make build'${NC}"
    exit 1
fi

# Очистка базы данных
echo -e "${YELLOW}Очистка базы данных...${NC}"
psql -h localhost -U calendar -d calendar -c "TRUNCATE events, notifications;" > /dev/null 2>&1

# Проверка Calendar API
echo -e "${YELLOW}Проверка Calendar API на порту 8080...${NC}"
if ! curl -s http://localhost:8080/api/v1/events > /dev/null 2>&1; then
    echo -e "${RED}Calendar API не запущен!${NC}"
    echo "Запустите в отдельном терминале:"
    echo "  ./bin/calendar -config ./configs/config.toml"
    exit 1
fi
echo -e "${GREEN}✓ Calendar API работает${NC}"

# Создание события
echo -e "${YELLOW}Создание тестового события...${NC}"
START_TIME=$(date -u -v+3M '+%Y-%m-%dT%H:%M:%SZ')
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d "{
    \"title\": \"Тестовое событие с уведомлением\",
    \"start_time\": \"$START_TIME\",
    \"duration\": 3600,
    \"description\": \"Проверка работы уведомлений\",
    \"user_id\": 1,
    \"notify_before\": 120
  }")

if echo "$RESPONSE" | grep -q "error"; then
    echo -e "${RED}Ошибка создания события: $RESPONSE${NC}"
    exit 1
fi

EVENT_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo -e "${GREEN}✓ Событие создано с ID: $EVENT_ID${NC}"

# Проверка в БД
echo -e "${YELLOW}Проверка события в БД...${NC}"
DB_CHECK=$(psql -h localhost -U calendar -d calendar -t -c "SELECT COUNT(*) FROM events WHERE id='$EVENT_ID';")
if [ "$DB_CHECK" -eq 1 ]; then
    echo -e "${GREEN}✓ Событие найдено в БД${NC}"
else
    echo -e "${RED}✗ Событие не найдено в БД${NC}"
    exit 1
fi

# Вывод информации о событии
echo ""
echo -e "${YELLOW}Информация о событии:${NC}"
psql -h localhost -U calendar -d calendar -c "
SELECT 
    id, 
    title, 
    start_time, 
    notify_before,
    (start_time - notify_before * INTERVAL '1 second') as notify_time,
    NOW() AT TIME ZONE 'UTC' as current_utc
FROM events 
WHERE id='$EVENT_ID';"

echo ""
echo -e "${GREEN}=== Событие успешно создано ===${NC}"
echo ""
echo -e "${YELLOW}Ожидайте ~1 минуту...${NC}"
echo "Scheduler должен обнаружить событие и отправить уведомление в Kafka"
echo "Storer должен получить уведомление и сохранить его в БД"
echo ""
echo "Проверьте логи:"
echo "  - Терминал 2 (Scheduler): должно появиться 'Found 1 events to notify'"
echo "  - Терминал 3 (Storer): должно появиться 'Received notification'"
echo ""
echo "Через 1-2 минуты проверьте уведомления:"
echo "  psql -h localhost -U calendar -d calendar -c 'SELECT * FROM notifications;'"
