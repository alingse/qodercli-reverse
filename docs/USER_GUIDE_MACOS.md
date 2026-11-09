# User Guide: Running qodercli-reverse on macOS

This guide covers installing, configuring, and using the reconstructed
`qodercli-reverse` CLI on macOS (Apple Silicon / arm64 and Intel / amd64).
The binary is a reverse-engineered build of Qoder CLI v0.1.30 and is provided
for learning and research only.

## 1. Prerequisites

- **macOS** 12 (Monterey) or later (tested on macOS 25.x, arm64).
- **Go 1.25+** (only if building from source): `brew install go`.
- **Git** (for cloning): `brew install git`.
- **A terminal** with UTF-8 support (Terminal.app, iTerm2, or WezTerm).
- **An LLM provider** (optional but recommended for AI features):
  - OpenAI-compatible endpoint, or
  - OpenRouter, or
  - any provider exposing an OpenAI chat-completions API.
- **GitHub token** (optional, only for GitHub integration): `GITHUB_TOKEN` / `GH_TOKEN`.

## 2. Installation

### Option A — One-command install (Makefile)

```bash
git clone git@github.com:sham435/qodercli-reverse.git
cd qodercli-reverse
make install            # builds + copies to ~/.local/bin
make doctor             # verify
```

Or with the install script:

```bash
./scripts/install.sh
```

### Option A2 — Build from source (manual)

```bash
# 1. Clone
git clone git@github.com:sham435/qodercli-reverse.git
cd qodercli-reverse

# 2. Build (production flags)
cd decompiled
VERSION="v0.1.30"
COMMIT="$(git rev-parse HEAD)"
BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
LDFLAGS="-X github.com/alingse/qodercli-reverse/decompiled/version.Version=${VERSION} \
         -X main.commit=${COMMIT} \
         -X main.buildTime=${BUILD_TIME}"
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o /tmp/qodercli-reverse .
# For Intel Macs use GOARCH=amd64

# 3. Install onto PATH
mkdir -p ~/.local/bin
cp /tmp/qodercli-reverse ~/.local/bin/qodercli-reverse
chmod +x ~/.local/bin/qodercli-reverse

# 4. Ensure PATH (add to ~/.zshrc if missing)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# 5. Verify
qodercli-reverse --version
```

Expected output:

```
qodercli-reverse version v0.1.30 (decompiled)
```

### Option B — Use a pre-built release tarball

If a release tarball is available (e.g. `qodercli-reverse_darwin_arm64.tar.gz`):

```bash
tar xzf qodercli-reverse_darwin_arm64.tar.gz
mkdir -p ~/.local/bin
mv qodercli-reverse ~/.local/bin/qodercli-reverse
chmod +x ~/.local/bin/qodercli-reverse
qodercli-reverse --version
```

### Option C — Homebrew (upstream cask, if published)

```bash
brew install --cask qodercli-reverse
qodercli-reverse --version
```

> Note: the cask in this repo points at the official Qoder download host and may
> not match this reconstructed build. Prefer Option A for this fork.

## 3. Configuration

### 3.1 LLM provider (required for AI features)

Set these environment variables in `~/.zshrc` (or `~/.bash_profile`):

```bash
# OpenAI (default base URL is https://api.openai.com/v1)
export OPENAI_API_KEY="sk-..."
# export OPENAI_BASE_URL="https://api.openai.com/v1"   # optional override
# export OPENAI_MODEL="gpt-4o"                        # optional

# OR OpenRouter
export OPENAI_API_KEY="sk-or-..."
export OPENAI_BASE_URL="https://openrouter.ai/api/v1"
export OPENAI_MODEL="openai/gpt-4o"

# OR any OpenAI-compatible endpoint
export OPENAI_API_KEY="..."
export OPENAI_BASE_URL="https://your-endpoint/v1"
```

Reload your shell: `source ~/.zshrc`.

Verify with:

```bash
qodercli-reverse doctor
```

Expected (healthy):

```
=== QoderCLI Reverse Doctor ===
API key present (OPENAI_API_KEY): yes
Base URL: https://openrouter.ai/api/v1
Model: auto
Provider creation: OK
Diagnostics complete.
```

> The provider key and base URL must match: an OpenAI key will not work against
> an OpenRouter base URL and vice-versa.

### 3.2 GitHub integration (optional)

```bash
export GITHUB_TOKEN="ghp_..."   # or GH_TOKEN
```

Used by `review-pr`, `assistant`, `run`, and the GitHub Actions helpers.

### 3.3 State / config directory

All local state lives in `~/.qodercli-reverse/` (isolated from official Qoder's
`~/.qoder` / `~/.qodercli`):

```
~/.qodercli-reverse/
├── qodercli-reverse.log    # default log file
└── ...                     # session/config state
```

Override the log file with `--log-file /path/to/file.log`. Enable debug with
`--debug`.

## 4. Running

### 4.1 Interactive TUI mode (default)

```bash
cd your-project
qodercli-reverse
```

