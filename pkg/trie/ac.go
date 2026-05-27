package trie

// EnsureBuilt 在自上次构建以来发生过 CRUD 操作时重建 Aho-Corasick 的
// fail/dictLink 结构。无修改时开销很小。调用方应在调用 Step 之前调用此方法。
// 可安全重复调用。
func (t *Trie) EnsureBuilt() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.dirty {
		t.buildLocked()
	}
}

// Build 是 EnsureBuilt 的别名，用于在初始化代码中更清晰地表达
// 预先构建自动机的意图。
func (t *Trie) Build() { t.EnsureBuilt() }

func (t *Trie) buildLocked() {
	queue := make([]*node, 0, 64)
	for _, child := range t.root.children {
		child.fail = t.root
		child.dictLink = nil
		queue = append(queue, child)
	}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for r, v := range u.children {
			f := u.fail
			for f != nil {
				if _, ok := f.children[r]; ok {
					break
				}
				f = f.fail
			}
			if f == nil {
				v.fail = t.root
			} else if next, ok := f.children[r]; ok && next != v {
				v.fail = next
			} else {
				v.fail = t.root
			}
			if v.fail.isTerminal() {
				v.dictLink = v.fail
			} else {
				v.dictLink = v.fail.dictLink
			}
			queue = append(queue, v)
		}
	}
	t.dirty = false
}

// Step 将游标推进一个字符，并报告所有在此位置结束的模式
// （当前节点加上所有 dictLink 链接的终止祖先节点）。
// 返回的匹配中 RuneLen 为模式长度；调用方需自行计算绝对字符位置。
//
// Step 不会获取 trie 锁。调用方必须保证在一系列 Step 调用期间
// 没有并发的 CRUD 操作。参见包文档。
func (t *Trie) Step(c Cursor, r rune) (Cursor, []Match) {
	cur := c.n
	if cur == nil {
		cur = t.root
	}
	for cur != t.root {
		if _, ok := cur.children[r]; ok {
			break
		}
		cur = cur.fail
		if cur == nil {
			cur = t.root
			break
		}
	}
	if next, ok := cur.children[r]; ok {
		cur = next
	} else {
		cur = t.root
	}

	var matches []Match
	if cur.isTerminal() {
		matches = append(matches, Match{WordID: cur.wordID, RuneLen: cur.runeLen})
	}
	for d := cur.dictLink; d != nil; d = d.dictLink {
		matches = append(matches, Match{WordID: d.wordID, RuneLen: d.runeLen})
	}
	return Cursor{n: cur}, matches
}
