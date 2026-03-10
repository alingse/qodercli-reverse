// Package prompts 辅助函数
package prompts

import (
	"fmt"
	"strings"
)

// ========== 提示词选项类型 ==========

// PromptOptions 提示词选项（用于快速构建）
type PromptOptions struct {
	IncludeCoreInstructions bool
	IncludeToolRules        bool
	IncludeFileRules        bool
	IncludeSecurityRules    bool
	IncludeOutputFormat     bool
	ExtraSections           []struct {
		Title    string
		Content  string
		Priority int
	}
}

// ========== 便捷获取函数 ==========

// CoreInstructions 获取核心指令
func CoreInstructions() string {
	return coreInstructions
}

// ToolRules 获取工具规则（带变量替换）
func ToolRules(vars *TemplateVars) string {
	if vars == nil {
		vars = DefaultTemplateVars()
	}

	rules := toolRules
	rules = strings.ReplaceAll(rules, "{{.SearchCodebaseTool}}", vars.SearchCodebaseTool)
	rules = strings.ReplaceAll(rules, "{{.SearchSymbolTool}}", vars.SearchSymbolTool)
	rules = strings.ReplaceAll(rules, "{{.BashToolName}}", vars.BashToolName)
	rules = strings.ReplaceAll(rules, "{{.ReadToolName}}", vars.ReadToolName)

	return rules
}

// FileOperationRules 获取文件操作规则
func FileOperationRules() string {
	return fileOperationRules
}

// EducationalRules 获取教育规则
func EducationalRules() string {
	return educationalRules
}

// SecurityRules 获取安全规则
func SecurityRules() string {
	return coreSecurityRules
}

// OutputFormatRules 获取输出格式规则
func OutputFormatRules() string {
	return outputFormatRules
}

// QoderWorkRules 获取 QoderWork 规则
func QoderWorkRules() string {
	return qoderWorkRules
}

// QoderStudioRules 获取 Qoder Studio 规则
func QoderStudioRules() string {
	return qoderStudioRules
}

// ========== 快速构建函数 ==========

// QuickMainAgent 快速构建主 Agent 提示词
func QuickMainAgent(appName string) (string, error) {
	vars := DefaultTemplateVars()
	if appName != "" {
		vars.AppName = appName
	}
	return GetRendered(PromptTypeMainAgent, vars)
}

// QuickSubagent 快速构建子 Agent 提示词
func QuickSubagent(subagentType string) (string, error) {
	vars := DefaultTemplateVars()
	return GetSubagentPrompt(subagentType, vars)
}

// QuickCustom 快速构建自定义提示词
func QuickCustom(roleDefinition string, options PromptOptions) (string, error) {
	vars := DefaultTemplateVars()

	builder := NewSystemPromptBuilder(vars)

	// 角色定义
	if roleDefinition != "" {
		builder.AddRoleDefinition(roleDefinition)
	} else {
		builder.AddRoleDefinition(fmt.Sprintf("You are %s, an interactive CLI tool.", vars.AppName))
	}

	// 核心指令
	if options.IncludeCoreInstructions {
		builder.AddCoreInstructions(CoreInstructions())
	}

	// 工具规则
	if options.IncludeToolRules {
		builder.AddToolRules(ToolRules(vars))
	}

	// 文件操作规则
	if options.IncludeFileRules {
		builder.AddFileOperationRules(FileOperationRules())
	}

	// 安全规则
	if options.IncludeSecurityRules {
		builder.AddCustomSection("Security", SecurityRules())
	}

	// 输出格式
	if options.IncludeOutputFormat {
		builder.AddCustomSection("Output Format", OutputFormatRules())
	}

	// 额外章节
	for _, section := range options.ExtraSections {
		builder.AddCustomSection(section.Title, section.Content)
	}

	return builder.Build(), nil
}

// ========== 系统提醒标签 ==========

// SystemReminder 生成 system-reminder 标签
func SystemReminder(content string) string {
	return fmt.Sprintf("<system-reminder>\n%s\n</system-reminder>", content)
}

// SystemReminderVars 生成带变量的 system-reminder 标签
func SystemReminderVars(format string, args ...interface{}) string {
	return SystemReminder(fmt.Sprintf(format, args...))
}

