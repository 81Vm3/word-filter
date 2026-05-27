package sanitizer_test

import (
	"strings"
	"testing"

	"sensitive-filter/pkg/sanitizer"
	"sensitive-filter/pkg/trie"
)

func makeSanitizer(words map[string]int, mode sanitizer.MatchMode, ignoreCase, ignoreWidth bool) *sanitizer.Sanitizer {
	tr := trie.New(trie.BuildNormalizer(trie.NormalizeOptions{IgnoreCase: ignoreCase, IgnoreWidth: ignoreWidth}))
	for w, sev := range words {
		tr.Add(w, trie.Metadata{Severity: sev})
	}
	return sanitizer.New(tr, sanitizer.Config{
		IgnoreCase:  ignoreCase,
		IgnoreWidth: ignoreWidth,
		Mode:        mode,
	})
}

func TestSanitizeBasic(t *testing.T) {
	san := makeSanitizer(map[string]int{
		"badword": 1,
		"傻逼":      2,
	}, sanitizer.ModeFullWord, false, false)

	if got := san.Sanitize("hello badword 傻逼 world"); got != "hello ******* ** world" {
		t.Errorf("got %q", got)
	}
}

func TestSanitizeWithDebug(t *testing.T) {
	san := makeSanitizer(map[string]int{
		"foo": 1,
		"bar": 2,
	}, sanitizer.ModeFullWord, false, false)

	dbg := san.SanitizeWithDebug("foo and bar and foo")
	if dbg.Count != 3 {
		t.Errorf("expected 3 hits, got %d (hits: %+v)", dbg.Count, dbg.Hits)
	}
	if dbg.Sanitized != "*** and *** and ***" {
		t.Errorf("sanitized: got %q", dbg.Sanitized)
	}
	if dbg.Elapsed <= 0 {
		t.Errorf("Elapsed should be > 0, got %v", dbg.Elapsed)
	}
	for _, h := range dbg.Hits {
		piece := "foo and bar and foo"[h.ByteStart:h.ByteEnd]
		if piece != h.Word {
			t.Errorf("byte slice %q != Word %q", piece, h.Word)
		}
	}
}

func TestSanitizeIgnoreCase(t *testing.T) {
	san := makeSanitizer(map[string]int{
		"badword": 1,
	}, sanitizer.ModeFullWord, true, false)

	if got := san.Sanitize("This is BADWORD!"); got != "This is *******!" {
		t.Errorf("got %q", got)
	}
	if got := san.Sanitize("BaDwOrD"); got != "*******" {
		t.Errorf("got %q", got)
	}
}

func TestSanitizeIgnoreWidth(t *testing.T) {
	san := makeSanitizer(map[string]int{
		"abc": 1,
	}, sanitizer.ModeFullWord, true, true)

	if got := san.Sanitize("xxＡＢＣyy"); got != "xx***yy" {
		t.Errorf("got %q", got)
	}
}

func TestSanitizeOverlap(t *testing.T) {
	san := makeSanitizer(map[string]int{
		"abc":  1,
		"bcd":  1,
		"abcd": 1,
	}, sanitizer.ModeFullWord, false, false)

	out := san.Sanitize("xxabcdxx")
	if out != "xx****xx" {
		t.Errorf("overlap merge: got %q want %q", out, "xx****xx")
	}
}

func TestSanitizePrefixMode(t *testing.T) {
	san := makeSanitizer(map[string]int{
		"http":     1,
		"https":    1,
		"http://":  1,
		"https://": 1,
	}, sanitizer.ModePrefix, false, false)

	dbg := san.SanitizeWithDebug("visit https://example.com today")
	if dbg.Count < 2 {
		t.Errorf("expected at least 2 hits for prefix overlap, got %d", dbg.Count)
	}
}

func TestSanitizeNoHits(t *testing.T) {
	san := makeSanitizer(map[string]int{"foo": 1}, sanitizer.ModeFullWord, false, false)
	in := "this string has nothing to hide"
	if out := san.Sanitize(in); out != in {
		t.Errorf("expected unchanged, got %q", out)
	}
}

func TestSanitizeReplaceChar(t *testing.T) {
	tr := trie.New(nil)
	tr.Add("foo", trie.DefaultMetadata())
	san := sanitizer.New(tr, sanitizer.Config{Mode: sanitizer.ModeFullWord, ReplaceRune: '#'})
	if got := san.Sanitize("xfooxfoo"); got != "x###x###" {
		t.Errorf("got %q", got)
	}
}

func TestStreamingBasic(t *testing.T) {
	san := makeSanitizer(map[string]int{
		"badword": 1,
		"傻逼":      2,
	}, sanitizer.ModeFullWord, false, false)

	ss := san.InitStream()
	ss.AppendToStream("hello ")
	ss.AppendToStream("badword 傻逼")
	ss.AppendToStream(" world")
	r := ss.GetStreamResult()

	if r.Sanitized != "hello ******* ** world" {
		t.Errorf("streaming sanitized: got %q", r.Sanitized)
	}
	if r.Count != 2 {
		t.Errorf("expected 2 hits, got %d", r.Count)
	}
}

