package print

import (
	"encoding/json"
	"fmt"
	"strings"
)

// formatToolCallArgs formats tool call arguments
// Extracts and formats key params based on tool type
func formatToolCallArgs(toolName, arguments string) string {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return truncateString(arguments, 100)
	}

	switch toolName {
	case "Read":
		if filePath, ok := args["file_path"].(string); ok {
			return filePath
		}
	case "Write", "Edit":
		if filePath, ok := args["file_path"].(string); ok {
			return filePath
		}
	case "Bash":
		// Show actual command
		if cmd, ok := args["command"].(string); ok {
			return truncateString(cmd, 80)
		}
		// If command missing, fall back to description
		if desc, ok := args["description"].(string); ok && desc != "" {
			return desc
		}
	case "Glob":
		if pattern, ok := args["pattern"].(string); ok {
			return pattern
		}
	case "Grep":
		if pattern, ok := args["pattern"].(string); ok {
			return pattern
		}
	}

	return truncateString(arguments, 100)
}

// formatToolResult formats tool result
// Extracts and formats result summary based on tool type
func formatToolResult(toolName, content string, isError bool) string {
	if isError {
		return fmt.Sprintf("✗ Error: %s", truncateString(content, 100))
	}

	switch toolName {
	case "Read":
		lineCount := strings.Count(content, "\n")
		if lineCount > 0 {
			return fmt.Sprintf("Read %d lines", lineCount)
		}
		return "Read completed"

	case "Write", "Edit":
		return "✓ Completed"

	case "Bash":
		// Parse JSON result
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(content), &result); err == nil {
			if exitCode, ok := result["exit_code"].(float64); ok {
				if exitCode == 0 {
					return "✓ Success"
				}
				return fmt.Sprintf("✗ Exit code: %d", int(exitCode))
			}
		}
		return truncateString(content, 100)

	default:
		// Other tools: show first line or first 100 chars
		firstLine := content
		if idx := strings.Index(content, "\n"); idx > 0 && idx < 100 {
			firstLine = content[:idx]
		}
		return truncateString(firstLine, 100)
	}
}

// truncateString truncates string (UTF-8 safe)
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
