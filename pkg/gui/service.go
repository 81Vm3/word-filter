package gui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"sensitive-filter/pkg/sanitizer"
	"sensitive-filter/pkg/store"
	"sensitive-filter/pkg/trie"
)

// Service 提供给 GUI 的统一测试入口，避免前端直接依赖底层包细节。
type Service struct {
	mu        sync.Mutex
	t         *trie.Trie
	wordsPath string
}

func NewService(wordsPath string) *Service {
	if wordsPath == "" {
		wordsPath = DefaultWordsPath
	}
	t := trie.New(trie.BuildNormalizer(trie.NormalizeOptions{IgnoreCase: true, IgnoreWidth: true}))
	s := &Service{t: t, wordsPath: wordsPath}
	s.loadDefaultWords()
	return s
}

func (s *Service) loadDefaultWords() {
	if _, err := os.Stat(s.wordsPath); err != nil {
		return
	}
	_, _, _ = store.ImportTxt(s.t, s.wordsPath)
}

func (s *Service) FilterTest(req FilterTestRequest) FilterTestResponse {
	start := time.Now()
	resp := FilterTestResponse{}
	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		resp.Timing = timingFrom(start)
	}()

	baseText, err := textFromInput(req.InputText, req.InputFileName, req.InputFileContent)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	text := scaledText(baseText, req.Scale)
	r := []rune(req.ReplaceRune)
	replaceRune := '*'
	if len(r) > 0 {
		replaceRune = r[0]
	}

	normTrie := trie.New(trie.BuildNormalizer(trie.NormalizeOptions{
		IgnoreCase:  req.IgnoreCase,
		IgnoreWidth: req.IgnoreWidth,
	}))
	_, _, _ = store.ImportEntries(normTrie, store.CollectFromTrie(s.t))

	san := sanitizer.New(normTrie, sanitizer.Config{
		IgnoreCase:  req.IgnoreCase,
		IgnoreWidth: req.IgnoreWidth,
		Mode:        sanitizer.ModeFullWord,
		ReplaceRune: replaceRune,
	})
	dbg := san.SanitizeWithDebug(text)

	hits := make([]FilterHit, 0, minInt(len(dbg.Hits), maxPreviewRows))
	truncated := false
	for i, h := range dbg.Hits {
		if i >= maxPreviewRows {
			truncated = true
			break
		}
		hits = append(hits, FilterHit{
			Word:      h.Word,
			Severity:  h.Meta.Severity,
			Extra:     h.Meta.Extra,
			RuneStart: h.RuneStart,
			RuneEnd:   h.RuneEnd,
			ByteStart: h.ByteStart,
			ByteEnd:   h.ByteEnd,
		})
	}

	resp.Result = FilterTestResult{
		InputRunes: runeCount(text),
		InputBytes: len(text),
		Sanitized:  dbg.Sanitized,
		HitCount:   dbg.Count,
		Hits:       hits,
		Truncated:  truncated,
	}
	return resp
}

func (s *Service) TrieTest(req TrieTestRequest) TrieTestResponse {
	start := time.Now()
	resp := TrieTestResponse{}
	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		resp.Timing = timingFrom(start)
	}()

	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action == "" {
		action = "walk"
	}
	scale := normalizeScale(req.Scale)
	meta := trie.Metadata{Severity: req.Severity}
	extra, err := parseExtra(req.ExtraJSON)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	meta.Extra = extra

	result := TrieTestResult{Action: action, Requested: scale}
	word := strings.TrimSpace(req.Word)

	switch action {
	case "add":
		if word == "" {
			resp.Error = "word is required for add"
			return resp
		}
		for _, w := range scaledWords(word, scale) {
			if _, ok := s.t.Add(w, meta); ok {
				result.Added++
			} else {
				result.Duplicated++
			}
		}
	case "update":
		if word == "" {
			resp.Error = "word is required for update"
			return resp
		}
		for _, w := range scaledWords(word, scale) {
			if s.t.Update(w, meta) {
				result.Updated++
			}
		}
	case "delete":
		if word == "" {
			resp.Error = "word is required for delete"
			return resp
		}
		for _, w := range scaledWords(word, scale) {
			if s.t.Delete(w) {
				result.Deleted++
			}
		}
	case "has":
		if word == "" {
			resp.Error = "word is required for has"
			return resp
		}
		for i := 0; i < scale; i++ {
			result.Exists = s.t.Has(word)
		}
	case "get":
		if word == "" {
			resp.Error = "word is required for get"
			return resp
		}
		for i := 0; i < scale; i++ {
			metaOut, ok := s.t.Get(word)
			result.Exists = ok
			if ok {
				result.Metadata = metaOut
			}
		}
	case "walk":
		for i := 0; i < scale; i++ {
			walked := 0
			s.t.Walk(func(_ string, _ trie.Metadata) {
				walked++
			})
			result.Walked = walked
		}
	default:
		resp.Error = fmt.Sprintf("unsupported trie action: %s", action)
		return resp
	}

	result.TotalEntries = s.t.Size()
	result.Sample = collectSample(s.t, 20)
	resp.Result = result
	return resp
}

