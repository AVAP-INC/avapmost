# Avapmost development Makefile
# Usage: make <target>

REPO_ROOT   := $(shell git rev-parse --show-toplevel)
SERVER_DIR  := $(REPO_ROOT)/server
WEBAPP_DIR  := $(REPO_ROOT)/webapp
PLUGIN_DIR  := $(REPO_ROOT)/plugin/avapmost-search

# Docker: exclude grafana (port 3000 conflicts with sourcebot)
DOCKER_SERVICES ?= postgres inbucket redis prometheus loki otel-collector minio

# Plugin signing
PLUGIN_ID      := com.avap.avapmost-search
PLUGIN_VERSION := $(shell python3 -c "import json; print(json.load(open('$(PLUGIN_DIR)/plugin.json'))['version'])" 2>/dev/null)
PLUGIN_TAR     := $(PLUGIN_DIR)/$(PLUGIN_ID)-$(PLUGIN_VERSION).tar.gz
PLUGIN_SIG_KEY := support@avap.jp
PREPACKAGED    := $(SERVER_DIR)/prepackaged_plugins

# Docker image
REGISTRY     := avap.plus/public
IMAGE_NAME   := avapmost
TAG          ?=
MM_PACKAGE   := dist/mattermost-team-linux-amd64.tar.gz

REMOTE_IMAGE  := $(REGISTRY)/$(IMAGE_NAME)
VERSIONS_FILE := $(REPO_ROOT)/VERSIONS.md

# Mattermost base version (first entry in server/public/model/version.go)
MM_VERSION := $(shell grep -m1 '"[0-9]\+\.[0-9]\+\.[0-9]\+"' $(SERVER_DIR)/public/model/version.go | tr -d '"\t ,')

.PHONY: help \
        dev-start dev-docker dev-server dev-webapp \
        plugin-build plugin-sign \
        image-build image-build-fast image-push _record-version \
        clean

# -----------------------------------------------------------------------
# Help
# -----------------------------------------------------------------------
help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# -----------------------------------------------------------------------
# Dev environment
# -----------------------------------------------------------------------
dev-start: dev-docker dev-server dev-webapp ## Start full dev environment (Docker + server + webapp in tmux)

dev-docker: ## Start Docker services (postgres, redis, etc.)
	cd $(SERVER_DIR) && \
	  ENABLED_DOCKER_SERVICES="$(DOCKER_SERVICES)" make start-docker

dev-server: ## Open tmux window and start dev server (port 8066)
	@tmux new-window -n "avap-server" \
	  "cd $(SERVER_DIR) && ENABLED_DOCKER_SERVICES=\"$(DOCKER_SERVICES)\" make run-server; exec bash" \
	  || (echo "No tmux session; starting server in foreground..." && \
	      cd $(SERVER_DIR) && ENABLED_DOCKER_SERVICES="$(DOCKER_SERVICES)" make run-server)

dev-webapp: ## Open tmux window and start webpack watch
	@tmux new-window -n "avap-webapp" \
	  "cd $(WEBAPP_DIR) && npm run dev; exec bash" \
	  || echo "No tmux session; run 'cd webapp && npm run dev' manually"

# -----------------------------------------------------------------------
# Plugin
# -----------------------------------------------------------------------
plugin-build: ## Build avapmost-search plugin
	cd $(PLUGIN_DIR) && make dist

plugin-sign: plugin-build ## Build, sign, and copy plugin to prepackaged_plugins
	@echo "==> Signing $(PLUGIN_TAR)"
	gpg --batch --yes --detach-sig \
	    -u "$(PLUGIN_SIG_KEY)" \
	    --output "$(PLUGIN_TAR).sig" \
	    "$(PLUGIN_TAR)"
	cp "$(PLUGIN_TAR)"     "$(PREPACKAGED)/"
	cp "$(PLUGIN_TAR).sig" "$(PREPACKAGED)/"
	@echo "==> Plugin signed and copied to prepackaged_plugins/"

# -----------------------------------------------------------------------
# Docker image build
# -----------------------------------------------------------------------
image-build: plugin-sign _server-build-linux _webapp-build _server-package _docker-build ## Full image build (plugin + server + webapp + docker)

image-build-fast: plugin-sign _server-build-linux _server-package _docker-build ## Image build skipping webapp (webapp/channels/dist must be current)

_server-build-linux:
	cd $(SERVER_DIR) && make build-linux-amd64

_webapp-build:
	cd $(WEBAPP_DIR) && npm run build

_server-package:
	cd $(SERVER_DIR) && make package-linux-amd64

_docker-build:
	docker build \
	  -f $(SERVER_DIR)/build/Dockerfile.avapmost \
	  --build-arg MM_PACKAGE=$(MM_PACKAGE) \
	  -t $(IMAGE_NAME):local \
	  $(SERVER_DIR)
	@echo "==> Built image: $(IMAGE_NAME):local"

image-push: ## Build and push image to registry  (TAG= required, e.g. make image-push TAG=1.0.0)
	@[ -n "$(TAG)" ] || (echo "ERROR: TAG is required. Usage: make image-push TAG=1.0.0" && exit 1)
	$(MAKE) image-build-fast
	docker tag $(IMAGE_NAME):local $(REMOTE_IMAGE):$(TAG)
	docker tag $(IMAGE_NAME):local $(REMOTE_IMAGE):latest
	docker push $(REMOTE_IMAGE):$(TAG)
	docker push $(REMOTE_IMAGE):latest
	$(MAKE) _record-version
	@echo "==> Pushed $(REMOTE_IMAGE):$(TAG) and $(REMOTE_IMAGE):latest"

_record-version:
	@[ -n "$(TAG)" ] || exit 0
	@echo "==> Recording version mapping to $(VERSIONS_FILE)"
	@if ! grep -qF '| Tag |' $(VERSIONS_FILE) 2>/dev/null; then \
	  printf '# Avapmost Version History\n\n| Tag | Mattermost base | Date | Image |\n|-----|-----------------|------|-------|\n' \
	  > $(VERSIONS_FILE); \
	fi
	@printf '| %s | %s | %s | `%s:%s` |\n' \
	  "$(TAG)" "$(MM_VERSION)" "$$(date +%Y-%m-%d)" "$(REMOTE_IMAGE)" "$(TAG)" \
	  >> $(VERSIONS_FILE)
	@echo "==> Recorded: avapmost $(TAG) based on Mattermost $(MM_VERSION)"

# -----------------------------------------------------------------------
# Misc
# -----------------------------------------------------------------------
clean: ## Remove server dist and plugin build artifacts
	rm -rf $(SERVER_DIR)/dist
	cd $(PLUGIN_DIR) && make clean
