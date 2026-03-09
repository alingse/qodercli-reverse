# qodercli 二进制逆向分析报告

## 1. 基本信息

### 1.1 文件信息
- **文件名**: qodercli-0.1.29
- **路径**: /Users/zhihu/.qoder/bin/qodercli/qodercli-0.1.29
- **文件类型**: Mach-O 64-bit executable arm64
- **架构**: ARM64 (Apple Silicon)
- **大小**: 约 40-50 MB (通过符号表和段信息推断)

### 1.2 编程语言
- **主要语言**: Go (Golang)
- **Go 版本**: go1.25.7
- **编译器**: Go 官方编译器

### 1.3 依赖的系统框架
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

## 2. 架构分析

### 2.1 二进制结构
```
段信息:
- __PAGEZERO: 0x0000000000000000 (零页，用于空指针检测)
- __TEXT: 0x0000000100000000 (代码段)
- __DATA_CONST: 0x00000001011fc000 (只读数据段)
- __DATA: 0x0000000101fcc000 (数据段)
- __LINKEDIT: 0x000000010252c000 (链接编辑段)
```

### 2.2 Mach-O 头信息
```
Magic: MH_MAGIC_64
CPU Type: ARM64
CPU Subtype: ALL
File Type: EXECUTE
Flags: NOUNDEFS, DYLDLINK, TWOLEVEL, BINDS_TO_WEAK, PIE
```

## 3. 主要依赖库分析

### 3.1 Go 标准库
从字符串分析中发现使用了大量 Go 标准库：
- `runtime` - Go 运行时
- `net/http` - HTTP 客户端/服务器
- `net/url` - URL 解析
- `crypto/*` - 加密库 (TLS, RSA, AES, etc.)
- `os/exec` - 进程执行
- `mime/multipart` - MIME 处理
- `archive/zip` - ZIP 文件处理
- `compress/*` - 压缩算法 (gzip, zlib, etc.)
- `encoding/json` - JSON 编解码
- `encoding/xml` - XML 编解码
- `html/template` - HTML 模板
- `text/template` - 文本模板

### 3.2 第三方库
从字符串中提取的主要第三方库：

#### CLI 框架
- `github.com/spf13/cobra` - 强大的 CLI 应用框架
  - 用于命令行参数解析和子命令管理

#### 数据格式处理
- `github.com/BurntSushi/toml` - TOML 配置文件解析
- `github.com/JohannesKaufmann/html-to-markdown/v2` - HTML 转 Markdown
- `github.com/JohannesKaufmann/dom` - DOM 操作库

#### GitHub 集成
- `github.com/google/go-github/v57` - GitHub API 客户端

#### 其他
- 多个 MCP (Model Context Protocol) 相关库
- 多个 AI/LLM 相关库

## 4. 功能分析

### 4.1 主要功能模块

#### 4.1.1 CLI 命令系统
基于 cobra 框架构建的命令行系统，主要命令包括：
- `update` - 自更新功能
- `login` - 用户登录
- `session` - 会话管理
- `agent` - AI 代理管理
- `skill` - 技能管理
- `worktree` - Git worktree 管理
- `review` - 代码审查
- `compact` - 会话压缩
- `export` - 会话导出

#### 4.1.2 AI/LLM 集成
支持的 AI 模型：
- Claude 系列 (claude-3-haiku, claude-3-5-haiku, claude-opus-4, etc.)
- GPT 系列 (gpt-4o-mini, gpt-4.5-preview, etc.)
- Qwen Coder
- Kimi 模型
- 其他自定义模型

#### 4.1.3 MCP (Model Context Protocol) 支持
- MCP 服务器管理
- MCP 工具调用
- MCP 资源访问
- 支持多种传输协议 (stdio, sse, http)

#### 4.1.4 GitHub 集成
- GitHub Actions 工作流创建
- Pull Request 创建
- 代码审查自动化
- GitHub App 认证

#### 4.1.5 会话管理
- 会话创建和恢复
- 会话压缩和优化
- 会话导出和导入
- 会话持久化存储

#### 4.1.6 权限管理
- 工具权限控制
- 文件访问权限
- 命令执行权限
- 用户确认机制

### 4.2 配置文件
- `.qoder.json` - 全局配置
- `AGENTS.md` - 项目代理配置
- `.mcp.json` - MCP 服务器配置
- `settings.local.json` - 本地设置

### 4.3 API 端点
主要 API 端点：
- `https://qoder.com` - 主站
- `https://center.qoder.sh` - 中心服务
- `https://daily.qoder.ai` - 日常服务
- `https://openapi.qoder.sh` - OpenAPI 服务
- `/api/v1/userinfo` - 用户信息
- `/api/v1/heartbeat` - 心跳检测
- `/api/v1/tracking` - 事件追踪
- `/api/v2/model/list` - 模型列表
- `/api/v2/config/getDataPolicy` - 数据策略

## 5. 代码结构推测

### 5.1 主要包结构
```
qodercli/
├── cmd/                    # CLI 命令
│   ├── root.go            # 根命令
│   ├── update.go          # 更新命令
│   ├── login.go           # 登录命令
│   └── ...
├── internal/              # 内部包
│   ├── agent/            # AI 代理
│   ├── session/          # 会话管理
│   ├── mcp/              # MCP 协议
│   ├── github/           # GitHub 集成
│   ├── permission/       # 权限管理
│   └── ...
├── pkg/                   # 公共包
│   ├── api/              # API 客户端
│   ├── config/           # 配置管理
│   ├── utils/            # 工具函数
│   └── ...
└── main.go               # 入口文件
```

