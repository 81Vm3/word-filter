package trie_test

import (
	"reflect"
	"testing"

	"sensitive-filter/pkg/trie"
)

func newTestTrie() *trie.Trie {
	return trie.New(trie.BuildNormalizer(trie.NormalizeOptions{IgnoreCase: true, IgnoreWidth: true}))
}

func TestTrieAddGet(t *testing.T) {
	tr := newTestTrie()
	id, added := tr.Add("badword", trie.Metadata{Severity: 2})
	if !added {
		t.Fatal("expected first Add to succeed")
	}
	if id != 0 {
		t.Errorf("expected first id = 0, got %d", id)
	}

	m, ok := tr.Get("BadWord")
	if !ok {
		t.Fatal("Get should be case-insensitive")
	}
	if m.Severity != 2 {
		t.Errorf("expected severity=2, got %d", m.Severity)
	}
}

func TestTrieAddDuplicate(t *testing.T) {
	tr := newTestTrie()
	id1, added1 := tr.Add("hello", trie.DefaultMetadata())
	id2, added2 := tr.Add("HELLO", trie.Metadata{Severity: 9})
	if !added1 {
		t.Fatal("first Add should succeed")
	}
	if added2 {
		t.Error("duplicate Add (after case-fold) should return added=false")
	}
	if id1 != id2 {
		t.Errorf("duplicate should return same id, got %d vs %d", id1, id2)
	}
}

func TestTrieUpdate(t *testing.T) {
	tr := newTestTrie()
	tr.Add("foo", trie.DefaultMetadata())
	if !tr.Update("FOO", trie.Metadata{Severity: 5}) {
		t.Fatal("Update should succeed")
	}
	m, _ := tr.Get("foo")
	if m.Severity != 5 {
		t.Errorf("expected severity=5 after Update, got %d", m.Severity)
	}
	if tr.Update("nonexistent", trie.Metadata{}) {
		t.Error("Update on missing word should return false")
	}
}

func TestTrieDelete(t *testing.T) {
	tr := newTestTrie()
	tr.Add("abc", trie.DefaultMetadata())
	tr.Add("abcd", trie.DefaultMetadata())
	if !tr.Delete("abc") {
		t.Fatal("Delete abc should succeed")
	}
	if tr.Has("abc") {
		t.Error("abc should be removed")
	}
	if !tr.Has("abcd") {
		t.Error("abcd should still exist after deleting abc")
	}
	if tr.Delete("abc") {
		t.Error("second Delete should return false")
	}
}

func TestTrieDeleteSharedPath(t *testing.T) {
	tr := newTestTrie()
	tr.Add("ab", trie.DefaultMetadata())
	tr.Add("abcd", trie.DefaultMetadata())
	if !tr.Delete("abcd") {
		t.Fatal("Delete abcd should succeed")
	}
	if !tr.Has("ab") {
		t.Error("ab should still exist")
	}
}

func TestTrieWalk(t *testing.T) {
	tr := newTestTrie()
	want := map[string]int{"a": 0, "bb": 1, "ccc": 2}
	for w, sev := range want {
		tr.Add(w, trie.Metadata{Severity: sev})
	}
	got := map[string]int{}
	tr.Walk(func(word string, m trie.Metadata) {
		got[word] = m.Severity
	})
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Walk mismatch: want %v got %v", want, got)
	}
}

func TestTrieChinese(t *testing.T) {
	tr := newTestTrie()
	tr.Add("傻逼", trie.Metadata{Severity: 2})
	tr.Add("测试敏感", trie.Metadata{Severity: 1})
	if !tr.Has("傻逼") {
		t.Error("should find 傻逼")
	}
	m, _ := tr.Get("测试敏感")
	if m.Severity != 1 {
		t.Errorf("expected severity=1, got %d", m.Severity)
	}
}

func TestTrieEmptyWord(t *testing.T) {
	tr := newTestTrie()
	id, added := tr.Add("", trie.DefaultMetadata())
	if added {
		t.Error("empty word should not be added")
	}
	if id != -1 {
		t.Errorf("expected id=-1 for empty word, got %d", id)
	}
}

func TestTrieSize(t *testing.T) {
	tr := newTestTrie()
	if tr.Size() != 0 {
		t.Fatal("empty trie size should be 0")
	}
	tr.Add("a", trie.DefaultMetadata())
	tr.Add("b", trie.DefaultMetadata())
	tr.Add("a", trie.DefaultMetadata()) // 重复
	if tr.Size() != 2 {
		t.Errorf("expected size=2, got %d", tr.Size())
	}
}

func TestTrieIDLookups(t *testing.T) {
	tr := newTestTrie()
	id, _ := tr.Add("foo", trie.Metadata{Severity: 7})
	if w, ok := tr.WordByID(id); !ok || w != "foo" {
		t.Errorf("WordByID(%d) = %q, %v", id, w, ok)
	}
	if m, ok := tr.MetaByID(id); !ok || m.Severity != 7 {
		t.Errorf("MetaByID(%d) = %v, %v", id, m, ok)
	}
	if _, ok := tr.WordByID(9999); ok {
		t.Error("WordByID for missing id should return false")
	}
}
