package trie

import (
	"sync"
)

type node struct {
	children map[rune]*node
	wordID   int
	runeLen  int
	fail     *node
	dictLink *node
}

func newNode() *node {
	return &node{
		children: make(map[rune]*node),
		wordID:   -1,
	}
}

func (n *node) isTerminal() bool { return n.wordID >= 0 }

// Cursor 是 trie 中的不透明位置，供执行流式或一次性匹配的调用方使用。
// 游标只能与其来源的 Trie 配合使用。
type Cursor struct {
	n *node
}

// Match 表示 Step 报告的一次模式命中。RuneLen 是模式的字符长度
// （调用方需通过从累计字符索引中减去 RuneLen 来计算绝对字符位置）。
type Match struct {
	WordID  int
	RuneLen int
}

// Trie 存储敏感词及其元数据，并作为 Aho-Corasick 自动机的底层存储。
// 内部的 RWMutex 保护 CRUD 并发安全；匹配侧 API（Root, Step, Normalize,
// MetaByID, WordByID）不保证与 CRUD 操作并发安全——调用方应通过
// EnsureBuilt 来协调匹配与修改操作。
type Trie struct {
	root  *node
	metas map[int]Metadata
	words map[string]int
	ids   map[int]string
	next  int
	dirty bool
	norm  NormFunc
	mu    sync.RWMutex
}

func New(norm NormFunc) *Trie {
	if norm == nil {
		norm = identityRune
	}
	return &Trie{
		root:  newNode(),
		metas: make(map[int]Metadata),
		words: make(map[string]int),
		ids:   make(map[int]string),
		norm:  norm,
	}
}

func (t *Trie) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.words)
}

func (t *Trie) Normalizer() NormFunc { return t.norm }

func (t *Trie) Normalize(r rune) rune {
	if t.norm == nil {
		return r
	}
	return t.norm(r)
}

func (t *Trie) Add(word string, m Metadata) (id int, added bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.addLocked(word, m)
}

func (t *Trie) addLocked(word string, m Metadata) (int, bool) {
	if word == "" {
		return -1, false
	}
	key := NormalizeString(word, t.norm)
	if key == "" {
		return -1, false
	}
	if existingID, ok := t.words[key]; ok {
		return existingID, false
	}
	cur := t.root
	runeLen := 0
	for _, r := range key {
		child, ok := cur.children[r]
		if !ok {
			child = newNode()
			cur.children[r] = child
		}
		cur = child
		runeLen++
	}
	id := t.next
	t.next++
	cur.wordID = id
	cur.runeLen = runeLen
	t.metas[id] = m.Clone()
	t.words[key] = id
	t.ids[id] = key
	t.dirty = true
	return id, true
}

func (t *Trie) Update(word string, m Metadata) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	key := NormalizeString(word, t.norm)
	id, ok := t.words[key]
	if !ok {
		return false
	}
	t.metas[id] = m.Clone()
	return true
}

func (t *Trie) Delete(word string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	key := NormalizeString(word, t.norm)
	id, ok := t.words[key]
	if !ok {
		return false
	}
	t.removePath(key)
	delete(t.words, key)
	delete(t.metas, id)
	delete(t.ids, id)
	t.dirty = true
	return true
}

func (t *Trie) removePath(key string) {
	runes := []rune(key)
	path := make([]*node, 0, len(runes)+1)
	path = append(path, t.root)
	cur := t.root
	for _, r := range runes {
		child, ok := cur.children[r]
		if !ok {
			return
		}
		path = append(path, child)
		cur = child
	}
	cur.wordID = -1
	cur.runeLen = 0
	for i := len(path) - 1; i > 0; i-- {
		n := path[i]
		if n.isTerminal() || len(n.children) > 0 {
			break
		}
		parent := path[i-1]
		delete(parent.children, runes[i-1])
	}
}

func (t *Trie) Get(word string) (Metadata, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	key := NormalizeString(word, t.norm)
	id, ok := t.words[key]
	if !ok {
		return Metadata{}, false
	}
	return t.metas[id].Clone(), true
}

func (t *Trie) Has(word string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, ok := t.words[NormalizeString(word, t.norm)]
	return ok
}

func (t *Trie) Walk(fn func(word string, m Metadata)) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for word, id := range t.words {
		fn(word, t.metas[id].Clone())
	}
}

func (t *Trie) WordByID(id int) (string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	w, ok := t.ids[id]
	return w, ok
}

func (t *Trie) MetaByID(id int) (Metadata, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	m, ok := t.metas[id]
	if !ok {
		return Metadata{}, false
	}
	return m.Clone(), true
}

// Root 返回指向 trie 根节点的游标。如果 trie 自上次匹配以来可能被修改过，
// 请先调用 EnsureBuilt。
func (t *Trie) Root() Cursor { return Cursor{n: t.root} }