func (s *Service) ACTest(req ACTestRequest) ACTestResponse {
	start := time.Now()
	resp := ACTestResponse{}
	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		resp.Timing = timingFrom(start)
	}()

	text := scaledText(req.Text, req.Scale)
	s.t.EnsureBuilt()
	cur := s.t.Root()
	idx := 0
	totalMatches := 0
	matches := make([]ACMatchView, 0, maxPreviewRows)
	truncated := false

	for _, r := range text {
		nr := s.t.Normalize(r)
		var ms []trie.Match
		cur, ms = s.t.Step(cur, nr)
		totalMatches += len(ms)
		for _, m := range ms {
			if len(matches) >= maxPreviewRows {
				truncated = true
				continue
			}
			word, _ := s.t.WordByID(m.WordID)
			matches = append(matches, ACMatchView{
				Word:      word,
				WordID:    m.WordID,
				RuneStart: idx - m.RuneLen + 1,
				RuneEnd:   idx + 1,
			})
		}
		idx++
	}

	resp.Result = ACTestResult{
		InputRunes:   runeCount(text),
		Steps:        idx,
		TotalMatches: totalMatches,
		Matches:      matches,
		Truncated:    truncated,
	}
	return resp
}

func (s *Service) NormalizeTest(req NormalizeTestRequest) NormalizeTestResponse {
	start := time.Now()
	resp := NormalizeTestResponse{}
	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		resp.Timing = timingFrom(start)
	}()

	text := scaledText(req.Text, req.Scale)
	norm := trie.BuildNormalizer(trie.NormalizeOptions{IgnoreCase: req.IgnoreCase, IgnoreWidth: req.IgnoreWidth})
	pairs := make([]NormalizePair, 0, minInt(runeCount(text), maxPreviewRows))
	changed := 0
	var out strings.Builder
	out.Grow(len(text))
	idx := 0
	truncated := false
	for _, r := range text {
		n := norm(r)
		if r != n {
			changed++
		}
		out.WriteRune(n)
		if idx < maxPreviewRows {
			pairs = append(pairs, NormalizePair{Index: idx, Original: string(r), Normalized: string(n), Changed: r != n})
		} else {
			truncated = true
		}
		idx++
	}
	resp.Result = NormalizeTestResult{
		InputRunes:     idx,
		OutputRunes:    utf8.RuneCountInString(out.String()),
		ChangedRunes:   changed,
		NormalizedText: out.String(),
		Pairs:          pairs,
		Truncated:      truncated,
	}
	return resp
}

func (s *Service) LexiconTest(req LexiconTestRequest) LexiconTestResponse {
	start := time.Now()
	resp := LexiconTestResponse{}
	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		resp.Timing = timingFrom(start)
	}()

	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action == "" {
		action = "list"
	}
	scale := normalizeScale(req.Scale)
	result := LexiconTestResult{Action: action}
	meta := trie.Metadata{Severity: req.Severity}
	extra, err := parseExtra(req.ExtraJSON)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	meta.Extra = extra
	word := strings.TrimSpace(req.Word)

	switch action {
	case "import":
		entries, err := wordEntriesFromFile(req.InputFileName, req.InputFileContent)
		if err != nil {
			resp.Error = err.Error()
			return resp
		}
		if scale > 1 && len(entries) > 0 {
			expanded := make([]store.WordEntry, 0, len(entries)*scale)
			for i := 0; i < scale; i++ {
				for _, e := range entries {
					wordValue := e.Word
					if i > 0 {
						wordValue = fmt.Sprintf("%s_%d", e.Word, i+1)
					}
					expanded = append(expanded, store.WordEntry{Word: wordValue, Meta: e.Meta})
				}
			}
			entries = expanded
		}
		added, dup, err := store.ImportEntries(s.t, entries)
		if err != nil {
			resp.Error = err.Error()
			return resp
		}
		result.Imported = len(entries)
		result.Added = added
		result.Duplicated = dup
		result.Message = fmt.Sprintf("imported=%d added=%d dup=%d", len(entries), added, dup)
	case "add":
		if word == "" {
			resp.Error = "word is required for add"
			return resp
		}
		for _, w := range scaledWords(word, scale) {
			if _, ok := s.t.Add(w, meta); ok {
				result.Added++
			} else {
				result.Duplicated++
			}
		}
	case "update":
		if word == "" {
			resp.Error = "word is required for update"
			return resp
		}
		for _, w := range scaledWords(word, scale) {
			if s.t.Update(w, meta) {
				result.Updated++
			}
		}
	case "delete":
		if word == "" {
			resp.Error = "word is required for delete"
			return resp
		}
		for _, w := range scaledWords(word, scale) {
			if s.t.Delete(w) {
				result.Deleted++
			}
		}
	case "save":
		if err := store.WriteWordsTxt(s.wordsPath, store.CollectFromTrie(s.t)); err != nil {
			resp.Error = err.Error()
			return resp
		}
		result.SavedPath = s.wordsPath
		result.Message = "saved"
	case "list":
		// handled below
	default:
		resp.Error = fmt.Sprintf("unsupported lexicon action: %s", action)
		return resp
	}

	allEntries := collectSample(s.t, s.t.Size())
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	startIdx := (page - 1) * pageSize
	if startIdx > len(allEntries) {
		startIdx = len(allEntries)
	}
	endIdx := startIdx + pageSize
	if endIdx > len(allEntries) {
		endIdx = len(allEntries)
	}
	result.Page = page
	result.PageSize = pageSize
	result.TotalEntries = s.t.Size()
	result.Entries = allEntries[startIdx:endIdx]
	resp.Result = result
	return resp
}

func (s *Service) WordsPath() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.wordsPath
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
