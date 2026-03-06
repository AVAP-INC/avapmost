package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

// reindexStatus tracks the progress of a full reindex operation.
type reindexStatus struct {
	running atomic.Bool
	total   atomic.Int64
	indexed atomic.Int64
	errMsg  atomic.Value // stores string
}

// handleReindexStart starts a full reindex job (admin only).
func (p *Plugin) handleReindexStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Require system admin role.
	user, appErr := p.API.GetUser(userID)
	if appErr != nil || !user.IsSystemAdmin() {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if p.reindexStatus.running.Load() {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "reindex is already running"})
		return
	}

	client := p.getESClient()
	if client == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "Elasticsearch not configured"})
		return
	}

	cfg := p.getConfig()

	// Reset status.
	p.reindexStatus.running.Store(true)
	p.reindexStatus.total.Store(0)
	p.reindexStatus.indexed.Store(0)
	p.reindexStatus.errMsg.Store("")

	go p.runReindex(client, cfg.ElasticsearchIndex)

	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// handleReindexStatus returns the current reindex progress.
func (p *Plugin) handleReindexStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	errMsg := ""
	if v := p.reindexStatus.errMsg.Load(); v != nil {
		errMsg, _ = v.(string)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"running": p.reindexStatus.running.Load(),
		"total":   p.reindexStatus.total.Load(),
		"indexed": p.reindexStatus.indexed.Load(),
		"error":   errMsg,
	})
}

// runReindex iterates all teams/channels/posts and bulk-indexes them.
func (p *Plugin) runReindex(client *elasticsearch.Client, index string) {
	defer p.reindexStatus.running.Store(false)

	start := time.Now()
	p.API.LogInfo("Reindex started", "index", index)

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         index,
		Client:        client,
		NumWorkers:    2,
		FlushBytes:    5 * 1024 * 1024, // 5 MB
		FlushInterval: 10 * time.Second,
	})
	if err != nil {
		p.reindexStatus.errMsg.Store(fmt.Sprintf("bulk indexer init: %v", err))
		return
	}

	// Count total posts for progress (approximate: we count as we go).
	var totalPosts int64

	// Get all teams.
	teams, appErr := p.API.GetTeams()
	if appErr != nil {
		p.reindexStatus.errMsg.Store(fmt.Sprintf("get teams: %v", appErr.Error()))
		return
	}

	for _, team := range teams {
		// Get all channels in this team, page by page.
		chPage := 0
		for {
			channels, chErr := p.API.GetPublicChannelsForTeam(team.Id, chPage, 200)
			if chErr != nil {
				break
			}
			if len(channels) == 0 {
				break
			}

			for _, ch := range channels {
				p.reindexChannel(client, bi, index, ch.Id, string(ch.Type), ch.TeamId, &totalPosts)
			}

			if len(channels) < 200 {
				break
			}
			chPage++
		}
	}

	if err := bi.Close(context.Background()); err != nil {
		p.reindexStatus.errMsg.Store(fmt.Sprintf("bulk indexer close: %v", err))
		return
	}

	stats := bi.Stats()
	p.reindexStatus.indexed.Store(int64(stats.NumFlushed))
	p.reindexStatus.total.Store(totalPosts)

	p.API.LogInfo("Reindex completed",
		"indexed", stats.NumFlushed,
		"failed", stats.NumFailed,
		"duration", time.Since(start).String(),
	)
}

// reindexChannel indexes all posts in a channel using the bulk indexer.
func (p *Plugin) reindexChannel(
	_ *elasticsearch.Client,
	bi esutil.BulkIndexer,
	index, channelID, channelType, teamID string,
	totalPosts *int64,
) {
	const batchSize = 500

	postPage := 0
	for {
		postList, appErr := p.API.GetPostsForChannel(channelID, postPage, batchSize)
		if appErr != nil {
			break
		}
		if postList == nil || len(postList.Order) == 0 {
			break
		}

		for _, postID := range postList.Order {
			post, ok := postList.Posts[postID]
			if !ok || !isIndexablePost(post) {
				continue
			}

			*totalPosts++
			p.reindexStatus.total.Add(1)

			// Resolve username (best effort).
			username := post.UserId
			if u, err := p.API.GetUser(post.UserId); err == nil {
				username = u.Username
			}

			doc := postDocument{
				PostID:      post.Id,
				Message:     post.Message,
				UserID:      post.UserId,
				Username:    username,
				ChannelID:   channelID,
				ChannelType: channelType,
				TeamID:      teamID,
				CreateAt:    post.CreateAt,
			}

			body, err := json.Marshal(doc)
			if err != nil {
				continue
			}

			if err := bi.Add(context.Background(), esutil.BulkIndexerItem{
				Action:     "index",
				Index:      index,
				DocumentID: post.Id,
				Body:       bytes.NewReader(body),
				OnSuccess: func(_ context.Context, _ esutil.BulkIndexerItem, _ esutil.BulkIndexerResponseItem) {
					p.reindexStatus.indexed.Add(1)
				},
				OnFailure: func(_ context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						p.API.LogError("Bulk index item failed", "id", item.DocumentID, "error", err.Error())
					} else {
						p.API.LogError("Bulk index item failed", "id", item.DocumentID, "error", res.Error.Reason)
					}
				},
			}); err != nil {
				p.API.LogError("Failed to add item to bulk indexer", "error", err.Error())
			}
		}

		if len(postList.Order) < batchSize {
			break
		}
		postPage++
	}
}
