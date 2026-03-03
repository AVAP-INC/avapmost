# Search UI Redirect to Plugin

## Goal
Redirect built-in search to the existing hybrid search plugin
that supports hybrid, keyword, and semantic search modes.

## Flow
1. User types in search bar
2. webapp calls Client4.searchPosts()
3. Instead of /api/v4/posts/search, redirect to plugin endpoint
4. Plugin returns PostList format response
5. Existing search result UI renders normally

## Files to Modify (webapp TypeScript)
- webapp/channels/src/client/client4.ts (searchPosts method)
- webapp/channels/src/actions/views/search.ts (performSearch action)
- webapp/channels/src/components/search_bar/ (optional mode toggle)

## Response Format
Must maintain PostList compatibility:
- order: array of post IDs
- posts: map of post_id to Post object
- matches: map of post_id to matched text snippets

## Plugin endpoint spec
- Path: /plugins/PLUGIN_ID/api/v1/search
- Method: POST
- Body: team_id, channel_id (opt), terms, is_or_search, mode (opt), page, per_page
- Response: PostList format

## Fallback
If plugin unavailable, fall back to default database search.
