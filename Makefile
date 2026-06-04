BINARY=gitback

build:
	go build -o ./tmp/$(BINARY) ./cmd/gitback

run:
	go run ./cmd/gitback

test:
	go test ./...

clean:
	rm -f ./tmp/$(BINARY)
