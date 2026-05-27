package store

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"sensitive-filter/pkg/trie"
)

// ImportCSV 批量导入 CSV 文件（列：word, severity[, extra_json]）。
// 自动检测并跳过表头行 "word,severity[,extra]"（不区分大小写）。
// 返回新增数和重复数。
func ImportCSV(t *trie.Trie, path string) (added, dup int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	rowIdx := 0
	headerSkipped := false
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return added, dup, fmt.Errorf("row %d: %w", rowIdx, err)
		}
		rowIdx++

		if len(rec) == 0 {
			continue
		}
		if !headerSkipped {
			headerSkipped = true
			if isHeaderRow(rec) {
				continue
			}
		}

		word := strings.TrimSpace(rec[0])
		if word == "" {
			continue
		}
		meta := trie.DefaultMetadata()
		if len(rec) >= 2 {
			sevStr := strings.TrimSpace(rec[1])
			if sevStr != "" {
				if sev, err := strconv.Atoi(sevStr); err == nil {
					meta.Severity = sev
				}
			}
		}
		if len(rec) >= 3 {
			extraStr := strings.TrimSpace(rec[2])
			if extraStr != "" {
				var extra map[string]string
				if err := json.Unmarshal([]byte(extraStr), &extra); err == nil {
					meta.Extra = extra
				}
			}
		}

		if _, ok := t.Add(word, meta); ok {
			added++
		} else {
			dup++
		}
	}
	return added, dup, nil
}

func isHeaderRow(rec []string) bool {
	if len(rec) == 0 {
		return false
	}
	first := strings.ToLower(strings.TrimSpace(rec[0]))
	if first == "word" || first == "keyword" || first == "term" {
		return true
	}
	if len(rec) >= 2 {
		second := strings.ToLower(strings.TrimSpace(rec[1]))
		if (first == "word" || first == "term" || first == "keyword") &&
			(second == "severity" || second == "level" || second == "sev") {
			return true
		}
	}
	return false
}
