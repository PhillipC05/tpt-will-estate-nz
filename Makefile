.PHONY: dev stop test test-race lint vet build migrate clean tools

APP_NAME := tpt-will-estate-nz

dev:
	docker compose up -d

stop:
	docker compose down

test:
	go test ./...

test-race:
	go test -race ./...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

build:
	go build -o bin/$(APP_NAME) ./cmd/server

migrate:
	atlas schema apply --dir "file://migrations" --url "$(DATABASE_URL)" --auto-approve

clean:
	rm -rf bin/ tmp/

tools:
	go install github.com/air-verse/air@latest
	go install ariga.io/atlas/cmd/atlas@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
