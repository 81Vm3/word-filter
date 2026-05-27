package sanitizer

import (
	"time"
	"unicode/utf8"

	"sensitive-filter/pkg/trie"
)

// StreamSession 是增量匹配的会话状态。由 Sanitizer.InitStream 返回，
// 不保证多 goroutine 并发安全。每次 AppendToStream 调用都会扩展流；
// 累计结果通过 GetStreamResult 读取；ClearStream 将会话重置为初始状态以便复用。
//
// 该会话不会持有 trie 的长期锁；调用方必须避免在 AppendToStream 调用之间
// 对 trie 进行 CRUD 操作，否则游标状态将引用过期的自动机（参见包文档）。
type StreamSession struct {
	s          *Sanitizer
	cur        trie.Cursor
	runeOff    int
	byteOff    int
	pendingBuf []byte
	hits       []matchPoint
	processed  []byte
	runeStarts []int
}

// InitStream 启动一个新的流式会话。会预先调用一次 EnsureBuilt，
// 使首次 AppendToStream 不需要承担自动机重建的开销。
func (s *Sanitizer) InitStream() *StreamSession {
	s.t.EnsureBuilt()
	return &StreamSession{
		s:          s,
		cur:        s.t.Root(),
		runeStarts: make([]int, 0, 128),
	}
}

// AppendToStream 消费下一个数据块。尾部未完整的 UTF-8 字节序列会被暂存，
// 并在下次调用时前置到下一个数据块之前。
func (ss *StreamSession) AppendToStream(chunk string) {
	if chunk == "" {
		return
	}

	var combined []byte
	if len(ss.pendingBuf) == 0 {
		combined = []byte(chunk)
	} else {
		combined = make([]byte, 0, len(ss.pendingBuf)+len(chunk))
		combined = append(combined, ss.pendingBuf...)
		combined = append(combined, chunk...)
		ss.pendingBuf = ss.pendingBuf[:0]
	}

	t := ss.s.t
	i := 0
	for i < len(combined) {
		if !utf8.FullRune(combined[i:]) {
			ss.pendingBuf = append(ss.pendingBuf[:0], combined[i:]...)
			combined = combined[:i]
			break
		}
		r, size := utf8.DecodeRune(combined[i:])
		byteStartAbs := ss.byteOff + i
		ss.runeStarts = append(ss.runeStarts, byteStartAbs)
		nr := t.Normalize(r)
		var matches []trie.Match
		ss.cur, matches = t.Step(ss.cur, nr)
		for _, m := range matches {
			start := ss.runeOff - m.RuneLen + 1
			byteStart := 0
			if start >= 0 && start < len(ss.runeStarts) {
				byteStart = ss.runeStarts[start]
			}
			ss.hits = append(ss.hits, matchPoint{
				runeStart: start,
				runeEnd:   ss.runeOff + 1,
				byteStart: byteStart,
				byteEnd:   byteStartAbs + size,
				wordID:    m.WordID,
			})
		}
		ss.runeOff++
		i += size
	}

	if i > 0 {
		ss.processed = append(ss.processed, combined[:i]...)
		ss.byteOff += i
	}
}

// GetStreamResult 从目前为止消费的所有数据中组装过滤后的文本和命中列表。
// 调用该方法不会重置会话——后续的 AppendToStream 调用将继续累积。
func (ss *StreamSession) GetStreamResult() DebugResult {
	start := time.Now()
	processedText := string(ss.processed)
	points := ss.s.filterMode(ss.hits)
	sanitized := replaceByPoints(processedText, points, ss.s.cfg.ReplaceRune)
	hits := ss.s.hitsFromPoints(points)
	return DebugResult{
		Sanitized: sanitized,
		Hits:      hits,
		Count:     len(hits),
		Elapsed:   time.Since(start),
	}
}

// ClearStream 将会话重置为 InitStream 时的状态。会话对象可以立即用于新的流。
func (ss *StreamSession) ClearStream() {
	ss.cur = ss.s.t.Root()
	ss.runeOff = 0
	ss.byteOff = 0
	ss.pendingBuf = ss.pendingBuf[:0]
	ss.hits = ss.hits[:0]
	ss.processed = ss.processed[:0]
	ss.runeStarts = ss.runeStarts[:0]
}

// PendingBytes 返回最近一次 AppendToStream 中尚未解码的尾部字节数
// （即不完整的 UTF-8 序列）。主要用于测试和诊断。
func (ss *StreamSession) PendingBytes() int { return len(ss.pendingBuf) }

// ProcessedRunes 返回自 InitStream / ClearStream 以来所有 AppendToStream 调用
// 成功消费的字符总数。
func (ss *StreamSession) ProcessedRunes() int { return ss.runeOff }
