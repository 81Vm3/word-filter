package store

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"sensitive-filter/pkg/trie"
)

// WordEntry 是磁盘与内存 trie 之间传递的 (word, metadata) 对。
// 它是 TXT 和 CSV 格式导入/导出的基本单元。
type WordEntry struct {
	Word string
	Meta trie.Metadata
}

// ReadWordsTxt 解析 words.txt，支持两种格式：
//
// 1) 计数头格式：
//
//	第 1 行：          总数 n
//	第 2..n+1 行：    n 个敏感词（每行一个）
//	第 n+2..2n+1 行：n 个严重度整数（按索引与词语对应）
//
// 如果末尾缺少元数据行，则回退到默认严重度。
//
// 2) 纯词条格式：
//
//	首行不是整数时，按“每行一个敏感词”读取整文件，
//	所有词条严重度均使用默认值（SeverityNormal）。
func ReadWordsTxt(path string) ([]WordEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024*16)

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}
	firstLine := scanner.Text()
	header := strings.TrimSpace(firstLine)
	if header == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(header)
	if err != nil {
		words := []string{firstLine}
		for scanner.Scan() {
			words = append(words, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		entries := make([]WordEntry, 0, len(words))
		for _, w := range words {
			if w == "" {
				continue
			}
			entries = append(entries, WordEntry{
				Word: w,
				Meta: trie.DefaultMetadata(),
			})
		}
		return entries, nil
	}
	if n < 0 {
		return nil, fmt.Errorf("invalid header: count must be non-negative, got %d", n)
	}

	words := make([]string, 0, n)
	for i := 0; i < n; i++ {
		if !scanner.Scan() {
			return nil, fmt.Errorf("expected %d words but only read %d", n, i)
		}
		words = append(words, scanner.Text())
	}

	metas := make([]trie.Metadata, n)
	for i := 0; i < n; i++ {
		if !scanner.Scan() {
			metas[i] = trie.DefaultMetadata()
			continue
		}
		line := strings.TrimSpace(scanner.Text())
		metas[i] = parseMetadataLine(line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	entries := make([]WordEntry, 0, n)
	for i, w := range words {
		if w == "" {
			continue
		}
		entries = append(entries, WordEntry{Word: w, Meta: metas[i]})
	}
	return entries, nil
}

func parseMetadataLine(line string) trie.Metadata {
	if line == "" {
		return trie.DefaultMetadata()
	}
	sev, err := strconv.Atoi(line)
	if err != nil {
		return trie.DefaultMetadata()
	}
	return trie.Metadata{Severity: sev}
}

// WriteWordsTxt 将词条序列化为 words.txt 格式。词语顺序保持不变；
// 元数据按相同顺序写入。
func WriteWordsTxt(path string, entries []WordEntry) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	if _, err := fmt.Fprintf(w, "%d\n", len(entries)); err != nil {
		return err
	}
	for _, e := range entries {
		if _, err := fmt.Fprintf(w, "%s\n", e.Word); err != nil {
			return err
		}
	}
	for _, e := range entries {
		if _, err := fmt.Fprintf(w, "%d\n", e.Meta.Severity); err != nil {
			return err
		}
	}
	return nil
}

// CollectFromTrie 遍历 trie 并返回其中所有的 (word, metadata) 对。
// 可在调用 WriteWordsTxt 之前使用此函数收集数据。
func CollectFromTrie(t *trie.Trie) []WordEntry {
	entries := make([]WordEntry, 0, t.Size())
	t.Walk(func(word string, m trie.Metadata) {
		entries = append(entries, WordEntry{Word: word, Meta: m})
	})
	return entries
}