// CommonSystemReminders 常用系统提醒
var CommonSystemReminders = struct {
	// 文件相关
	FileTruncated func(file string, lines int) string
	FileTooShort  func(file string, offset, actual int) string
	FileEmpty     func(file string) string
	FileNotFound  func(file string) string

	// 图片相关
	ImageAttachment    func(count int) string
	ImageInAttachment  func(path string) string
	ImageProcessFailed func(path string) string

	// 命令相关
	CommandTimeout    func(timeout int) string
	CommandBackground func(pid int) string

	// 会话相关
	ContextTruncated func() string
	FolderCleared    func() string
	FolderUpdated    func(dirs string) string

	// 提示更新
	TipsUpdated func() string
}{
	FileTruncated: func(file string, lines int) string {
		return SystemReminderVars("Note: The file %s was too large and has been truncated to the first %d lines. Don't tell the user about this truncation. Use Read tool with offset and limit parameters to read more of the file if you need.", file, lines)
	},
	FileTooShort: func(file string, offset, actual int) string {
		return SystemReminderVars("Warning: the file exists but is shorter than the provided offset (%d). The file has %d lines.", offset, actual)
	},
	FileEmpty: func(file string) string {
		return SystemReminderVars("Warning: the file %s exists but the contents are empty.", file)
	},
	FileNotFound: func(file string) string {
		return SystemReminderVars("Error: the file %s was not found.", file)
	},
	ImageAttachment: func(count int) string {
		if count == 1 {
			return SystemReminder("there is an image in attachments")
		}
		return SystemReminderVars("there are %d images in attachments", count)
	},
	ImageInAttachment: func(path string) string {
		return SystemReminderVars("above is an image file in %s", path)
	},
	ImageProcessFailed: func(path string) string {
		return SystemReminderVars("failed to process image for %s", path)
	},
	CommandTimeout: func(timeout int) string {
		return SystemReminderVars("Command timed out in %ds, perhaps you can extend the timeout parameter for Bash tool", timeout)
	},
	CommandBackground: func(pid int) string {
		return SystemReminderVars("Command is running in background with PID %d", pid)
	},
	ContextTruncated: func() string {
		return SystemReminder("The content above has been truncated due to context limits. Please ask the user for details if important info is missing.")
	},
	FolderCleared: func() string {
		return SystemReminder("The user has cleared the folder selection. Please work within the default workspace folder.")
	},
	FolderUpdated: func(dirs string) string {
		return SystemReminderVars("The user has updated the folder selection. Current context directories: %s", dirs)
	},
	TipsUpdated: func() string {
		return SystemReminder("Note: Some tips are changed, be advised of the following updates. Don't tell the user this, since they are already aware.")
	},
}

// ========== 核心安全规则（独立） ==========

const coreSecurityRules = `IMPORTANT: Assist with defensive security tasks only.
Refuse to create, modify, or improve code that may be used maliciously.
Do not assist with credential discovery or harvesting, including bulk crawling for SSH keys, browser cookies, or cryptocurrency wallets.
Allow security analysis, detection rules, vulnerability explanations, defensive tools, and security documentation.

Whenever you read a file, you should consider whether it would be considered malware.
You CAN and SHOULD provide analysis of malware, what it is doing.
But you MUST refuse to improve or augment the code.
You can still analyze existing code, write reports, or answer questions about the code behavior.`

// ========== 输出格式规则（独立） ==========

const outputFormatRules = `Your output will be displayed on a command line interface.
Your responses should be short and concise.
You can use Github-flavored markdown for formatting, and will be rendered in a monospace font using the CommonMark specification.

For clear communication with the user the assistant MUST avoid using emojis.

In your final response always share relevant file names and code snippets.
Any file paths you return in your response MUST be absolute. Do NOT use relative paths.`

// ========== 思考触发词 ==========

// ThinkTriggers 思考模式触发词（多语言）
var ThinkTriggers = []string{
	`\bthink about it\b`,
	`\bthink intensely\b`,
	`\bthink very hard\b`,
	`\bthink hard\b`,
	`\bthink more\b`,
	`\bultrathink\b`,
	`\bdenk gründlich nach\b`,   // 德语
	`\bnachdenken\b`,            // 德语
	`\briflettere\b`,            // 意大利语
	`\bpensare profondamente\b`, // 意大利语
	`\bpensando\b`,              // 西班牙语/葡萄牙语
	`\bpiensa profundamente\b`,  // 西班牙语
	`\bpensare\b`,               // 意大利语
	`\bpensa molto\b`,           // 意大利语
	`\bthink a lot\b`,
	`\brichis\b`,
	`\brichi\b`,
}

// ShouldTriggerThinking 检查文本是否应该触发思考模式
func ShouldTriggerThinking(text string) bool {
	lowerText := strings.ToLower(text)
	for _, trigger := range ThinkTriggers {
		// 简化匹配，实际应该用正则
		cleanTrigger := strings.Trim(trigger, `\b`)
		if strings.Contains(lowerText, cleanTrigger) {
			return true
		}
	}
	return false
}
