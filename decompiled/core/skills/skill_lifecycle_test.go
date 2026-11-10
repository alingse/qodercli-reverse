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

func TestPathTraversal(t *testing.T) {
	s := &Skill{Name: "../evil", Description: "x"}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidate(t *testing.T) {
	valid := &Skill{Name: "valid", Description: "a valid skill", Version: "1.0.0"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid skill should not error: %v", err)
	}
	noName := &Skill{Description: "no name"}
	if err := noName.Validate(); err == nil {
		t.Fatal("expected validation error for missing name")
	}
	traversal := &Skill{Name: "../evil", Description: "evil", Version: "1.0.0"}
	if err := traversal.Validate(); err == nil {
		t.Fatal("expected validation error for path traversal")
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
}

func TestSearch(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "s1"), 0755)
	os.WriteFile(filepath.Join(dir, "s1", "SKILL.md"), []byte("---\nname: my-skill\ndescription: a test for search\ntags: [test]\n---\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "s2"), 0755)
	os.WriteFile(filepath.Join(dir, "s2", "SKILL.md"), []byte("---\nname: another-skill\ndescription: a different test\ntags: [another]\n---\n"), 0644)
	sks, err := DiscoverSkills(dir)
	if err != nil {
		t.Fatal(err)
	}
	res := Search(sks, "my-skill")
	if len(res) != 1 || res[0].Name != "my-skill" {
		t.Fatalf("expected 1 skill named 'my-skill', got %v", res)
	}
	res = Search(sks, "test")
	if len(res) != 2 {
		t.Fatalf("expected 2 skills with tag 'test', got %d", len(res))
	}
}

func TestSkillDirs(t *testing.T) {
	dir := t.TempDir()
	dirs := SkillDirs(dir)
	if len(dirs) != 2 {
		t.Fatalf("expected 2 skill dirs, got %d", len(dirs))
	}
}
