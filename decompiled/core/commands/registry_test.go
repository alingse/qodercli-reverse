package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverProjectExtensions(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, ".qoder", "commands")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "foo.md"), []byte("# Foo\ndescription: do foo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	subDir := filepath.Join(dir, ".qoder", "subagents")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "bar.md"), []byte("description: bar agent\n"), 0644); err != nil {
		t.Fatal(err)
	}

	reg := NewRegistry()
	reg.Discover(dir)

	foundCmd := false
	for _, c := range reg.Commands {
		if c.Name == "/foo" {
			foundCmd = true
		}
	}
	if !foundCmd {
		t.Errorf("expected /foo command discovered")
	}
	foundSub := false
	for _, s := range reg.Subagents {
		if s.Name == "bar" {
			foundSub = true
		}
	}
	if !foundSub {
		t.Errorf("expected bar subagent discovered")
	}
}
