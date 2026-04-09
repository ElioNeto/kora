.PHONY: help dev desktop compiler apk clean build

# Variaveis
APP_NAME := kora
EDITOR_PORT := 8080
DESKTOP_PORT := 5173

## Help: Show available commands
help:
	@echo "Kora Engine - Available Commands"
	@echo "=================================="
	@echo ""
	@echo "Development:"
	@echo "  make dev          - Run editor in development mode"
	@echo "  make desktop      - Start desktop app (Electron)"
	@echo "  make compiler     - Build KScript compiler"
	@echo "  make run          - Run sample game"
	@echo ""
	@echo "Build:"
	@echo "  make build        - Build all artifacts"
	@echo "  make apk          - Build Android APK"
	@echo "  make dist         - Create distribution package"
	@echo ""
	@echo "Tests:"
	@echo "  make test         - Run all tests"
	@echo "  make test-editor  - Run editor tests"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make dist-clean   - Remove dist directories"
	@echo ""

## dev: Run editor in development mode
dev:
	cd editor && python -m http.server $(EDITOR_PORT)

## desktop: Start Electron desktop app
desktop:
	cd apps/desktop && npm run dev

## compiler: Build KScript compiler
compiler:
	cd compiler && go build -o ../bin/kora-compiler .

## run: Run sample game
run:
	go run main.go examples/

## test: Run all tests
test:
	go test ./...
	cd compiler && go test ./...
	@if [ -f "editor/serializer.test.js" ]; then \
		node editor/serializer.test.js; \
	fi

## test-editor: Run editor tests
test-editor:
	@if [ -f "editor/serializer.test.js" ]; then \
		node editor/serializer.test.js; \
	else \
		echo "No editor tests found"; \
	fi

## apk: Build Android APK
apk:
	@if [ ! -f "android/build.sh" ]; then \
		echo "Error: android/build.sh not found"; \
		exit 1; \
	fi
	cd android && ./build.sh release

## build: Build all artifacts
build: compiler apk

## dist: Create distribution package
dist:
	@echo "Creating distribution package..."
	mkdir -p dist
	cp -r bin/ dist/
	cp -r examples/ dist/
	cp README.md dist/
	cp LICENSE dist/
	tar -czf dist-$(shell date +%Y%m%d).tar.gz dist

## clean: Remove build artifacts
clean:
	rm -rf bin/
	rm -rf dist/
	find . -name "*.apk" -delete
	find . -name "*.ks" -type f ! -path "./examples/*" -delete

## dist-clean: Remove dist directories
dist-clean:
	rm -rf apps/desktop/dist/
	rm -rf apps/desktop/dist-electron/

## install-deps: Install development dependencies
install-deps:
	@echo "Installing dependencies..."
	cd compiler && go mod download
	cd apps/desktop && npm install

## setup: Full setup
setup: install-deps clean
	@echo "Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  make dev       - Run web editor"
	@echo "  make desktop   - Run desktop app"
	@echo "  make apk       - Build Android APK"

## docs: Preview documentation
docs:
	python -m http.server 3000 --directory docs

## release: Create release
release: clean build dist

## lint: Run linters
lint:
	@gofmt -l . | grep -v ".pb.go" | grep -v "vendor"
	@if [ -f "apps/desktop/package.json" ]; then \
		cd apps/desktop && npm run lint 2>/dev/null || true; \
	fi

## format: Format code
format:
	gofmt -w .
	@if [ -f "apps/desktop/package.json" ]; then \
		cd apps/desktop && npm run format 2>/dev/null || true; \
	fi

## version: Show version
version:
	@echo "@$(shell date +%Y.%m.%d)"
