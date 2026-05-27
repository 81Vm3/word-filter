package main

import (
	"fmt"
	"strings"

	"sensitive-filter/pkg/sanitizer"
	"sensitive-filter/pkg/store"
	"sensitive-filter/pkg/trie"
)

func main() {
	// 1. 从 words.txt 加载敏感词到 Trie
	tr := trie.New(trie.BuildNormalizer(trie.NormalizeOptions{IgnoreCase: true, IgnoreWidth: true}))
	added, dup, err := store.ImportTxt(tr, "words.txt")
	if err != nil {
		fmt.Printf("import words.txt: %v\n", err)
	} else {
		fmt.Printf("[txt] added=%d dup=%d total=%d\n", added, dup, tr.Size())
	}

	// 2. 运行时增量 CRUD
	tr.Add("BadCase", trie.Metadata{Severity: 0})
	tr.Add("badword", trie.Metadata{Severity: 9}) // 已存在，会被去重
	tr.Update("傻逼", trie.Metadata{Severity: 3})   // 调整严重度
	tr.Delete("不存在")                              // 不报错

	// 3. 构造过滤器
	san := sanitizer.New(tr, sanitizer.Config{
		IgnoreCase:  true,
		IgnoreWidth: true,
		Mode:        sanitizer.ModeFullWord,
		ReplaceRune: '*',
	})

	// 4. 一次性 sanitize
	fmt.Println(strings.Repeat("-", 60))
	in := "你这个ＢＡＤＷＯＲＤ的BadCase！傻逼真的测试敏感"
	fmt.Printf("input:     %q\n", in)
	fmt.Printf("sanitized: %q\n", san.Sanitize(in))

	// 5. 带 debug 信息的 sanitize
	fmt.Println(strings.Repeat("-", 60))
	dbg := san.SanitizeWithDebug("测试敏感词 badword 重叠 BadCaseBADWORD")
	fmt.Printf("count=%d elapsed=%s\n", dbg.Count, dbg.Elapsed)
	fmt.Printf("sanitized: %q\n", dbg.Sanitized)
	for _, h := range dbg.Hits {
		fmt.Printf("  hit %q  rune[%d,%d) byte[%d,%d) sev=%d\n",
			h.Word, h.RuneStart, h.RuneEnd, h.ByteStart, h.ByteEnd, h.Meta.Severity)
	}

	// 6. 流式接口：跨 chunk 命中（"bad"+"word"、"傻"+"逼"）
	fmt.Println(strings.Repeat("-", 60))
	ss := san.InitStream()
	for _, chunk := range []string{"hello bad", "word world 傻", "逼!"} {
		ss.AppendToStream(chunk)
	}
	r := ss.GetStreamResult()
	fmt.Printf("[stream] hits=%d sanitized=%q\n", r.Count, r.Sanitized)
	for _, h := range r.Hits {
		fmt.Printf("  stream-hit %q rune[%d,%d)\n", h.Word, h.RuneStart, h.RuneEnd)
	}
	ss.ClearStream()

	// 7. 持久化回 words.txt（开发者决定何时保存）
	fmt.Println(strings.Repeat("-", 60))
	if err := store.WriteWordsTxt("words.out.txt", store.CollectFromTrie(tr)); err != nil {
		fmt.Printf("save: %v\n", err)
	} else {
		fmt.Printf("saved %d entries to words.out.txt\n", tr.Size())
	}
}
