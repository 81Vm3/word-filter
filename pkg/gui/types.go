package gui

import "sensitive-filter/pkg/trie"

const (
	DefaultWordsPath = "words.txt"
	maxPreviewRows   = 200
)

type Timing struct {
	ElapsedNS int64   `json:"elapsed_ns"`
	ElapsedMS float64 `json:"elapsed_ms"`
}

type FilterTestRequest struct {
	InputText        string `json:"input_text"`
	InputFileName    string `json:"input_file_name"`
	InputFileContent string `json:"input_file_content"`
	IgnoreCase       bool   `json:"ignore_case"`
	IgnoreWidth      bool   `json:"ignore_width"`
	ReplaceRune      string `json:"replace_rune"`
	Scale            int    `json:"scale"`
}

type FilterHit struct {
	Word      string            `json:"word"`
	Severity  int               `json:"severity"`
	Extra     map[string]string `json:"extra,omitempty"`
	RuneStart int               `json:"rune_start"`
	RuneEnd   int               `json:"rune_end"`
	ByteStart int               `json:"byte_start"`
	ByteEnd   int               `json:"byte_end"`
}

type FilterTestResult struct {
	InputRunes int         `json:"input_runes"`
	InputBytes int         `json:"input_bytes"`
	Sanitized  string      `json:"sanitized"`
	HitCount   int         `json:"hit_count"`
	Hits       []FilterHit `json:"hits"`
	Truncated  bool        `json:"truncated"`
}

type FilterTestResponse struct {
	Timing
	Result FilterTestResult `json:"result"`
	Error  string           `json:"error,omitempty"`
}

type TrieTestRequest struct {
	Action    string `json:"action"`
	Word      string `json:"word"`
	Severity  int    `json:"severity"`
	ExtraJSON string `json:"extra_json"`
	Scale     int    `json:"scale"`
}

type WordEntryView struct {
	Word     string            `json:"word"`
	Severity int               `json:"severity"`
	Extra    map[string]string `json:"extra,omitempty"`
}

type TrieTestResult struct {
	Action       string          `json:"action"`
	Requested    int             `json:"requested"`
	Added        int             `json:"added"`
	Updated      int             `json:"updated"`
	Deleted      int             `json:"deleted"`
	Duplicated   int             `json:"duplicated"`
	Exists       bool            `json:"exists"`
	Metadata     trie.Metadata   `json:"metadata"`
	Walked       int             `json:"walked"`
	TotalEntries int             `json:"total_entries"`
	Sample       []WordEntryView `json:"sample"`
}

type TrieTestResponse struct {
	Timing
	Result TrieTestResult `json:"result"`
	Error  string         `json:"error,omitempty"`
}

type ACTestRequest struct {
	Text  string `json:"text"`
	Scale int    `json:"scale"`
}

type ACMatchView struct {
	Word      string `json:"word"`
	WordID    int    `json:"word_id"`
	RuneStart int    `json:"rune_start"`
	RuneEnd   int    `json:"rune_end"`
}

type ACTestResult struct {
	InputRunes   int           `json:"input_runes"`
	Steps        int           `json:"steps"`
	TotalMatches int           `json:"total_matches"`
	Matches      []ACMatchView `json:"matches"`
	Truncated    bool          `json:"truncated"`
}

type ACTestResponse struct {
	Timing
	Result ACTestResult `json:"result"`
	Error  string       `json:"error,omitempty"`
}

type NormalizeTestRequest struct {
	Text        string `json:"text"`
	IgnoreCase  bool   `json:"ignore_case"`
	IgnoreWidth bool   `json:"ignore_width"`
	Scale       int    `json:"scale"`
}

type NormalizePair struct {
	Index      int    `json:"index"`
	Original   string `json:"original"`
	Normalized string `json:"normalized"`
	Changed    bool   `json:"changed"`
}

type NormalizeTestResult struct {
	InputRunes     int             `json:"input_runes"`
	OutputRunes    int             `json:"output_runes"`
	ChangedRunes   int             `json:"changed_runes"`
	NormalizedText string          `json:"normalized_text"`
	Pairs          []NormalizePair `json:"pairs"`
	Truncated      bool            `json:"truncated"`
}

type NormalizeTestResponse struct {
	Timing
	Result NormalizeTestResult `json:"result"`
	Error  string              `json:"error,omitempty"`
}

type LexiconTestRequest struct {
	Action           string `json:"action"`
	Word             string `json:"word"`
	Severity         int    `json:"severity"`
	ExtraJSON        string `json:"extra_json"`
	Scale            int    `json:"scale"`
	InputFileName    string `json:"input_file_name"`
	InputFileContent string `json:"input_file_content"`
	Page             int    `json:"page"`
	PageSize         int    `json:"page_size"`
}

type LexiconTestResult struct {
	Action       string          `json:"action"`
	Added        int             `json:"added"`
	Updated      int             `json:"updated"`
	Deleted      int             `json:"deleted"`
	Duplicated   int             `json:"duplicated"`
	Imported     int             `json:"imported"`
	Page         int             `json:"page"`
	PageSize     int             `json:"page_size"`
	TotalEntries int             `json:"total_entries"`
	SavedPath    string          `json:"saved_path"`
	Entries      []WordEntryView `json:"entries"`
	Message      string          `json:"message"`
}

type LexiconTestResponse struct {
	Timing
	Result LexiconTestResult `json:"result"`
	Error  string            `json:"error,omitempty"`
}
