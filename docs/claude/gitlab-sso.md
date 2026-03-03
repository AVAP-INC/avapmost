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
