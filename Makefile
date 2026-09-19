BINARY := gitback
VERSION ?= dev

BUILD_DIR := /tmp/gitback
RELEASE_DIR := $(BUILD_DIR)/gitback-release

LDFLAGS := -s -w -X github.com/flarexes/gitback/internal/version.Version=$(VERSION)

.PHONY: build run test test-verbose test-cover clean release

build:
	@mkdir -p $(BUILD_DIR)

	go build \
		-trimpath \
		-ldflags="$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY) \
		./cmd/gitback

	@echo
	@echo "Built binary:"
	@echo "  $(BUILD_DIR)/$(BINARY)"

run:
	go run ./cmd/gitback

# Runs every *_test.go file in the module. -race catches data races —
# cheap to run and directly relevant here given mirror's worker pool
# and logging's shared mutex-guarded encoder.
test:
	go test -race ./...

# Same as above, but prints every test name as it runs (-v) instead of
# only failures — useful when narrowing down which specific case in a
# table-driven test is failing.
test-verbose:
	go test -race -v ./...

# Generates a coverage profile and prints a human-readable summary.
# Run `go tool cover -html=$(BUILD_DIR)/coverage.out` afterward to see
# an annotated, line-by-line view in your browser.
test-cover:
	@mkdir -p $(BUILD_DIR)
	go test -race -coverprofile=$(BUILD_DIR)/coverage.out ./...
	go tool cover -func=$(BUILD_DIR)/coverage.out

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
		-ldflags="$(LDFLAGS)" \
		-o $(RELEASE_DIR)/$(BINARY)-linux-amd64 \
		./cmd/gitback

	GOOS=linux GOARCH=arm64 \
	go build \
		-trimpath \
		-ldflags="$(LDFLAGS)" \
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
