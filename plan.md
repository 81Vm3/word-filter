     海量敏感词过滤系统（Go + Trie + AC 自动机）

     Context

     /home/p3g4s/system/ 仅有需求文档 1.md，是一个绿地项目。需要从零构建一套基于
     Trie + AC 自动机的高性能敏感词过滤引擎：
     - 自定义磁盘格式（words.txt）持久化 + 内存 Trie
     - 词库 CRUD + 元数据（严重系数，可扩展）
     - TXT/CSV 批量导入（自动去重）
     - 过滤引擎：全词/前缀两种多模式匹配、大小写/全半角归一化、脱敏替换、debug 信息
     - 流式接口：把 AC 单步匹配拆出来跨 chunk 调用

     已对齐的设计选择（来自前序问答）：
     1. Sanitize 返回脱敏后文本（敏感词每 rune 替换为 *）
     2. 匹配语义：全词匹配 = AC 标准多模式找出所有完整子串命中；前缀匹配 =
     文本任意位置以某敏感词为前缀的命中
     3. Metadata 是可扩展 struct（Severity int + Extra map[string]string）
     4. 单模块 sensitive-filter，源码扁平放根目录

     ---
     设计要点

     Trie 节点（trie.go）

     - children map[rune]*Node：中文键空间大、节点稀疏，map O(1)
     平均；数组排序在中文场景增删差
     - 终结节点存 wordID int + runeLen int；非终结 wordID = -1
     - Metadata 集中存在 Trie.metas map[int]Metadata（按 wordID 索引），便于
     Update/Delete 不动 Trie 结构
     - 去重表 words map[string]int：归一化后的词 → wordID
     - dirty bool：CRUD 后置位，下次匹配前重建 AC

     AC 自动机（ac_automaton.go）

     - BFS 构建 fail 指针 + dictLink（沿 fail
     链最近的终结祖先），匹配时报告所有命中只需一次 dictLink 跳跃
     - 惰性构建：CRUD 只置 dirty=true，第一次匹配（或显式 Build()）才一次性
     BFS。理由：增量 AC 复杂易错；词库通常批量导入后才开始用
     - 匹配核心是 step(cur, r) (next *Node, hits []Hit)，复用于一次性 + 流式

     归一化（normalize.go）

     - IgnoreCase：A-Z → a-z
     - IgnoreWidth：U+FF01..FF5E → U+0021..007E、U+3000 → U+0020
     - 1:1 rune 映射，rune 位置守恒；byte 位置需独立维护
     - 构建 Trie 时归一化词；匹配时归一化输入

     流式状态（StreamSession）

     - 持有 cur *Node、runeOff/byteOff int、pendingBuf []byte（处理跨 chunk
     的不完整 UTF-8 序列）、hits []Hit、chunks []string（用于重构脱敏文本）
     - InitStream 重置；ClearStream 等价 InitStream；GetStreamResult
     返回累计结果但不重置（可继续喂数据）

     重叠命中与替换

     - AC 报告全部命中（含子串与重叠）
     - 替换时按 [runeStart, runeEnd) 排序后合并为最小覆盖区间，每个 rune 置为
     ReplaceRune（默认 '*'，可配）

     位置语义

     - Hit 同时给 RuneStart/RuneEnd 和 ByteStart/ByteEnd（rune 用于流式累计、byte
     用于切原文）
     - 流式时是从 InitStream 起的累计偏移

     并发

     - Trie 内置 sync.RWMutex：CRUD 写锁，Build/匹配读锁
     - Sanitizer 无状态，不加锁
     - StreamSession per-instance，非并发安全（注释明示）

     words.txt 格式解释

     原文："第1行为词总数n，第2~2+n行为敏感词，第2+n+1~2+2n行为敏感词的元数据"——按
     数学严格读 (2~2+n) 是 n+1 行，存在轻微 off-by-one。采用最自然解释：line1 =
     count n；接下来 n 行为词；再接下来 n 行为对应严重系数。实现时按"逐行读取，先读
      n，再读 n 个词，再读 n 个数字"，对 off-by-one 表述天然容忍。

     ---
     关键类型签名（已与 Plan 代理推敲过）

     // metadata.go
     type Metadata struct {
         Severity int
         Extra    map[string]string
     }

     // trie.go
     type Node struct {
         children map[rune]*Node
         wordID   int   // -1 非终结
         runeLen  int
         fail     *Node
         dictLink *Node
     }
     type Trie struct {
         root  *Node
         metas map[int]Metadata
         words map[string]int  // 归一化词 → wordID
         next  int
         dirty bool
         norm  NormFunc        // 构建时使用的归一化函数（与 sanitizer 共用）
         mu    sync.RWMutex
     }
     func New(norm NormFunc) *Trie
     func (t *Trie) Add(word string, m Metadata) (id int, added bool)
     func (t *Trie) Update(word string, m Metadata) bool
     func (t *Trie) Delete(word string) bool
     func (t *Trie) Get(word string) (Metadata, bool)
     func (t *Trie) Walk(fn func(word string, m Metadata)) // 持久化辅助
     func (t *Trie) Build()

     // word_reader.go
     type WordEntry struct { Word string; Meta Metadata }
     func ReadWordsTxt(path string) ([]WordEntry, error)
     func WriteWordsTxt(path string, entries []WordEntry) error

     // word_importer_txt.go / word_importer_csv.go
     func ImportTxt(t *Trie, path string) (added, dup int, err error)
     func ImportCSV(t *Trie, path string) (added, dup int, err error) // 列:
     word,severity[,extra_json]

     // sanitizer.go
     type MatchMode int
     const (ModeFullWord MatchMode = iota; ModePrefix)
     type Config struct {
         IgnoreCase  bool
         IgnoreWidth bool
         Mode        MatchMode
         ReplaceRune rune // 默认 '*'
     }
     type Hit struct {
         Word     string
         WordID   int
         Meta     Metadata
         RuneStart, RuneEnd int
         ByteStart, ByteEnd int
     }
     type DebugResult struct {
         Sanitized string
         Hits      []Hit
         Count     int
         Elapsed   time.Duration
     }
     type Sanitizer struct {
         trie *Trie
         cfg  Config
     }
     func NewSanitizer(t *Trie, cfg Config) *Sanitizer
     func (s *Sanitizer) Sanitize(text string) string
     func (s *Sanitizer) SanitizeWithDebug(text string) DebugResult

     // 流式
     type StreamSession struct { /* 内部字段见上文 */ }
     func (s *Sanitizer) InitStream() *StreamSession
     func (ss *StreamSession) AppendToStream(chunk string)
     func (ss *StreamSession) GetStreamResult() DebugResult
     func (ss *StreamSession) ClearStream()

     ---
     文件清单

     /home/p3g4s/system/
     ├── go.mod                                     module sensitive-filter
     ├── metadata.go
     ├── normalize.go            normalize_test.go
     ├── trie.go                 trie_test.go
     ├── ac_automaton.go         ac_automaton_test.go
     ├── word_reader.go          word_reader_test.go
     ├── word_importer_txt.go    word_importer_txt_test.go
     ├── word_importer_csv.go    word_importer_csv_test.go
     ├── sanitizer.go            sanitizer_test.go
     ├── main.go                                    demo
     └── words.txt                                  示例数据

     ---
     实现顺序

     1. go.mod (go mod init sensitive-filter) → metadata.go → normalize.go
     2. trie.go（先实现 CRUD + Walk，无 AC）
     3. ac_automaton.go（Build() BFS + step()）
     4. word_reader.go → 两个 importer（importer 调 Trie.Add 自动去重）
     5. sanitizer.go：一次性 Sanitize/SanitizeWithDebug → 区间合并替换 → 流式 4
     方法
     6. main.go 端到端 demo + 全部 *_test.go
     7. 生成虚构 words.txt，跑 go vet ./... && go test ./... -race，gofmt -w .

     ---
     关键算法伪代码

     AC BFS 构建：
     queue := [root.children]
     for u := queue.pop():
         for r, v := range u.children:
             f := u.fail
             for f != nil && f.children[r] == nil: f = f.fail
             v.fail = f.children[r] if f else root
             v.dictLink = v.fail if v.fail.wordID >= 0 else v.fail.dictLink
             queue.push(v)

     单 rune 推进（一次性匹配与流式共用）：
     func step(cur *Node, r rune, runeIdx int) (next *Node, hits []*Node):
         for cur != root && cur.children[r] == nil: cur = cur.fail
         if c, ok := cur.children[r]; ok: cur = c
         n := cur if cur.wordID >= 0 else cur.dictLink
         for n != nil: append hits, n = n.dictLink
         return cur, hits

     前缀模式过滤：复用 AC 标准命中，仅保留每个 RuneStart
     的全部命中（自然就是"从该位置起始的所有前缀"）。

     区间合并替换：sort by RuneStart → 扫描合并相邻/重叠 → rebuild rune slice
     内区间全置 ReplaceRune。

     流式跨 chunk UTF-8：AppendToStream 先 pendingBuf+chunk 解码
     rune，末尾若是不完整序列回填 pendingBuf；剩余 rune 逐个 step 并累加偏移。非法
     UTF-8 推进 1 字节避免死循环。

     ---
     测试矩阵

     ┌─────────────────────────┬───────────────────────────────────────────────────
     ┐
     │        测试文件         │                      覆盖点
     │
     ├─────────────────────────┼───────────────────────────────────────────────────
     ┤
     │ trie_test.go            │ Add/Get/Update/Delete/重复 Add、Walk 完整性
     │
     ├─────────────────────────┼───────────────────────────────────────────────────
     ┤
     │ ac_automaton_test.go    │ BFS 正确性、单词 + 重叠子串命中、dictLink
     │
     │                         │ 链报告所有后缀
     │
     ├─────────────────────────┼───────────────────────────────────────────────────
     ┤
     │ normalize_test.go       │ 大小写、全半角、ASCII / 中文 / 数字混合
     │
     ├─────────────────────────┼───────────────────────────────────────────────────
     ┤
     │ word_reader_test.go     │ 读、写、空文件、计数不一致容错
     │
     ├─────────────────────────┼───────────────────────────────────────────────────
     ┤
     │ word_importer_*_test.go │ 去重、错误行容错、CSV 表头、Extra JSON
     │
     ├─────────────────────────┼───────────────────────────────────────────────────
     ┤
     │                         │ 一次性脱敏、debug
     │
     │ sanitizer_test.go       │ 元信息、全词/前缀模式差异、归一化、流式跨
     │
     │                         │ chunk（"bad" + "word"）、Init/Clear/GetResult
     │
     │                         │ 顺序、多 chunk 累积偏移
     │
     └─────────────────────────┴───────────────────────────────────────────────────
     ┘

     虚构数据集：傻逼(sev 2) / 测试敏感(sev 1) / badword(sev 1) / BadCase(sev 0) /
     ＡＢＣ(全角，测试归一化)。

     ---
     Verification

     1. cd /home/p3g4s/system && go mod init sensitive-filter && go build ./...
     2. go vet ./... — 静态检查无 warning
     3. go test ./... -race -v — 全绿，覆盖以上测试矩阵
     4. go run . — main.go 演示输出符合预期（脱敏文本、debug 信息、流式跨 chunk
     命中）
     5. gofmt -l . 输出为空（无未格式化文件）
     6. 任务完成后通过 mcp-notifier 发通知

     ---
     Critical Files

     - /home/p3g4s/system/trie.go — 字典树 + CRUD + 元数据存储
     - /home/p3g4s/system/ac_automaton.go — AC 自动机 BFS 构建 + step
     - /home/p3g4s/system/sanitizer.go — 过滤引擎主入口 + 流式 session
     - /home/p3g4s/system/word_reader.go — words.txt 编解码
     - /home/p3g4s/system/normalize.go — 大小写/全半角归一化（构建与匹配共用）
     - /home/p3g4s/system/main.go — 端到端 demo
