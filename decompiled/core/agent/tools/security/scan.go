package security

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Tool interface compliance
func (s *Scanner) Name() string {
	return "security-scan"
}

func (s *Scanner) Description() string {
	return "Scan codebase for potential security issues (hardcoded keys, passwords, dangerous patterns)"
}

func (s *Scanner) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Root directory to scan (defaults to current workspace)",
			},
		},
		"required": []string{"path"},
	}
}

// Scanner security scanner for codebase analysis.
type Scanner struct {
	// root directory to scan
	root string
	// results collected
	results []Result
}

// Result security scan result.
type Result struct {
	// severity: high, medium, low
	Severity string `json:"severity"`
	// file path
	File string `json:"file"`
	// line number
	Line int `json:"line"`
	// message
	Message string `json:"message"`
	// pattern matched
	Pattern string `json:"pattern,omitempty"`
}

// NewScanner creates a new scanner.
func NewScanner(root string) *Scanner {
	return &Scanner{
		root:    root,
		results: []Result{},
	}
}

// Scan performs a security scan of the codebase.
func (s *Scanner) Scan() []Result {
	s.results = nil
	s.scanDirectory(s.root)
	return s.results
}

func (s *Scanner) scanDirectory(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Skip .git and .qodercli-reverse directories
			if entry.Name() == ".git" || entry.Name() == ".qodercli-reverse" {
				continue
			}
			s.scanDirectory(filepath.Join(dir, entry.Name()))
			continue
		}

		// Only scan source files
		ext := filepath.Ext(entry.Name())
		switch ext {
		case ".go", ".md", ".yaml", ".yml", ".json", ".env", ".toml", ".cfg":
			s.scanFile(filepath.Join(dir, entry.Name()))
		}
	}
}

func (s *Scanner) scanFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		s.checkLine(path, lineNum, line)
	}
}

func (s *Scanner) checkLine(filePath string, lineNum int, line string) {
	// Check for hardcoded API keys
	if matched, _ := regexp.MatchString(`(?:api[_-]?key|apikey|secret[_-]?key|auth[_-]?token)\s*[:=]\s*["'][^"']{8,}["']`, line); matched {
		s.results = append(s.results, Result{
			Severity: "high",
			File:     filePath,
			Line:     lineNum,
			Message:  "Potential hardcoded API key or token detected",
			Pattern:  "hardcoded-key",
		})
	}

	// Check for password patterns
	if matched, _ := regexp.MatchString(`(?:password|passwd|pwd)\s*[:=]\s*["'][^"']{4,}["']`, line); matched {
		s.results = append(s.results, Result{
			Severity: "high",
			File:     filePath,
			Line:     lineNum,
			Message:  "Potential hardcoded password detected",
			Pattern:  "hardcoded-password",
		})
	}

	// Check for SSH private key patterns
	if matched, _ := regexp.MatchString(`-----BEGIN (?:RSA |EC |OPEN )?PRIVATE KEY-----`, line); matched {
		s.results = append(s.results, Result{
			Severity: "high",
			File:     filePath,
			Line:     lineNum,
			Message:  "Potential SSH private key detected",
			Pattern:  "ssh-private-key",
		})
	}

	// Check for AWS key patterns
	if matched, _ := regexp.MatchString(`AKIA[0-9A-Z]{16}`, line); matched {
		s.results = append(s.results, Result{
			Severity: "high",
			File:     filePath,
			Line:     lineNum,
			Message:  "Potential AWS access key ID detected",
			Pattern:  "aws-access-key",
		})
	}

	// Check for database connection strings
	if matched, _ := regexp.MatchString(`(?:postgres|postgresql|mysql|mongodb)://[^\s]+`, line); matched {
		s.results = append(s.results, Result{
			Severity: "medium",
			File:     filePath,
			Line:     lineNum,
			Message:  "Potential database connection string detected",
			Pattern:  "db-connection-string",
		})
	}

	// Check for bare environment variable exposure
	if strings.Contains(line, "export ") && strings.Contains(line, "=") {
		// Skip known safe vars
		safe := []string{"PATH", "HOME", "SHELL", "USER", "LOGNAME"}
		var isSafe bool
		for _, s := range safe {
			if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(strings.Split(line, "=")[0])), s) {
				isSafe = true
				break
			}
		}
		if !isSafe {
			// Check if it's a secret var
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				if strings.Contains(strings.ToLower(key), "key") ||
					strings.Contains(strings.ToLower(key), "secret") ||
					strings.Contains(strings.ToLower(key), "token") ||
					strings.Contains(strings.ToLower(key), "credential") {
					s.results = append(s.results, Result{
						Severity: "medium",
						File:     filePath,
						Line:     lineNum,
						Message:  fmt.Sprintf("Potential secret environment variable: %s", key),
						Pattern:  "secret-env-var",
					})
				}
			}
		}
	}

	// Check for dangerous bash patterns
	dangerous := []*regexp.Regexp{
		regexp.MustCompile(`rm\s+-\s*rf\s+`),
		regexp.MustCompile(`>\s*/dev`),
		regexp.MustCompile(`mkfs`),
		regexp.MustCompile(`dd\s+if=`),
	}

	for _, pat := range dangerous {
		if pat.MatchString(line) {
			s.results = append(s.results, Result{
				Severity: "medium",
				File:     filePath,
				Line:     lineNum,
				Message:  "Potentially dangerous bash pattern detected",
				Pattern:  "dangerous-bash",
			})
			// Only add one result per line
			break
		}
	}
}

// Execute implements the Tool interface.
func (s *Scanner) Execute(ctx context.Context, input string) (string, error) {
	// Run scan
	results := s.Scan()

	// Format results
	if len(results) == 0 {
		return "No security issues found.", nil
	}

	builder := strings.Builder{}
	builder.WriteString("Security Scan Results:\n")
	builder.WriteString("====================\n\n")

	// Count by severity
	high := 0
	medium := 0
	low := 0
	for _, r := range results {
		switch r.Severity {
		case "high":
			high++
		case "medium":
			medium++
		case "low":
			low++
		}
	}

	builder.WriteString(fmt.Sprintf("High: %d, Medium: %d, Low: %d\n\n", high, medium, low))

	for _, r := range results {
		builder.WriteString(fmt.Sprintf("[%s] %s:%d - %s\n", r.Severity, r.File, r.Line, r.Message))
	}

	return builder.String(), nil
}
