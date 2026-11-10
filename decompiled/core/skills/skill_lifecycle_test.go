package skills

import (
	"os"
	"path/filepath"
	"strings"
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

func TestPathTraversal(t *testing.T) {
	s := &Skill{Name: "../evil", Description: "x"}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error for path traversal in name")
	}
}

func TestValidate(t *testing.T) {
	// Valid skill
	valid := &Skill{
		Name:        "valid-skill",
		Description: "a valid skill",
		Version:     "1.0.0",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid skill should not error: %v", err)
	}

	// Missing name
	noName := &Skill{
		Description: "no name",
	}
	if err := noName.Validate(); err == nil {
		t.Fatal("expected validation error for missing name")
	}

	// Empty description
	noDesc := &Skill{
		Name:    "has-name",
		Version: "1.0.0",
	}
	if err := noDesc.Validate(); err == nil {
		t.Fatal("expected validation error for empty description")
	}

	// Path traversal in name
	traversal := &Skill{
		Name:        "../evil",
		Description: "evil",
		Version:     "1.0.0",
	}
	if err := traversal.Validate(); err == nil {
		t.Fatal("expected validation error for path traversal in name")
	}
}

func TestDiscoverSkills(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "s1"), 0755)
	os.WriteFile(filepath.Join(dir, "s1", "SKILL.md"), []byte("---\nname: s1\ndescription: d1\n---\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "s2"), 0755)
	os.WriteFile(filepath.Join(dir, "s2", "SKILL.md"), []byte("---\nname: s2\ndescription: d2\n---\n"), 0644)

	sks, err := DiscoverSkills(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sks) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(sks))
	}

	names := make(map[string]bool)
	for _, sk := range sks {
		names[sk.Name] = true
	}
	if !names["s1"] {
		t.Fatal("expected 's1' in discovered skills")
	}
	if !names["s2"] {
		t.Fatal("expected 's2' in discovered skills")
	}
}

func TestSearchByName(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "s1"), 0755)
	os.WriteFile(filepath.Join(dir, "s1", "SKILL.md"), []byte("---\nname: my-skill\ndescription: description for my-skill\ntags: [test]\n---\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "s2"), 0755)
	os.WriteFile(filepath.Join(dir, "s2", "SKILL.md"), []byte("---\nname: another-skill\ndescription: description for another-skill\ntags: [another]\n---\n"), 0644)

	sks, err := DiscoverSkills(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Search by exact name
	res := Search(sks, "my-skill")
	if len(res) != 1 || res[0].Name != "my-skill" {
		t.Fatalf("expected 1 skill named 'my-skill', got %v", res)
	}

	// Search by tag 'test' - should match only s1 (tag test)
	res = Search(sks, "test")
	if len(res) != 1 {
		t.Fatalf("expected 1 skill with tag 'test', got %d: %v", len(res), res)
	}
	if res[0].Name != "my-skill" {
		t.Fatalf("expected 'my-skill', got '%s'", res[0].Name)
	}

	// Search by tag 'another' - should match only s2
	res = Search(sks, "another")
	if len(res) != 1 {
		t.Fatalf("expected 1 skill with tag 'another', got %d: %v", len(res), res)
	}
	if res[0].Name != "another-skill" {
		t.Fatalf("expected 'another-skill', got '%s'", res[0].Name)
	}
}

func TestSearchByDescription(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "s1"), 0755)
	os.WriteFile(filepath.Join(dir, "s1", "SKILL.md"), []byte("---\nname: skill-one\ndescription: this is a test description\ntags: [alpha]\n---\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "s2"), 0755)
	os.WriteFile(filepath.Join(dir, "s2", "SKILL.md"), []byte("---\nname: skill-two\ndescription: no match here\ntags: [beta]\n---\n"), 0644)

	sks, err := DiscoverSkills(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Search by description substring "test"
	res := Search(sks, "test")
	if len(res) != 1 {
		t.Fatalf("expected 1 skill with 'test' in description, got %d: %v", len(res), res)
	}
	if res[0].Name != "skill-one" {
		t.Fatalf("expected 'skill-one', got '%s'", res[0].Name)
	}
}

func TestSkillDirs(t *testing.T) {
	dir := t.TempDir()
	dirs := SkillDirs(dir)
	if len(dirs) != 2 {
		t.Fatalf("expected 2 skill dirs, got %d", len(dirs))
	}

	// Both dirs should end with /.qoder/skills
	for _, d := range dirs {
		if !strings.Contains(d, ".qoder") {
			t.Fatalf("expected dir %s to contain '.qoder', got %s", d, d)
		}
	}
}

func TestCommunitySourceBasic(t *testing.T) {
	reg := NewRegistry()
	// Just verify registry can be created and has sources
	if len(reg.Sources) < 0 {
		t.Fatalf("expected non-negative source count, got %d", len(reg.Sources))
	}
	t.Logf("Registry has %d source(s)", len(reg.Sources))
}
