#!/bin/bash
# PostToolUse auto-save: commit after every Write/Edit so work is never lost on quota limit

input=$(cat)
tool_name=$(echo "$input" | jq -r '.tool_name // ""')

# Only trigger on file write tools
if [[ "$tool_name" != "Write" && "$tool_name" != "Edit" && "$tool_name" != "NotebookEdit" ]]; then
  exit 0
fi

# Must be inside a git repo
if ! git rev-parse --git-dir >/dev/null 2>&1; then
  exit 0
fi

# Skip if nothing changed
if git diff --quiet && git diff --cached --quiet && [ -z "$(git ls-files --others --exclude-standard)" ]; then
  exit 0
fi

file_path=$(echo "$input" | jq -r '.tool_input.file_path // ""')
file_name=$(basename "$file_path" 2>/dev/null || echo "files")

# Stage all changes and commit (--no-verify skips git hooks, not Claude hooks)
git add -A
git commit --no-verify -m "wip: auto-save $file_name" -m "$(date '+%Y-%m-%d %H:%M:%S')" > /dev/null 2>&1 || exit 0

# Push every 3 wip commits to ensure remote has latest
wip_count=$(git log --oneline | grep -cE "^[a-f0-9]+ wip:" 2>/dev/null || echo 0)
if (( wip_count > 0 && wip_count % 3 == 0 )); then
  current_branch=$(git branch --show-current 2>/dev/null)
  if [[ -n "$current_branch" ]]; then
    git push -u origin "$current_branch" --no-verify > /dev/null 2>&1 || true
  fi
fi

exit 0