func TestStreamingCrossChunkMatch(t *testing.T) {
	san := makeSanitizer(map[string]int{
		"badword": 1,
	}, sanitizer.ModeFullWord, false, false)

	ss := san.InitStream()
	ss.AppendToStream("bad")
	ss.AppendToStream("word")
	r := ss.GetStreamResult()

	if r.Count != 1 {
		t.Errorf("expected 1 hit across chunks, got %d", r.Count)
	}
	if r.Sanitized != "*******" {
		t.Errorf("got %q", r.Sanitized)
	}
}

func TestStreamingCrossChunkUTF8(t *testing.T) {
	san := makeSanitizer(map[string]int{
		"傻逼": 2,
	}, sanitizer.ModeFullWord, false, false)

	full := []byte("傻逼")
	if len(full) != 6 {
		t.Fatalf("unexpected utf-8 length %d", len(full))
	}

	ss := san.InitStream()
	ss.AppendToStream(string(full[:4]))
	if ss.PendingBytes() == 0 {
		t.Error("expected pending bytes after partial UTF-8")
	}
	ss.AppendToStream(string(full[4:]))
	r := ss.GetStreamResult()

	if r.Count != 1 {
		t.Errorf("expected 1 hit, got %d", r.Count)
	}
	if r.Sanitized != "**" {
		t.Errorf("expected '**', got %q", r.Sanitized)
	}
}

func TestStreamingClear(t *testing.T) {
	san := makeSanitizer(map[string]int{"foo": 1}, sanitizer.ModeFullWord, false, false)
	ss := san.InitStream()
	ss.AppendToStream("foo")
	if ss.GetStreamResult().Count != 1 {
		t.Error("expected 1 hit before clear")
	}
	ss.ClearStream()
	if ss.GetStreamResult().Count != 0 {
		t.Error("expected 0 hits after clear")
	}
	ss.AppendToStream("foo")
	if ss.GetStreamResult().Count != 1 {
		t.Error("expected 1 hit after clear+append")
	}
}

func TestStreamingMultipleHits(t *testing.T) {
	san := makeSanitizer(map[string]int{"abc": 1}, sanitizer.ModeFullWord, false, false)
	ss := san.InitStream()
	for i := 0; i < 5; i++ {
		ss.AppendToStream("xabcy")
	}
	r := ss.GetStreamResult()
	if r.Count != 5 {
		t.Errorf("expected 5 hits, got %d", r.Count)
	}
	if r.Sanitized != strings.Repeat("x***y", 5) {
		t.Errorf("got %q", r.Sanitized)
	}
}

func TestStreamingGetResultDoesNotReset(t *testing.T) {
	san := makeSanitizer(map[string]int{"foo": 1}, sanitizer.ModeFullWord, false, false)
	ss := san.InitStream()
	ss.AppendToStream("foo ")
	r1 := ss.GetStreamResult()
	ss.AppendToStream("foo")
	r2 := ss.GetStreamResult()

	if r1.Count != 1 {
		t.Errorf("r1.Count = %d", r1.Count)
	}
	if r2.Count != 2 {
		t.Errorf("r2.Count = %d (should accumulate)", r2.Count)
	}
}

func TestStreamingPositions(t *testing.T) {
	san := makeSanitizer(map[string]int{"foo": 1}, sanitizer.ModeFullWord, false, false)
	ss := san.InitStream()
	ss.AppendToStream("a")
	ss.AppendToStream("foo")
	ss.AppendToStream("bfoo")
	r := ss.GetStreamResult()

	if r.Count != 2 {
		t.Fatalf("expected 2 hits, got %d", r.Count)
	}
	want := []struct{ start, end int }{{1, 4}, {5, 8}}
	for i, h := range r.Hits {
		if h.RuneStart != want[i].start || h.RuneEnd != want[i].end {
			t.Errorf("hit %d: got rune [%d,%d), want [%d,%d)", i, h.RuneStart, h.RuneEnd, want[i].start, want[i].end)
		}
	}
}

func TestSanitizeChinesePositions(t *testing.T) {
	san := makeSanitizer(map[string]int{"傻逼": 2}, sanitizer.ModeFullWord, false, false)
	in := "你好傻逼啊"
	dbg := san.SanitizeWithDebug(in)
	if dbg.Count != 1 {
		t.Fatalf("expected 1 hit, got %d", dbg.Count)
	}
	h := dbg.Hits[0]
	if h.RuneStart != 2 || h.RuneEnd != 4 {
		t.Errorf("rune range got [%d,%d) want [2,4)", h.RuneStart, h.RuneEnd)
	}
	if in[h.ByteStart:h.ByteEnd] != "傻逼" {
		t.Errorf("byte slice got %q", in[h.ByteStart:h.ByteEnd])
	}
}

func TestEmptyInput(t *testing.T) {
	san := makeSanitizer(map[string]int{"foo": 1}, sanitizer.ModeFullWord, false, false)
	if got := san.Sanitize(""); got != "" {
		t.Errorf("got %q", got)
	}
}
