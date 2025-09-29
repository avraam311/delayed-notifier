FROM golang:alpine AS build_base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o worker ./cmd/app/main.go

FROM alpine AS runner

COPY --from=build_base /app/worker .
COPY config ./config

CMD ["./worker"]