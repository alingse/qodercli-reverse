package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Skill represents a community/agent skill.
type Skill struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Tools       []string `json:"tools"`
	Source      string   `json:"source"`
	Path        string   `json:"path"`
	Content     string   `json:"-"`
}

// ParseFrontmatter parses a SKILL.md file into a Skill.
func ParseSkill(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := string(data)
	s := &Skill{Path: path}
	// frontmatter between --- fences
	if strings.HasPrefix(strings.TrimSpace(content), "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			fm := parts[1]
			s.Content = strings.TrimSpace(parts[2])
			parseFrontmatter(fm, s)
		} else {
			s.Content = content
		}
	} else {
		s.Content = content
	}
	if s.Name == "" {
		s.Name = strings.TrimSuffix(filepath.Base(filepath.Dir(path)), string(filepath.Separator))
		if s.Name == "" || s.Name == "." {
			s.Name = filepath.Base(path)
		}
	}
	if s.Source == "" {
		s.Source = "local"
	}
	return s, nil
}

func parseFrontmatter(fm string, s *Skill) {
	for _, line := range strings.Split(fm, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.Contains(line, ":") {
			continue
		}
		idx := strings.Index(line, ":")
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, `"'`)
		switch strings.ToLower(key) {
		case "name":
			s.Name = val
		case "version":
			s.Version = val
		case "description":
			s.Description = val
		case "category":
			s.Category = val
		case "tags":
			s.Tags = splitList(val)
		case "tools":
			s.Tools = splitList(val)
		case "source":
			s.Source = val
		}
	}
}

func splitList(val string) []string {
	val = strings.Trim(val, "[]")
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"'`)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// DiscoverSkills finds SKILL.md files under root dirs.
func DiscoverSkills(roots ...string) ([]*Skill, error) {
	var result []*Skill
	seen := map[string]bool{}
	for _, root := range roots {
		filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			if filepath.Base(p) != "SKILL.md" {
				return nil
			}
			sk, err := ParseSkill(p)
			if err != nil {
				return nil
			}
			if seen[sk.Name] {
				return nil
			}
			seen[sk.Name] = true
			result = append(result, sk)
			return nil
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

// Search filters skills by query.
func Search(skills []*Skill, query string) []*Skill {
	q := strings.ToLower(query)
	var out []*Skill
	for _, s := range skills {
		if strings.Contains(strings.ToLower(s.Name), q) ||
			strings.Contains(strings.ToLower(s.Description), q) ||
			strings.Contains(strings.ToLower(s.Category), q) {
			out = append(out, s)
		}
		for _, t := range s.Tags {
			if strings.Contains(strings.ToLower(t), q) {
				out = append(out, s)
				break
			}
		}
	}
	return out
}

// Validate checks required fields.
func (s *Skill) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("skill name is required")
	}
	if s.Description == "" {
		return fmt.Errorf("skill %q: description is required", s.Name)
	}
	// path traversal check
	if strings.Contains(s.Name, "..") || strings.ContainsAny(s.Name, "/\\") {
		return fmt.Errorf("invalid skill name: %q", s.Name)
	}
	return nil
}
