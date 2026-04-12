BINARY_NAME=gestic
INSTALL_DIR=$(HOME)/.local/bin

VERSION    := $(shell git describe --tags --always --dirty)
COMMIT     := $(shell git rev-parse HEAD)
LDFLAGS    = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)"

GO=go

.DEFAULT_GOAL := build


build:
	@echo "==> Building $(BINARY_NAME)..."
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) .

install: build
	@echo "==> Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@install -d "$(INSTALL_DIR)"
	@install "$(BINARY_NAME)" "$(INSTALL_DIR)"
	@echo "Installed successfully!"

clean: 
	@echo "==> Removing binary..."
	@rm -f "$(BINARY_NAME)"

manpage: build
	@echo "==> Generating manpage..."
	@mkdir -p docs
	@help2man --no-info --name="A diff tool for restic snapshots" ./$(BINARY_NAME) > docs/$(BINARY_NAME).1
	@gzip -f -9 docs/$(BINARY_NAME).1