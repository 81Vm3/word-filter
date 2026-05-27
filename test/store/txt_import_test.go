package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"sensitive-filter/pkg/store"
	"sensitive-filter/pkg/trie"
)

func TestImportTxtBasic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "words.txt")
	entries := []store.WordEntry{
		{Word: "foo", Meta: trie.Metadata{Severity: 1}},
		{Word: "bar", Meta: trie.Metadata{Severity: 0}},
		{Word: "测试", Meta: trie.Metadata{Severity: 2}},
	}
	if err := store.WriteWordsTxt(path, entries); err != nil {
		t.Fatal(err)
	}

	tr := trie.New(nil)
	added, dup, err := store.ImportTxt(tr, path)
	if err != nil {
		t.Fatal(err)
	}
	if added != 3 || dup != 0 {
		t.Errorf("expected added=3 dup=0, got added=%d dup=%d", added, dup)
	}
	if tr.Size() != 3 {
		t.Errorf("expected size=3, got %d", tr.Size())
	}
}

func TestImportTxtDeduplicate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "words.txt")
	entries := []store.WordEntry{
		{Word: "dup", Meta: trie.Metadata{Severity: 1}},
		{Word: "uniq", Meta: trie.Metadata{Severity: 0}},
	}
	store.WriteWordsTxt(path, entries)

	tr := trie.New(nil)
	tr.Add("dup", trie.Metadata{Severity: 9})
	added, dup, err := store.ImportTxt(tr, path)
	if err != nil {
		t.Fatal(err)
	}
	if added != 1 || dup != 1 {
		t.Errorf("expected added=1 dup=1, got added=%d dup=%d", added, dup)
	}
}

func TestImportTxtNonExistent(t *testing.T) {
	tr := trie.New(nil)
	_, _, err := store.ImportTxt(tr, "/no/such/file.txt")
	if err == nil {
		t.Error("expected error for missing file")
	}
	if !os.IsNotExist(err) {
		t.Logf("got error: %v (acceptable)", err)
	}
}
