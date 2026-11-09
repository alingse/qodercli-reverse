# Live AI Inference Plan

Branch: feat/qoder-community-integration

## Provider inspection

- Provider: OpenAI-compatible (OPENAI_API_KEY / OPENAI_BASE_URL / OPENAI_MODEL)
- Auth: Bearer token in Authorization header
- Streaming: SSE via OpenAI chat completions
- Tool calling: OpenAI function calling schema

## Diagnostic command

`qodercli-reverse doctor` reports provider configuration without exposing secrets.

## Live inference status

Blocked in this environment due to provider credits/model availability (HTTP 401/402/410 observed). Transport and request path verified.