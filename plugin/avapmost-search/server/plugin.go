package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// Plugin implements the Mattermost plugin interface.
type Plugin struct {
	plugin.MattermostPlugin

	configMu sync.RWMutex
	config   configuration

	esMu     sync.RWMutex
	esClient *elasticsearch.Client

	reindexStatus reindexStatus
}

// getConfig returns a snapshot of the current configuration.
func (p *Plugin) getConfig() configuration {
	p.configMu.RLock()
	defer p.configMu.RUnlock()
	return p.config
}

// getESClient returns the current ES client (may be nil if not configured).
func (p *Plugin) getESClient() *elasticsearch.Client {
	p.esMu.RLock()
	defer p.esMu.RUnlock()
	return p.esClient
}

// OnConfigurationChange is called when plugin settings change.
func (p *Plugin) OnConfigurationChange() error {
	var cfg configuration
	if err := p.API.LoadPluginConfiguration(&cfg); err != nil {
		return fmt.Errorf("failed to load plugin configuration: %w", err)
	}
	cfg.setDefaults()

	p.configMu.Lock()
	p.config = cfg
	p.configMu.Unlock()

	// Reinitialize ES client on config change.
	if err := p.initESClient(); err != nil {
		p.API.LogWarn("ES client initialization failed", "error", err.Error())
	}

	return nil
}

// OnActivate initializes the plugin.
func (p *Plugin) OnActivate() error {
	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	client := p.getESClient()
	if client != nil {
		cfg := p.getConfig()
		if err := p.ensureIndex(client, cfg.ElasticsearchIndex); err != nil {
			p.API.LogWarn("Failed to create/verify ES index", "error", err.Error())
		}
	}

	p.API.LogInfo("Avapmost Search plugin activated")
	return nil
}

// MessageHasBeenPosted indexes a new post.
func (p *Plugin) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {
	if !isIndexablePost(post) {
		return
	}
	client := p.getESClient()
	if client == nil {
		return
	}
	cfg := p.getConfig()
	if err := p.indexPost(client, cfg.ElasticsearchIndex, post); err != nil {
		p.API.LogError("Failed to index post", "post_id", post.Id, "error", err.Error())
	}
}

// MessageHasBeenUpdated re-indexes an updated post.
func (p *Plugin) MessageHasBeenUpdated(c *plugin.Context, newPost, oldPost *model.Post) {
	if !isIndexablePost(newPost) {
		return
	}
	client := p.getESClient()
	if client == nil {
		return
	}
	cfg := p.getConfig()
	if err := p.indexPost(client, cfg.ElasticsearchIndex, newPost); err != nil {
		p.API.LogError("Failed to re-index updated post", "post_id", newPost.Id, "error", err.Error())
	}
}

// MessageHasBeenDeleted removes a post from the index.
func (p *Plugin) MessageHasBeenDeleted(c *plugin.Context, post *model.Post) {
	client := p.getESClient()
	if client == nil {
		return
	}
	cfg := p.getConfig()
	if err := p.deletePost(client, cfg.ElasticsearchIndex, post.Id); err != nil {
		p.API.LogError("Failed to delete post from index", "post_id", post.Id, "error", err.Error())
	}
}

// ServeHTTP handles plugin HTTP requests.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/v1/config":
		p.handleConfig(w, r)
	case "/api/v1/search":
		p.handleSearch(w, r)
	case "/api/v1/search/webapp":
		p.handleSearchWebapp(w, r)
	case "/api/v1/ask":
		p.handleAsk(w, r)
	case "/api/v1/reindex/start":
		p.handleReindexStart(w, r)
	case "/api/v1/reindex/status":
		p.handleReindexStatus(w, r)
	default:
		http.NotFound(w, r)
	}
}

// isIndexablePost returns true if the post should be indexed.
func isIndexablePost(post *model.Post) bool {
	if post == nil {
		return false
	}
	// Skip system messages.
	if post.Type != "" {
		return false
	}
	// Skip empty messages.
	if strings.TrimSpace(post.Message) == "" {
		return false
	}
	return true
}

// handleConfig returns plugin capability flags to the webapp.
func (p *Plugin) handleConfig(w http.ResponseWriter, r *http.Request) {
	cfg := p.getConfig()
	writeJSON(w, http.StatusOK, map[string]bool{
		"semantic_enabled":  cfg.semanticEnabled(),
		"anthropic_enabled": cfg.AnthropicAPIKey != "",
	})
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
