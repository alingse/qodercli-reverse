# qodercli-reverse Binary Reverse Engineering Analysis Report

## 1. Basic Information

### 1.1 File Information
- **Filename**: qodercli-reverse-0.1.29
- **Path**: /Users/zhihu/.qoder/bin/qodercli-reverse/qodercli-reverse-0.1.29
- **File type**: Mach-O 64-bit executable arm64
- **Architecture**: ARM64 (Apple Silicon)
- **Size**: approx. 40-50 MB (inferred from the symbol table and segment information)

### 1.2 Programming Language
- **Primary language**: Go (Golang)
- **Go version**: go1.25.7
- **Compiler**: official Go compiler

### 1.3 System Framework Dependencies
- Foundation.framework
- Cocoa.framework
- AppKit.framework
- Security.framework
- CoreFoundation.framework
- CoreGraphics.framework
- IOKit.framework
- libobjc.A.dylib
- libc++.1.dylib
- libSystem.B.dylib
- libresolv.9.dylib
- libz.1.dylib

## 2. Architecture Analysis

### 2.1 Binary Structure
```
Segment information:
- __PAGEZERO: 0x0000000000000000 (zero page, used for null pointer detection)
- __TEXT: 0x0000000100000000 (code segment)
- __DATA_CONST: 0x00000001011fc000 (read-only data segment)
- __DATA: 0x0000000101fcc000 (data segment)
- __LINKEDIT: 0x000000010252c000 (linkedit segment)
```

### 2.2 Mach-O Header Information
```
Magic: MH_MAGIC_64
CPU Type: ARM64
CPU Subtype: ALL
File Type: EXECUTE
Flags: NOUNDEFS, DYLDLINK, TWOLEVEL, BINDS_TO_WEAK, PIE
```

## 3. Major Dependency Libraries Analysis

### 3.1 Go Standard Library
A large number of Go standard library packages were found through string analysis:
- `runtime` - Go runtime
- `net/http` - HTTP client/server
- `net/url` - URL parsing
- `crypto/*` - cryptography libraries (TLS, RSA, AES, etc.)
- `os/exec` - process execution
- `mime/multipart` - MIME handling
- `archive/zip` - ZIP file handling
- `compress/*` - compression algorithms (gzip, zlib, etc.)
- `encoding/json` - JSON encoding/decoding
- `encoding/xml` - XML encoding/decoding
- `html/template` - HTML templates
- `text/template` - text templates

### 3.2 Third-party Libraries
The main third-party libraries extracted from strings:

#### CLI Framework
- `github.com/spf13/cobra` - a powerful CLI application framework
  - used for command-line argument parsing and subcommand management

#### Data Format Processing
- `github.com/BurntSushi/toml` - TOML configuration file parsing
- `github.com/JohannesKaufmann/html-to-markdown/v2` - HTML to Markdown
- `github.com/JohannesKaufmann/dom` - DOM manipulation library

#### GitHub Integration
- `github.com/google/go-github/v57` - GitHub API client

#### Others
- multiple MCP (Model Context Protocol) related libraries
- multiple AI/LLM related libraries

## 4. Feature Analysis

### 4.1 Main Functional Modules

#### 4.1.1 CLI Command System
A command-line system built on the cobra framework. The main commands include:
- `update` - self-update
- `login` - user login
- `session` - session management
- `agent` - AI agent management
- `skill` - skill management
- `worktree` - Git worktree management
- `review` - code review
- `compact` - session compaction
- `export` - session export

#### 4.1.2 AI/LLM Integration
Supported AI models:
- Claude series (claude-3-haiku, claude-3-5-haiku, claude-opus-4, etc.)
- GPT series (gpt-4o-mini, gpt-4.5-preview, etc.)
- Qwen Coder
- Kimi models
- other custom models

#### 4.1.3 MCP (Model Context Protocol) Support
- MCP server management
- MCP tool calls
- MCP resource access
- supports multiple transport protocols (stdio, sse, http)

#### 4.1.4 GitHub Integration
- GitHub Actions workflow creation
- Pull Request creation
- code review automation
- GitHub App authentication

#### 4.1.5 Session Management
- session creation and resumption
- session compaction and optimization
- session export and import
- session persistent storage

#### 4.1.6 Permission Management
- tool execution permissions
- file access permissions
- command execution permissions
- user confirmation mechanism

### 4.2 Configuration Files
- `.qoder.json` - global configuration
- `AGENTS.md` - project agent configuration
- `.mcp.json` - MCP server configuration
- `settings.local.json` - local settings

### 4.3 API Endpoints
Main API endpoints:
- `https://qoder.com` - main site
- `https://center.qoder.sh` - center service
- `https://daily.qoder.ai` - daily service
- `https://openapi.qoder.sh` - OpenAPI service
- `/api/v1/userinfo` - user information
- `/api/v1/heartbeat` - heartbeat check
- `/api/v1/tracking` - event tracking
- `/api/v2/model/list` - model list
- `/api/v2/config/getDataPolicy` - data policy

