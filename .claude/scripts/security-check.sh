#!/bin/bash
# PreToolUse security hook: block commands that exfiltrate data or load untrusted content

input=$(cat)
tool_name=$(echo "$input" | jq -r '.tool_name // ""')

# Only check Bash tool
if [[ "$tool_name" != "Bash" ]]; then
  exit 0
fi

command=$(echo "$input" | jq -r '.tool_input.command // ""')

# ── 1. Pipe to shell (remote code execution) ─────────────────────────────────
# Block: curl ... | bash, wget ... | sh, etc.
if echo "$command" | grep -qE '(curl|wget|fetch|nc|cat|python|node)\b.*\|\s*(bash|sh|zsh|fish|dash)\b'; then
  echo "BLOCKED [security]: pipe-to-shell detected — remote code execution risk." >&2
  echo "  Command: $command" >&2
  exit 2
fi

# Block: eval with network fetch
if echo "$command" | grep -qE 'eval\s*\$\(\s*(curl|wget|fetch)\b'; then
  echo "BLOCKED [security]: eval + network fetch — remote code execution risk." >&2
  echo "  Command: $command" >&2
  exit 2
fi

# Block: source/. with remote URL or /tmp script
if echo "$command" | grep -qE '(source|\.)\s+(https?://|/tmp/)'; then
  echo "BLOCKED [security]: sourcing remote/temp script — untrusted content risk." >&2
  echo "  Command: $command" >&2
  exit 2
fi

# ── 2. Data exfiltration via curl/wget ───────────────────────────────────────
# Block curl/wget sending data to non-localhost URLs
# Allow: localhost, 127.0.0.1, ::1
if echo "$command" | grep -qE '(curl|wget)\b'; then
  if echo "$command" | grep -qE '(-X\s*(POST|PUT|PATCH|DELETE)|--data|-d\s|--data-raw|--data-binary|--upload-file|--post-data)\b'; then
    if ! echo "$command" | grep -qE '(localhost|127\.0\.0\.1|::1|0\.0\.0\.0)'; then
      echo "BLOCKED [security]: curl/wget sending data to external host — data exfiltration risk." >&2
      echo "  Command: $command" >&2
      exit 2
    fi
  fi
fi

# ── 3. System package manager installs ───────────────────────────────────────
if echo "$command" | grep -qE '(^|\s)(sudo\s+)?(apt-get|apt|yum|dnf|pacman|apk|zypper)\s+install\b'; then
  echo "BLOCKED [security]: system package install — untrusted content risk." >&2
  echo "  Command: $command" >&2
  exit 2
fi

# ── 4. Python/Ruby/Perl package installs from internet ───────────────────────
if echo "$command" | grep -qE '(^|\s)pip[23]?\s+install\b'; then
  echo "BLOCKED [security]: pip install — use requirements.txt with pinned versions." >&2
  echo "  Command: $command" >&2
  exit 2
fi

if echo "$command" | grep -qE '(^|\s)(gem|bundle)\s+install\b'; then
  echo "BLOCKED [security]: gem/bundle install — untrusted content risk." >&2
  echo "  Command: $command" >&2
  exit 2
fi

# ── 5. Executing downloaded files directly ───────────────────────────────────
# Block: curl/wget download then execute in same command
if echo "$command" | grep -qE '(curl|wget)\b.*(-o|-O)\b.*&&.*(bash|sh|node|python|ruby|perl|chmod\s*\+x)'; then
  echo "BLOCKED [security]: download then execute — untrusted content risk." >&2
  echo "  Command: $command" >&2
  exit 2
fi

# Block: chmod +x on /tmp files then execute
if echo "$command" | grep -qE 'chmod\s*\+x\s*/tmp/'; then
  echo "BLOCKED [security]: making temp file executable — untrusted content risk." >&2
  echo "  Command: $command" >&2
  exit 2
fi

# ── 6. Netcat / reverse shell patterns ───────────────────────────────────────
if echo "$command" | grep -qE '\b(nc|ncat|netcat)\b.*(-e|-c|--exec)\b'; then
  echo "BLOCKED [security]: netcat with exec — reverse shell risk." >&2
  echo "  Command: $command" >&2
  exit 2
fi

exit 0
