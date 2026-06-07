BINARY := gitback
VERSION ?= dev

BUILD_DIR := /tmp/gitback
RELEASE_DIR := $(BUILD_DIR)/gitback-release

.PHONY: build run test clean release

build:
	@mkdir -p $(BUILD_DIR)

	go build \
		-trimpath \
		-ldflags="-s -w -X github.com/flarexes/gitback/internal/version.Version=$(VERSION)" \
		-o $(BUILD_DIR)/$(BINARY) \
		./cmd/gitback

	@echo
	@echo "Built binary:"
	@echo "  $(BUILD_DIR)/$(BINARY)"

run:
	go run ./cmd/gitback

test:
	go test ./...

clean:
	rm -f $(BUILD_DIR)/$(BINARY)
	rm -rf $(RELEASE_DIR)

	@echo
	@echo "Cleaned up artifacts:"
	@echo "  $(BUILD_DIR)/$(BINARY)"
	@echo "  $(RELEASE_DIR)"

release:
	@mkdir -p $(RELEASE_DIR)

	GOOS=linux GOARCH=amd64 \
	go build \
		-trimpath \
		-ldflags="-s -w -X github.com/flarexes/gitback/internal/version.Version=$(VERSION)" \
		-o $(RELEASE_DIR)/$(BINARY)-linux-amd64 \
		./cmd/gitback

	GOOS=linux GOARCH=arm64 \
	go build \
		-trimpath \
		-ldflags="-s -w -X github.com/flarexes/gitback/internal/version.Version=$(VERSION)" \
		-o $(RELEASE_DIR)/$(BINARY)-linux-arm64 \
		./cmd/gitback

	cd $(RELEASE_DIR) && \
	sha256sum \
		$(BINARY)-linux-amd64 \
		$(BINARY)-linux-arm64 \
		> $(BINARY)-$(VERSION)-checksums.txt

	@echo
	@echo "Release artifacts:"
	@echo "  $(RELEASE_DIR)"
