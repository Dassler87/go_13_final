# Сборка на Go 1.25
FROM golang:1.25 AS builder

# Внутри контейнера создаём папку /build и выполняем оттуда все команды
WORKDIR /build

# Копируем зависимости и скачиваем модули
COPY go.mod go.sum ./
RUN go mod download

# Копирование проекта в контейнер
COPY . .

# Компиляция проекта, отключаем CGO
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o scheduler .

# Образ для запуска
FROM ubuntu:latest

WORKDIR /app

# Копируем готовый бинарник из стадии builder
COPY --from=builder /build/scheduler .

# Копируем web 
COPY web ./web

# Переменные окружения по умолчанию
ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/data/scheduler.db

# Порт, который будет использовать приложение
EXPOSE 7540

# Запуск приложения
CMD ["./scheduler"]
