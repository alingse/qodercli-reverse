package print

import (
	"encoding/json"
	"fmt"
	"strings"
)

// formatToolCallArgs 格式化工具调用参数
// 根据不同工具类型，提取并格式化关键参数
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
		// 显示实际命令
		if cmd, ok := args["command"].(string); ok {
			return truncateString(cmd, 80)
		}
		// 如果命令缺失，回退到描述
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

// formatToolResult 格式化工具结果
// 根据不同工具类型，提取并格式化结果摘要
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
		// 解析 JSON 结果
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
		// 其他工具：显示第一行或前100字符
		firstLine := content
		if idx := strings.Index(content, "\n"); idx > 0 && idx < 100 {
			firstLine = content[:idx]
		}
		return truncateString(firstLine, 100)
	}
}

// truncateString 截断字符串（UTF-8 安全）
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
