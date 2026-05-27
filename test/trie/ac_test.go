package trie_test

import (
	"sort"
	"testing"

	"sensitive-filter/pkg/trie"
)

// matchOnce 通过公开的 Step API 手动驱动 AC 自动机。它是 sanitizer 内部
// 逻辑的简化重写，放在 trie 包的测试中以实现黑盒覆盖，避免依赖 sanitizer 包。
func matchOnce(tr *trie.Trie, text string) []string {
	tr.EnsureBuilt()
	cur := tr.Root()
	runeIdx := 0
	hits := []string{}
	for _, r := range text {
		var matches []trie.Match
		cur, matches = tr.Step(cur, tr.Normalize(r))
		for _, m := range matches {
			w, _ := tr.WordByID(m.WordID)
			hits = append(hits, w)
		}
		runeIdx++
	}
	_ = runeIdx
	sort.Strings(hits)
	return hits
}

func TestACMatchOverlap(t *testing.T) {
	tr := trie.New(nil)
	tr.Add("he", trie.DefaultMetadata())
	tr.Add("she", trie.DefaultMetadata())
	tr.Add("his", trie.DefaultMetadata())
	tr.Add("hers", trie.DefaultMetadata())

	got := matchOnce(tr, "ushers")
	want := []string{"he", "hers", "she"}
	if !equalSorted(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestACDictLinkChain(t *testing.T) {
	tr := trie.New(nil)
	tr.Add("a", trie.DefaultMetadata())
	tr.Add("ab", trie.DefaultMetadata())
	tr.Add("bab", trie.DefaultMetadata())
	tr.Add("bc", trie.DefaultMetadata())
	tr.Add("bca", trie.DefaultMetadata())
	tr.Add("c", trie.DefaultMetadata())
	tr.Add("caa", trie.DefaultMetadata())

	got := matchOnce(tr, "abcabbcabca")
	// 至少单字母模式应命中多次；只断言非空且文本中包含的每个已定义词都存在。
	if len(got) == 0 {
		t.Fatal("expected non-zero hits")
	}
	must := []string{"a", "ab", "bc", "c"}
	for _, w := range must {
		found := false
		for _, h := range got {
			if h == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing expected hit %q in %v", w, got)
		}
	}
}

func TestACAfterDeletion(t *testing.T) {
	tr := trie.New(nil)
	tr.Add("foo", trie.DefaultMetadata())
	tr.Add("bar", trie.DefaultMetadata())
	tr.EnsureBuilt()

	got1 := matchOnce(tr, "foobar")
	if !equalSorted(got1, []string{"bar", "foo"}) {
		t.Errorf("before delete: got %v", got1)
	}

	tr.Delete("foo")
	got2 := matchOnce(tr, "foobar")
	if !equalSorted(got2, []string{"bar"}) {
		t.Errorf("after delete: got %v", got2)
	}
}

func TestACAfterAddition(t *testing.T) {
	tr := trie.New(nil)
	tr.Add("foo", trie.DefaultMetadata())
	tr.EnsureBuilt()
	got1 := matchOnce(tr, "foobar")
	if !equalSorted(got1, []string{"foo"}) {
		t.Errorf("initial: got %v", got1)
	}
	tr.Add("bar", trie.DefaultMetadata())
	got2 := matchOnce(tr, "foobar")
	if !equalSorted(got2, []string{"bar", "foo"}) {
		t.Errorf("after add: got %v", got2)
	}
}

func TestACEmptyTrie(t *testing.T) {
	tr := trie.New(nil)
	tr.EnsureBuilt()
	if got := matchOnce(tr, "anything"); len(got) != 0 {
		t.Errorf("empty trie: got %v", got)
	}
}

func TestACChinesePatternBuild(t *testing.T) {
	tr := trie.New(trie.BuildNormalizer(trie.NormalizeOptions{IgnoreCase: true, IgnoreWidth: true}))
	tr.Add("傻逼", trie.Metadata{Severity: 2})
	tr.Add("测试敏感", trie.Metadata{Severity: 1})
	got := matchOnce(tr, "你这个傻逼别说测试敏感的话")
	if !equalSorted(got, []string{"傻逼", "测试敏感"}) {
		t.Errorf("got %v", got)
	}
}

func TestACStepRunes(t *testing.T) {
	tr := trie.New(nil)
	tr.Add("ab", trie.DefaultMetadata())
	tr.EnsureBuilt()

	cur := tr.Root()
	_, matches := tr.Step(cur, 'a')
	if len(matches) != 0 {
		t.Errorf("expected no hits at 'a', got %v", matches)
	}
	cur, matches = tr.Step(cur, 'a')
	cur, matches = tr.Step(cur, 'b')
	if len(matches) != 1 {
		t.Fatalf("expected 1 hit at 'ab', got %d", len(matches))
	}
	if matches[0].RuneLen != 2 {
		t.Errorf("RuneLen got %d want 2", matches[0].RuneLen)
	}
	_ = cur
}

func equalSorted(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x := append([]string(nil), a...)
	y := append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}
