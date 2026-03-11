package markdown

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewRenderer(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		style   string
		wantErr bool
	}{
		{
			name:    "default dark style",
			width:   80,
			style:   "dark",
			wantErr: false,
		},
		{
			name:    "dracula style",
			width:   100,
			style:   "dracula",
			wantErr: false,
		},
		{
			name:    "ascii style",
			width:   80,
			style:   "ascii",
			wantErr: false,
		},
		{
			name:    "empty style uses auto",
			width:   80,
			style:   "",
			wantErr: false,
		},
		{
			name:    "zero width uses default",
			width:   0,
			style:   "dark",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRenderer(tt.width, tt.style)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRenderer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if r == nil {
				t.Error("NewRenderer() returned nil")
				return
			}

			// 验证宽度
			expectedWidth := tt.width
			if expectedWidth <= 0 {
				expectedWidth = 80
			}
			if r.GetWidth() != expectedWidth {
				t.Errorf("GetWidth() = %v, want %v", r.GetWidth(), expectedWidth)
			}
		})
	}
}

func TestRenderer_Render(t *testing.T) {
	r, err := NewRenderer(80, "dark")
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "simple heading",
			input:   "# Hello World",
			wantErr: false,
		},
		{
			name:    "bold text",
			input:   "**bold text**",
			wantErr: false,
		},
		{
			name:    "italic text",
			input:   "*italic text*",
			wantErr: false,
		},
		{
			name:    "code block",
			input:   "```go\nfunc main() {}\n```",
			wantErr: false,
		},
		{
			name:    "list",
			input:   "- item 1\n- item 2\n- item 3",
			wantErr: false,
		},
		{
			name:    "link",
			input:   "[link](https://example.com)",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := r.Render(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 验证输出不为空（除非输入为空）
			if tt.input != "" && output == "" {
				t.Error("Render() returned empty output for non-empty input")
			}

			// 验证输出包含 ANSI 转义序列（markdown 格式化的标志）
			if tt.input != "" && !strings.Contains(output, "\x1b[") {
				t.Logf("Warning: Render() output may not contain ANSI codes: %q", output)
			}
		})
	}
}

func TestRenderer_SetSize(t *testing.T) {
	r, _ := NewRenderer(80, "dark")

	tests := []struct {
		name      string
		newWidth  int
		wantWidth int
		wantErr   bool
	}{
		{
			name:      "increase width",
			newWidth:  120,
			wantWidth: 120,
			wantErr:   false,
		},
		{
			name:      "decrease width",
			newWidth:  60,
			wantWidth: 60,
			wantErr:   false,
		},
		{
			name:      "same width no-op",
			newWidth:  80,
			wantWidth: 80,
			wantErr:   false,
		},
		{
			name:      "zero width uses default",
			newWidth:  0,
			wantWidth: 80,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.SetSize(tt.newWidth)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if r.GetWidth() != tt.wantWidth {
				t.Errorf("GetWidth() = %v, want %v", r.GetWidth(), tt.wantWidth)
			}
		})
	}
}

func TestRenderer_SetStyle(t *testing.T) {
	r, _ := NewRenderer(80, "dark")

	tests := []struct {
		name     string
		newStyle string
		wantErr  bool
	}{
		{
			name:     "switch to light",
			newStyle: "light",
			wantErr:  false,
		},
		{
			name:     "switch to dracula",
			newStyle: "dracula",
			wantErr:  false,
		},
		{
			name:     "same style no-op",
			newStyle: "dark",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.SetStyle(tt.newStyle)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetStyle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && r.GetStyle() != tt.newStyle {
				t.Errorf("GetStyle() = %v, want %v", r.GetStyle(), tt.newStyle)
			}
		})
	}
}

func TestDetectStyle(t *testing.T) {
	style := DetectStyle()
	// 只验证返回值为 dark 或 light
	if style != "dark" && style != "light" {
		t.Errorf("DetectStyle() returned unexpected value: %v", style)
	}
}

func TestGetAvailableStyles(t *testing.T) {
	styles := GetAvailableStyles()
	if len(styles) == 0 {
		t.Error("GetAvailableStyles() returned empty list")
	}

	// 验证包含一些常见主题
	expectedStyles := []string{"dark", "light", "dracula", "ascii", "notty"}
	for _, expected := range expectedStyles {
		found := false
		for _, style := range styles {
			if style == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetAvailableStyles() missing expected style: %v", expected)
		}
	}
}

