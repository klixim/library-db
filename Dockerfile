# Этап 1: Сборка
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем файлы модулей и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код и собираем бинарник
COPY . .
RUN go build -o main .

# Этап 2: Финальный легковесный образ
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/static ./static

EXPOSE 8080
CMD ["./main"]