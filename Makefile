.PHONY: build run test tidy lint docker-up docker-down clean

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

test:
	go test -v -coverprofile=coverage.out ./...

tidy:
	go mod tidy

lint:
	golangci-lint run

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

clean:
	rm -rf bin coverage.out
	rm -f data/database.sqlite

deps:
	go mod download

fmt:
	go fmt ./...

vet:
	go vet ./...