PLUGIN_ID    := com.avap.avapmost-search
PLUGIN_VERSION := $(shell python3 -c "import sys,json; print(json.load(open('plugin.json'))['version'])")
DIST_FILE    := $(PLUGIN_ID)-$(PLUGIN_VERSION).tar.gz

.PHONY: all dist server webapp setup clean

all: dist

## dist: Build and package the plugin tar.gz
dist: server webapp
	@rm -f $(DIST_FILE)
	@rm -rf .dist_tmp/
	@mkdir -p .dist_tmp/$(PLUGIN_ID)/server/dist
	@mkdir -p .dist_tmp/$(PLUGIN_ID)/webapp/dist
	@cp plugin.json .dist_tmp/$(PLUGIN_ID)/
	@cp server/dist/plugin-linux-amd64 .dist_tmp/$(PLUGIN_ID)/server/dist/
	@cp webapp/dist/main.js .dist_tmp/$(PLUGIN_ID)/webapp/dist/
	@if [ -d public ]; then cp -r public .dist_tmp/$(PLUGIN_ID)/; fi
	@cd .dist_tmp && tar -czf ../$(DIST_FILE) $(PLUGIN_ID)
	@rm -rf .dist_tmp/
	@echo "Built: $(DIST_FILE)"

## server: Cross-compile the Go plugin server for Linux amd64 and arm64
server:
	@echo "==> Building Go server (linux/amd64)..."
	@mkdir -p server/dist
	@cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o dist/plugin-linux-amd64 .
	@echo "==> Building Go server (linux/arm64)..."
	@cd server && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o dist/plugin-linux-arm64 .

## webapp: Build the React webapp bundle
webapp:
	@echo "==> Building webapp..."
	@cd webapp && npm run build

## setup: Install all dependencies (run once before first build)
setup:
	@echo "==> Installing Go dependencies..."
	@cd server && go mod tidy && go mod download
	@echo "==> Installing npm dependencies..."
	@cd webapp && npm install

## clean: Remove build artifacts
clean:
	@rm -rf server/dist/ webapp/dist/ .dist_tmp/ *.tar.gz
	@echo "Cleaned."
