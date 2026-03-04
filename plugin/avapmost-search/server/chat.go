package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/mattermost/mattermost/server/public/model"
)

const (
	askContextSize        = 20  // 直近メッセージ数 (channel モード)
	askContextPerChannel  = 5   // team/all モードで1チャンネルあたりのメッセージ数
	askContextMaxChannels = 10  // team/all モードでサンプルするチャンネル数上限
)

type askChannelContext struct {
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	ChannelType string `json:"channel_type"`
	TeamID      string `json:"team_id,omitempty"`
}

type askRequest struct {
	Question string             `json:"question"`
	Context  *askChannelContext `json:"context,omitempty"`
	Target   string             `json:"target"` // "channel" (default) | "team" | "all"
}

// handleAsk receives a question and responds using the Anthropic Claude API.
func (p *Plugin) handleAsk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	cfg := p.getConfig()
	if cfg.AnthropicAPIKey == "" {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "Anthropic API キーが設定されていません",
		})
		return
	}

	var req askRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Question) == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Target == "" {
		req.Target = "channel"
	}

	// Collect context posts respecting ACL.
	var channelName, channelHeader string
	if req.Context != nil && req.Context.ChannelID != "" {
		if ch, appErr := p.API.GetChannel(req.Context.ChannelID); appErr == nil {
			channelName = ch.DisplayName
			channelHeader = ch.Header
		}
	}

	recentPosts := p.getContextPosts(userID, req.Question, req.Target, req.Context)

	userName := ""
	if u, appErr := p.API.GetUser(userID); appErr == nil {
		userName = "@" + u.Username
	}

	prompt := buildPrompt(channelName, channelHeader, recentPosts, req.Target, userID, userName, req.Question)

	answer, err := callClaude(cfg.AnthropicAPIKey, prompt)
	if err != nil {
		p.API.LogError("Claude API call failed", "error", err.Error())
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "AI サービスエラー"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"answer": answer})
}

// getContextPosts returns posts relevant to the question, scoped to the target and the
// user's accessible channels (ACL enforcement).
func (p *Plugin) getContextPosts(userID, question, target string, ctx *askChannelContext) []*model.Post {
	switch target {
	case "team":
		teamID := ""
		if ctx != nil {
			teamID = ctx.TeamID
		}
		return p.getContextPostsForTeam(userID, question, teamID)
	case "all":
		return p.getContextPostsAll(userID, question)
	default: // "channel"
		channelID := ""
		if ctx != nil {
			channelID = ctx.ChannelID
		}
		return p.getContextPostsForChannel(userID, channelID)
	}
}

// getContextPostsForChannel returns recent posts from a single channel, verifying ACL.
func (p *Plugin) getContextPostsForChannel(userID, channelID string) []*model.Post {
	if channelID == "" {
		return nil
	}
	// ACL: verify user is a member of this channel.
	if _, appErr := p.API.GetChannelMember(channelID, userID); appErr != nil {
		p.API.LogWarn("User not in channel, skipping context", "user_id", userID, "channel_id", channelID)
		return nil
	}
	postList, appErr := p.API.GetPostsForChannel(channelID, 0, askContextSize)
	if appErr != nil {
		return nil
	}
	return orderedPosts(postList)
}

// getContextPostsForTeam returns question-relevant posts from the user's channels in a team.
func (p *Plugin) getContextPostsForTeam(userID, question, teamID string) []*model.Post {
	if teamID == "" {
		return nil
	}
	// Get channels the user can access in this team.
	members, appErr := p.API.GetChannelMembersForUser(teamID, userID, 0, 200)
	if appErr != nil {
		return nil
	}
	channelIDs := make([]string, 0, len(members))
	for _, m := range members {
		channelIDs = append(channelIDs, m.ChannelId)
	}
	return p.sampleContextPosts(channelIDs, question)
}

// getContextPostsAll returns question-relevant posts from all channels accessible to the user.
func (p *Plugin) getContextPostsAll(userID, question string) []*model.Post {
	channelIDs, err := p.getAccessibleChannelIDs(userID)
	if err != nil || len(channelIDs) == 0 {
		return nil
	}
	return p.sampleContextPosts(channelIDs, question)
}

