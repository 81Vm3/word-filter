# AGENTS.md

Instructions for AI coding agents working in this repository.

## Project

Go library implementing a high-performance sensitive-word filter using a Trie and an Aho-Corasick automaton. Supports CRUD on the word library, TXT/CSV bulk import, one-shot and streaming matching, case/full-width normalization, and per-word metadata (severity + extensible extras).

Module path: `sensitive-filter` (`go.mod` at repo root).

## Layout

```
pkg/
  trie/       Trie + AC automaton + metadata + normalization (the data layer)
  store/      words.txt read/write + TXT/CSV importers (depends on trie)
  sanitizer/  Filter engine + StreamSession (depends on trie)
test/
  trie/       black-box tests for pkg/trie       (package trie_test)
  store/      black-box tests for pkg/store      (package store_test)
  sanitizer/  black-box tests for pkg/sanitizer  (package sanitizer_test)
cmd/demo/     End-to-end demo binary
words.txt     Sample sensitive-word library (count + words + severities)
```

Source files live under `pkg/` only. Tests live under `test/` only. Do not mix.

## Build, test, run

```bash
go build ./...                                   # compile everything
go vet ./...                                     # static checks
go test ./test/... -race -count=1                # full test suite (must stay green)
go test ./test/... -coverpkg=./pkg/...           # coverage (test/ targets pkg/)
go run ./cmd/demo                                # end-to-end demo (reads words.txt)
```

After editing any `.go` file, format it:

```bash
gofmt -w <files>
```

This is a hard requirement — do not commit or finish a task with unformatted Go.

## Architecture notes

### `pkg/trie` is the only package with internal mutability

- `trie.Trie` owns the rune-keyed map structure (`map[rune]*node`), per-word `Metadata`, the word ↔ wordID maps, and a `sync.RWMutex` guarding CRUD.
- CRUD methods (`Add`/`Update`/`Delete`/`Get`/`Has`/`Walk`/`Size`/`WordByID`/`MetaByID`) acquire the appropriate lock.
- The match-side surface (`Root`/`Step`/`Normalize`) is intentionally **lock-free** — calling it concurrently with CRUD is a data race. The library's documented contract is: callers invoke `EnsureBuilt()` once, then drive `Step` repeatedly without mutating the trie. `pkg/sanitizer` follows this pattern.
- `EnsureBuilt()` rebuilds fail/dictLink pointers if any CRUD has occurred since the last build. It is idempotent and cheap when clean. Call it from any new entry point that performs matching.

### Cross-package match API

Sanitizer/store never reach into trie internals. They use:

- `trie.Cursor` — opaque position pointer
- `trie.Match{ WordID, RuneLen }` — emitted at each Step that closes a pattern
- `trie.Trie.Root() Cursor`
- `trie.Trie.Step(cur Cursor, r rune) (Cursor, []Match)`
- `trie.Trie.Normalize(r rune) rune`
- `trie.Trie.MetaByID(id) / WordByID(id)`

If you find yourself wanting access to `node`, fail/dictLink fields, or the children map from outside `pkg/trie`, extend the public API instead.

### Streaming session

`pkg/sanitizer/stream.go` holds the StreamSession state: cursor + absolute rune/byte offsets + a `pendingBuf` for UTF-8 byte sequences split across chunks + accumulated original bytes for sanitized-text reconstruction. `InitStream` calls `EnsureBuilt` once and caches `Root()`. `AppendToStream` drives `Step` per decoded rune. `GetStreamResult` does not reset; `ClearStream` does.

### Overlap handling

The AC pass reports every match (including overlapping substrings). Replacement merges hit intervals into minimum-cover ranges before substituting `ReplaceRune` (default `'*'`) per rune. Don't deduplicate matches before reporting — DebugResult should expose every hit so callers can inspect them.

### words.txt format

```
<n>
<word_1>
<word_2>
...
<word_n>
<severity_1>
<severity_2>
...
<severity_n>
```

Line 1 is the integer count `n`. Then `n` word lines, then `n` severity lines (integers). Missing trailing severity lines fall back to `SeverityNormal` (0).

## Conventions

- Tests are **black-box only**. Each `test/<pkg>/` directory uses `package <pkg>_test` and imports `sensitive-filter/pkg/<pkg>`. This forces matchability through public API and keeps `pkg/` internals out of test concerns. Do not add `*_test.go` files inside `pkg/`.
- When adding a new package method that needs test coverage, the test goes in the matching `test/<pkg>/` directory. If your test needs access to an unexported field, that's a design signal — expose what's needed via a new public accessor or restructure the test as behavioral.
- Names: use `New` (not `NewFoo`) when the package name already conveys the type — e.g. `sanitizer.New(t, cfg)`, `trie.New(norm)`. Avoid stutter.
- `Metadata` carries `Severity int` plus an `Extra map[string]string` for forward-compat fields. Extending the struct is preferred over parallel maps.
- Match positions are returned in both **rune** (`RuneStart`/`RuneEnd`) and **byte** (`ByteStart`/`ByteEnd`) coordinates. Rune positions are stable under width/case normalization (all foldings are 1:1 rune mappings); byte positions index the original input string.

## Common pitfalls

- **Forgetting `EnsureBuilt`** in a new matching entry point → matches use stale (or missing) fail/dictLink pointers. Always call it before the first `Step` of a session.
- **Treating Trie as immutable** during a stream → if a CRUD happens between `AppendToStream` calls, the cursor references a node that may have been deleted or pointed at a now-invalid fail chain. Document and respect the contract: no CRUD during a match.
- **Adding tests inside `pkg/`** to access internals → instead, expose a minimal accessor or restructure the test to verify behavior through the public surface.
- **Skipping `gofmt`** → CI/reviewers will reject. Run `gofmt -w` on every changed file.
- **Off-by-one on `words.txt`** → the spec text reads "lines 2..2+n" which over-counts by 1; the implementation treats it as `n` word lines followed by `n` severity lines. Don't "fix" the loop to read n+1 lines.

## When extending the system

- New match mode: add a `MatchMode` const in `pkg/sanitizer/sanitizer.go` and extend `Sanitizer.filterMode` rather than branching inside `scan`. Add tests covering the new mode's hit-filtering behavior in `test/sanitizer/`.
- New importer format (e.g. JSON): add `pkg/store/<format>_import.go` exposing `Import<Format>(t *trie.Trie, path string) (added, dup int, err error)`. Reuse `WordEntry` and the trie's dedupe-on-Add behavior.
- New metadata field: add it to `Metadata` in `pkg/trie/metadata.go`, update `Metadata.Clone()`, then propagate through the CSV importer (`pkg/store/csv_import.go`) and `WriteWordsTxt` if it needs to persist.
- New persistence format: add to `pkg/store/`, keep `WriteWordsTxt`/`ReadWordsTxt` as the canonical TXT shape so the importer roundtrips.
