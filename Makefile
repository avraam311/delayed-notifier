.PHONY: test, lint, up, down

test:
	go test -v ./...

lint:
	go vet ./... 
	golangci-lint run ./...

up:
	docker compose up --build

down:
	docker compose down -v