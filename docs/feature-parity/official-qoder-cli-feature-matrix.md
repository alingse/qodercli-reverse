# Official Qoder CLI Feature Matrix

| Feature | Official documentation | Existing implementation | Status | Gap | Dependencies | Security | Tests | Priority |
|---|---|---|---|---|---|---|---|---|
| Natural language tasks | https://docs.qoder.com/cli/overview.md | decompiled/core/agent/agent/agent.go | Partial | Agent runtime exists; needs provider wiring | Provider, tools | Permission checks | go test | P0 |
| Plan mode | https://docs.qoder.com/cli/plan-mode.md | not implemented | Missing | No plan mode | Agent | Safe defaults | none | P1 |
| Goal mode | https://docs.qoder.com/cli/goal.md | not implemented | Missing | No goal mode | Agent | Limits | none | P1 |
| Headless mode | https://docs.qoder.com/cli/run-in-scripts.md | decompiled/cmd/print/print.go | Partial | stream-json added | CLI | Exit codes | manual | P0 |
| Sessions | https://docs.qoder.com/cli/sessions.md | decompiled/core/agent/state/ | Partial | Persistence incomplete | Storage | Retention | none | P1 |
| Skills | https://docs.qoder.com/cli/Skills.md | decompiled/core/skills/ | Partial | Community registry added | Network | Validation | skill_test | P1 |
| MCP | https://docs.qoder.com/cli/mcp-servers.md | decompiled/core/resource/mcp/mcp.go | Partial | Optional GitHub MCP | MCP config | Permissions | none | P1 |
| Security scan | https://docs.qoder.com/cli/security.md | not implemented | Missing | No scanner | Tools | Safe | none | P2 |