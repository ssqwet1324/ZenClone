FROM golang:1.24.5-alpine

WORKDIR /app

# Копируем модули и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник
RUN go build -o main ./main.go

EXPOSE 8082

CMD ["./main"]