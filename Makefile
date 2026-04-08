BINARY_NAME=aim

all: build

# Build frontend then backend
build:
	@cd web && pnpm install && pnpm run build && cd ..
	go build -o $(BINARY_NAME) ./cmd/aim

# Build frontend only
frontend:
	cd web && pnpm run build

# Run backend only (requires frontend already built)
backend:
	go build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/aim
