.PHONY: build run test lint clean

build:
	go build -o bin/tmux-tui .

run:
	go run .

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
