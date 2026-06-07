BINARY=gitback
RELEASE_DIR := /tmp/gitback/gitback-release


build:
	go build -o /tmp/gitback/$(BINARY) ./cmd/gitback

run:
	go run ./cmd/gitback

test:
	go test ./...

clean:
	rm -f /tmp/gitback/$(BINARY)

release:
	@mkdir -p $(RELEASE_DIR)

	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" \
		-o $(RELEASE_DIR)/$(BINARY)-linux-amd64 ./cmd/gitback
	GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" \
		-o $(RELEASE_DIR)/$(BINARY)-linux-arm64 ./cmd/gitback

	cd $(RELEASE_DIR) && \
	sha256sum \
		$(BINARY)-linux-amd64 \
		$(BINARY)-linux-arm64 \
		> $(BINARY)-v0.2.0-checksums.txt