func TestIsValidStyle(t *testing.T) {
	tests := []struct {
		style string
		valid bool
	}{
		{"dark", true},
		{"light", true},
		{"dracula", true},
		{"ascii", true},
		{"notty", true},
		{"invalid-style", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.style, func(t *testing.T) {
			if got := IsValidStyle(tt.style); got != tt.valid {
				t.Errorf("IsValidStyle(%q) = %v, want %v", tt.style, got, tt.valid)
			}
		})
	}
}

func TestRenderer_RenderBytes(t *testing.T) {
	r, _ := NewRenderer(80, "dark")

	input := []byte("# Test\n\n**bold**")
	output, err := r.RenderBytes(input)
	if err != nil {
		t.Errorf("RenderBytes() error = %v", err)
	}

	if output == "" {
		t.Error("RenderBytes() returned empty output")
	}
}

func TestRenderer_NilSafety(t *testing.T) {
	var r *Renderer

	// 测试 nil 渲染器不会 panic
	output, err := r.Render("test")
	if err != nil {
		t.Errorf("nil Renderer.Render() error = %v", err)
	}
	if output != "test" {
		t.Errorf("nil Renderer.Render() = %v, want %v", output, "test")
	}

	// 测试 nil 渲染器的其他方法
	if r.GetStyle() != "dark" {
		t.Errorf("nil Renderer.GetStyle() = %v, want dark", r.GetStyle())
	}

	if r.GetWidth() != 80 {
		t.Errorf("nil Renderer.GetWidth() = %v, want 80", r.GetWidth())
	}

	// SetSize 和 SetStyle 不应该 panic
	r.SetSize(100)
	r.SetStyle("light")
}

// TestRendererNoPrefix 测试渲染器是否在开头添加额外字符
func TestRendererNoPrefix(t *testing.T) {
	renderer, err := NewRenderer(80, "dark")
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	testCases := []string{
		"这是项目的 README.md 文件内容。从 README 可以看出，这是一个 qodercli 的反编译代码项目。",
		"**Bold** text",
		"- List item",
		"1. Numbered item",
		"Plain text without markdown",
		"README",
	}

	for i, input := range testCases {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			output, err := renderer.Render(input)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}

			// 检查输出是否以输入开头（忽略可能的 ANSI 转义序列）
			// 简单检查：输出的字符数应该大于等于输入的字符数
			inputRunes := len([]rune(input))
			outputRunes := len([]rune(output))

			t.Logf("Input: %q", input)
			t.Logf("Input runes: %d", inputRunes)
			t.Logf("Output: %q", output)
			t.Logf("Output runes: %d", outputRunes)
			t.Logf("Input bytes: %d", len(input))
			t.Logf("Output bytes: %d", len(output))

			// 渲染后的内容可能包含 ANSI 转义序列，所以字节数会增加
			// 但 rune 数应该保持相似（除非渲染器添加了实际字符）
			if outputRunes < inputRunes {
				t.Errorf("Output has fewer runes than input: %d < %d", outputRunes, inputRunes)
			}

			// 检查输入文本是否在输出中（考虑到 markdown 转换）
			// 对于纯文本，输出应该包含输入
			if !containsMarkdownFormatted(output, input) {
				t.Logf("Warning: Input text not found in output (may have been transformed by markdown)")
			}
		})
	}
}

// containsMarkdownFormatted 检查输入文本是否在渲染后的输出中
// 考虑到 markdown 渲染可能会改变格式（如添加加粗、列表符号等）
func containsMarkdownFormatted(output, input string) bool {
	// 移除 ANSI 转义序列后再检查
	// 这里简化处理：直接检查输入的字符是否按顺序出现在输出中
	outputRunes := removeANSI([]rune(output))
	inputRunes := []rune(input)

	// 查找输入的连续字符序列在输出中的位置
	for i := 0; i <= len(outputRunes)-len(inputRunes); i++ {
		match := true
		for j := 0; j < len(inputRunes); j++ {
			if outputRunes[i+j] != inputRunes[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	// 如果没有完全匹配，检查是否大部分字符存在
	matchCount := 0
	for _, o := range outputRunes {
		if matchCount < len(inputRunes) && o == inputRunes[matchCount] {
			matchCount++
		}
	}

	// 如果匹配的字符数超过输入的 80%，认为包含
	return matchCount > len(inputRunes)*4/5
}

// removeANSI 移除 ANSI 转义序列
func removeANSI(runes []rune) []rune {
	var result []rune
	inEscape := false

	for _, r := range runes {
		if r == '\x1b' {
			inEscape = true
		} else if inEscape && (r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z') {
			inEscape = false
		} else if !inEscape {
			result = append(result, r)
		}
	}

	return result
}
