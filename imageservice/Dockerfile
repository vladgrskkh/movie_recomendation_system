FROM golang:1.25 AS builder

WORKDIR /app

ARG LINKER_FLAGS=""

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="${LINKER_FLAGS}" -o server ./cmd/api

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

ENTRYPOINT ["./server"]