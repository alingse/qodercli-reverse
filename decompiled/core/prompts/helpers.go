// Package prompts Helper functions
package prompts

import (
	"fmt"
	"strings"
)

// ========== Prompt option types ==========

// PromptOptions Prompt options (for quick building)
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

// ========== Convenience retrieval functions ==========

// CoreInstructions Gets core instructions
func CoreInstructions() string {
	return coreInstructions
}

// ToolRules Gets tool rules (with variable substitution)
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

// FileOperationRules Gets file operation rules
func FileOperationRules() string {
	return fileOperationRules
}

// EducationalRules Gets educational rules
func EducationalRules() string {
	return educationalRules
}

// SecurityRules Gets security rules
func SecurityRules() string {
	return coreSecurityRules
}

// OutputFormatRules Gets output format rules
func OutputFormatRules() string {
	return outputFormatRules
}

// QoderWorkRules Gets QoderWork rules
func QoderWorkRules() string {
	return qoderWorkRules
}

// QoderStudioRules Gets Qoder Studio rules
func QoderStudioRules() string {
	return qoderStudioRules
}

// ========== Quick build functions ==========

// QuickMainAgent Quickly builds the main Agent prompt
func QuickMainAgent(appName string) (string, error) {
	vars := DefaultTemplateVars()
	if appName != "" {
		vars.AppName = appName
	}
	return GetRendered(PromptTypeMainAgent, vars)
}

// QuickSubagent Quickly builds a sub-agent prompt
func QuickSubagent(subagentType string) (string, error) {
	vars := DefaultTemplateVars()
	return GetSubagentPrompt(subagentType, vars)
}

// QuickCustom Quickly builds a custom prompt
func QuickCustom(roleDefinition string, options PromptOptions) (string, error) {
	vars := DefaultTemplateVars()

	builder := NewSystemPromptBuilder(vars)

	// Role definition
	if roleDefinition != "" {
		builder.AddRoleDefinition(roleDefinition)
	} else {
		builder.AddRoleDefinition(fmt.Sprintf("You are %s, an interactive CLI tool.", vars.AppName))
	}

	// Core instructions
	if options.IncludeCoreInstructions {
		builder.AddCoreInstructions(CoreInstructions())
	}

	// Tool rules
	if options.IncludeToolRules {
		builder.AddToolRules(ToolRules(vars))
	}

	// File operation rules
	if options.IncludeFileRules {
		builder.AddFileOperationRules(FileOperationRules())
	}

	// Security rules
	if options.IncludeSecurityRules {
		builder.AddCustomSection("Security", SecurityRules())
	}

	// Output format
	if options.IncludeOutputFormat {
		builder.AddCustomSection("Output Format", OutputFormatRules())
	}

	// Extra sections
	for _, section := range options.ExtraSections {
		builder.AddCustomSection(section.Title, section.Content)
	}

	return builder.Build(), nil
}

// ========== System reminder tags ==========

// SystemReminder Generates a system-reminder tag
func SystemReminder(content string) string {
	return fmt.Sprintf("<system-reminder>\n%s\n</system-reminder>", content)
}

// SystemReminderVars Generates a system-reminder tag with variables
func SystemReminderVars(format string, args ...interface{}) string {
	return SystemReminder(fmt.Sprintf(format, args...))
}

// CommonSystemReminders Common system reminders
var CommonSystemReminders = struct {
	// File related
	FileTruncated func(file string, lines int) string
	FileTooShort  func(file string, offset, actual int) string
	FileEmpty     func(file string) string
	FileNotFound  func(file string) string

	// Image related
	ImageAttachment    func(count int) string
	ImageInAttachment  func(path string) string
	ImageProcessFailed func(path string) string

	// Command related
	CommandTimeout    func(timeout int) string
	CommandBackground func(pid int) string

	// Session related
	ContextTruncated func() string
	FolderCleared    func() string
	FolderUpdated    func(dirs string) string

	// Prompt updates
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

// ========== Core security rules (standalone) ==========

const coreSecurityRules = `IMPORTANT: Assist with defensive security tasks only.
Refuse to create, modify, or improve code that may be used maliciously.
Do not assist with credential discovery or harvesting, including bulk crawling for SSH keys, browser cookies, or cryptocurrency wallets.
Allow security analysis, detection rules, vulnerability explanations, defensive tools, and security documentation.

Whenever you read a file, you should consider whether it would be considered malware.
You CAN and SHOULD provide analysis of malware, what it is doing.
But you MUST refuse to improve or augment the code.
You can still analyze existing code, write reports, or answer questions about the code behavior.`

// ========== Output format rules (standalone) ==========

const outputFormatRules = `Your output will be displayed on a command line interface.
Your responses should be short and concise.
You can use Github-flavored markdown for formatting, and will be rendered in a monospace font using the CommonMark specification.

For clear communication with the user the assistant MUST avoid using emojis.

In your final response always share relevant file names and code snippets.
Any file paths you return in your response MUST be absolute. Do NOT use relative paths.`

// ========== Thinking triggers ==========

// ThinkTriggers Thinking mode trigger words (multilingual)
var ThinkTriggers = []string{
	`\bthink about it\b`,
	`\bthink intensely\b`,
	`\bthink very hard\b`,
	`\bthink hard\b`,
	`\bthink more\b`,
	`\bultrathink\b`,
	`\bdenk gründlich nach\b`,   // German
	`\bnachdenken\b`,            // German
	`\briflettere\b`,            // Italian
	`\bpensare profondamente\b`, // Italian
	`\bpensando\b`,              // Spanish/Portuguese
	`\bpiensa profundamente\b`,  // Spanish
	`\bpensare\b`,               // Italian
	`\bpensa molto\b`,           // Italian
	`\bthink a lot\b`,
	`\brichis\b`,
	`\brichi\b`,
}

// ShouldTriggerThinking Checks whether the text should trigger thinking mode
func ShouldTriggerThinking(text string) bool {
	lowerText := strings.ToLower(text)
	for _, trigger := range ThinkTriggers {
		// Simplified matching; regex should be used in practice
		cleanTrigger := strings.Trim(trigger, `\b`)
		if strings.Contains(lowerText, cleanTrigger) {
			return true
		}
	}
	return false
}
