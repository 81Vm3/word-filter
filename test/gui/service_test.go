package gui_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sensitive-filter/pkg/gui"
	"sensitive-filter/pkg/store"
	"sensitive-filter/pkg/trie"
)

func makeService(t *testing.T) (*gui.Service, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "words.txt")
	entries := []store.WordEntry{
		{Word: "badword", Meta: trie.Metadata{Severity: 1}},
		{Word: "傻逼", Meta: trie.Metadata{Severity: 2}},
	}
	if err := store.WriteWordsTxt(path, entries); err != nil {
		t.Fatalf("seed words: %v", err)
	}
	return gui.NewService(path), path
}

func TestFilterTest(t *testing.T) {
	svc, _ := makeService(t)
	resp := svc.FilterTest(gui.FilterTestRequest{
		InputText:   "BADWORD and badword",
		IgnoreCase:  true,
		IgnoreWidth: true,
		ReplaceRune: "*",
		Scale:       2,
	})
	if resp.Error != "" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
	if resp.ElapsedNS < 0 {
		t.Fatalf("elapsed_ns should be >= 0, got %d", resp.ElapsedNS)
	}
	if resp.Result.HitCount < 2 {
		t.Fatalf("expected >= 2 hits, got %d", resp.Result.HitCount)
	}
	if !strings.Contains(resp.Result.Sanitized, "*******") {
		t.Fatalf("sanitized output unexpected: %q", resp.Result.Sanitized)
	}
}

func TestTrieAndACAndNormalize(t *testing.T) {
	svc, _ := makeService(t)

	trResp := svc.TrieTest(gui.TrieTestRequest{
		Action:   "add",
		Word:     "hello",
		Severity: 3,
		Scale:    3,
	})
	if trResp.Error != "" {
		t.Fatalf("trie add error: %s", trResp.Error)
	}
	if trResp.Result.Added != 3 {
		t.Fatalf("expected added=3, got %d", trResp.Result.Added)
	}

	acResp := svc.ACTest(gui.ACTestRequest{Text: "xxbadwordyy", Scale: 1})
	if acResp.Error != "" {
		t.Fatalf("ac error: %s", acResp.Error)
	}
	if acResp.Result.TotalMatches < 1 {
		t.Fatalf("expected matches, got %d", acResp.Result.TotalMatches)
	}

	normResp := svc.NormalizeTest(gui.NormalizeTestRequest{
		Text:        "ＡB",
		IgnoreCase:  true,
		IgnoreWidth: true,
		Scale:       2,
	})
	if normResp.Error != "" {
		t.Fatalf("normalize error: %s", normResp.Error)
	}
	if normResp.Result.NormalizedText != "ab\nab" {
		t.Fatalf("unexpected normalized text: %q", normResp.Result.NormalizedText)
	}
}

func TestLexiconImportListSave(t *testing.T) {
	svc, path := makeService(t)

	impResp := svc.LexiconTest(gui.LexiconTestRequest{
		Action:           "import",
		InputFileName:    "lexicon.csv",
		InputFileContent: "word,severity\nfoo,1\nbar,2\n",
		Scale:            1,
		Page:             1,
		PageSize:         50,
	})
	if impResp.Error != "" {
		t.Fatalf("import error: %s", impResp.Error)
	}
	if impResp.Result.Added < 2 {
		t.Fatalf("expected imported words added, got %d", impResp.Result.Added)
	}

	listResp := svc.LexiconTest(gui.LexiconTestRequest{Action: "list", Page: 1, PageSize: 10})
	if listResp.Error != "" {
		t.Fatalf("list error: %s", listResp.Error)
	}
	if listResp.Result.TotalEntries < 4 {
		t.Fatalf("expected >=4 entries, got %d", listResp.Result.TotalEntries)
	}

	saveResp := svc.LexiconTest(gui.LexiconTestRequest{Action: "save", Page: 1, PageSize: 10})
	if saveResp.Error != "" {
		t.Fatalf("save error: %s", saveResp.Error)
	}
	if saveResp.Result.SavedPath != path {
		t.Fatalf("saved path mismatch: got %q want %q", saveResp.Result.SavedPath, path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("saved file missing: %v", err)
	}
}

func TestLexiconImportTxtWithoutNumericHeader(t *testing.T) {
	svc, _ := makeService(t)

	impResp := svc.LexiconTest(gui.LexiconTestRequest{
		Action:           "import",
		InputFileName:    "lexicon.txt",
		InputFileContent: "田惠宇\nbadword\n", // badword duplicated with seeded words
		Scale:            1,
		Page:             1,
		PageSize:         50,
	})
	if impResp.Error != "" {
		t.Fatalf("import txt error: %s", impResp.Error)
	}
	if impResp.Result.Imported != 2 {
		t.Fatalf("expected imported=2, got %d", impResp.Result.Imported)
	}
	if impResp.Result.Added != 1 {
		t.Fatalf("expected added=1, got %d", impResp.Result.Added)
	}
	if impResp.Result.Duplicated != 1 {
		t.Fatalf("expected duplicated=1, got %d", impResp.Result.Duplicated)
	}

	listResp := svc.LexiconTest(gui.LexiconTestRequest{Action: "list", Page: 1, PageSize: 50})
	if listResp.Error != "" {
		t.Fatalf("list error: %s", listResp.Error)
	}
	found := false
	for _, e := range listResp.Result.Entries {
		if e.Word == "田惠宇" {
			found = true
			if e.Severity != trie.SeverityNormal {
				t.Fatalf("expected severity normal, got %d", e.Severity)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected imported word 田惠宇 in list")
	}
}

func TestInvalidActionReturnsError(t *testing.T) {
	svc, _ := makeService(t)
	resp := svc.TrieTest(gui.TrieTestRequest{Action: "unknown", Scale: 1})
	if resp.Error == "" {
		t.Fatal("expected error for unsupported action")
	}
}