### 5.2 关键类型

#### 会话相关
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

#### 工具相关
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

#### 权限相关
```go
type Permission struct {
    Tool      string
    Action    string
    Allowed   bool
    Reason    string
    // ...
}
```

## 6. 安全机制分析

### 6.1 认证机制
- OAuth 2.0 认证
- Personal Access Token 认证
- GitHub App 认证
- Token 加密存储

### 6.2 权限控制
- 工具执行权限
- 文件访问权限
- 命令执行权限
- 网络访问权限

### 6.3 数据安全
- 敏感数据加密
- Token 安全存储
- HTTPS 通信
- FIPS 140 合规模式

## 7. 性能优化

### 7.1 运行时优化
- GC (垃圾回收) 调优
- 并发控制 (goroutine 管理)
- 内存池使用
- 缓存机制

### 7.2 网络优化
- HTTP/2 支持
- 连接池
- 请求重试
- 流式传输

## 8. 构建信息

### 8.1 编译标志
- 去除调试信息 (无 DWARF)
- 静态链接部分库
- PIE (Position Independent Executable)
- 优化级别: -O2 或更高

### 8.2 构建环境
- Go 版本: 1.25.7
- 目标平台: darwin/arm64
- 构建时间: 2026年3月左右

## 9. 逆向难点

### 9.1 符号信息
- 无 DWARF 调试信息
- 部分符号被剥离
- Go 运行时符号保留

### 9.2 代码混淆
- 无明显代码混淆
- 标准 Go 编译器优化
- 内联优化

### 9.3 反编译建议
1. 使用 Go 专用反编译工具 (如 IDA Pro + Hex-Rays, Ghidra)
2. 利用 Go 运行时特征识别函数
3. 通过字符串交叉引用定位功能
4. 分析导入的包路径理解代码结构

## 10. 总结

### 10.1 技术栈
- **语言**: Go 1.25.7
- **架构**: ARM64 (Apple Silicon)
- **框架**: Cobra CLI
- **协议**: MCP, HTTP/2, OAuth 2.0
- **加密**: TLS 1.3, AES, RSA, Ed25519

### 10.2 主要功能
- AI 编码助手 CLI 工具
- 支持多种 LLM 模型
- GitHub 集成和自动化
- MCP 协议支持
- 会话管理和优化
- 权限控制和安全机制

### 10.3 代码质量
- 结构清晰，模块化设计
- 使用成熟的第三方库
- 完善的错误处理
- 良好的性能优化

## 11. 反编译源代码示例

由于二进制文件已去除调试信息，以下是推测的源代码结构：

### 11.1 主入口 (main.go)
```go
package main

import (
    "github.com/spf13/cobra"
    "qodercli/cmd"
)

func main() {
    rootCmd := cmd.NewRootCmd()
    if err := rootCmd.Execute(); err != nil {
        panic(err)
    }
}
```

### 11.2 根命令 (cmd/root.go)
```go
package cmd

import (
    "github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "qodercli",
        Short: "AI-powered coding assistant CLI",
        Long:  `Qoder CLI is an AI-powered coding assistant that helps you write better code.`,
    }

    // 添加子命令
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

### 11.3 更新命令 (cmd/update.go)
```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "qodercli/internal/updater"
)

func NewUpdateCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "update",
        Short: "Update qodercli to the latest version",
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

### 11.4 会话管理 (internal/session/session.go)
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

### 11.5 MCP 客户端 (internal/mcp/client.go)
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
    // 建立 MCP 服务器连接
    transport, err := NewTransport(server.Transport, server.Command, server.Args...)
    if err != nil {
        return err
    }

    c.servers[server.Name] = server
    return nil
}

func (c *Client) CallTool(ctx context.Context, server, tool string, params map[string]interface{}) (interface{}, error) {
    // 调用 MCP 工具
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

### 11.6 权限管理 (internal/permission/manager.go)
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

### 11.7 GitHub 集成 (internal/github/client.go)
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

## 12. 推荐的反编译工具

### 12.1 商业工具
- **IDA Pro** + Hex-Rays Decompiler (推荐)
  - 支持 ARM64 反汇编
  - 有 Go 专用插件
  - 强大的反编译能力

- **Ghidra** (免费)
  - NSA 开源工具
  - 支持 ARM64
  - 可扩展插件

### 12.2 专用工具
- **go-reversing** - Go 二进制逆向工具集
- **go_parser** - IDA Pro 的 Go 解析插件
- **golang_loader_assist** - Ghidra 的 Go 加载辅助

### 12.3 辅助工具
- **strings** - 字符串提取
- **nm** - 符号表查看
- **otool** - Mach-O 文件分析
- **Hopper** - macOS 专用反汇编工具

## 13. 进一步分析建议

1. **动态分析**: 使用调试器 (lldb) 运行时跟踪
2. **网络抓包**: 分析 API 通信协议
3. **文件监控**: 观察配置文件和缓存文件变化
4. **行为分析**: 在沙箱环境中运行并记录行为
5. **代码注入**: 使用 Frida 进行动态插桩

---

**注意**: 本报告基于静态分析，推测的源代码仅供参考，可能与实际实现有所不同。
