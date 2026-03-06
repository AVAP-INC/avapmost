package main

import "strconv"

// configuration holds plugin settings loaded from the Mattermost admin UI.
type configuration struct {
	ElasticsearchURL      string `json:"ElasticsearchURL"`
	ElasticsearchIndex    string `json:"ElasticsearchIndex"`
	ElasticsearchUsername string `json:"ElasticsearchUsername"`
	ElasticsearchPassword string `json:"ElasticsearchPassword"`
	AnthropicAPIKey       string `json:"AnthropicAPIKey"`

	// Embedding API (OpenAI-compatible format). Leave empty to disable semantic search.
	EmbeddingAPIURL        string `json:"EmbeddingAPIURL"`
	EmbeddingAPIKey        string `json:"EmbeddingAPIKey"`
	EmbeddingModel         string `json:"EmbeddingModel"`
	EmbeddingDimensions    string `json:"EmbeddingDimensions"` // stored as string for Mattermost settings
}

func (c *configuration) setDefaults() {
	if c.ElasticsearchURL == "" {
		c.ElasticsearchURL = "http://localhost:9200"
	}
	if c.ElasticsearchIndex == "" {
		c.ElasticsearchIndex = "avapmost_posts"
	}
	if c.EmbeddingModel == "" {
		c.EmbeddingModel = "text-embedding-3-small"
	}
	if c.EmbeddingDimensions == "" {
		c.EmbeddingDimensions = "1536"
	}
}

// embeddingDims returns EmbeddingDimensions as int, defaulting to 1536.
func (c *configuration) embeddingDims() int {
	if n, err := strconv.Atoi(c.EmbeddingDimensions); err == nil && n > 0 {
		return n
	}
	return 1536
}

// semanticEnabled returns true when the embedding API is configured.
func (c *configuration) semanticEnabled() bool {
	return c.EmbeddingAPIURL != "" && c.EmbeddingAPIKey != ""
}
