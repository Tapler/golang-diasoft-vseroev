# Dockerfile для интеграционных тестов
FROM golang:1.23-alpine

RUN apk add --no-cache git ca-certificates curl

WORKDIR /tests

# Копирование зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование тестов
COPY . .

RUN chmod +x wait-and-test.sh

CMD ["./wait-and-test.sh"]
