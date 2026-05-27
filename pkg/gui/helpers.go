package gui

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"sensitive-filter/pkg/store"
	"sensitive-filter/pkg/trie"
)

func timingFrom(start time.Time) Timing {
	d := time.Since(start)
	if d <= 0 {
		d = time.Nanosecond
	}
	return Timing{ElapsedNS: d.Nanoseconds(), ElapsedMS: float64(d.Nanoseconds()) / 1e6}
}

func normalizeScale(scale int) int {
	if scale < 1 {
		return 1
	}
	if scale > 10000 {
		return 10000
	}
	return scale
}

func scaledText(input string, scale int) string {
	scale = normalizeScale(scale)
	if scale == 1 || input == "" {
		return input
	}
	var b strings.Builder
	b.Grow(len(input)*scale + scale - 1)
	for i := 0; i < scale; i++ {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(input)
	}
	return b.String()
}

func textFromInput(input, fileName, fileContent string) (string, error) {
	if fileContent == "" {
		return input, nil
	}
	if strings.HasSuffix(strings.ToLower(fileName), ".csv") {
		r := csv.NewReader(strings.NewReader(fileContent))
		r.FieldsPerRecord = -1
		rows, err := r.ReadAll()
		if err != nil {
			return "", fmt.Errorf("read csv input: %w", err)
		}
		parts := make([]string, 0, len(rows))
		headerChecked := false
		for _, row := range rows {
			if len(row) == 0 {
				continue
			}
			if !headerChecked {
				headerChecked = true
				first := strings.ToLower(strings.TrimSpace(row[0]))
				if first == "word" || first == "text" || first == "content" {
					continue
				}
			}
			cell := strings.TrimSpace(row[0])
			if cell == "" {
				continue
			}
			parts = append(parts, cell)
		}
		if len(parts) == 0 {
			return "", nil
		}
		return strings.Join(parts, "\n"), nil
	}
	return fileContent, nil
}

func parseExtra(extraJSON string) (map[string]string, error) {
	extraJSON = strings.TrimSpace(extraJSON)
	if extraJSON == "" {
		return nil, nil
	}
	var out map[string]string
	if err := json.Unmarshal([]byte(extraJSON), &out); err != nil {
		return nil, fmt.Errorf("parse extra_json: %w", err)
	}
	return out, nil
}

func scaledWords(base string, scale int) []string {
	scale = normalizeScale(scale)
	if scale == 1 {
		return []string{base}
	}
	words := make([]string, 0, scale)
	for i := 0; i < scale; i++ {
		words = append(words, fmt.Sprintf("%s_%d", base, i+1))
	}
	return words
}

func collectSample(t *trie.Trie, limit int) []WordEntryView {
	if limit < 1 {
		limit = maxPreviewRows
	}
	entries := make([]store.WordEntry, 0, t.Size())
	t.Walk(func(word string, m trie.Metadata) {
		entries = append(entries, store.WordEntry{Word: word, Meta: m})
	})
	if len(entries) == 0 {
		return nil
	}
	sortEntries(entries)
	if len(entries) > limit {
		entries = entries[:limit]
	}
	out := make([]WordEntryView, 0, len(entries))
	for _, e := range entries {
		out = append(out, WordEntryView{Word: e.Word, Severity: e.Meta.Severity, Extra: e.Meta.Clone().Extra})
	}
	return out
}

func sortEntries(entries []store.WordEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Word < entries[j].Word
	})
}

func wordEntriesFromFile(name, content string) ([]store.WordEntry, error) {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, ".csv") {
		return parseCSVEntries(content)
	}
	return parseTxtEntries(content)
}

func parseTxtEntries(content string) ([]store.WordEntry, error) {
	lines := splitLines(content)
	if len(lines) == 0 {
		return nil, nil
	}
	firstLine := lines[0]
	header := strings.TrimSpace(firstLine)
	if header == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(header)
	if err != nil {
		entries := make([]store.WordEntry, 0, len(lines))
		for _, word := range lines {
			word = strings.TrimSpace(word)
			if word == "" {
				continue
			}
			entries = append(entries, store.WordEntry{
				Word: word,
				Meta: trie.DefaultMetadata(),
			})
		}
		return entries, nil
	}
	if n < 0 {
		return nil, fmt.Errorf("invalid txt header: count must be non-negative")
	}
	if len(lines) < n+1 {
		return nil, fmt.Errorf("expected %d word lines, got %d", n, len(lines)-1)
	}
	words := lines[1 : n+1]
	severities := lines[n+1:]
	entries := make([]store.WordEntry, 0, n)
	for i, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}
		meta := trie.DefaultMetadata()
		if i < len(severities) {
			if sev, err := strconv.Atoi(strings.TrimSpace(severities[i])); err == nil {
				meta.Severity = sev
			}
		}
		entries = append(entries, store.WordEntry{Word: word, Meta: meta})
	}
	return entries, nil
}

func parseCSVEntries(content string) ([]store.WordEntry, error) {
	r := csv.NewReader(strings.NewReader(content))
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	entries := make([]store.WordEntry, 0, len(rows))
	headerChecked := false
	for _, row := range rows {
		if len(row) == 0 {
			continue
		}
		if !headerChecked {
			headerChecked = true
			first := strings.ToLower(strings.TrimSpace(row[0]))
			if first == "word" || first == "keyword" || first == "term" {
				continue
			}
		}
		word := strings.TrimSpace(row[0])
		if word == "" {
			continue
		}
		meta := trie.DefaultMetadata()
		if len(row) >= 2 {
			if sev, err := strconv.Atoi(strings.TrimSpace(row[1])); err == nil {
				meta.Severity = sev
			}
		}
		if len(row) >= 3 {
			extra, err := parseExtra(strings.TrimSpace(row[2]))
			if err == nil {
				meta.Extra = extra
			}
		}
		entries = append(entries, store.WordEntry{Word: word, Meta: meta})
	}
	return entries, nil
}

func splitLines(content string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.TrimSuffix(content, "\n")
	if content == "" {
		return nil
	}
	return strings.Split(content, "\n")
}

func runeCount(s string) int {
	return utf8.RuneCountInString(s)
}
