# TodoWrite 工具逆向调研报告

> 调研目标: qodercli v0.1.29 官方二进制中的 TodoWrite 工具实现
> 调研日期: 2026-03-10
> 调研工具: strings, go tool nm, go tool objdump

---

## 一、核心发现

### 1.1 文件位置
```
code.alibaba-inc.com/qoder-core/qodercli/core/agent/tools/todowrite.go
```

### 1.2 主要类型定义

#### 1.2.1 TodoWriteParams (参数类型)
```go
// TodoWriteParams TodoWrite 工具参数
type TodoWriteParams struct {
    Todos []state.Todo `json:"todos"`  // 任务列表
}
```

#### 1.2.2 TodoWriteResponseMetadata (响应元数据)
```go
// TodoWriteResponseMetadata 响应元数据
type TodoWriteResponseMetadata struct {
    OldTodos []state.Todo `json:"oldTodos"`  // 更新前的任务列表
    NewTodos []state.Todo `json:"newTodos"`  // 更新后的任务列表
}
```

#### 1.2.3 Todo 结构体 (定义在 state 包)
```go
// code.alibaba-inc.com/qoder-core/qodercli/core/agent/state.Todo

type Todo struct {
    ID         string `json:"id"`         // 任务唯一标识
    Content    string `json:"content"`    // 任务内容描述
    Status     string `json:"status"`     // 任务状态
    ActiveForm string `json:"activeForm"` // 执行时的进行态描述 (如 "Running tests")
}
```

#### 1.2.4 NotPresent (权限参数占位)
```go
// NotPresent 表示该工具不需要权限参数
type NotPresent struct{}
```

### 1.3 工具类定义
```go
// todoWriteTool TodoWrite 工具实现
type todoWriteTool struct {
    // 嵌入 Spec 泛型结构
    *Spec[TodoWriteParams, NotPresent, TodoWriteResponseMetadata]
    
    // TodoState 接口
    todoState TodoState
}
```

---

## 二、状态枚举值

根据二进制字符串提取，Todo 的 Status 字段支持以下值:

| 状态值 | JSON 标签 | 含义 |
|--------|-----------|------|
| `in_progress` | `"in_progress"` | 进行中 |
| `completed` | `"completed"` | 已完成 |
| `pending` | `"pending"` | 待处理 |
| `done` | `"done"` | 完成 |
| `cancelled` | `"cancelled"` | 已取消 |

---

## 三、主要方法

### 3.1 构造函数
```go
func NewTodoWriteTool(todoState TodoState) Tool
```

### 3.2 工具信息方法
```go
func (t *todoWriteTool) Info() ToolInfo
```

### 3.3 执行方法
```go
func (t *todoWriteTool) Run(ctx context.Context, params TodoWriteParams) (*ToolResult, error)
```

### 3.4 验证方法
```go
func (t *todoWriteTool) validateTodos(todos []Todo) error
```

### 3.5 自动修复方法
```go
func autoFixEmptyTodos(todos []Todo) []Todo
```

---

## 四、关键实现规则

### 4.1 空数组处理规则
**重要**: 绝不允许传递空数组来清除任务列表。当所有任务完成时，应保持任务在列表中，并将状态设置为 `"completed"`。

错误示例:
```json
{ "todos": [] }  // 这是不允许的！
```

正确示例:
```json
{
  "todos": [
    { "id": "1", "content": "任务1", "status": "completed", "activeForm": "完成任务1" }
  ]
}
```

### 4.2 ActiveForm 字段规则
- **必需字段**: `activeForm` 不能为空
- **用途**: 在执行任务时显示进行中的描述
- **示例**: 
  - Content: "Implement user authentication"
  - ActiveForm: "Implementing user authentication"

### 4.3 验证规则
根据错误信息推断的验证规则:

```
todo[%d]: content cannot be empty        // content 不能为空
todo[%d]: invalid status '%s'            // status 必须是有效值
todo[%d]: activeForm cannot be empty     // activeForm 不能为空
```

### 4.4 自动修复逻辑
当检测到空数组时，`autoFixEmptyTodos` 会自动将所有任务标记为 `completed`，而不是清空列表。

---

## 五、使用场景提示 (从 System Prompt 提取)

### 5.1 何时使用 TodoWrite
- **高阶段追踪**: 只追踪高级阶段（需求、设计、审查、编码、测试、交付）
- **复杂任务**: 将大功能分解为可管理的任务
- **进度可见**: 让用户看到任务进展

### 5.2 何时不使用 TodoWrite
- **简单任务**: 单次操作就能完成的简单任务
- **探索性工作**: 不确定步骤的探索阶段

### 5.3 语言对齐规则
**所有 TodoList 条目的语言必须与用户在对话中使用的语言一致。**

---

## 六、JSON Schema 推断

```json
{
  "type": "object",
  "properties": {
    "todos": {
      "type": "array",
      "description": "The updated todo list",
      "items": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "description": "The ID of the TODO item"
          },
          "content": {
            "type": "string",
            "description": "The content of the TODO item"
          },
          "status": {
            "type": "string",
            "enum": ["in_progress", "completed", "pending", "done", "cancelled"],
            "description": "The status of the TODO item"
          },
          "activeForm": {
            "type": "string",
            "description": "The activeForm of the TODO item - present continuous form shown during execution (e.g., 'Running tests')"
          }
        },
        "required": ["id", "content", "status", "activeForm"]
      }
    }
  },
  "required": ["todos"]
}
```

---

## 七、实现建议

### 7.1 需要实现的接口
```go
// TodoState 接口 - 管理待办事项状态
type TodoState interface {
    LoadTodos() ([]Todo, error)
    SaveTodos(todos []Todo) error
    todosToText(todos []Todo) string  // 将 todos 转换为文本描述
}
```

### 7.2 实现步骤
1. 在 `core/agent/state/` 中添加 `Todo` 结构体和 `TodoState` 接口
2. 在 `core/agent/tools/` 中创建 `todowrite.go`
3. 实现 `todoWriteTool` 结构体和所有方法
4. 在 Agent 初始化时注册 TodoWrite 工具
5. 在 Agent 的 State 中集成 TodoState 实现

### 7.3 关键验证点
- [ ] 空数组自动修复为全部 completed
- [ ] 所有字段验证 (content, status, activeForm)
- [ ] 状态值必须是枚举之一
- [ ] 响应包含 OldTodos 和 NewTodos

---

## 八、参考符号表

```
code.alibaba-inc.com/qoder-core/qodercli/core/agent/tools.NewTodoWriteTool
code.alibaba-inc.com/qoder-core/qodercli/core/agent/tools.(*todoWriteTool).Info
code.alibaba-inc.com/qoder-core/qodercli/core/agent/tools.(*todoWriteTool).Run
code.alibaba-inc.com/qoder-core/qodercli/core/agent/tools.(*todoWriteTool).validateTodos
code.alibaba-inc.com/qoder-core/qodercli/core/agent/tools.autoFixEmptyTodos
code.alibaba-inc.com/qoder-core/qodercli/core/agent/state.(*defaultTodoState).todosToText
```

---

## 九、总结

TodoWrite 工具是一个**状态管理工具**，而非外部调用工具。它的核心功能是:

1. **维护任务列表**: 在 Agent 会话中持久化任务状态
2. **状态流转**: 支持任务状态的更新和追踪
3. **自动修复**: 智能处理边界情况（如空数组）
4. **用户可见**: 通过 TUI 展示当前任务进度

实现时需要特别注意空数组的处理逻辑和 activeForm 字段的生成规则。
