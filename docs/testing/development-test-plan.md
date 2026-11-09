# Development Test Plan

Branch: feat/qoder-community-integration
Binary: .dev/qodercli (not installed)

## Preflight
- git status: clean (after .DS_Store cleanup)
- branches: 9 local branches, main stable
- open PRs: none

## Development build
- go mod verify: all modules verified
- go test ./...: pass
- go vet ./...: pass
- go build: success
- binary starts: --version prints "qodercli-reverse version 0.1.30 (decompiled)"
- --help lists commands: assistant, commands, community, completion, review-pr, run, skill

## Provider/AI connectivity
- Provider: OpenAI-compatible (OPENAI_API_KEY/OPENAI_BASE_URL)
- Auth path verified (401 when key mismatched, 402 on credits, 410 on EOL model)
- Successful end-to-end AI response blocked by provider credit/availability in this environment
- Streaming, tool calls, permissions code paths exercised via agent runtime

## Known limitations
- No funded provider available in this environment for a successful live AI response
- Feature branches not merged; main stable