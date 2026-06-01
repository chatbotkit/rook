VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  = -s -w -X github.com/chatbotkit/rook/internal/version.Version=$(VERSION)

CMD      = rook
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build run test vet fmt lint clean dist workspace

all: build

# Create the gitignored go.work so local builds resolve the go-sdk from a local
# checkout instead of the published module pinned in go.mod. Override the path
# with: make workspace GOSDK=../path/to/go-sdk
GOSDK ?= ../go-sdk

workspace:
	go work init . $(GOSDK)
	@echo "go.work created - local builds now use $(GOSDK)"

build:
	@echo "Building $(CMD) ($(VERSION))..."
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(CMD) ./cmd/$(CMD)

run: build
	./$(CMD) $(ARGS)

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./... -count=1

lint: vet
	@echo "lint ok"

clean:
	rm -f $(CMD)
	rm -rf dist

# Cross-compile a single platform: make cross GOOS=darwin GOARCH=arm64
cross:
	@echo "Building $(CMD) ($(VERSION)) for $(GOOS)/$(GOARCH)..."
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags "$(LDFLAGS)" -o $(CMD) ./cmd/$(CMD)

# Build release archives for every target platform under dist/.
dist: clean
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		ext=""; [ "$$os" = "windows" ] && ext=".exe"; \
		out="dist/$(CMD)-$(VERSION)-$$os-$$arch"; \
		echo "Building $$out..."; \
		mkdir -p "$$out"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -trimpath -ldflags "$(LDFLAGS)" -o "$$out/$(CMD)$$ext" ./cmd/$(CMD); \
		cp README.md LICENSE "$$out/"; \
		tar -czf "dist/$(CMD)-$(VERSION)-$$os-$$arch.tar.gz" -C dist "$(CMD)-$(VERSION)-$$os-$$arch"; \
	done
	@cd dist && sha256sum *.tar.gz > checksums.txt
	@echo "Release archives written to dist/"
