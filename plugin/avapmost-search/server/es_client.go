package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
)

// initESClient creates or replaces the Elasticsearch client based on current config.
func (p *Plugin) initESClient() error {
	cfg := p.getConfig()
	if cfg.ElasticsearchURL == "" {
		return fmt.Errorf("ElasticsearchURL is not configured")
	}

	esCfg := elasticsearch.Config{
		Addresses: []string{cfg.ElasticsearchURL},
	}
	if cfg.ElasticsearchUsername != "" {
		esCfg.Username = cfg.ElasticsearchUsername
		esCfg.Password = cfg.ElasticsearchPassword
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return fmt.Errorf("failed to create ES client: %w", err)
	}

	p.esMu.Lock()
	p.esClient = client
	p.esMu.Unlock()

	return nil
}

// ensureIndex creates the index with mappings if it does not exist, then updates
// the mapping to add the embedding dense_vector field if semantic search is enabled.
func (p *Plugin) ensureIndex(client *elasticsearch.Client, index string) error {
	// Check if index exists.
	res, err := client.Indices.Exists([]string{index})
	if err != nil {
		return fmt.Errorf("indices.exists: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		// Index does not exist — create it.
		if err := p.createIndex(client, index, true); err != nil {
			p.API.LogWarn("kuromoji analyzer unavailable, falling back to standard", "error", err.Error())
			if err := p.createIndex(client, index, false); err != nil {
				return err
			}
		}
	}

	// Ensure dense_vector field exists if semantic search is enabled.
	cfg := p.getConfig()
	if cfg.semanticEnabled() {
		if err := p.ensureEmbeddingMapping(client, index, cfg.embeddingDims()); err != nil {
			p.API.LogWarn("Failed to add embedding mapping", "error", err.Error())
		}
	}

	return nil
}

// createIndex creates the index with the given analyzer choice.
func (p *Plugin) createIndex(client *elasticsearch.Client, index string, useKuromoji bool) error {
	analyzer := "standard"
	var settingsBlock string
	if useKuromoji {
		analyzer = "kuromoji"
		settingsBlock = `"settings": {"analysis": {"analyzer": {"default": {"type": "kuromoji"}}}},`
	}

	mapping := fmt.Sprintf(`{
		%s
		"mappings": {
			"properties": {
				"post_id":      {"type": "keyword"},
				"message":      {"type": "text", "analyzer": "%s"},
				"user_id":      {"type": "keyword"},
				"username":     {"type": "keyword"},
				"channel_id":   {"type": "keyword"},
				"channel_type": {"type": "keyword"},
				"team_id":      {"type": "keyword"},
				"create_at":    {"type": "date", "format": "epoch_millis"}
			}
		}
	}`, settingsBlock, analyzer)

	res, err := client.Indices.Create(
		index,
		client.Indices.Create.WithBody(bytes.NewReader([]byte(mapping))),
		client.Indices.Create.WithContext(context.Background()),
	)
	if err != nil {
		return fmt.Errorf("indices.create: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		var esErr struct {
			Error struct {
				Type string `json:"type"`
			} `json:"error"`
		}
		if jsonErr := json.Unmarshal(body, &esErr); jsonErr == nil &&
			esErr.Error.Type == "resource_already_exists_exception" {
			return nil
		}
		return fmt.Errorf("indices.create failed (%d): %s", res.StatusCode, body)
	}

	p.API.LogInfo("ES index created", "index", index, "kuromoji", useKuromoji)
	return nil
}

// ensureEmbeddingMapping adds the dense_vector field to the index mapping if not present.
func (p *Plugin) ensureEmbeddingMapping(client *elasticsearch.Client, index string, dims int) error {
	mapping := fmt.Sprintf(`{
		"properties": {
			"embedding": {
				"type":       "dense_vector",
				"dims":       %d,
				"index":      true,
				"similarity": "cosine"
			}
		}
	}`, dims)

	res, err := client.Indices.PutMapping(
		[]string{index},
		bytes.NewReader([]byte(mapping)),
		client.Indices.PutMapping.WithContext(context.Background()),
	)
	if err != nil {
		return fmt.Errorf("put_mapping: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		// Ignore errors if the field already exists with same dims.
		var esErr struct {
			Error struct {
				Type string `json:"type"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &esErr) == nil && esErr.Error.Type == "illegal_argument_exception" {
			// Field already exists (possibly with different dims) — warn only.
			p.API.LogWarn("embedding mapping update skipped (field may already exist)", "index", index)
			return nil
		}
		return fmt.Errorf("put_mapping failed (%d): %s", res.StatusCode, body)
	}

	p.API.LogInfo("ES embedding mapping updated", "index", index, "dims", dims)
	return nil
}
