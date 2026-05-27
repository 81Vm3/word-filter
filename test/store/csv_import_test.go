package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"sensitive-filter/pkg/store"
	"sensitive-filter/pkg/trie"
)

func TestImportCSVBasic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "words.csv")
	content := "word,severity,extra\n" +
		`badword,2,"{""category"":""abuse""}"` + "\n" +
		"傻逼,3,\n" +
		`hello,0,"{""source"":""manual""}"` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	tr := trie.New(nil)
	added, dup, err := store.ImportCSV(tr, path)
	if err != nil {
		t.Fatal(err)
	}
	if added != 3 || dup != 0 {
		t.Errorf("expected added=3 dup=0, got added=%d dup=%d", added, dup)
	}
	m, ok := tr.Get("badword")
	if !ok {
		t.Fatal("badword missing")
	}
	if m.Severity != 2 {
		t.Errorf("severity got %d", m.Severity)
	}
	if m.Extra["category"] != "abuse" {
		t.Errorf("Extra: got %v", m.Extra)
	}
}

func TestImportCSVNoHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "words.csv")
	content := "foo,1\nbar,0\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	tr := trie.New(nil)
	added, _, err := store.ImportCSV(tr, path)
	if err != nil {
		t.Fatal(err)
	}
	if added != 2 {
		t.Errorf("expected added=2, got %d", added)
	}
}

func TestImportCSVWordOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "words.csv")
	content := "alpha\nbeta\ngamma\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	tr := trie.New(nil)
	added, _, err := store.ImportCSV(tr, path)
	if err != nil {
		t.Fatal(err)
	}
	if added != 3 {
		t.Errorf("expected added=3, got %d", added)
	}
	m, _ := tr.Get("alpha")
	if m.Severity != trie.SeverityNormal {
		t.Errorf("expected default severity, got %d", m.Severity)
	}
}

func TestImportCSVDeduplicate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "words.csv")
	content := "foo,1\nFOO,2\nbar,0\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	tr := trie.New(trie.BuildNormalizer(trie.NormalizeOptions{IgnoreCase: true}))
	added, dup, err := store.ImportCSV(tr, path)
	if err != nil {
		t.Fatal(err)
	}
	if added != 2 || dup != 1 {
		t.Errorf("expected added=2 dup=1, got added=%d dup=%d", added, dup)
	}
}

func TestImportCSVMalformedSeverity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "words.csv")
	content := "foo,notanumber\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	tr := trie.New(nil)
	added, _, err := store.ImportCSV(tr, path)
	if err != nil {
		t.Fatal(err)
	}
	if added != 1 {
		t.Errorf("expected added=1, got %d", added)
	}
	m, _ := tr.Get("foo")
	if m.Severity != trie.SeverityNormal {
		t.Errorf("malformed severity should fallback, got %d", m.Severity)
	}
}
