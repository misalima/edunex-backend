# Stage 1: Build
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copia arquivos de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código
COPY . .

# Compila o binário (Ajuste o caminho se o seu main.go não estiver em cmd/app/main.go)
RUN go build -o edunex-api cmd/app/main.go

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copia o binário do builder
COPY --from=builder /app/edunex-api .

# Expõe a porta da API
EXPOSE 8080

CMD ["./edunex-api"]