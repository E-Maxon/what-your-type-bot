# Устанавливаем базовый образ для Go
FROM golang:1.19-alpine as builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем все файлы в контейнер
COPY . .

# Загружаем зависимости и собираем приложение
RUN go mod tidy
RUN go build -o what-your-type-bot .

# Второй этап для минимизации образа
FROM alpine:latest

WORKDIR /root/

# Копируем собранное приложение
COPY --from=builder /app/what-your-type-bot .

# Команда для запуска приложения
CMD ["./what-your-type-bot"]
