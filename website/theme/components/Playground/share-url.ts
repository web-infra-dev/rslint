import { deflateSync, inflateSync } from 'fflate';
import { SHARE_DICT_V1 } from './share-dict';

export const DEFAULT_CODE = ['let a: any;', 'a.b = 10;'].join('\n');

export const DEFAULT_RSLINT_CONFIG = `import { defineConfig, js, ts } from '@rslint/core';

export default defineConfig([
  js.configs.recommended,
  ts.configs.recommendedTypeChecked,
  {
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.json'],
      },
    },
    rules: {
      '@typescript-eslint/no-unsafe-member-access': 'error',
    },
  },
]);`;

export const DEFAULT_TSCONFIG = `{
  "compilerOptions": {
    "target": "ESNext",
    "module": "ESNext",
    "strict": true,
    "strictNullChecks": true
  }
}`;

export interface ShareState {
  code: string;
  rslintConfig: string;
  tsconfig: string;
  /**
   * The `@rslint/wasm` version the link runs against. Always written once the
   * version list has loaded, so a shared link reproduces what its author saw
   * even after a newer version ships; `undefined` only on a link that predates
   * version pinning, or before the list arrives.
   */
  wasmVersion?: string;
}

/**
 * Bumped together with the frame layout, the compression settings or the
 * dictionary. Links carry it as their first character so an older link can
 * still be decoded the way it was written.
 */
const SHARE_VERSION = '1';

/**
 * The payload owns the whole fragment rather than sitting in a `name=` pair:
 * the playground has no anchors to share it with, and two characters of URL are
 * two characters. A fragment that does not start with the version digit is an
 * old plain-text link instead.
 */
const LEGACY_CODE_PARAM = 'code';

const FLAG_CODE = 1;
const FLAG_RSLINT_CONFIG = 2;
const FLAG_TSCONFIG = 4;
const FLAG_WASM_VERSION = 8;
const ALL_FLAGS =
  FLAG_CODE | FLAG_RSLINT_CONFIG | FLAG_TSCONFIG | FLAG_WASM_VERSION;

const encoder = new TextEncoder();
const decoder = new TextDecoder();

let dictionary: Uint8Array | undefined;
function getDictionary(): Uint8Array {
  dictionary ??= encoder.encode(SHARE_DICT_V1);
  return dictionary;
}

function toBase64Url(bytes: Uint8Array): string {
  let binary = '';
  for (const byte of bytes) binary += String.fromCharCode(byte);
  return btoa(binary)
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
}

function fromBase64Url(value: string): Uint8Array {
  const binary = atob(value.replace(/-/g, '+').replace(/_/g, '/'));
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
  return bytes;
}

/**
 * The frame is a flags byte naming the parts that differ from their default,
 * followed by the pinned wasm version as three LEB128 numbers when it is
 * present, then the text sections in flag order. Every text section but the
 * last is prefixed with its LEB128 byte length; the last one runs to the end.
 */
function pushVarint(bytes: number[], value: number) {
  let rest = value;
  do {
    const byte = rest & 0x7f;
    rest >>>= 7;
    bytes.push(rest > 0 ? byte | 0x80 : byte);
  } while (rest > 0);
}

function readVarint(frame: Uint8Array, cursor: { offset: number }): number {
  let value = 0;
  let shift = 0;
  let byte: number;
  do {
    if (cursor.offset >= frame.length || shift > 28) throw new RangeError();
    byte = frame[cursor.offset++];
    value |= (byte & 0x7f) << shift;
    shift += 7;
  } while (byte & 0x80);
  return value;
}

function parseVersion(version: string): number[] | null {
  const match = /^(\d+)\.(\d+)\.(\d+)$/.exec(version);
  return match === null ? null : [+match[1], +match[2], +match[3]];
}

function encodeFrame(
  flags: number,
  version: number[] | null,
  sections: Uint8Array[],
): Uint8Array {
  const bytes: number[] = [flags];
  if (version !== null) for (const part of version) pushVarint(bytes, part);
  sections.forEach((section, index) => {
    if (index < sections.length - 1) pushVarint(bytes, section.length);
    for (const byte of section) bytes.push(byte);
  });
  return new Uint8Array(bytes);
}

