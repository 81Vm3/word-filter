export interface Timing {
  elapsed_ns: number;
  elapsed_ms: number;
}

export interface WordEntryView {
  word: string;
  severity: number;
  extra?: Record<string, string>;
}

export interface FilterHit {
  word: string;
  severity: number;
  extra?: Record<string, string>;
  rune_start: number;
  rune_end: number;
  byte_start: number;
  byte_end: number;
}

export interface FilterTestResponse extends Timing {
  error?: string;
  result: {
    input_runes: number;
    input_bytes: number;
    sanitized: string;
    hit_count: number;
    hits: FilterHit[];
    truncated: boolean;
  };
}

export interface TrieTestResponse extends Timing {
  error?: string;
  result: {
    action: string;
    requested: number;
    added: number;
    updated: number;
    deleted: number;
    duplicated: number;
    exists: boolean;
    metadata: { severity: number; extra?: Record<string, string> };
    walked: number;
    total_entries: number;
    sample: WordEntryView[];
  };
}

export interface ACTestResponse extends Timing {
  error?: string;
  result: {
    input_runes: number;
    steps: number;
    total_matches: number;
    matches: Array<{
      word: string;
      word_id: number;
      rune_start: number;
      rune_end: number;
    }>;
    truncated: boolean;
  };
}

export interface NormalizeTestResponse extends Timing {
  error?: string;
  result: {
    input_runes: number;
    output_runes: number;
    changed_runes: number;
    normalized_text: string;
    pairs: Array<{
      index: number;
      original: string;
      normalized: string;
      changed: boolean;
    }>;
    truncated: boolean;
  };
}

export interface LexiconTestResponse extends Timing {
  error?: string;
  result: {
    action: string;
    added: number;
    updated: number;
    deleted: number;
    duplicated: number;
    imported: number;
    page: number;
    page_size: number;
    total_entries: number;
    saved_path: string;
    entries: WordEntryView[];
    message: string;
  };
}
