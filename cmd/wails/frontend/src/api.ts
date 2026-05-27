import type {
  ACTestResponse,
  FilterTestResponse,
  LexiconTestResponse,
  NormalizeTestResponse,
  TrieTestResponse,
} from './types';

type BackendMethod =
  | 'FilterTest'
  | 'TrieTest'
  | 'ACTest'
  | 'NormalizeTest'
  | 'LexiconTest'
  | 'WordsPath';

type AnyRecord = Record<string, unknown>;

function makeFallbackError<T>(message: string): T {
  return {
    elapsed_ns: 0,
    elapsed_ms: 0,
    error: message,
    result: {},
  } as T;
}

function asNumber(v: unknown): number | null {
  if (typeof v === 'number' && Number.isFinite(v)) {
    return v;
  }
  if (typeof v === 'string') {
    const n = Number(v);
    if (Number.isFinite(n)) {
      return n;
    }
  }
  return null;
}

function asRecord(v: unknown): AnyRecord {
  if (v && typeof v === 'object') {
    return v as AnyRecord;
  }
  return {};
}

function firstNumber(...values: unknown[]): number | null {
  for (const value of values) {
    const n = asNumber(value);
    if (n !== null) {
      return n;
    }
  }
  return null;
}

function normalizeResponse(raw: unknown, elapsedLocalMs: number): unknown {
  const obj = asRecord(raw);
  if (Object.keys(obj).length === 0) {
    return raw;
  }

  const timing = asRecord(obj.timing ?? obj.Timing);
  let elapsedNs = firstNumber(
    obj.elapsed_ns,
    obj.elapsedNs,
    obj.ElapsedNS,
    timing.elapsed_ns,
    timing.elapsedNs,
    timing.ElapsedNS
  );
  let elapsedMs = firstNumber(
    obj.elapsed_ms,
    obj.elapsedMs,
    obj.ElapsedMS,
    timing.elapsed_ms,
    timing.elapsedMs,
    timing.ElapsedMS
  );

  if ((elapsedMs === null || elapsedMs <= 0) && elapsedNs !== null && elapsedNs > 0) {
    elapsedMs = elapsedNs / 1e6;
  }
  if ((elapsedNs === null || elapsedNs <= 0) && elapsedMs !== null && elapsedMs > 0) {
    elapsedNs = Math.round(elapsedMs * 1e6);
  }
  if ((elapsedMs === null || elapsedMs <= 0) && elapsedLocalMs > 0) {
    elapsedMs = elapsedLocalMs;
    elapsedNs = Math.round(elapsedLocalMs * 1e6);
  }
  if (elapsedMs === null) {
    elapsedMs = 0;
  }
  if (elapsedNs === null) {
    elapsedNs = 0;
  }

  const result = obj.result ?? obj.Result ?? {};
  const errorValue = obj.error ?? obj.Error ?? '';
  const error = typeof errorValue === 'string' ? errorValue : String(errorValue ?? '');

  return {
    ...obj,
    elapsed_ns: elapsedNs,
    elapsed_ms: elapsedMs,
    result,
    error,
  };
}

async function invoke<T>(method: BackendMethod, payload?: unknown): Promise<T> {
  const fn = window.go?.main?.App?.[method];
  if (!fn) {
    return makeFallbackError<T>('Wails runtime is not attached. Start via `wails dev` or built binary.');
  }
  const started = performance.now();
  const raw = await fn(payload);
  if (method === 'WordsPath') {
    return raw as T;
  }
  const elapsedLocalMs = Math.max(0, performance.now() - started);
  return normalizeResponse(raw, elapsedLocalMs) as T;
}

export function filterTest(payload: unknown): Promise<FilterTestResponse> {
  return invoke<FilterTestResponse>('FilterTest', payload);
}

export function trieTest(payload: unknown): Promise<TrieTestResponse> {
  return invoke<TrieTestResponse>('TrieTest', payload);
}

export function acTest(payload: unknown): Promise<ACTestResponse> {
  return invoke<ACTestResponse>('ACTest', payload);
}

export function normalizeTest(payload: unknown): Promise<NormalizeTestResponse> {
  return invoke<NormalizeTestResponse>('NormalizeTest', payload);
}

export function lexiconTest(payload: unknown): Promise<LexiconTestResponse> {
  return invoke<LexiconTestResponse>('LexiconTest', payload);
}

export function wordsPath(): Promise<string> {
  return invoke<string>('WordsPath');
}
