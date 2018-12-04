package notes

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func configForCategoryTest(subdir string) *Config {
	cwd, err := os.Getwd()
	panicIfErr(err)
	return &Config{
		HomePath: filepath.Join(cwd, "testdata", "category", subdir),
	}
}

func TestCollectCategoriesOnlyFirstCategory(t *testing.T) {
	cfg := configForCategoryTest("normal")
	cats, err := CollectCategories(cfg, OnlyFirstCategory)
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 1 {
		t.Fatal(len(cats), "categories found", cats)
	}
	var name string
	var cat *Category
	for name, cat = range cats {
		break
	}
	if cat == nil || name == "" {
		t.Fatal(name, cat)
	}
	if len(cat.NotePaths) != 1 {
		t.Fatal("Only first note should be collected", cat.NotePaths)
	}
}

func TestCollectCategoriesOK(t *testing.T) {
	cfg := configForCategoryTest("normal")
	cats, err := CollectCategories(cfg, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Check category names
	{
		want := []string{
			"a",
			"a/c",
			"b",
		}

		have := cats.Names()
		sort.Strings(have)
		if !reflect.DeepEqual(want, have) {
			t.Fatal("All categories are not detected. Wanted", want, "but have", have)
		}

		for n, c := range cats {
			if n != c.Name {
				t.Fatal("Name and its name mismatches", n, c.Name)
			}
			p := filepath.Join(cfg.HomePath, filepath.FromSlash(c.Name))
			if p != c.Path {
				t.Fatal("Name does not match to its path", c.Name, c.Path)
			}
		}
	}

	// Check note paths
	{
		want := []string{
			"a/1.md",
			"a/4.md",
			"a/c/3.md",
			"a/c/5.md",
			"b/2.md",
			"b/6.md",
		}
		for i, w := range want {
			want[i] = filepath.Join(cfg.HomePath, filepath.FromSlash(w))
		}

		have := make([]string, 0, 6)
		for _, c := range cats {
			have = append(have, c.NotePaths...)
		}
		sort.Strings(have)

		if !reflect.DeepEqual(want, have) {
			t.Fatal("Note paths are unexpected. Wanted", want, "but have", have)
		}
	}
}

func TestCategoryNotesOK(t *testing.T) {
	name := "a/c"
	cfg := configForCategoryTest("normal")
	dir := filepath.Join(cfg.HomePath, filepath.FromSlash(name))
	cat := &Category{
		Path: dir,
		Name: name,
		NotePaths: []string{
			filepath.Join(dir, "3.md"),
			filepath.Join(dir, "5.md"),
		},
	}

	notes, err := cat.Notes(cfg)
	if err != nil {
		t.Fatal(err)
	}

	for _, note := range notes {
		if note.Category != name {
			t.Fatal("Category mismatch", note.Category, name)
		}
		if f := note.File; f != "3.md" && f != "5.md" {
			t.Fatal("Unexpected note file loaded", f)
		}
	}
}

func TestCategoriesNotesOK(t *testing.T) {
	cfg := configForCategoryTest("normal")
	cats, err := CollectCategories(cfg, 0)
	if err != nil {
		t.Fatal(err)
	}

	want := make([]string, 0, 6)
	for _, c := range cats {
		want = append(want, c.NotePaths...)
	}
	sort.Strings(want)

	notes, err := cats.Notes(cfg)
	if err != nil {
		t.Fatal(err)
	}

	have := make([]string, 0, len(notes))
	for _, note := range notes {
		c := filepath.Join(cfg.HomePath, filepath.FromSlash(note.Category), note.File)
		have = append(have, c)
	}
	sort.Strings(have)

	if !reflect.DeepEqual(want, have) {
		t.Fatal("Wanted", want, "but have", have)
	}
}

func TestCollectCategoriesNoHome(t *testing.T) {
	cfg := &Config{HomePath: "/path/to/somewhere/unknown"}
	_, err := CollectCategories(cfg, 0)
	if err == nil || !strings.Contains(err.Error(), "Cannot read home") {
		t.Fatal("Got unexpected", err)
	}
}

func TestCategoryNotesError(t *testing.T) {
	cfg := configForCategoryTest("fail")
	name := "missing-created"
	dir := filepath.Join(cfg.HomePath, name)
	cat := &Category{
		Path: dir,
		Name: name,
		NotePaths: []string{
			filepath.Join(dir, "1.md"),
		},
	}
	if _, err := cat.Notes(cfg); err == nil || !strings.Contains(err.Error(), "Missing metadata") {
		t.Fatal("Got unexpected", err)
	}
}

func TestCategoriesNotesError(t *testing.T) {
	cfg := configForCategoryTest("fail")
	cats, err := CollectCategories(cfg, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cats.Notes(cfg); err == nil || !strings.Contains(err.Error(), "Missing metadata") {
		t.Fatal("Got unexpected", err)
	}
}

func TestCategoriesNoNote(t *testing.T) {
	cfg := configForCategoryTest("empty")
	cats, err := CollectCategories(cfg, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) > 0 {
		t.Fatal("No note should mean no category:", cats)
	}
}