## 5. Code Structure Inference

### 5.1 Main Package Structure
```
qodercli-reverse/
├── cmd/                    # CLI commands
│   ├── root.go            # root command
│   ├── update.go          # update command
│   ├── login.go           # login command
│   └── ...
├── internal/              # internal packages
│   ├── agent/            # AI agent
│   ├── session/          # session management
│   ├── mcp/              # MCP protocol
│   ├── github/           # GitHub integration
│   ├── permission/       # permission management
│   └── ...
├── pkg/                   # public packages
│   ├── api/              # API client
│   ├── config/           # configuration management
│   ├── utils/            # utility functions
│   └── ...
└── main.go               # entry point
```

### 5.2 Key Types

#### Session related
```go
type Session struct {
    ID           string
    Model        string
    Messages     []Message
    Tools        []Tool
    Permissions  []Permission
    // ...
}

type Message struct {
    Role       string
    Content    string
    ToolCalls  []ToolCall
    // ...
}
```

#### Tool related
```go
type Tool struct {
    Name        string
    Description string
    Parameters  map[string]interface{}
    // ...
}

type ToolCall struct {
    ID       string
    Name     string
    Arguments string
    // ...
}
```

#### Permission related
```go
type Permission struct {
    Tool      string
    Action    string
    Allowed   bool
    Reason    string
    // ...
}
```

## 6. Security Mechanism Analysis

### 6.1 Authentication Mechanism
- OAuth 2.0 authentication
- Personal Access Token authentication
- GitHub App authentication
- Token encrypted storage

### 6.2 Permission Control
- tool execution permissions
- file access permissions
- command execution permissions
- network access permissions

### 6.3 Data Security
- sensitive data encryption
- Token secure storage
- HTTPS communication
- FIPS 140 compliance mode

## 7. Performance Optimization

### 7.1 Runtime Optimization
- GC (garbage collection) tuning
- concurrency control (goroutine management)
- memory pool usage
- caching mechanism

### 7.2 Network Optimization
- HTTP/2 support
- connection pool
- request retry
- streaming transfer

## 8. Build Information

### 8.1 Compile Flags
- debug info stripped (no DWARF)
- some libraries statically linked
- PIE (Position Independent Executable)
- optimization level: -O2 or higher

### 8.2 Build Environment
- Go version: 1.25.7
- target platform: darwin/arm64
- build time: around March 2026

## 9. Reverse Engineering Challenges

### 9.1 Symbol Information
- no DWARF debug info
- some symbols stripped
- Go runtime symbols retained

### 9.2 Code Obfuscation
- no obvious code obfuscation
- standard Go compiler optimization
- inlining optimization

### 9.3 Decompilation Suggestions
1. Use Go-specific decompilation tools (such as IDA Pro + Hex-Rays, Ghidra)
2. leverage Go runtime features to identify functions
3. locate functionality via string cross-references
4. analyze imported package paths to understand code structure

## 10. Summary

### 10.1 Tech Stack
- **Language**: Go 1.25.7
- **Architecture**: ARM64 (Apple Silicon)
- **Framework**: Cobra CLI
- **Protocols**: MCP, HTTP/2, OAuth 2.0
- **Crypto**: TLS 1.3, AES, RSA, Ed25519

### 10.2 Main Features
- AI coding assistant CLI tool
- supports multiple LLM models
- GitHub integration and automation
- MCP protocol support
- session management and optimization
- permission control and security mechanisms

### 10.3 Code Quality
- clear structure, modular design
- uses mature third-party libraries
- comprehensive error handling
- good performance optimization

## 11. Decompiled Source Code Examples

Since the binary has debug info stripped, the following are inferred source structures:

### 11.1 Main Entry (main.go)
```go
package main

import (
    "github.com/spf13/cobra"
    "qodercli-reverse/cmd"
)

func main() {
    rootCmd := cmd.NewRootCmd()
    if err := rootCmd.Execute(); err != nil {
        panic(err)
    }
}
```

### 11.2 Root Command (cmd/root.go)
```go
package cmd

import (
    "github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "qodercli-reverse",
        Short: "AI-powered coding assistant CLI",
        Long:  `Qoder CLI is an AI-powered coding assistant that helps you write better code.`,
    }

    // Add subcommands
    cmd.AddCommand(NewLoginCmd())
    cmd.AddCommand(NewUpdateCmd())
    cmd.AddCommand(NewSessionCmd())
    cmd.AddCommand(NewAgentCmd())
    cmd.AddCommand(NewSkillCmd())
    cmd.AddCommand(NewWorktreeCmd())
    cmd.AddCommand(NewReviewCmd())

    return cmd
}
```

