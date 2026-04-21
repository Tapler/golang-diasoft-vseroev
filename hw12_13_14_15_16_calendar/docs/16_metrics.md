## Обзор метрик

Мониторинг сервиса календаря с использованием Prometheus. Добавлены метрики для отслеживания ключевых показателей работы всех компонентов системы: HTTP API, бизнес-логики, scheduler и storer.

---

## Реализованные метрики

### 1. HTTP метрики (RED метрики)

#### `calendar_http_requests_total`
- **Тип**: Counter
- **Описание**: Общее количество HTTP запросов
- **Labels**: 
  - `method` - HTTP метод (GET, POST, PUT, DELETE)
  - `endpoint` - путь эндпоинта
  - `status` - HTTP статус код
- **Назначение**: Отслеживание Rate (частоты запросов) и Errors (количества ошибок)

#### `calendar_http_request_duration_seconds`
- **Описание**: Время обработки HTTP запросов в секундах
- **Labels**: 
  - `method` - HTTP метод
  - `endpoint` - путь эндпоинта
- **Назначение**: Отслеживание Duration (латентности запросов)
- **Buckets**: Используются стандартные бакеты Prometheus (0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10)

### 2. Бизнес-метрики событий календаря

#### `calendar_events_created_total`
- **Тип**: Counter
- **Описание**: Общее количество созданных событий
- **Назначение**: Отслеживание активности создания событий

#### `calendar_events_updated_total`
- **Тип**: Counter
- **Описание**: Общее количество обновленных событий
- **Назначение**: Отслеживание активности изменения событий

#### `calendar_events_deleted_total`
- **Тип**: Counter
- **Описание**: Общее количество удаленных событий
- **Назначение**: Отслеживание активности удаления событий

#### `calendar_events_retrieved_total`
- **Тип**: Counter
- **Описание**: Общее количество полученных событий (по ID)
- **Назначение**: Отслеживание запросов на получение отдельных событий

#### `calendar_events_listed_total`
- **Тип**: Counter
- **Описание**: Общее количество запросов на получение списка событий
- **Labels**: 
  - `period` - период (day, week, month)
- **Назначение**: Отслеживание запросов на получение списков событий по периодам

### 3. Метрики Scheduler (фоновые задачи)

#### `calendar_notifications_sent_total`
- **Тип**: Counter
- **Описание**: Общее количество отправленных уведомлений в очередь
- **Назначение**: Отслеживание успешной отправки уведомлений

#### `calendar_notifications_send_errors_total`
- **Тип**: Counter
- **Описание**: Общее количество ошибок при отправке уведомлений
- **Назначение**: Отслеживание проблем с отправкой уведомлений

#### `calendar_events_scanned_total`
- **Тип**: Counter
- **Описание**: Общее количество сканирований событий для уведомлений
- **Назначение**: Отслеживание работы планировщика

#### `calendar_old_events_deleted_total`
- **Тип**: Counter
- **Описание**: Общее количество операций очистки старых событий
- **Назначение**: Отслеживание работы процесса очистки

### 4. Метрики Storer (обработка уведомлений)

#### `calendar_notifications_received_total`
- **Тип**: Counter
- **Описание**: Общее количество полученных уведомлений из очереди
- **Назначение**: Отслеживание получения уведомлений

#### `calendar_notifications_saved_total`
- **Тип**: Counter
- **Описание**: Общее количество сохраненных уведомлений в БД
- **Назначение**: Отслеживание успешного сохранения уведомлений

#### `calendar_notifications_save_errors_total`
- **Тип**: Counter
- **Описание**: Общее количество ошибок при сохранении уведомлений
- **Назначение**: Отслеживание проблем с сохранением уведомлений

### 5. Метрики ошибок

#### `calendar_errors_total`
- **Тип**: Counter
- **Описание**: Общее количество ошибок по типам
- **Labels**: 
  - `type` - тип операции (create_event, update_event, delete_event, get_event, list_events_day, list_events_week, list_events_month)
- **Назначение**: Детальное отслеживание ошибок в бизнес-логике

---

### Эндпоинт /metrics

- **URL**: `http://localhost:8080/metrics`
- **Формат**: Prometheus text format
- **Содержимое**: Все метрики приложения + стандартные метрики Go runtime
- **Использование**: Prometheus scraper может собирать метрики с этого эндпоинта

---


## Использование метрик для анализа

### 1. Мониторинг производительности

**Запросы в секунду (RPS)**:
```promql
rate(calendar_http_requests_total[5m])
```

**Процент ошибок**:
```promql
sum(rate(calendar_http_requests_total{status=~"5.."}[5m])) 
/ 
sum(rate(calendar_http_requests_total[5m])) * 100
```

### 2. Бизнес-метрики

**Активность пользователей**:
```promql
sum(rate(calendar_events_created_total[1h]))
```

**Соотношение операций**:
```promql
calendar_events_created_total / calendar_events_deleted_total
```

### 3. Мониторинг фоновых задач

**Успешность отправки уведомлений**:
```promql
calendar_notifications_sent_total 
/ 
(calendar_notifications_sent_total + calendar_notifications_send_errors_total) * 100
```

**Скорость обработки уведомлений**:
```promql
rate(calendar_notifications_saved_total[5m])
```

### 4. Выявление узких мест

**Самые медленные эндпоинты**:
```promql
topk(5, 
  histogram_quantile(0.95, 
    rate(calendar_http_request_duration_seconds_bucket[5m])
  ) by (endpoint)
)
```

**Эндпоинты с наибольшим количеством ошибок**:
```promql
topk(5, 
  sum(rate(calendar_http_requests_total{status=~"5.."}[5m])) by (endpoint)
)
```

---