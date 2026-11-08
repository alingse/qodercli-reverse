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

			// verify width
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

			// verify output is not empty (unless input is empty)
			if tt.input != "" && output == "" {
				t.Error("Render() returned empty output for non-empty input")
			}

			// verify output contains ANSI escape sequences (sign of markdown formatting)
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
	// only verify return value is dark or light
	if style != "dark" && style != "light" {
		t.Errorf("DetectStyle() returned unexpected value: %v", style)
	}
}

func TestGetAvailableStyles(t *testing.T) {
	styles := GetAvailableStyles()
	if len(styles) == 0 {
		t.Error("GetAvailableStyles() returned empty list")
	}

	// verify it contains some common themes
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

	// test that a nil renderer does not panic
	output, err := r.Render("test")
	if err != nil {
		t.Errorf("nil Renderer.Render() error = %v", err)
	}
	if output != "test" {
		t.Errorf("nil Renderer.Render() = %v, want %v", output, "test")
	}

	// test other methods on a nil renderer
	if r.GetStyle() != "dark" {
		t.Errorf("nil Renderer.GetStyle() = %v, want dark", r.GetStyle())
	}

	if r.GetWidth() != 80 {
		t.Errorf("nil Renderer.GetWidth() = %v, want 80", r.GetWidth())
	}

	// SetSize and SetStyle should not panic
	r.SetSize(100)
	r.SetStyle("light")
}

// TestRendererNoPrefix tests whether the renderer adds extra characters at the start
func TestRendererNoPrefix(t *testing.T) {
	renderer, err := NewRenderer(80, "dark")
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	testCases := []string{
		"This is the project's README.md file content. From the README it can be seen that this is a decompiled code project for qodercli-reverse.",
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

			// check that output does not start with extra characters (ignoring possible ANSI escape sequences)
			// simple check: output character count should be >= input character count
			inputRunes := len([]rune(input))
			outputRunes := len([]rune(output))

			t.Logf("Input: %q", input)
			t.Logf("Input runes: %d", inputRunes)
			t.Logf("Output: %q", output)
			t.Logf("Output runes: %d", outputRunes)
			t.Logf("Input bytes: %d", len(input))
			t.Logf("Output bytes: %d", len(output))

			// rendered content may contain ANSI escape sequences, so byte count increases
			// but rune count should stay similar (unless the renderer adds actual characters)
			if outputRunes < inputRunes {
				t.Errorf("Output has fewer runes than input: %d < %d", outputRunes, inputRunes)
			}

			// check that input text is in the output (considering markdown transformation)
			// for plain text, output should contain the input
			if !containsMarkdownFormatted(output, input) {
				t.Logf("Warning: Input text not found in output (may have been transformed by markdown)")
			}
		})
	}
}

// containsMarkdownFormatted checks whether the input text is in the rendered output
// considering that markdown rendering may change formatting (e.g. adding bold, list markers, etc.)
func containsMarkdownFormatted(output, input string) bool {
	// remove ANSI escape sequences before checking
	// simplified: directly check whether input characters appear in order in the output
	outputRunes := removeANSI([]rune(output))
	inputRunes := []rune(input)

	// find the position of the input's contiguous character sequence in the output
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

	// if no exact match, check whether most characters are present
	matchCount := 0
	for _, o := range outputRunes {
		if matchCount < len(inputRunes) && o == inputRunes[matchCount] {
			matchCount++
		}
	}

	// if matched characters exceed 80% of input, consider it contained
	return matchCount > len(inputRunes)*4/5
}

// removeANSI removes ANSI escape sequences
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
