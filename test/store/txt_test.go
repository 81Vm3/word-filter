package store_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"sensitive-filter/pkg/store"
	"sensitive-filter/pkg/trie"
)

func TestReadWriteRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "words.txt")

	entries := []store.WordEntry{
		{Word: "badword", Meta: trie.Metadata{Severity: 1}},
		{Word: "傻逼", Meta: trie.Metadata{Severity: 2}},
		{Word: "测试", Meta: trie.Metadata{Severity: 0}},
	}

	if err := store.WriteWordsTxt(path, entries); err != nil {
		t.Fatal(err)
	}
	got, err := store.ReadWordsTxt(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(got))
	}
	sortEntries(got)
	sortEntries(entries)
	for i := range got {
		if got[i].Word != entries[i].Word {
			t.Errorf("[%d] word: got %q want %q", i, got[i].Word, entries[i].Word)
		}
		if got[i].Meta.Severity != entries[i].Meta.Severity {
			t.Errorf("[%d] severity: got %d want %d", i, got[i].Meta.Severity, entries[i].Meta.Severity)
		}
	}
}

func TestReadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := store.ReadWordsTxt(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestReadZeroCount(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "zero.txt")
	if err := os.WriteFile(path, []byte("0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := store.ReadWordsTxt(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestReadLineBasedWordsWhenHeaderIsNotNumber(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.txt")
	if err := os.WriteFile(path, []byte("notanumber\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := store.ReadWordsTxt(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].Word != "notanumber" {
		t.Fatalf("unexpected word: got %q", got[0].Word)
	}
	if got[0].Meta.Severity != trie.SeverityNormal {
		t.Fatalf("unexpected severity: got %d", got[0].Meta.Severity)
	}
}

func TestReadLineBasedWordsAllSeverityNormal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "line_based.txt")
	content := "badword\n傻逼\nBADWORD\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := store.ReadWordsTxt(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	for i, e := range got {
		if e.Meta.Severity != trie.SeverityNormal {
			t.Fatalf("[%d] expected severity normal, got %d", i, e.Meta.Severity)
		}
	}
}

func TestReadMissingMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "partial.txt")
	content := "2\nfoo\nbar\n5\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := store.ReadWordsTxt(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	for _, e := range got {
		if e.Word == "bar" && e.Meta.Severity != trie.SeverityNormal {
			t.Errorf("bar should default to severity=0, got %d", e.Meta.Severity)
		}
	}
}

func TestReadFewerWordsThanHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "short.txt")
	content := "3\nonly\none\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := store.ReadWordsTxt(path)
	if err == nil {
		t.Error("expected error when words are fewer than declared")
	}
}

func TestCollectFromTrie(t *testing.T) {
	tr := trie.New(nil)
	tr.Add("a", trie.Metadata{Severity: 1})
	tr.Add("b", trie.Metadata{Severity: 2})

	got := store.CollectFromTrie(tr)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func sortEntries(e []store.WordEntry) {
	sort.Slice(e, func(i, j int) bool { return e[i].Word < e[j].Word })
}
