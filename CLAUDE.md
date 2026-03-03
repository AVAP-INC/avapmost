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
