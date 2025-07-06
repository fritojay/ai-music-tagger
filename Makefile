BINARY_NAME=ai-music-tagger

build: lint
	go build -o ${BINARY_NAME} main.go

lint:
	golangci-lint run

run:
	go run main.go

clean:
	go clean
