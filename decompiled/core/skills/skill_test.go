package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSkill(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "SKILL.md")
	content := "---\nname: foo\nversion: 1.0.0\ndescription: bar\ntags: [a, b]\n---\n# Body\n"
	os.WriteFile(p, []byte(content), 0644)
	sk, err := ParseSkill(p)
	if err != nil {
		t.Fatal(err)
	}
	if sk.Name != "foo" || sk.Version != "1.0.0" || sk.Description != "bar" {
		t.Fatalf("bad parse: %+v", sk)
	}
	if err := sk.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverSkills(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "s1"), 0755)
	os.WriteFile(filepath.Join(dir, "s1", "SKILL.md"), []byte("---\nname: s1\ndescription: d1\n---\n"), 0644)
	sks, err := DiscoverSkills(dir)
	if err != nil || len(sks) != 1 {
		t.Fatalf("discover failed: %v %d", err, len(sks))
	}
}

func TestPathTraversal(t *testing.T) {
	s := &Skill{Name: "../evil", Description: "x"}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error")
	}
}