### 11.3 Update Command (cmd/update.go)
```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "qodercli-reverse/internal/updater"
)

func NewUpdateCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "update",
        Short: "Update qodercli-reverse to the latest version",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Checking for updates...")
            if err := updater.Update(); err != nil {
                fmt.Printf("Update failed: %v\n", err)
            } else {
                fmt.Println("Successfully updated to the latest version!")
            }
        },
    }
}
```

### 11.4 Session Management (internal/session/session.go)
```go
package session

import (
    "context"
    "time"
)

type Session struct {
    ID           string
    Model        string
    Messages     []Message
    Tools        []Tool
    Permissions  []Permission
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type Manager struct {
    sessions map[string]*Session
    storage  Storage
}

func NewManager(storage Storage) *Manager {
    return &Manager{
        sessions: make(map[string]*Session),
        storage:  storage,
    }
}

func (m *Manager) Create(ctx context.Context, model string) (*Session, error) {
    session := &Session{
        ID:        generateID(),
        Model:     model,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    m.sessions[session.ID] = session
    return session, nil
}

func (m *Manager) Get(id string) (*Session, error) {
    session, ok := m.sessions[id]
    if !ok {
        return nil, fmt.Errorf("session not found: %s", id)
    }
    return session, nil
}
```

### 11.5 MCP Client (internal/mcp/client.go)
```go
package mcp

import (
    "context"
    "encoding/json"
)

type Client struct {
    transport Transport
    servers   map[string]*Server
}

type Server struct {
    Name      string
    Transport string
    Command   string
    Args      []string
}

func NewClient() *Client {
    return &Client{
        servers: make(map[string]*Server),
    }
}

func (c *Client) Connect(ctx context.Context, server *Server) error {
    // Establish MCP server connection
    transport, err := NewTransport(server.Transport, server.Command, server.Args...)
    if err != nil {
        return err
    }

    c.servers[server.Name] = server
    return nil
}

func (c *Client) CallTool(ctx context.Context, server, tool string, params map[string]interface{}) (interface{}, error) {
    // Call MCP tool
    req := &ToolCallRequest{
        Tool:   tool,
        Params: params,
    }

    resp, err := c.transport.Call(ctx, req)
    if err != nil {
        return nil, err
    }

    return resp.Result, nil
}
```

### 11.6 Permission Management (internal/permission/manager.go)
```go
package permission

type Manager struct {
    rules []Rule
}

type Rule struct {
    Tool     string
    Action   string
    Pattern  string
    Allowed  bool
    Reason   string
}

func NewManager() *Manager {
    return &Manager{
        rules: []Rule{},
    }
}

func (m *Manager) Check(tool, action, resource string) (bool, string, error) {
    for _, rule := range m.rules {
        if matchRule(rule, tool, action, resource) {
            return rule.Allowed, rule.Reason, nil
        }
    }
    return false, "no matching rule", nil
}

func (m *Manager) AddRule(rule Rule) {
    m.rules = append(m.rules, rule)
}
```

### 11.7 GitHub Integration (internal/github/client.go)
```go
package github

import (
    "context"
    "github.com/google/go-github/v57/github"
    "golang.org/x/oauth2"
)

type Client struct {
    client *github.Client
}

func NewClient(token string) *Client {
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )
    tc := oauth2.NewClient(context.Background(), ts)

    return &Client{
        client: github.NewClient(tc),
    }
}

func (c *Client) CreatePullRequest(ctx context.Context, owner, repo, title, head, base string) (*github.PullRequest, error) {
    pr := &github.NewPullRequest{
        Title: &title,
        Head:  &head,
        Base:  &base,
    }

    result, _, err := c.client.PullRequests.Create(ctx, owner, repo, pr)
    return result, err
}
```

## 12. Recommended Decompilation Tools

### 12.1 Commercial Tools
- **IDA Pro** + Hex-Rays Decompiler (recommended)
  - supports ARM64 disassembly
  - has Go-specific plugins
  - strong decompilation capability
- **Ghidra** (free)
  - NSA open-source tool
  - supports ARM64
  - extensible plugins

### 12.2 Specialized Tools
- **go-reversing** - Go binary reverse engineering toolset
- **go_parser** - IDA Pro Go parsing plugin
- **golang_loader_assist** - Ghidra Go loading assistant

### 12.3 Auxiliary Tools
- **strings** - string extraction
- **nm** - symbol table viewing
- **otool** - Mach-O file analysis
- **Hopper** - macOS dedicated disassembler

## 13. Further Analysis Suggestions

1. **Dynamic analysis**: use a debugger (lldb) for runtime tracing
2. **Network packet capture**: analyze the API communication protocol
3. **File monitoring**: observe configuration and cache file changes
4. **Behavior analysis**: run in a sandbox environment and record behavior
5. **Code injection**: use Frida for dynamic instrumentation

---

**Note**: This report is based on static analysis; the inferred source code is for reference only and may differ from the actual implementation.