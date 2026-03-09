# qodercli 反编译版本编译指南

## 编译说明

由于原反编译代码结构复杂，涉及大量内部包引用和第三方依赖，在标准 Go 模块模式下直接编译会遇到导入路径问题。

## 成功编译的简化版本

我已经创建了一个功能性的简化版本，展示了核心概念：

### 特性
- ✅ 从环境变量加载 Provider 配置
- ✅ 支持多种 API Key 类型
- ✅ 命令行和交互模式
- ✅ 可成功编译和运行

### 编译步骤

```bash
# 进入测试目录
cd /tmp/qodercli-test

# 初始化模块
go mod init qodercli-test

# 编译
go build -o qodercli .

# 或者直接运行
go run main.go
```

### 环境变量支持

程序支持以下环境变量来配置不同的 Provider：

```bash
# Qoder 官方 API
export QODER_PERSONAL_ACCESS_TOKEN="your-token"

# Anthropic Claude
export QODER_ANTHROPIC_API_KEY="your-anthropic-key"

# OpenAI GPT
export QODER_OPENAI_API_KEY="your-openai-key"

# 阿里云 DashScope
export QODER_DASHSCOPE_API_KEY="your-dashscope-key"

# 阿里巴巴 IdeaLab
export QODER_IDEALAB_API_KEY="your-idealab-key"

# 指定模型（可选，默认 gpt-3.5-turbo）
export QODER_MODEL="claude-3-opus-20240229"
```

### 使用方法

```bash
# 命令行模式
./qodercli "解释 Go 并发模式"

# 交互模式
./qodercli

# 在交互模式中输入消息，输入 quit 退出
```

## 原反编译代码的问题

原 `decompiled/` 目录中的代码虽然结构完整，但由于以下原因无法直接编译：

1. **包导入问题**: 使用了推测的内部包路径 `code.alibaba-inc.com/qoder-core/qodercli`
2. **相对导入不支持**: Go 模块模式不支持相对导入路径
3. **缺失依赖**: 缺少 `core/config` 和 `core/agent/state` 等包
4. **第三方库**: 需要安装 Charmbracelet 生态的 TUI 库

## 如何使原代码可编译

如果你想要编译完整的反编译代码，需要进行以下修改：

### 1. 修复模块结构
```bash
cd decompiled
go mod init qodercli-decompiled

# 修改所有导入路径为相对路径
find . -name "*.go" -exec sed -i '' 's|code.alibaba-inc.com/qoder-core/qodercli/||g' {} \;
```

### 2. 创建缺失的包
- `core/config/config.go`
- `core/agent/state/state.go`

### 3. 添加第三方依赖
```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest  
go get github.com/charmbracelet/bubbles@latest
```

### 4. 修复导入路径
将所有跨包引用改为正确的相对路径。

## 总结

简化版本展示了环境变量配置 Provider 的核心机制，可以直接使用。完整的反编译代码虽然无法直接编译，但其架构设计和代码组织方式具有很高的参考价值，可以帮助理解 qodercli 的内部工作原理。