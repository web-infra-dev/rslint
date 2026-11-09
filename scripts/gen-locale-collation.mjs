// Regenerate with: node scripts/gen-locale-collation.mjs
//
// Derives the printable-ASCII part of ICU collation from the running Node.js
// build and writes it as compact tables for internal/utils/ecmascript. Each
// locale is checked against Intl.Collator on a randomized corpus before it is
// emitted; a locale whose ASCII order the table model cannot reproduce falls
// back to the root order instead of shipping a wrong tailoring.
import { execFileSync } from 'node:child_process';
import { writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';

const outDir = resolve(import.meta.dirname, '../internal/utils/ecmascript');

// Candidate locales. Anything that does not change printable-ASCII order is
// dropped after derivation, so this list may be generous.
const candidates = [
  ...'af am ar as az be bg bn bo br bs ca ceb chr cs cy da de dsb dz ee el en eo es et eu fa ff fi fil fo fr fy ga gl gu ha haw he hi hr hsb hu hy id ig is it ja ka kk kl km kn ko kok ku ky lb lkt ln lo lt lv mk ml mn mr ms mt my nb ne nl nn no om or pa pl ps pt ro ru sa se si sk sl smn sq sr sv sw ta te th tk to tr ug uk ur uz vi wo xh yi yo zh zu'.split(
    ' ',
  ),
  'sr-Latn',
  'bs-Cyrl',
  'uz-Cyrl',
  'az-Cyrl',
  'es-u-co-trad',
  'sv-u-co-trad',
  'fi-u-co-trad',
  'de-u-co-phonebk',
  'ln-u-co-phonetic',
  'zh-u-co-pinyin',
  'zh-u-co-stroke',
];

const CGJ = '͏';
const printable = Array.from({ length: 0x7f - 0x20 }, (_, i) =>
  String.fromCharCode(0x20 + i),
);
const letters = printable.filter((c) => /[A-Za-z]/.test(c));

function lookupKey(tag) {
  const locale = new Intl.Locale(tag);
  const max = locale.maximize();
  const co = locale.collation;
  return `${max.language}-${max.script}${co ? `-u-co-${co}` : ''}`;
}

// --- derivation -----------------------------------------------------------

function segment(text, contractions) {
  // Greedy longest match, mirroring the comparer.
  const parts = [];
  for (let i = 0; i < text.length;) {
    let len = 1;
    for (const unit of contractions) {
      if (unit.length > len && text.startsWith(unit, i)) len = unit.length;
    }
    parts.push(text.slice(i, i + len));
    i += len;
  }
  return parts;
}

function findContractions(compare) {
  const found = [];
  // A contraction collates differently once a combining grapheme joiner,
  // which is completely ignorable, separates its characters.
  for (const a of printable) {
    for (const b of printable) {
      if (compare(a + b, a + CGJ + b) !== 0) found.push(a + b);
    }
  }
  // Longer contractions are letter sequences, plus the apostrophe of forms
  // such as Breton c'h.
  const pool = [...letters, "'"];
  const test = (text) => {
    if (
      !found.includes(text) &&
      compare(text, segment(text, found).join(CGJ)) !== 0
    ) {
      found.push(text);
    }
  };
  for (const a of pool)
    for (const b of pool) for (const c of pool) test(a + b + c);
  const shorter = [...found];
  for (const unit of shorter) {
    for (const a of letters) {
      test(a + unit);
      test(unit + a);
      if (unit.length === 2) for (const b of letters) test(a + unit + b);
    }
    for (const other of shorter)
      if (unit.length === 2 && other.length === 2) test(unit + other);
  }
  return found;
}

function rank(units, compare, groupOf) {
  const sorted = [...units].sort(compare);
  const ranks = new Map();
  let current = 0;
  let previous;
  for (const unit of sorted) {
    if (previous !== undefined) {
      if (groupOf(previous) !== groupOf(unit)) current = 0;
      else if (compare(previous, unit) !== 0) current++;
    }
    ranks.set(unit, current);
    previous = unit;
  }
  return ranks;
}

function derive(tag) {
  const make = (options) => new Intl.Collator(tag, options).compare;
  const full = make({ caseFirst: 'false' });
  const contractions = findContractions(full);
  const units = [...printable, ...contractions];

  const base = make({ sensitivity: 'base' });
  const accent = make({ sensitivity: 'accent' });
  const sorted = [...units].sort(full);
  const primary = new Map();
  const secondary = new Map();
  let p = 0;
  let s = 0;
  for (let i = 0; i < sorted.length; i++) {
    if (i > 0) {
      if (base(sorted[i - 1], sorted[i]) !== 0) {
        p++;
        s = 0;
      } else if (accent(sorted[i - 1], sorted[i]) !== 0) s++;
    }
    primary.set(sorted[i], p);
    secondary.set(sorted[i], s);
  }
  if (p > 0xff || s > 0xff) throw new Error(`${tag}: weights overflow`);
  const group = (u) => `${primary.get(u)}/${secondary.get(u)}`;
  const tertiary = {};
  for (const mode of ['false', 'lower', 'upper']) {
    tertiary[mode] = rank(units, make({ caseFirst: mode }), group);
  }
  for (const unit of units) {
    if (tertiary.false.get(unit) !== tertiary.lower.get(unit)) {
      throw new Error(`${tag}: caseFirst false and lower disagree on ${unit}`);
    }
  }
  const elem = (u) => [
    primary.get(u),
    secondary.get(u),
    tertiary.lower.get(u),
    tertiary.upper.get(u),
  ];
  const digitZero = primary.get('0');
  for (let d = 1; d <= 9; d++) {
    if (primary.get(String(d)) !== digitZero + d) {
      throw new Error(`${tag}: digits are not consecutive primaries`);
    }
  }
  return {
    tag,
    key: lookupKey(tag),
    upperFirst: new Intl.Collator(tag).resolvedOptions().caseFirst === 'upper',
    singles: printable.map(elem),
    contractions: contractions.map((u) => [u, elem(u)]),
  };
}

// --- model comparer (the ASCII part of locale_compare.go) -----------------

function elements(text, table, { numeric, upperFirst }) {
  const out = [];
  const points = [...text];
  const tIndex = upperFirst ? 3 : 2;
  const zero = table.singles['0'.charCodeAt(0) - 0x20];
  for (let i = 0; i < points.length;) {
    const ch = points[i];
    const code = ch.codePointAt(0);
    if (code < 0x20 || code === 0x7f) {
      i++;
      continue;
    }
    // The corpus is printable ASCII; the comparer's handling of other text is
    // not derived from ICU and is tested separately.
    if (code >= 0x80)
      throw new Error(`unexpected non-ASCII ${JSON.stringify(text)}`);
    if (numeric && ch >= '0' && ch <= '9') {
      let j = i;
      while (j < points.length && points[j] >= '0' && points[j] <= '9') j++;
      let digits = points.slice(i, j).join('').replace(/^0+/, '');
      if (digits === '') digits = '0';
      out.push([zero[0], zero[1], zero[tIndex]], [digits.length, 0, 0]);
      for (const d of digits) {
        const e = table.singles[d.charCodeAt(0) - 0x20];
        out.push([e[0], e[1], e[tIndex]]);
      }
      i = j;
      continue;
    }
    let matched;
    for (const [unit, e] of table.contractions) {
      if (
        (!matched || unit.length > matched[0].length) &&
        points.slice(i, i + unit.length).join('') === unit
      ) {
        matched = [unit, e];
      }
    }
    if (matched) {
      const e = matched[1];
      out.push([e[0], e[1], e[tIndex]]);
      i += matched[0].length;
      continue;
    }
    const e = table.singles[code - 0x20];
    out.push([e[0], e[1], e[tIndex]]);
    i++;
  }
  return out;
}

function modelCompare(table, options) {
  return (left, right) => {
    const a = elements(left, table, options);
    const b = elements(right, table, options);
    for (let level = 0; level < 3; level++) {
      const n = Math.min(a.length, b.length);
      for (let i = 0; i < n; i++) {
        if (a[i][level] !== b[i][level])
          return a[i][level] < b[i][level] ? -1 : 1;
      }
      if (a.length !== b.length) return a.length < b.length ? -1 : 1;
    }
    return 0;
  };
}

// --- verification ---------------------------------------------------------

function mulberry32(seed) {
  return () => {
    seed |= 0;
    seed = (seed + 0x6d2b79f5) | 0;
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

function corpus(table, seed, size) {
  const random = mulberry32(seed);
  const pick = (list) => list[Math.floor(random() * list.length)];
  const pieces = [
    ...printable,
    ...letters,
    ...letters,
    ...letters,
    ...table.contractions.map(([u]) => u),
    ...table.contractions.map(([u]) => u),
  ];
  const out = new Set();
  while (out.size < size) {
    const length = 1 + Math.floor(random() * 5);
    let text = '';
    for (let i = 0; i < length; i++) text += pick(pieces);
    out.add(text);
  }
  return [...out];
}

function sign(value) {
  return value < 0 ? -1 : value > 0 ? 1 : 0;
}

function verify(table, tag) {
  const variants = [
    { suffix: '', options: { numeric: false, upperFirst: table.upperFirst } },
    { suffix: 'kf-upper', options: { numeric: false, upperFirst: true } },
    { suffix: 'kf-lower', options: { numeric: false, upperFirst: false } },
    { suffix: 'kn', options: { numeric: true, upperFirst: table.upperFirst } },
  ];
  const strings = corpus(table, 0x5eed, 4000);
  const random = mulberry32(0xc0ffee);
  for (const { suffix, options } of variants) {
    const oracleTag = suffix
      ? `${tag}${tag.includes('-u-') ? '-' : '-u-'}${suffix}`
      : tag;
    const oracle = new Intl.Collator(oracleTag).compare;
    const model = modelCompare(table, options);
    for (let n = 0; n < 60000; n++) {
      const a = strings[Math.floor(random() * strings.length)];
      const b = strings[Math.floor(random() * strings.length)];
      if (sign(oracle(a, b)) !== sign(model(a, b))) {
        return `${oracleTag}: ${JSON.stringify(a)} vs ${JSON.stringify(b)} oracle=${sign(oracle(a, b))} model=${sign(model(a, b))}`;
      }
    }
  }
  return undefined;
}

// --- emit -----------------------------------------------------------------

function sameTable(a, b) {
  return (
    JSON.stringify([a.singles, a.contractions]) ===
    JSON.stringify([b.singles, b.contractions])
  );
}

// "en" uses the CLDR root collation; "und" would resolve to the host locale.
const root = derive('en');
const rootFailure = verify(root, 'en');
if (rootFailure) throw new Error(`root model mismatch: ${rootFailure}`);
if (root.contractions.length) throw new Error('root has ASCII contractions');

const tables = [root];
const locales = new Map();
const skipped = [];
for (const tag of candidates) {
  // An unsupported tag would silently resolve to the host default locale.
  if (Intl.Collator.supportedLocalesOf([tag]).length === 0) continue;
  const table = derive(tag);
  const failure = verify(table, tag);
  if (failure) {
    skipped.push(`${tag} (${failure})`);
    continue;
  }
  let index = tables.findIndex((t) => sameTable(t, table));
  if (index === 0 && !table.upperFirst) continue;
  if (index < 0) {
    index = tables.length;
    tables.push(table);
  }
  const previous = locales.get(table.key);
  if (
    previous &&
    (previous.index !== index || previous.upperFirst !== table.upperFirst)
  ) {
    throw new Error(`${tag}: lookup key ${table.key} is ambiguous`);
  }
  locales.set(table.key, {
    index,
    upperFirst: table.upperFirst,
    tags: [...(previous?.tags ?? []), tag],
  });
}

const hex = (bytes) =>
  `"${bytes.map((b) => `\\x${b.toString(16).padStart(2, '0')}`).join('')}"`;

let go = `// Code generated by scripts/gen-locale-collation.mjs; DO NOT EDIT.

package ecmascript

// Derived from Node.js ${process.version} (ICU ${process.versions.icu}, CLDR ${process.versions.cldr}).
const collationICUVersion = ${JSON.stringify(process.versions.icu)}

// collationTables[0] is the root collation. Each element is four bytes:
// primary, secondary, tertiary (lower first), tertiary (upper first).
var collationTables = [...]collationTable{
`;
for (const [index, table] of tables.entries()) {
  go += `\t{ // ${index === 0 ? 'root' : tables[index].tag}\n`;
  go += `\t\tsingles: ${hex(table.singles.flat())},\n`;
  if (table.contractions.length) {
    go += `\t\tcontractions: []collationContraction{\n`;
    for (const [unit, e] of [...table.contractions].sort(([a], [b]) =>
      a < b ? -1 : a > b ? 1 : 0,
    )) {
      go += `\t\t\t{${JSON.stringify(unit)}, ${hex(e)}},\n`;
    }
    go += `\t\t},\n`;
  }
  go += `\t},\n`;
}
go += `}

// collationLocales maps language-Script[-u-co-type] to a table. Locales that
// keep the root order for printable ASCII are absent.
var collationLocales = map[string]collationLocale{
`;
for (const [key, { index, upperFirst, tags }] of [...locales].sort(
  ([a], [b]) => (a < b ? -1 : 1),
)) {
  go += `\t${JSON.stringify(key)}: {table: ${index}, upperFirst: ${upperFirst}}, // ${tags.join(', ')}\n`;
}
go += `}
`;
if (skipped.length) {
  go += `
// Locales whose ASCII order this model cannot reproduce and that therefore use
// the root order:
${skipped.map((s) => `//   - ${s.split(' (')[0]}`).join('\n')}
`;
}
const goFile = resolve(outDir, 'locale_collation_table.go');
await writeFile(goFile, go);
execFileSync('gofmt', ['-w', goFile]);

// Golden data for the Go test: per lookup key, a corpus sorted by Intl.Collator
// with the locale's default and opposite case order, plus numeric collation
// for the root and one upper-first tailoring. Strings that tie are excluded so
// the expected order is unique.
const golden = [];
for (const [key, { index, tags }] of [
  ['root', { index: 0, tags: ['en'] }],
  ...locales,
]) {
  const tag = tags[0];
  const strings = corpus(tables[index], 0xbeef + index, 100).filter(
    (s) => !/[\x00-\x1f\x7f]/.test(s),
  );
  const suffixes = ['', 'kf-upper', 'kf-lower'];
  if (key === 'root' || tag === 'da') suffixes.push('kn');
  for (const suffix of suffixes) {
    const oracleTag = suffix
      ? `${tag}${tag.includes('-u-') ? '-' : '-u-'}${suffix}`
      : tag;
    const compare = new Intl.Collator(oracleTag).compare;
    const sorted = [...strings]
      .sort(compare)
      .filter((s, i, all) => i === 0 || compare(all[i - 1], s) !== 0);
    golden.push({ locale: oracleTag, sorted });
  }
}
await writeFile(
  resolve(outDir, 'testdata/locale_collation_golden.json'),
  `[\n${golden.map((entry) => JSON.stringify(entry)).join(',\n')}\n]\n`,
);

console.log(`tables: ${tables.length}, locales: ${locales.size}`);
if (skipped.length) console.log(`skipped:\n  ${skipped.join('\n  ')}`);
