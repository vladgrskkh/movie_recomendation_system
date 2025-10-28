# ---------- Stage 1: Build ----------
FROM golang:1.25 AS builder

WORKDIR /app

ARG LINKER_FLAGS=""

# Копируем go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем бинарь
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="${LINKER_FLAGS}" -o server ./cmd/api

# ---------- Stage 2: Runtime ----------
FROM alpine:3.20

# Добавим сертификаты (для HTTPS-запросов, SMTP и т.п.)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /app/server .

EXPOSE 8080

ENTRYPOINT ["./server"]