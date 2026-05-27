package sanitizer

import (
	"sort"
	"time"
	"unicode/utf8"

	"sensitive-filter/pkg/trie"
)

// MatchMode 控制命中结果的呈现方式。ModeFullWord 是标准的
// Aho-Corasick 多模式匹配语义：文本中每个模式的每次出现都会被报告。
// ModePrefix 使用相同的命中收集机制，但适用于将命中解释为
// 每个字符位置"前缀匹配"的调用方。
type MatchMode int

const (
	ModeFullWord MatchMode = iota
	ModePrefix
)

// Config 封装了传递给 New 的可调参数。零值即为安全的默认值
// （ReplaceRune 默认为 '*'，Mode 默认为 ModeFullWord）。
type Config struct {
	IgnoreCase  bool
	IgnoreWidth bool
	Mode        MatchMode
	ReplaceRune rune
}

func (c Config) normalized() Config {
	if c.ReplaceRune == 0 {
		c.ReplaceRune = '*'
	}
	return c
}

// Hit 描述一次模式匹配。位置同时以字符索引（便于流式偏移计算）
// 和原始输入的字节偏移量给出（以便调用方直接切片源字符串）。
type Hit struct {
	Word      string
	WordID    int
	Meta      trie.Metadata
	RuneStart int
	RuneEnd   int
	ByteStart int
	ByteEnd   int
}

// DebugResult 是 SanitizeWithDebug 和 StreamSession.GetStreamResult
// 的详细返回结构。
type DebugResult struct {
	Sanitized string
	Hits      []Hit
	Count     int
	Elapsed   time.Duration
}

// Sanitizer 将 trie 与匹配配置组合在一起。它是无状态的；多个 Sanitizer
// 可以共享同一个 trie，只要底层 trie 未被修改，单个 Sanitizer 可以安全地
// 从多个 goroutine 并发使用。
type Sanitizer struct {
	t   *trie.Trie
	cfg Config
}

// New 基于 t 构建一个 Sanitizer。如果 t 为 nil 则 panic。
func New(t *trie.Trie, cfg Config) *Sanitizer {
	if t == nil {
		panic("sanitizer: New called with nil Trie")
	}
	return &Sanitizer{t: t, cfg: cfg.normalized()}
}

func (s *Sanitizer) Trie() *trie.Trie { return s.t }

func (s *Sanitizer) Config() Config { return s.cfg }

type matchPoint struct {
	runeStart int
	runeEnd   int
	byteStart int
	byteEnd   int
	wordID    int
}

func (s *Sanitizer) scan(text string) []matchPoint {
	s.t.EnsureBuilt()

	var points []matchPoint
	cur := s.t.Root()
	runeIdx := 0
	runeBytePos := make([]int, 0, len(text)/2+1)
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		runeBytePos = append(runeBytePos, i)
		nr := s.t.Normalize(r)
		var matches []trie.Match
		cur, matches = s.t.Step(cur, nr)
		for _, m := range matches {
			start := runeIdx - m.RuneLen + 1
			points = append(points, matchPoint{
				runeStart: start,
				runeEnd:   runeIdx + 1,
				byteStart: runeBytePos[start],
				byteEnd:   i + size,
				wordID:    m.WordID,
			})
		}
		runeIdx++
		i += size
	}
	return points
}

func (s *Sanitizer) filterMode(points []matchPoint) []matchPoint {
	switch s.cfg.Mode {
	case ModePrefix, ModeFullWord:
		fallthrough
	default:
		return points
	}
}

func (s *Sanitizer) hitsFromPoints(points []matchPoint) []Hit {
	hits := make([]Hit, 0, len(points))
	for _, p := range points {
		word, _ := s.t.WordByID(p.wordID)
		meta, _ := s.t.MetaByID(p.wordID)
		hits = append(hits, Hit{
			Word:      word,
			WordID:    p.wordID,
			Meta:      meta,
			RuneStart: p.runeStart,
			RuneEnd:   p.runeEnd,
			ByteStart: p.byteStart,
			ByteEnd:   p.byteEnd,
		})
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].RuneStart != hits[j].RuneStart {
			return hits[i].RuneStart < hits[j].RuneStart
		}
		return hits[i].RuneEnd < hits[j].RuneEnd
	})
	return hits
}

// Sanitize 返回将所有命中替换为配置的 ReplaceRune 后的文本，
// 每个匹配的字符对应一个替换字符。重叠的匹配在替换前会被合并为最小覆盖区间。
func (s *Sanitizer) Sanitize(text string) string {
	points := s.scan(text)
	points = s.filterMode(points)
	return replaceByPoints(text, points, s.cfg.ReplaceRune)
}

// SanitizeWithDebug 在 Sanitize 的基础上额外返回每次命中的元数据、命中次数和
// 匹配耗时。适用于工具/调试场景，不建议用于热路径。
func (s *Sanitizer) SanitizeWithDebug(text string) DebugResult {
	start := time.Now()
	points := s.scan(text)
	points = s.filterMode(points)
	sanitized := replaceByPoints(text, points, s.cfg.ReplaceRune)
	hits := s.hitsFromPoints(points)
	return DebugResult{
		Sanitized: sanitized,
		Hits:      hits,
		Count:     len(hits),
		Elapsed:   time.Since(start),
	}
}

func replaceByPoints(text string, points []matchPoint, replace rune) string {
	if len(points) == 0 {
		return text
	}
	runes := []rune(text)
	intervals := make([][2]int, 0, len(points))
	for _, p := range points {
		intervals = append(intervals, [2]int{p.runeStart, p.runeEnd})
	}
	sort.Slice(intervals, func(i, j int) bool {
		if intervals[i][0] != intervals[j][0] {
			return intervals[i][0] < intervals[j][0]
		}
		return intervals[i][1] < intervals[j][1]
	})
	merged := intervals[:0]
	curStart, curEnd := intervals[0][0], intervals[0][1]
	for _, iv := range intervals[1:] {
		if iv[0] <= curEnd {
			if iv[1] > curEnd {
				curEnd = iv[1]
			}
		} else {
			merged = append(merged, [2]int{curStart, curEnd})
			curStart, curEnd = iv[0], iv[1]
		}
	}
	merged = append(merged, [2]int{curStart, curEnd})
	for _, iv := range merged {
		for i := iv[0]; i < iv[1] && i < len(runes); i++ {
			runes[i] = replace
		}
	}
	return string(runes)
}
