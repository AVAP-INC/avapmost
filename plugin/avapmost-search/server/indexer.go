package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mattermost/mattermost/server/public/model"
)

// postDocument is the Elasticsearch document structure for a Mattermost post.
type postDocument struct {
	PostID      string    `json:"post_id"`
	Message     string    `json:"message"`
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	ChannelID   string    `json:"channel_id"`
	ChannelType string    `json:"channel_type"`
	TeamID      string    `json:"team_id"`
	CreateAt    int64     `json:"create_at"`
	Embedding   []float64 `json:"embedding,omitempty"`
}

// indexPost upserts a post document into Elasticsearch.
func (p *Plugin) indexPost(client *elasticsearch.Client, index string, post *model.Post) error {
	doc := postDocument{
		PostID:    post.Id,
		Message:   post.Message,
		UserID:    post.UserId,
		CreateAt:  post.CreateAt,
		ChannelID: post.ChannelId,
	}

	// Resolve username.
	if u, appErr := p.API.GetUser(post.UserId); appErr == nil {
		doc.Username = u.Username
	}

	// Resolve channel info.
	if ch, appErr := p.API.GetChannel(post.ChannelId); appErr == nil {
		doc.ChannelType = string(ch.Type)
		doc.TeamID = ch.TeamId
	}

	// Generate embedding if configured.
	cfg := p.getConfig()
	if cfg.semanticEnabled() {
		if vec, err := p.generateEmbedding(cfg, post.Message); err == nil {
			doc.Embedding = vec
		} else {
			p.API.LogWarn("Failed to generate embedding", "post_id", post.Id, "error", err.Error())
		}
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal document: %w", err)
	}

	res, err := client.Index(
		index,
		bytes.NewReader(body),
		client.Index.WithDocumentID(post.Id),
		client.Index.WithContext(context.Background()),
	)
	if err != nil {
		return fmt.Errorf("index request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		errBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("index error (%d): %s", res.StatusCode, errBody)
	}

	return nil
}

// deletePost removes a post document from Elasticsearch.
func (p *Plugin) deletePost(client *elasticsearch.Client, index, postID string) error {
	res, err := client.Delete(
		index,
		postID,
		client.Delete.WithContext(context.Background()),
	)
	if err != nil {
		return fmt.Errorf("delete request: %w", err)
	}
	defer res.Body.Close()

	// 404 is acceptable (post was never indexed).
	if res.IsError() && res.StatusCode != 404 {
		errBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("delete error (%d): %s", res.StatusCode, errBody)
	}

	return nil
}

// generateEmbedding calls an OpenAI-compatible embedding API and returns the vector.
func (p *Plugin) generateEmbedding(cfg configuration, text string) ([]float64, error) {
	if text = strings.TrimSpace(text); text == "" {
		return nil, fmt.Errorf("empty text")
	}

	reqBody, err := json.Marshal(map[string]string{
		"input": text,
		"model": cfg.EmbeddingModel,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
		cfg.EmbeddingAPIURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.EmbeddingAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding API error (%d): %s", resp.StatusCode, body)
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}
	if len(result.Data) == 0 || len(result.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding in response")
	}

	return result.Data[0].Embedding, nil
}
