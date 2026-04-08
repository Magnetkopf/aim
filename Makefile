BINARY_NAME=aim

# Frontend source files
WEB_SRC_FILES := $(shell find web/src -type f 2>/dev/null)
WEB_PUBLIC_FILES := $(shell find web/public -type f 2>/dev/null)
WEB_DEPS := web/package.json web/pnpm-lock.yaml $(WEB_SRC_FILES) $(WEB_PUBLIC_FILES)

.PHONY: all build backend frontend build-all clean

all: build

# Build frontend then backend
build: backend

# Build frontend only if source files changed
web/dist: $(WEB_DEPS)
	@cd web && pnpm install && pnpm run build

frontend: web/dist

# Run backend only (requires frontend already built)
backend: web/dist
	go build -o $(BINARY_NAME) ./cmd/aim

# Build all architectures
build-all: build-linux-386 build-linux-amd64 build-linux-arm build-linux-arm64 build-linux-riscv64

# Cross-compile for Linux (all architectures) - reuse built frontend
build-linux-%: web/dist
	GOOS=linux GOARCH=$* go build -o $(BINARY_NAME)-linux-$* ./cmd/aim

clean:
	rm -rf web/dist $(BINARY_NAME) $(BINARY_NAME)-linux-*


