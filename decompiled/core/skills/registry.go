package skills

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Source represents a community skill source.
type Source struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Ref  string `json:"ref,omitempty"` // branch/tag/commit
}

// Registry manages community sources and installed skills.
type Registry struct {
	ConfigPath string
	Sources    []Source
}

func NewRegistry() *Registry {
	home, _ := os.UserHomeDir()
	cfg := filepath.Join(home, ".qoder", "community-sources.json")
	r := &Registry{ConfigPath: cfg}
	r.load()
	if len(r.Sources) == 0 {
		r.Sources = []Source{{Name: "qoder-community", URL: "https://github.com/Qoder-AI/qoder-community", Ref: "main"}}
	}
	return r
}

func (r *Registry) load() {
	data, err := os.ReadFile(r.ConfigPath)
	if err != nil {
		return
	}
	json.Unmarshal(data, &r.Sources)
}

func (r *Registry) save() error {
	os.MkdirAll(filepath.Dir(r.ConfigPath), 0755)
	data, _ := json.MarshalIndent(r.Sources, "", "  ")
	return os.WriteFile(r.ConfigPath, data, 0644)
}

func (r *Registry) AddSource(name, url, ref string) error {
	for _, s := range r.Sources {
		if s.Name == name {
			return fmt.Errorf("source %q already exists", name)
		}
	}
	r.Sources = append(r.Sources, Source{Name: name, URL: url, Ref: ref})
	return r.save()
}

func (r *Registry) RemoveSource(name string) error {
	var out []Source
	found := false
	for _, s := range r.Sources {
		if s.Name == name {
			found = true
			continue
		}
		out = append(out, s)
	}
	if !found {
		return fmt.Errorf("source %q not found", name)
	}
	r.Sources = out
	return r.save()
}

// SkillDirs returns candidate skill directories (project + user).
func SkillDirs(workDir string) []string {
	var dirs []string
	if workDir != "" {
		dirs = append(dirs, filepath.Join(workDir, ".qoder", "skills"))
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		dirs = append(dirs, filepath.Join(home, ".qoder", "skills"))
	}
	return dirs
}