// sampleContextPosts retrieves relevant posts from a sample of channels.
// Uses ES keyword search when available; falls back to recent posts.
func (p *Plugin) sampleContextPosts(channelIDs []string, question string) []*model.Post {
	esClient := p.getESClient()

	// Prefer ES-based relevance if available.
	if esClient != nil && strings.TrimSpace(question) != "" {
		if posts := p.searchContextPosts(channelIDs, question); len(posts) > 0 {
			return posts
		}
	}

	// Fallback: sample recent posts from a few channels.
	limit := askContextMaxChannels
	if len(channelIDs) < limit {
		limit = len(channelIDs)
	}
	var posts []*model.Post
	for _, chID := range channelIDs[:limit] {
		pl, appErr := p.API.GetPostsForChannel(chID, 0, askContextPerChannel)
		if appErr != nil {
			continue
		}
		posts = append(posts, orderedPosts(pl)...)
	}
	return posts
}

// searchContextPosts runs an ES keyword search to find the most relevant context posts.
func (p *Plugin) searchContextPosts(channelIDs []string, question string) []*model.Post {
	esClient := p.getESClient()
	if esClient == nil {
		return nil
	}
	cfg := p.getConfig()

	query := p.buildESQuery(question, "keyword", channelIDs, 0, askContextSize, nil)
	bodyBytes, err := json.Marshal(query)
	if err != nil {
		return nil
	}

	res, err := esClient.Search(
		esClient.Search.WithIndex(cfg.ElasticsearchIndex),
		esClient.Search.WithBody(bytes.NewReader(bodyBytes)),
		esClient.Search.WithContext(context.Background()),
	)
	if err != nil || res.IsError() {
		if res != nil {
			res.Body.Close()
		}
		return nil
	}
	defer res.Body.Close()

	var esResp struct {
		Hits struct {
			Hits []struct {
				Source postDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil
	}

	posts := make([]*model.Post, 0, len(esResp.Hits.Hits))
	for _, hit := range esResp.Hits.Hits {
		post, appErr := p.API.GetPost(hit.Source.PostID)
		if appErr == nil {
			posts = append(posts, post)
		}
	}
	return posts
}

// orderedPosts converts a PostList to a slice ordered oldest-first, excluding deleted posts.
func orderedPosts(pl *model.PostList) []*model.Post {
	if pl == nil {
		return nil
	}
	posts := make([]*model.Post, 0, len(pl.Order))
	for i := len(pl.Order) - 1; i >= 0; i-- {
		if post, ok := pl.Posts[pl.Order[i]]; ok && post.DeleteAt == 0 {
			posts = append(posts, post)
		}
	}
	return posts
}

// buildPrompt constructs a prompt that includes channel context.
func buildPrompt(channelName, channelHeader string, posts []*model.Post, target, userID, userName, question string) string {
	var sb strings.Builder

	sb.WriteString("## コンテキスト範囲\n")
	targetLabel := map[string]string{
		"channel": "現在のチャンネル",
		"team":    "現在のチーム全体",
		"all":     "アクセス可能な全チャンネル",
	}[target]
	if targetLabel == "" {
		targetLabel = target
	}
	fmt.Fprintf(&sb, "スコープ: %s\n", targetLabel)
	if channelName != "" {
		fmt.Fprintf(&sb, "チャンネル名: %s\n", channelName)
	}
	if channelHeader != "" {
		fmt.Fprintf(&sb, "チャンネルヘッダー: %s\n", channelHeader)
	}
	fmt.Fprintf(&sb, "質問者: %s (user_id: %s)\n\n", userName, userID)

	if len(posts) > 0 {
		fmt.Fprintf(&sb, "## 参照メッセージ (%d件)\n", len(posts))
		for _, post := range posts {
			t := time.Unix(post.CreateAt/1000, 0).UTC()
			fmt.Fprintf(&sb, "[%s] %s: %s\n",
				t.Format("01/02 15:04"),
				post.UserId,
				truncate(post.Message, 200),
			)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## 質問\n")
	sb.WriteString(question)

	return sb.String()
}

// callClaude sends prompt to Anthropic Messages API and returns the text reply.
func callClaude(apiKey, prompt string) (string, error) {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	msg, err := client.Messages.New(
		context.Background(),
		anthropic.MessageNewParams{
			Model:     anthropic.Model("claude-sonnet-4-6"),
			MaxTokens: 1024,
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			},
		},
	)
	if err != nil {
		return "", err
	}
	for _, block := range msg.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}
	return "", fmt.Errorf("no text content in response")
}

// truncate cuts a string to max runes and appends "…".
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…"
}
