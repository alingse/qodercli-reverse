# qodercli GitHub 集成和 CI/CD 功能分析

## 1. GitHub 集成架构

### 1.1 包结构

```
core/utils/qoder/github/
└── GitHub API 集成

tui/components/command/github/
├── assets/              # GitHub Actions 模板
└── installer/           # 安装器
```

## 2. GitHub 认证

### 2.1 认证方式

1. **Qoder Token**: 通过 Qoder 服务代理
2. **GitHub App**: OAuth 应用集成
3. **Personal Access Token**: 直接使用 GitHub PAT

### 2.2 GitHub App 集成

```
ExchangeGitHubAppTokenParam
ExchangeGitHubAppTokenResponse
GetGitHubAppRepoStatusParam
GitHubAppRepoStatusResponse
```

## 3. GitHub CLI 命令

### 3.1 设置命令

```bash
/setup-github     # 设置 GitHub 集成
/github           # GitHub 操作
```

### 3.2 设置流程

```
1. 检查 GitHub App 安装状态
2. 获取仓库访问权限
3. 配置 GitHub Actions
4. 设置 Secrets
```

## 4. GitHub Actions 集成

### 4.1 工作流模板

```
.github/workflows/qoder-assistant.yml
github-workflows/qoder-assistant.yaml
```

### 4.2 所需 Secrets

```yaml
QODER_PERSONAL_ACCESS_TOKEN: ${{ secrets.QODER_PERSONAL_ACCESS_TOKEN }}
```

### 4.3 工作流功能

- **代码审查**: 自动 PR 审查
- **问题诊断**: 错误分析
- **持续集成**: 测试和构建

## 5. PR 集成

### 5.1 PR 创建

```
Preparing Pull Request
Creating new branch
```

### 5.2 PR 审查

```
code_review           # 代码审查
Code Review           # 审查模式
```

### 5.3 PR 相关 API

```
repos/%v/%v/git/refs          # Git 引用
repos/%v/%v/actions/secrets   # Actions Secrets
pull_request_review_comment   # PR 审查评论
pull_request_review_thread    # PR 审查线程
```

## 6. GitHub API 使用

### 6.1 依赖库

```
github.com/google/go-github/v57
```

### 6.2 主要功能

- 仓库操作
- PR 管理
- Issue 管理
- Actions 管理
- Secret 管理

## 7. CI/CD 功能

### 7.1 自动化任务

- **代码扫描**: 安全漏洞检测
- **测试运行**: 自动测试
- **构建检查**: 编译验证
- **部署**: 自动部署

### 7.2 事件处理

```
GITHUB_EVENT_PATH          # GitHub 事件路径
github_env_result_type     # 环境结果类型
github_action_task_start   # Action 任务开始
GITHUB_ACTIONS             # Actions 标识
```

## 8. Git 操作

### 8.1 常用操作

```bash
git status
git diff
git log
git add
git commit
git push
git checkout
git branch
```

### 8.2 Worktree 管理

```
qodercli --worktree --branch feature-x
```

创建隔离的 Git worktree 进行并行开发。

## 9. 代码审查功能

### 9.1 审查流程

```
1. 分析代码变更
2. 检查代码风格
3. 发现潜在问题
4. 生成审查意见
5. 发布 PR 评论
```

### 9.2 审查配置

```json
{
  "review": {
    "autoReview": true,
    "checkStyle": true,
    "securityScan": true
  }
}
```

## 10. 错误处理

### 10.1 常见错误

| 错误 | 说明 |
|------|------|
| `GitHub App Permission Error` | GitHub App 权限错误 |
| `Failed to get GitHub token` | 获取 Token 失败 |
| `repository not found` | 仓库未找到 |

### 10.2 错误恢复

- 重新认证
- 检查权限
- 验证仓库访问
