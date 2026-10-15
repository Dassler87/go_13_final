# Сборка на Go 1.26
FROM golang:1.26 AS builder

# Внутри контейнера создаём папку  и выполняем оттуда все команды
WORKDIR /app

# Копируем зависимости и скачиваем модули
COPY go.mod go.sum ./
RUN go mod download

# Копирование проекта в контейнер
COPY . .

# Компиляция проекта, отключаем CGO
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o scheduler .

# Образ для запуска
FROM alpine:3.22

WORKDIR /app

# Создаём папку под БД
RUN mkdir -p /app/data

# Копируем бинарник
COPY --from=builder /app/scheduler .

# Копируем web 
COPY web ./web

# Переменные окружения по умолчанию
ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/data/scheduler.db

# Запуск приложения
CMD ["./scheduler"]
