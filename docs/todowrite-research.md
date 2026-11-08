# TodoWrite Tool Reverse-Engineering Research Report

> Research target: TodoWrite tool implementation in the official qodercli-reverse v0.1.29 binary
> Research date: 2026-03-10
> Research tools: strings, go tool nm, go tool objdump

---

## 1. Core Findings

### 1.1 File Location
```
code.alibaba-inc.com/qoder-core/qodercli-reverse/core/agent/tools/todowrite.go
```

### 1.2 Main Type Definitions

#### 1.2.1 TodoWriteParams (Parameters Type)
```go
// TodoWriteParams TodoWrite tool parameters
type TodoWriteParams struct {
    Todos []state.Todo `json:"todos"`  // todo list
}
```

#### 1.2.2 TodoWriteResponseMetadata (Response Metadata)
```go
// TodoWriteResponseMetadata response metadata
type TodoWriteResponseMetadata struct {
    OldTodos []state.Todo `json:"oldTodos"`  // task list before the update
    NewTodos []state.Todo `json:"newTodos"`  // task list after the update
}
```

#### 1.2.3 Todo Struct (Defined in the state package)
```go
// code.alibaba-inc.com/qoder-core/qodercli-reverse/core/agent/state.Todo

type Todo struct {
    ID         string `json:"id"`         // unique task ID
    Content    string `json:"content"`    // task content description
    Status     string `json:"status"`     // task status
    ActiveForm string `json:"activeForm"` // in-progress description while executing (e.g., "Running tests")
}
```

#### 1.2.4 NotPresent (Permission Parameter Placeholder)
```go
// NotPresent means the tool does not require a permission parameter
type NotPresent struct{}
```

### 1.3 Tool Class Definition
```go
// todoWriteTool TodoWrite tool implementation
type todoWriteTool struct {
    // embed the Spec generic structure
    *Spec[TodoWriteParams, NotPresent, TodoWriteResponseMetadata]
    
    // TodoState interface
    todoState TodoState
}
```

---

## 2. Status Enum Values

According to strings extracted from the binary, the Todo Status field supports the following values:

| Status Value | JSON Label | Meaning |
|--------|-----------|------|
| `in_progress` | `"in_progress"` | In Progress |
| `completed` | `"completed"` | Completed |
| `pending` | `"pending"` | Pending |
| `done` | `"done"` | Done |
| `cancelled` | `"cancelled"` | Cancelled |

---

## 3. Main Methods

### 3.1 Constructor
```go
func NewTodoWriteTool(todoState TodoState) Tool
```

### 3.2 Tool Info Method
```go
func (t *todoWriteTool) Info() ToolInfo
```

### 3.3 Execution Method
```go
func (t *todoWriteTool) Run(ctx context.Context, params TodoWriteParams) (*ToolResult, error)
```

### 3.4 Validation Method
```go
func (t *todoWriteTool) validateTodos(todos []Todo) error
```

### 3.5 Auto-Repair Method
```go
func autoFixEmptyTodos(todos []Todo) []Todo
```

---

## 4. Key Implementation Rules

### 4.1 Empty Array Handling Rule
**Important**: An empty array must not be passed to clear the task list. When all tasks are complete, keep the tasks in the list and set their status to `"completed"`.

Invalid Example:
```json
{ "todos": [] }  // this is not allowed!
```

Valid Example:
```json
{
  "todos": [
    { "id": "1", "content": "Task 1", "status": "completed", "activeForm": "Completed task 1" }
  ]
}
```

### 4.2 ActiveForm Field Rules
- **Required field**: `activeForm` must not be empty
- **Purpose**: Displays the in-progress description while the task runs
- **Example**: 
  - Content: "Implement user authentication"
  - ActiveForm: "Implementing user authentication"

### 4.3 Validation Rules
Validation rules inferred from error messages:

```
todo[%d]: content cannot be empty        // content must not be empty
todo[%d]: invalid status '%s'            // status must be a valid value
todo[%d]: activeForm cannot be empty     // activeForm must not be empty
```

### 4.4 Auto-Repair Logic
When an empty array is detected, `autoFixEmptyTodos` automatically marks all tasks as `completed` instead of clearing the list.

---

## 5. Usage Scenarios (Extracted from the System Prompt)

### 5.1 When to Use TodoWrite
- **High-level tracking**: Track only high-level stages (requirements, design, review, coding, testing, delivery)
- **Complex tasks**: Break large features into manageable tasks
- **Visible progress**: Show the user the task progress

### 5.2 When Not to Use TodoWrite
- **Simple tasks**: Simple tasks that can be completed in a single operation
- **Exploratory work**: The exploration phase where the steps are uncertain

### 5.3 Language Alignment Rule
**All TodoList entries must be in the same language the user uses in the conversation.**

---

## 6. JSON Schema Inference

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

## 7. Implementation Suggestions

### 7.1 Interfaces to Implement
```go
// TodoState interface - manages todo state
type TodoState interface {
    LoadTodos() ([]Todo, error)
    SaveTodos(todos []Todo) error
    todosToText(todos []Todo) string  // convert todos to text description
}
```

### 7.2 Implementation Steps
1. Add the `Todo` struct and the `TodoState` interface in `core/agent/state/`
2. Create `todowrite.go` in `core/agent/tools/`
3. Implement the `todoWriteTool` struct and all methods
4. Register the TodoWrite tool when initializing the Agent
5. Integrate the TodoState implementation into the Agent's State

### 7.3 Key Validation Points
- [ ] Empty array auto-repaired to completed
- [ ] Validate all fields (content, status, activeForm)
- [ ] Status values must be one of the enum values
- [ ] Response includes OldTodos and NewTodos

---

## 8. Reference Symbols

```
code.alibaba-inc.com/qoder-core/qodercli-reverse/core/agent/tools.NewTodoWriteTool
code.alibaba-inc.com/qoder-core/qodercli-reverse/core/agent/tools.(*todoWriteTool).Info
code.alibaba-inc.com/qoder-core/qodercli-reverse/core/agent/tools.(*todoWriteTool).Run
code.alibaba-inc.com/qoder-core/qodercli-reverse/core/agent/tools.(*todoWriteTool).validateTodos
code.alibaba-inc.com/qoder-core/qodercli-reverse/core/agent/tools.autoFixEmptyTodos
code.alibaba-inc.com/qoder-core/qodercli-reverse/core/agent/state.(*defaultTodoState).todosToText
```

---

## 9. Summary

The TodoWrite tool is a **state management tool**, not an external invocation tool. Its core functionality is:

1. **Maintain the task list**: Persist task status in the Agent session
2. **State transitions**: Support updates and tracking of task status
3. **Auto repair**: Handle edge cases intelligently (e.g., empty arrays)
4. **User visible**: Show the current task progress via the TUI

Special attention should be paid during implementation to the empty-array handling logic and the rules for generating the activeForm field.
