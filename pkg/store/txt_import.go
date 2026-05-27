package store

import (
	"sensitive-filter/pkg/trie"
)

// ImportTxt 读取路径为 path 的 words.txt 并将所有词条添加到 t 中。
// 返回新增词数和跳过的重复词数。导入器不会回写磁盘；持久化由调用方决定。
func ImportTxt(t *trie.Trie, path string) (added, dup int, err error) {
	entries, err := ReadWordsTxt(path)
	if err != nil {
		return 0, 0, err
	}
	return ImportEntries(t, entries)
}

// ImportEntries 批量将词条添加到 trie 中，按归一化后的词语去重
// （应用 trie 的归一化器）。返回新增数和重复数。
func ImportEntries(t *trie.Trie, entries []WordEntry) (added, dup int, err error) {
	for _, e := range entries {
		if e.Word == "" {
			continue
		}
		_, ok := t.Add(e.Word, e.Meta)
		if ok {
			added++
		} else {
			dup++
		}
	}
	return added, dup, nil
}
