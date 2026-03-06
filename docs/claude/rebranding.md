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
