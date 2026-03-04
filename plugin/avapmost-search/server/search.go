package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
)

const (
	defaultSearchPerPage = 20
	maxSearchPerPage     = 100
)

type searchRequest struct {
	Query     string `json:"query"`
	Page      int    `json:"page"`     // 0-indexed
	PerPage   int    `json:"per_page"` // default 20
	QueryType string `json:"query_type"` // "keyword" (default) | "phrase" | "semantic"
}

type searchResult struct {
	PostID      string  `json:"post_id"`
	Message     string  `json:"message"`
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	ChannelID   string  `json:"channel_id"`
	ChannelName string  `json:"channel_name"`
	ChannelType string  `json:"channel_type"`
	TeamName    string  `json:"team_name"`
	TeamSlug    string  `json:"team_slug"`
	CreateAt    int64   `json:"create_at"`
	Score       float64 `json:"score"`
}

// handleSearch performs an Elasticsearch query filtered to the user's accessible channels.
func (p *Plugin) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	client := p.getESClient()
	if client == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Elasticsearch が設定されていません",
		})
		return
	}

	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{"results": []searchResult{}, "total": 0})
		return
	}

	perPage := req.PerPage
	if perPage <= 0 || perPage > maxSearchPerPage {
		perPage = defaultSearchPerPage
	}
	from := req.Page * perPage

	// Get user's accessible channel IDs for ACL filtering.
	channelIDs, err := p.getAccessibleChannelIDs(userID)
	if err != nil || len(channelIDs) == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{"results": []searchResult{}, "total": 0})
		return
	}

	cfg := p.getConfig()

	// Generate query embedding for semantic search.
	var queryVec []float64
	if req.QueryType == "semantic" {
		if !cfg.semanticEnabled() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"error": "セマンティック検索の設定がありません",
			})
			return
		}
		vec, err := p.generateEmbedding(cfg, req.Query)
		if err != nil {
			p.API.LogError("Failed to generate query embedding", "error", err.Error())
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "埋め込み生成エラー"})
			return
		}
		queryVec = vec
	}

	query := p.buildESQuery(req.Query, req.QueryType, channelIDs, from, perPage, queryVec)

	body, err := json.Marshal(query)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	res, err := client.Search(
		client.Search.WithIndex(cfg.ElasticsearchIndex),
		client.Search.WithBody(bytes.NewReader(body)),
		client.Search.WithContext(context.Background()),
	)
	if err != nil {
		p.API.LogError("ES search failed", "error", err.Error())
		http.Error(w, "search unavailable", http.StatusBadGateway)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		errBody, _ := io.ReadAll(res.Body)
		p.API.LogError("ES search error", "status", res.StatusCode, "body", string(errBody))
		http.Error(w, "search error", http.StatusBadGateway)
		return
	}

	var esResp struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string       `json:"_id"`
				Score  float64      `json:"_score"`
				Source postDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		p.API.LogError("ES response decode failed", "error", err.Error())
		http.Error(w, "search error", http.StatusBadGateway)
		return
	}

	results := make([]searchResult, 0, len(esResp.Hits.Hits))
	for _, hit := range esResp.Hits.Hits {
		src := hit.Source
		sr := searchResult{
			PostID:    src.PostID,
			Message:   src.Message,
			UserID:    src.UserID,
			Username:  src.Username,
			ChannelID: src.ChannelID,
			CreateAt:  src.CreateAt,
			Score:     hit.Score,
		}

		// Enrich with channel/team display names.
		if src.ChannelID != "" {
			if ch, appErr := p.API.GetChannel(src.ChannelID); appErr == nil {
				sr.ChannelName = ch.DisplayName
				sr.ChannelType = string(ch.Type)
				if ch.TeamId != "" {
					if team, appErr := p.API.GetTeam(ch.TeamId); appErr == nil {
						sr.TeamName = team.DisplayName
						sr.TeamSlug = team.Name
					}
				}
			}
		}

		results = append(results, sr)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"total":   esResp.Hits.Total.Value,
	})
}

// buildESQuery constructs the Elasticsearch query DSL.
// For "semantic" query type, vec must be non-nil; the knn block bypasses from/size paging
// (knn uses k + num_candidates instead).
func (p *Plugin) buildESQuery(query, queryType string, channelIDs []string, from, size int, vec []float64) map[string]interface{} {
	if queryType == "semantic" && len(vec) > 0 {
		return map[string]interface{}{
			"knn": map[string]interface{}{
				"field":          "embedding",
				"query_vector":   vec,
				"k":              size,
				"num_candidates": size * 5,
				"filter": map[string]interface{}{
					"terms": map[string]interface{}{
						"channel_id": channelIDs,
					},
				},
			},
		}
	}

	var mustClause map[string]interface{}
	switch queryType {
	case "phrase":
		mustClause = map[string]interface{}{
			"match_phrase": map[string]interface{}{
				"message": query,
			},
		}
	default: // "keyword"
		mustClause = map[string]interface{}{
			"match": map[string]interface{}{
				"message": map[string]interface{}{
					"query":    query,
					"operator": "or",
				},
			},
		}
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClause,
				"filter": map[string]interface{}{
					"terms": map[string]interface{}{
						"channel_id": channelIDs,
					},
				},
			},
		},
		"sort": []interface{}{
			map[string]interface{}{"_score": map[string]interface{}{"order": "desc"}},
			map[string]interface{}{"create_at": map[string]interface{}{"order": "desc"}},
		},
		"from": from,
		"size": size,
	}
}

