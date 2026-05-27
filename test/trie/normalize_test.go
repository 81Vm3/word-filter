package trie_test

import (
	"testing"

	"sensitive-filter/pkg/trie"
)

func TestNormalizeCase(t *testing.T) {
	n := trie.BuildNormalizer(trie.NormalizeOptions{IgnoreCase: true})
	if got := trie.NormalizeString("HelloWorld", n); got != "helloworld" {
		t.Errorf("got %q", got)
	}
}

func TestNormalizeWidth(t *testing.T) {
	n := trie.BuildNormalizer(trie.NormalizeOptions{IgnoreWidth: true})
	if got := trie.NormalizeString("ＡＢＣ１２３", n); got != "ABC123" {
		t.Errorf("got %q", got)
	}
	if got := trie.NormalizeString("ａ　ｂ", n); got != "a b" {
		t.Errorf("got %q", got)
	}
}

func TestNormalizeCombined(t *testing.T) {
	n := trie.BuildNormalizer(trie.NormalizeOptions{IgnoreCase: true, IgnoreWidth: true})
	if got := trie.NormalizeString("ＨＥＬＬＯ", n); got != "hello" {
		t.Errorf("got %q", got)
	}
}

func TestNormalizeIdentity(t *testing.T) {
	n := trie.BuildNormalizer(trie.NormalizeOptions{})
	if got := trie.NormalizeString("HelloＡ", n); got != "HelloＡ" {
		t.Errorf("got %q", got)
	}
}

func TestNormalizeChineseUntouched(t *testing.T) {
	n := trie.BuildNormalizer(trie.NormalizeOptions{IgnoreCase: true, IgnoreWidth: true})
	if got := trie.NormalizeString("傻逼测试", n); got != "傻逼测试" {
		t.Errorf("got %q", got)
	}
}

func TestNormalizeNilFunc(t *testing.T) {
	if got := trie.NormalizeString("hello", nil); got != "hello" {
		t.Errorf("nil normalizer should pass-through, got %q", got)
	}
}
