BINARY=gitback

build:
	go build -o $(BINARY) ./cmd/gitback

run:
	go run ./cmd/gitback

test:
	go test ./...

clean:
	rm -f $(BINARY)