Launches the Bubble Tea terminal UI. Type natural-language requests; the agent
will use tools (Read/Write/Edit/Grep/Bash/Glob/...) with permission prompts.

Keyboard shortcuts and slash commands (`/review-pr`, `/explain-code`, etc.) are
available inside the TUI; run `qodercli-reverse commands` to list them.

### 4.2 Non-interactive / print mode

```bash
# Single prompt, text output
qodercli-reverse -p "Summarize the README.md"

# JSON output
qodercli-reverse -p "List all Go files" -f json

# Stream-JSON (for CI/CD)
qodercli-reverse -p "Run the tests" -f stream-json

# Quiet (no spinner)
qodercli-reverse -p "..." -q

# Pick a model / tune generation
qodercli-reverse -p "..." --model auto --max-tokens 4096 --temperature 0.7
```

### 4.3 Common subcommands

```bash
qodercli-reverse doctor            # provider/environment diagnostics
qodercli-reverse commands          # list slash commands + subagents
qodercli-reverse commands --json   # JSON form

qodercli-reverse skill list        # list installed skills
qodercli-reverse skill install <path-or-source>
qodercli-reverse skill inspect <name>
qodercli-reverse skill search <query>
qodercli-reverse skill remove <name>

qodercli-reverse community sources # manage community skill sources

qodercli-reverse review-pr --repo owner/repo --pr 123
qodercli-reverse assistant --repo owner/repo --issue 42 --comment "@qoder fix" --post
qodercli-reverse run               # GitHub Actions compatible mode

qodercli-reverse --version         # show version
qodercli-reverse --help            # global help
```

### 4.4 Useful global flags

| Flag | Description |
|---|---|
| `--model` | Model selector (auto/efficient/.../ultimate) |
| `--max-tokens` | Max tokens per response (default 4096) |
| `--temperature` | Sampling temperature (default 0.7) |
| `--max-turns` | Max agent turns (default 25) |
| `--permission-mode` | ask / allow / deny (default ask) |
| `--system-prompt` | Override the system prompt |
| `--workspace` / `-w` | Working directory |
| `--continue` / `-c` | Continue last session |
| `--resume` / `-r` | Resume a specific session |
| `--allowed-tools` / `--disallowed-tools` | Tool allow/deny lists |
| `--dangerously-skip-permissions` / `--yolo` | Skip permission checks |
| `--debug` | Enable debug logging |
| `--log-file` | Log file path |
| `--worktree` | Start a concurrent task via git worktree |
| `--with-claude-config` | Load `.claude` configuration |

### 4.5 Worktree / concurrent task mode

```bash
qodercli-reverse --worktree --branch feature/x --path ../wt-x
```

## 5. First-run smoke test

```bash
# 1. Version
qodercli-reverse --version

# 2. Diagnostics
qodercli-reverse doctor

# 3. List commands
qodercli-reverse commands

# 4. A real (provider-dependent) round trip
qodercli-reverse -p "Reply with exactly: OK"
```

If step 4 fails with 401/402/410, check that `OPENAI_API_KEY` matches
`OPENAI_BASE_URL` (see §3.1). The CLI itself is healthy as long as steps 1–3
succeed.

## 6. Troubleshooting

| Symptom | Cause / Fix |
|---|---|
| `command not found: qodercli-reverse` | `~/.local/bin` not on PATH. Add `export PATH="$HOME/.local/bin:$PATH"` to `~/.zshrc` and `source ~/.zshrc`. |
| `API key present: no` | `OPENAI_API_KEY` not set in the current shell. Export it and reload. |
| `Provider creation: OK` but print mode 401 | Key/base-URL mismatch. Align them (OpenAI key ↔ OpenAI base URL). |
| `402` / credits error | Provider account out of credits. Top up or switch provider. |
| `410` / EOL model | Model retired. Set `OPENAI_MODEL` to a current model. |
| TUI renders garbled | Terminal lacks UTF-8 / color support. Use iTerm2 or set `LANG=en_US.UTF-8`. |
| `go: command not found` (build) | Install Go: `brew install go`. |
| Build fails on `go.mod` | Run `cd decompiled && go mod tidy` then rebuild. |
| Permission prompts block automation | Use `--permission-mode allow` or `--dangerously-skip-permissions` (use with care). |

## 7. Uninstall

```bash
rm -f ~/.local/bin/qodercli-reverse
rm -rf ~/.qodercli-reverse   # removes logs/state (optional)
```

## 8. Notes and limitations

- This is a **reconstructed** binary, not official Qoder source. Behavior may
  diverge; some features (session persistence, plan/goal mode, security scan)
  are partially implemented or missing — see
  `docs/feature-parity/official-qoder-cli-feature-matrix.md`.
- For learning and research only; not for production redistribution.
- Track upstream `alingse/qodercli-reverse` for improvements and PRs.

---

See also: `DEVELOPMENT.md` (build/test), `README.md` (project overview),
`docs/` (architecture), `milestone.md` (project history).
