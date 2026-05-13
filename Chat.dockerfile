FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copiar go.mod y go.sum desde la raíz
COPY go.mod go.sum ./
RUN go mod download

# Copiar todo el código
COPY . .

# Compilar desde cmd/api/main.go
WORKDIR /app/cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/main .

# Etapa final
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl

RUN adduser -D -g '' appuser

WORKDIR /app

# Copiar el binario compilado
COPY --from=builder /app/main .

# Crear directorio para uploads
RUN mkdir -p /app/uploads && chown -R appuser:appuser /app/uploads

USER appuser

# EXPOSE es solo documentación - Render usará la variable PORT
EXPOSE 8085

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:${PORT:-8085}/health || exit 1

CMD ["./main"]