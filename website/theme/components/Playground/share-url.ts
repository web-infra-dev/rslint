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
}

/**
 * Bumped together with the frame layout, the compression settings or the
 * dictionary. Links carry it as their first character so an older link can
 * still be decoded the way it was written.
 */
const SHARE_VERSION = '1';

/** Holds the whole state; `code` alone still lives in `?code=` on old links. */
const SHARE_PARAM = 's';
/** Plain-text parameter used before share links were compressed. */
const LEGACY_CODE_PARAM = 'code';

const FLAG_CODE = 1;
const FLAG_RSLINT_CONFIG = 2;
const FLAG_TSCONFIG = 4;

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
 * The frame is a flags byte naming the sections that differ from their default,
 * followed by those sections in flag order. Every section but the last is
 * prefixed with its LEB128 byte length; the last one runs to the end.
 */
function encodeFrame(sections: Uint8Array[], flags: number): Uint8Array {
  const bytes: number[] = [flags];
  sections.forEach((section, index) => {
    if (index < sections.length - 1) {
      let length = section.length;
      do {
        const byte = length & 0x7f;
        length >>>= 7;
        bytes.push(length > 0 ? byte | 0x80 : byte);
      } while (length > 0);
    }
    for (const byte of section) bytes.push(byte);
  });
  return new Uint8Array(bytes);
}

function decodeFrame(
  frame: Uint8Array,
): { flags: number; sections: string[] } | null {
  if (frame.length === 0) return null;
  const flags = frame[0];
  const count =
    (flags & FLAG_CODE ? 1 : 0) +
    (flags & FLAG_RSLINT_CONFIG ? 1 : 0) +
    (flags & FLAG_TSCONFIG ? 1 : 0);
  if (count === 0 || flags > (FLAG_CODE | FLAG_RSLINT_CONFIG | FLAG_TSCONFIG)) {
    return null;
  }

  const sections: string[] = [];
  let offset = 1;
  for (let index = 0; index < count; index++) {
    if (index === count - 1) {
      sections.push(decoder.decode(frame.subarray(offset)));
      return { flags, sections };
    }
    let length = 0;
    let shift = 0;
    let byte: number;
    do {
      if (offset >= frame.length || shift > 28) return null;
      byte = frame[offset++];
      length |= (byte & 0x7f) << shift;
      shift += 7;
    } while (byte & 0x80);
    if (offset + length > frame.length) return null;
    sections.push(decoder.decode(frame.subarray(offset, offset + length)));
    offset += length;
  }
  return { flags, sections };
}

/**
 * Returns the payload for the share parameter, or `null` when every editor
 * holds its default and the link needs no parameter at all.
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
  if (flags === 0) return null;

  const compressed = deflateSync(encodeFrame(sections, flags), {
    level: 9,
    mem: 12,
    dictionary: getDictionary(),
  });
  return SHARE_VERSION + toBase64Url(compressed);
}

/**
 * Decodes a share payload. A truncated, corrupt or future-version link decodes
 * to `null` so the caller can fall back to the defaults instead of failing.
 */
export function decodeShareState(payload: string): ShareState | null {
  if (payload[0] !== SHARE_VERSION) return null;
  let decoded: { flags: number; sections: string[] } | null;
  try {
    const frame = inflateSync(fromBase64Url(payload.slice(1)), {
      dictionary: getDictionary(),
    });
    decoded = decodeFrame(frame);
  } catch {
    return null;
  }
  if (decoded === null) return null;

  const { flags, sections } = decoded;
  let index = 0;
  return {
    code: flags & FLAG_CODE ? sections[index++] : DEFAULT_CODE,
    rslintConfig:
      flags & FLAG_RSLINT_CONFIG ? sections[index++] : DEFAULT_RSLINT_CONFIG,
    tsconfig: flags & FLAG_TSCONFIG ? sections[index++] : DEFAULT_TSCONFIG,
  };
}

function readParam(name: string): string | null {
  const { search, hash } = window.location;
  const fromSearch = new URLSearchParams(search).get(name);
  if (fromSearch !== null) return fromSearch;
  if (hash.startsWith('#')) {
    return new URLSearchParams(hash.slice(1)).get(name);
  }
  return null;
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

  const payload = readParam(SHARE_PARAM);
  if (payload !== null) return decodeShareState(payload) ?? defaults;

  const legacyCode = readParam(LEGACY_CODE_PARAM);
  if (legacyCode !== null) return { ...defaults, code: legacyCode };

  return defaults;
}

/**
 * Rewrites the current URL to carry `state`. Nothing but the share parameter is
 * left behind: an unmodified playground gets a bare URL, and a link opened in
 * the legacy plain-text form is upgraded in place on the first edit.
 */
export function writeShareState(state: ShareState) {
  if (typeof window === 'undefined') return;
  try {
    const url = new URL(window.location.href);
    url.searchParams.delete(SHARE_PARAM);
    url.searchParams.delete(LEGACY_CODE_PARAM);
    const payload = encodeShareState(state);
    url.hash = payload === null ? '' : `${SHARE_PARAM}=${payload}`;
    window.history.replaceState(null, '', url.toString());
  } catch {
    // A URL we cannot rewrite is not worth failing an edit over.
  }
}
