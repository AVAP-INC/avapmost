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
