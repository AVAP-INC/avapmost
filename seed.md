# Avapmost - Claude Code Setup Guide

## Overview

This document contains all files needed to set up Claude Code for the
avapmost fork (https://github.com/AVAP-INC/avapmost).

Copy each section into the corresponding file in your repository.

---

## File 1: CLAUDE.md (repository root)

```markdown
# Avapmost

Mattermost v11 fork by AVAP-INC. Rebranding + feature restoration.
Repo: https://github.com/AVAP-INC/avapmost

## WHY

Mattermost v11 removed critical features from the free edition:
- GitLab SSO (removed from Team Edition)
- Bleve search (the only free option for Japanese full-text search)
- MySQL support ended

## WHAT

- **server/** - Go backend (core logic in server/channels/app/)
- **webapp/** - React/TypeScript frontend (webapp/channels/src/)
- **api/** - OpenAPI schema
- **e2e-tests/** - E2E tests (Playwright)

## Three modification areas

### 1. Rebranding (trademark compliance)
- Rename "Mattermost" to "Avapmost" in UI and i18n
- Replace logo, icon, favicon
- IMPORTANT: DO NOT remove Mattermost Inc. copyright from login footer and About dialog
- IMPORTANT: DO NOT modify NOTICE.txt or LICENSE.txt
- Details: @docs/claude/rebranding.md

### 2. Search UI redirect to plugin
- Redirect webapp Client4.searchPosts() to plugin API
- Leverage existing search plugin (hybrid/keyword/semantic)
- Maintain PostList format: order[], posts{}, matches{}
- Details: @docs/claude/search-redirect.md

### 3. GitLab SSO restoration or Generic OAuth2
- Restore auth flow referencing v10 code
- Key files: server/channels/app/oauth.go, server/public/model/config.go
- Login UI: webapp/channels/src/components/login/
- Details: @docs/claude/gitlab-sso.md

## Build and test

Server:       cd server && make build
Webapp:       cd webapp && npm ci && npm run build
Go tests:     cd server && make test
JS tests:     cd webapp && npm test
Go lint:      cd server && make golangci-lint
JS lint:      cd webapp && npm run check
Full package: make package

## Coding conventions
- Go: Follow existing golangci-lint config
- TypeScript: Follow existing ESLint/prettier config
- Commit messages: [area] description (e.g. [rebrand] Replace logo assets)
- Branches: feature/rebrand, feature/search-redirect, feature/gitlab-sso

## DO NOT
- Modify NOTICE.txt, LICENSE.txt, LICENSE.enterprise
- Touch server/enterprise/ directory (Source Available License)
- Remove Mattermost Inc. copyright from login footer / About dialog
- Break existing plugin API compatibility
```

---

## File 2: docs/claude/rebranding.md

```markdown
# Rebranding Guide: Mattermost -> Avapmost

## Overview
Mattermost trademark policy requires derivative versions to replace name and logo.
However, Mattermost Inc. copyright notices MUST remain in login footer and About dialog.

## Phase 1: Identify all brand assets

Discovery commands:
- Logo/image files: search webapp/channels/src/images/
- i18n strings: grep "Mattermost" in webapp/channels/src/i18n/en.json
- Email templates: grep in server/templates/
- PWA manifest and favicon locations
- HTML title and meta tags

## Phase 2: Name replacement rules

### REPLACE (UI-facing display text):
- Window/page titles
- i18n strings in en.json and ja.json (display text only, not keys)
- Email template visible text
- System Console visible labels
- About dialog product name (NOT the copyright line)
- README.md

### DO NOT REPLACE:
- Mattermost, Inc. copyright notices (login footer, About dialog)
- NOTICE.txt, LICENSE.txt, LICENSE.enterprise
- Go package names and import paths (functional, not branding)
- API endpoint paths (/api/v4/...)
- Plugin API interface names
- Config key names in config.json (e.g., ServiceSettings)
- Internal variable names in Go/TypeScript code
- server/enterprise/ directory

### HANDLE WITH CARE:
- User-Agent strings
- Server header values
- Error messages with product name (replace display, keep internal)

## Phase 3: Logo and icon replacement

Replace these with Avapmost assets:
- webapp/channels/src/images/logo.svg
- webapp/channels/src/images/favicon/
- webapp/channels/src/images/icon.png
- server/templates/ (email header logos)

## Phase 4: Verification
- Check no Mattermost branding remains in UI (excluding allowed locations)
- Verify copyright notices intact in login footer and About dialog
- Build webapp and visual check
```

---

## File 3: docs/claude/search-redirect.md

```markdown
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
```

---

## File 4: docs/claude/gitlab-sso.md

```markdown
# GitLab SSO Restoration

## Background
Mattermost v11 removed GitLab SSO from Team Edition.
Restore by referencing v10.x code.

## Server side (Go) changes
1. server/public/model/config.go - Restore GitLabSettings struct
2. server/channels/app/oauth.go - Restore GitLab OAuth provider
3. server/channels/app/login.go - Re-enable GitLab auth service
4. server/channels/api4/oauth.go - API endpoints for auth flow

## Webapp side (TypeScript) changes
1. webapp/channels/src/components/login/login.tsx - Restore login button
2. webapp/channels/src/components/admin_console/ - Config panel

## GitLabSettings config fields
- Enable (bool)
- Id (OAuth App ID)
- Secret (OAuth App Secret)
- Scope (default: read_user)
- AuthEndpoint, TokenEndpoint, UserAPIEndpoint

## Alternative: Generic OAuth2 Provider
More future-proof: implement generic OAuth2 for any IdP
(GitLab, Gitea, Keycloak, etc.)

## Migration
Preserve existing AuthData for users with gitlab auth service.
```

---

## File 5: .claude/commands/rebrand.md

```markdown
---
description: Run rebranding tasks - replace Mattermost name/logo with Avapmost
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
---

Read @docs/claude/rebranding.md for detailed instructions.

Your task: $ARGUMENTS

Steps:
1. First scan the codebase to understand current brand asset locations
2. Make the requested changes following the replacement rules
3. NEVER touch copyright notices, NOTICE.txt, or LICENSE files
4. Verify changes with grep to ensure consistency
5. Run the webapp build to check for errors
```

---

## File 6: .claude/commands/search-redirect.md

```markdown
---
description: Work on redirecting search UI to the hybrid search plugin
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
---

Read @docs/claude/search-redirect.md for architecture details.

Your task: $ARGUMENTS

Steps:
1. Understand the current search flow in the webapp
2. Make the requested changes to redirect search to plugin
3. Maintain PostList response format compatibility
4. Include fallback to default search if plugin unavailable
5. Run TypeScript type checks after changes
```

---

## File 7: .claude/commands/gitlab-sso.md

```markdown
---
description: Work on GitLab SSO restoration or Generic OAuth2
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
---

Read @docs/claude/gitlab-sso.md for implementation details.

Your task: $ARGUMENTS

Steps:
1. Reference Mattermost v10 code for the original implementation
2. Make the requested changes
3. Ensure backward compatibility with existing auth data
4. Run Go tests for auth-related packages
```

---

## File 8: .claude/commands/investigate.md

```markdown
---
description: Investigate a specific area of the Mattermost codebase using subagents
allowed-tools: Read, Grep, Glob
---

Use subagents to investigate: $ARGUMENTS

Spawn subagents to explore different parts of the codebase in parallel.
Report back a summary including:
- Relevant file paths
- Key functions/components
- Dependencies and call chains
- Suggested modification approach
```

---

## File 9: .claude/settings.json

```json
{
  "permissions": {
    "allow": [
      "Bash(git:*)",
      "Bash(make:*)",
      "Bash(npm:*)",
      "Bash(go:*)",
      "Bash(grep:*)",
      "Bash(find:*)",
      "Bash(cat:*)",
      "Bash(head:*)",
      "Read",
      "Write",
      "Edit",
      "Glob",
      "Grep"
    ],
    "deny": [
      "Bash(rm -rf /)",
      "Bash(git push --force)"
    ]
  }
}
```

---

## Quick Start: How to use with Claude Code

### Step 1: Clone and setup

```bash
git clone https://github.com/AVAP-INC/avapmost.git
cd avapmost

# Copy CLAUDE.md to repo root
# Copy docs/claude/*.md files
# Copy .claude/ directory

git add CLAUDE.md docs/claude/ .claude/
git commit -m "[setup] Add Claude Code configuration"
```

### Step 2: Start Claude Code

```bash
cd avapmost
claude
```

### Step 3: Use commands for each phase

```
# Phase 1: Rebranding
/project:rebrand Scan all Mattermost brand references and list them

/project:rebrand Replace product name in en.json and ja.json

/project:rebrand Replace logo SVG files with the new Avapmost logo

# Phase 2: Search redirect
/project:search-redirect Analyze the current Client4.searchPosts flow

/project:search-redirect Implement the redirect to plugin API with fallback

# Phase 3: GitLab SSO
/project:gitlab-sso Investigate the v10 GitLab SSO code that was removed

/project:gitlab-sso Restore GitLabSettings in config.go

# General investigation
/project:investigate How does the login page render SSO buttons?

/project:investigate What is the full search flow from search bar to API?
```

### Step 4: Branching strategy

```bash
# Work on each area in separate branches
git checkout -b feature/rebrand
# ... do rebranding work with Claude Code ...
git commit -m "[rebrand] Replace Mattermost branding with Avapmost"

git checkout -b feature/search-redirect
# ... do search redirect work ...

git checkout -b feature/gitlab-sso
# ... do SSO restoration work ...
```

### Tips
- Use separate Claude Code sessions for each feature branch
- Run /compact when context gets long, with instruction to preserve file list
- Use /project:investigate with subagents to explore without filling context
- Check the webapp build after rebranding changes to catch missed references
- Keep commits granular: one logical change per commit
