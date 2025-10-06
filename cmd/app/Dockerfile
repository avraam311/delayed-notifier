FROM golang:alpine AS build_base

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o app cmd/app/main.go

FROM alpine AS runner

COPY --from=build_base app/app .
COPY config config

CMD ["app"]