// getAccessibleChannelIDs returns the list of channel IDs the user has access to.
func (p *Plugin) getAccessibleChannelIDs(userID string) ([]string, error) {
	seen := make(map[string]bool)
	var channelIDs []string

	// Get all teams the user belongs to.
	page := 0
	for {
		teamMembers, appErr := p.API.GetTeamMembersForUser(userID, page, 200)
		if appErr != nil {
			return nil, fmt.Errorf("get team members: %v", appErr.Error())
		}
		if len(teamMembers) == 0 {
			break
		}

		for _, tm := range teamMembers {
			chPage := 0
			for {
				channelMembers, chErr := p.API.GetChannelMembersForUser(tm.TeamId, userID, chPage, 200)
				if chErr != nil {
					break
				}
				if len(channelMembers) == 0 {
					break
				}
				for _, cm := range channelMembers {
					if !seen[cm.ChannelId] {
						seen[cm.ChannelId] = true
						channelIDs = append(channelIDs, cm.ChannelId)
					}
				}
				if len(channelMembers) < 200 {
					break
				}
				chPage++
			}
		}

		if len(teamMembers) < 200 {
			break
		}
		page++
	}

	return channelIDs, nil
}

// postListResponse is the PostSearchResults-compatible format consumed by the Mattermost webapp.
type postListResponse struct {
	Order   []string               `json:"order"`
	Posts   map[string]*model.Post `json:"posts"`
	Matches map[string][]string    `json:"matches"`
}

// handleSearchWebapp performs an ES search and returns results in PostList format so the
// standard Mattermost search UI can render them directly (header search bar integration).
func (p *Plugin) handleSearchWebapp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	client := p.getESClient()
	if client == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Elasticsearch が設定されていません",
		})
		return
	}

	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		writeJSON(w, http.StatusOK, postListResponse{
			Order:   []string{},
			Posts:   map[string]*model.Post{},
			Matches: map[string][]string{},
		})
		return
	}

	perPage := req.PerPage
	if perPage <= 0 || perPage > maxSearchPerPage {
		perPage = defaultSearchPerPage
	}
	from := req.Page * perPage

	channelIDs, err := p.getAccessibleChannelIDs(userID)
	if err != nil || len(channelIDs) == 0 {
		writeJSON(w, http.StatusOK, postListResponse{
			Order:   []string{},
			Posts:   map[string]*model.Post{},
			Matches: map[string][]string{},
		})
		return
	}

	cfg := p.getConfig()

	var queryVec []float64
	if req.QueryType == "semantic" && cfg.semanticEnabled() {
		if vec, err := p.generateEmbedding(cfg, req.Query); err == nil {
			queryVec = vec
		}
	}

	query := p.buildESQuery(req.Query, req.QueryType, channelIDs, from, perPage, queryVec)

	body, err := json.Marshal(query)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	res, err := client.Search(
		client.Search.WithIndex(cfg.ElasticsearchIndex),
		client.Search.WithBody(bytes.NewReader(body)),
		client.Search.WithContext(context.Background()),
	)
	if err != nil {
		p.API.LogError("ES search failed", "error", err.Error())
		http.Error(w, "search unavailable", http.StatusBadGateway)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		errBody, _ := io.ReadAll(res.Body)
		p.API.LogError("ES search error", "status", res.StatusCode, "body", string(errBody))
		http.Error(w, "search error", http.StatusBadGateway)
		return
	}

	var esResp struct {
		Hits struct {
			Hits []struct {
				Source postDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		p.API.LogError("ES response decode failed", "error", err.Error())
		http.Error(w, "search error", http.StatusBadGateway)
		return
	}

	queryTerms := strings.Fields(req.Query)
	resp := postListResponse{
		Order:   make([]string, 0, len(esResp.Hits.Hits)),
		Posts:   make(map[string]*model.Post),
		Matches: make(map[string][]string),
	}

	for _, hit := range esResp.Hits.Hits {
		postID := hit.Source.PostID
		if postID == "" {
			continue
		}
		post, appErr := p.API.GetPost(postID)
		if appErr != nil {
			p.API.LogWarn("Post not found for search result", "post_id", postID)
			continue
		}
		resp.Order = append(resp.Order, postID)
		resp.Posts[postID] = post
		resp.Matches[postID] = queryTerms
	}

	writeJSON(w, http.StatusOK, resp)
}
