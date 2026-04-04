#!/bin/bash
# Stop hook: remind Claude to update plan.md only when session had actual changes

input=$(cat)
stop_hook_active=$(echo "$input" | python3 -c "import sys,json; d=json.load(sys.stdin); print(str(d.get('stop_hook_active',False)).lower())" 2>/dev/null || echo "false")
if [[ "$stop_hook_active" = "true" ]]; then
  exit 0
fi

PLAN="plans/plan.md"
if [ ! -f "$PLAN" ]; then
  exit 0
fi

# Skip if no changes this session: clean working tree AND no commits in last 2 hours
has_uncommitted=$(git status --porcelain 2>/dev/null)
recent_commits=$(git log --since="2 hours ago" --oneline 2>/dev/null | head -1)

if [[ -z "$has_uncommitted" && -z "$recent_commits" ]]; then
  exit 0
fi

echo "Session ending. Update $PLAN before finishing: move completed tasks to 'Đã hoàn thành', add new TODOs, update Notes. Then commit and push plan.md." >&2
exit 2
