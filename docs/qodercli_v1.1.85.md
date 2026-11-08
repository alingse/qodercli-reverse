# qodercli v1.1.85 — Observed CLI Surface

This note captures the command-line surface of a locally installed `qodercli`
v1.1.85 (Mach-O 64-bit arm64), collected via `qodercli --help` and
`qodercli --version`.

## Version

```
Version: 1.1.85
```

## Top-level usage

```
Usage: qodercli [options] [command] [query...]

Qoder CLI - Defaults to interactive mode. Use -p/--print for non-interactive
output.
```

## Commands

- `hooks|hook` — Manage hooks
- `update [options]` — Update to the latest version
- `rollback [options]` — Rollback to a previous version
- `plugins|plugin` — Manage plugins
- `mcp` — Configure and manage MCP servers
- `commit [options]` — Generate a commit message and commit changes
- `agents|agent [options]` — Manage agents
- `skills|skill` — Manage agent skills
- `feedback [options]` — Submit feedback
- `security` — Scan source code for security issues
- `login` — Sign in to your account
- `status [options]` — Show session status
- `remote-control [options]` — Start the remote-control daemon
- `sites [args...]` — Build, publish and manage Qoder Sites
- `wiki [args...]` — Generate wiki documentation for projects

## Notable options

- `-p, --print` — Print response and exit (non-interactive)
- `-m, --model <model>` — Model for the current session
- `-w, --cwd <dir>` — Change working directory before startup
- `--worktree [name]` — Start in a new git worktree
- `--permission-mode <mode>` — default | accept_edits | bypass_permissions | dont_ask | auto
- `--remote [task]` — Create a cloud remote session and print its access URL
- `-c, --continue` / `-r, --resume [id]` — Session resume
- `--mcp-config <config>` — Load MCP servers from JSON

## commit subcommand

```
Usage: qodercli commit [options]
  -m, --message <message>      Commit message (required without --hook)
  -w, --workspace <workspace>  Working directory (defaults to current directory)
  --hook                       Hook mode: skip git commit, only verify AI stats
                               for the latest commit
```

## Notes

- Account state: requires `qodercli login` (browser device flow).
- This document is an observation of the published binary's CLI surface and
  is provided for research/documentation purposes only.