function decodeFrame(frame: Uint8Array): ShareState | null {
  if (frame.length === 0) return null;
  const flags = frame[0];
  if (flags === 0 || flags > ALL_FLAGS) return null;
  const cursor = { offset: 1 };

  const state: ShareState = {
    code: DEFAULT_CODE,
    rslintConfig: DEFAULT_RSLINT_CONFIG,
    tsconfig: DEFAULT_TSCONFIG,
  };
  if (flags & FLAG_WASM_VERSION) {
    const parts = [
      readVarint(frame, cursor),
      readVarint(frame, cursor),
      readVarint(frame, cursor),
    ];
    state.wasmVersion = parts.join('.');
  }

  const keys = (
    [
      [FLAG_CODE, 'code'],
      [FLAG_RSLINT_CONFIG, 'rslintConfig'],
      [FLAG_TSCONFIG, 'tsconfig'],
    ] as const
  ).filter(([flag]) => flags & flag);
  keys.forEach(([, key], index) => {
    if (index === keys.length - 1) {
      state[key] = decoder.decode(frame.subarray(cursor.offset));
      cursor.offset = frame.length;
      return;
    }
    const length = readVarint(frame, cursor);
    if (cursor.offset + length > frame.length) throw new RangeError();
    state[key] = decoder.decode(
      frame.subarray(cursor.offset, cursor.offset + length),
    );
    cursor.offset += length;
  });
  return state;
}

/**
 * Returns the fragment for a link to `state`, or `null` when nothing has been
 * touched and no version is known yet, in which case the link needs no
 * fragment at all.
 */
export function encodeShareState(state: ShareState): string | null {
  let flags = 0;
  const sections: Uint8Array[] = [];
  if (state.code !== DEFAULT_CODE) {
    flags |= FLAG_CODE;
    sections.push(encoder.encode(state.code));
  }
  if (state.rslintConfig !== DEFAULT_RSLINT_CONFIG) {
    flags |= FLAG_RSLINT_CONFIG;
    sections.push(encoder.encode(state.rslintConfig));
  }
  if (state.tsconfig !== DEFAULT_TSCONFIG) {
    flags |= FLAG_TSCONFIG;
    sections.push(encoder.encode(state.tsconfig));
  }
  const version =
    state.wasmVersion === undefined ? null : parseVersion(state.wasmVersion);
  if (version !== null) flags |= FLAG_WASM_VERSION;
  if (flags === 0) return null;

  const compressed = deflateSync(encodeFrame(flags, version, sections), {
    level: 9,
    mem: 12,
    dictionary: getDictionary(),
  });
  return SHARE_VERSION + toBase64Url(compressed);
}

/**
 * Decodes a share fragment. A truncated, corrupt or future-version link decodes
 * to `null` so the caller can fall back to the defaults instead of failing.
 */
export function decodeShareState(payload: string): ShareState | null {
  if (payload[0] !== SHARE_VERSION) return null;
  try {
    return decodeFrame(
      inflateSync(fromBase64Url(payload.slice(1)), {
        dictionary: getDictionary(),
      }),
    );
  } catch {
    return null;
  }
}

/**
 * Reads the state a link carries. Links that predate compressed sharing only
 * ever carried plain-text code, so they resolve to the default configs.
 */
export function readShareState(): ShareState {
  const defaults: ShareState = {
    code: DEFAULT_CODE,
    rslintConfig: DEFAULT_RSLINT_CONFIG,
    tsconfig: DEFAULT_TSCONFIG,
  };
  if (typeof window === 'undefined') return defaults;

  const { search, hash } = window.location;
  const fragment = hash.startsWith('#') ? hash.slice(1) : '';
  if (fragment.startsWith(SHARE_VERSION)) {
    return decodeShareState(fragment) ?? defaults;
  }

  const legacyCode =
    new URLSearchParams(search).get(LEGACY_CODE_PARAM) ??
    new URLSearchParams(fragment).get(LEGACY_CODE_PARAM);
  if (legacyCode !== null) return { ...defaults, code: legacyCode };

  return defaults;
}

/**
 * Rewrites the current URL to carry `state`. Nothing but the fragment is left
 * behind: an unmodified playground gets a bare URL, and a link opened in the
 * legacy plain-text form is upgraded in place on the first edit.
 */
export function writeShareState(state: ShareState) {
  if (typeof window === 'undefined') return;
  try {
    const url = new URL(window.location.href);
    url.searchParams.delete(LEGACY_CODE_PARAM);
    url.hash = encodeShareState(state) ?? '';
    window.history.replaceState(null, '', url.toString());
  } catch {
    // A URL we cannot rewrite is not worth failing an edit over.
  }
}
