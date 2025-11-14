FROM golang:1.25.3-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o ./avito-intership-2025 ./cmd/server

CMD ["./avito-intership-2025"]
