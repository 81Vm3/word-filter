export namespace gui {
	
	export class ACMatchView {
	    word: string;
	    word_id: number;
	    rune_start: number;
	    rune_end: number;
	
	    static createFrom(source: any = {}) {
	        return new ACMatchView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.word = source["word"];
	        this.word_id = source["word_id"];
	        this.rune_start = source["rune_start"];
	        this.rune_end = source["rune_end"];
	    }
	}
	export class ACTestRequest {
	    text: string;
	    scale: number;
	
	    static createFrom(source: any = {}) {
	        return new ACTestRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.scale = source["scale"];
	    }
	}
	export class ACTestResult {
	    input_runes: number;
	    steps: number;
	    total_matches: number;
	    matches: ACMatchView[];
	    truncated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ACTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.input_runes = source["input_runes"];
	        this.steps = source["steps"];
	        this.total_matches = source["total_matches"];
	        this.matches = this.convertValues(source["matches"], ACMatchView);
	        this.truncated = source["truncated"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ACTestResponse {
	    elapsed_ns: number;
	    elapsed_ms: number;
	    result: ACTestResult;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ACTestResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.elapsed_ns = source["elapsed_ns"];
	        this.elapsed_ms = source["elapsed_ms"];
	        this.result = this.convertValues(source["result"], ACTestResult);
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class FilterHit {
	    word: string;
	    severity: number;
	    extra?: Record<string, string>;
	    rune_start: number;
	    rune_end: number;
	    byte_start: number;
	    byte_end: number;
	
	    static createFrom(source: any = {}) {
	        return new FilterHit(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.word = source["word"];
	        this.severity = source["severity"];
	        this.extra = source["extra"];
	        this.rune_start = source["rune_start"];
	        this.rune_end = source["rune_end"];
	        this.byte_start = source["byte_start"];
	        this.byte_end = source["byte_end"];
	    }
	}
	export class FilterTestRequest {
	    input_text: string;
	    input_file_name: string;
	    input_file_content: string;
	    ignore_case: boolean;
	    ignore_width: boolean;
	    replace_rune: string;
	    scale: number;
	
	    static createFrom(source: any = {}) {
	        return new FilterTestRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.input_text = source["input_text"];
	        this.input_file_name = source["input_file_name"];
	        this.input_file_content = source["input_file_content"];
	        this.ignore_case = source["ignore_case"];
	        this.ignore_width = source["ignore_width"];
	        this.replace_rune = source["replace_rune"];
	        this.scale = source["scale"];
	    }
	}
	export class FilterTestResult {
	    input_runes: number;
	    input_bytes: number;
	    sanitized: string;
	    hit_count: number;
	    hits: FilterHit[];
	    truncated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FilterTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.input_runes = source["input_runes"];
	        this.input_bytes = source["input_bytes"];
	        this.sanitized = source["sanitized"];
	        this.hit_count = source["hit_count"];
	        this.hits = this.convertValues(source["hits"], FilterHit);
	        this.truncated = source["truncated"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FilterTestResponse {
	    elapsed_ns: number;
	    elapsed_ms: number;
	    result: FilterTestResult;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new FilterTestResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.elapsed_ns = source["elapsed_ns"];
	        this.elapsed_ms = source["elapsed_ms"];
	        this.result = this.convertValues(source["result"], FilterTestResult);
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class LexiconTestRequest {
	    action: string;
	    word: string;
	    severity: number;
	    extra_json: string;
	    scale: number;
	    input_file_name: string;
	    input_file_content: string;
	    page: number;
	    page_size: number;
	
	    static createFrom(source: any = {}) {
	        return new LexiconTestRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action = source["action"];
	        this.word = source["word"];
	        this.severity = source["severity"];
	        this.extra_json = source["extra_json"];
	        this.scale = source["scale"];
	        this.input_file_name = source["input_file_name"];
	        this.input_file_content = source["input_file_content"];
	        this.page = source["page"];
	        this.page_size = source["page_size"];
	    }
	}
	export class WordEntryView {
	    word: string;
	    severity: number;
	    extra?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new WordEntryView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.word = source["word"];
	        this.severity = source["severity"];
	        this.extra = source["extra"];
	    }
	}
	export class LexiconTestResult {
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
	
	    static createFrom(source: any = {}) {
	        return new LexiconTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action = source["action"];
	        this.added = source["added"];
	        this.updated = source["updated"];
	        this.deleted = source["deleted"];
	        this.duplicated = source["duplicated"];
	        this.imported = source["imported"];
	        this.page = source["page"];
	        this.page_size = source["page_size"];
	        this.total_entries = source["total_entries"];
	        this.saved_path = source["saved_path"];
	        this.entries = this.convertValues(source["entries"], WordEntryView);
	        this.message = source["message"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LexiconTestResponse {
	    elapsed_ns: number;
	    elapsed_ms: number;
	    result: LexiconTestResult;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new LexiconTestResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.elapsed_ns = source["elapsed_ns"];
	        this.elapsed_ms = source["elapsed_ms"];
	        this.result = this.convertValues(source["result"], LexiconTestResult);
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class NormalizePair {
	    index: number;
	    original: string;
	    normalized: string;
	    changed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NormalizePair(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.original = source["original"];
	        this.normalized = source["normalized"];
	        this.changed = source["changed"];
	    }
	}
	export class NormalizeTestRequest {
	    text: string;
	    ignore_case: boolean;
	    ignore_width: boolean;
	    scale: number;
	
	    static createFrom(source: any = {}) {
	        return new NormalizeTestRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.ignore_case = source["ignore_case"];
	        this.ignore_width = source["ignore_width"];
	        this.scale = source["scale"];
	    }
	}
	export class NormalizeTestResult {
	    input_runes: number;
	    output_runes: number;
	    changed_runes: number;
	    normalized_text: string;
	    pairs: NormalizePair[];
	    truncated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NormalizeTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.input_runes = source["input_runes"];
	        this.output_runes = source["output_runes"];
	        this.changed_runes = source["changed_runes"];
	        this.normalized_text = source["normalized_text"];
	        this.pairs = this.convertValues(source["pairs"], NormalizePair);
	        this.truncated = source["truncated"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class NormalizeTestResponse {
	    elapsed_ns: number;
	    elapsed_ms: number;
	    result: NormalizeTestResult;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new NormalizeTestResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.elapsed_ns = source["elapsed_ns"];
	        this.elapsed_ms = source["elapsed_ms"];
	        this.result = this.convertValues(source["result"], NormalizeTestResult);
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class TrieTestRequest {
	    action: string;
	    word: string;
	    severity: number;
	    extra_json: string;
	    scale: number;
	
	    static createFrom(source: any = {}) {
	        return new TrieTestRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action = source["action"];
	        this.word = source["word"];
	        this.severity = source["severity"];
	        this.extra_json = source["extra_json"];
	        this.scale = source["scale"];
	    }
	}
	export class TrieTestResult {
	    action: string;
	    requested: number;
	    added: number;
	    updated: number;
	    deleted: number;
	    duplicated: number;
	    exists: boolean;
	    metadata: trie.Metadata;
	    walked: number;
	    total_entries: number;
	    sample: WordEntryView[];
	
	    static createFrom(source: any = {}) {
	        return new TrieTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action = source["action"];
	        this.requested = source["requested"];
	        this.added = source["added"];
	        this.updated = source["updated"];
	        this.deleted = source["deleted"];
	        this.duplicated = source["duplicated"];
	        this.exists = source["exists"];
	        this.metadata = this.convertValues(source["metadata"], trie.Metadata);
	        this.walked = source["walked"];
	        this.total_entries = source["total_entries"];
	        this.sample = this.convertValues(source["sample"], WordEntryView);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TrieTestResponse {
	    elapsed_ns: number;
	    elapsed_ms: number;
	    result: TrieTestResult;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TrieTestResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.elapsed_ns = source["elapsed_ns"];
	        this.elapsed_ms = source["elapsed_ms"];
	        this.result = this.convertValues(source["result"], TrieTestResult);
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace trie {
	
	export class Metadata {
	    Severity: number;
	    Extra: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Metadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Severity = source["Severity"];
	        this.Extra = source["Extra"];
	    }
	}

}

