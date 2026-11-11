// Regenerate with: node scripts/gen-locale-collation.mjs
//
// Derives the printable-ASCII part of ICU collation from the running Node.js
// build and writes it as compact tables for internal/utils/ecmascript.
//
// Intl.Collator selects collation data by canonicalizing the tag and then
// truncating it until an available locale matches; it never adds likely
// subtags, so "sr-ME" keeps the Serbian default rather than Serbian Latin and
// "az-IR" keeps Azerbaijani. Instead of reimplementing that resolution, the
// generator enumerates every language code, every script and region under the
// languages ICU knows, and every collation type, records the order Intl.Collator
// actually uses, and emits the sparse set of lookup keys that the Go resolver
// (truncation of the tag as written) needs to reproduce all of them. Subtags
// outside the enumerated sets are unknown to ICU and truncated by both sides.
//
// Each distinct order is derived into a table and checked against
// Intl.Collator on a randomized corpus; one the table model cannot reproduce
// falls back to the root order instead of shipping a wrong tailoring.
// cspell:ignore Qzzz aaland alalc arevela arevmda bokmal heploc polyton polytoni saaho
import { execFileSync } from 'node:child_process';
import { writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';

const outDir = resolve(import.meta.dirname, '../internal/utils/ecmascript');

const CGJ = '͏';
const printable = Array.from({ length: 0x7f - 0x20 }, (_, i) =>
  String.fromCharCode(0x20 + i),
);
const letters = printable.filter((c) => /[A-Za-z]/.test(c));

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
  // A locale that ignores punctuation by default, such as Thai, makes space
  // and punctuation completely ignorable.
  const ignorableChars = printable.filter((c) => full(c, '') === 0);
  const contractions = findContractions(full);
  const units = [
    ...printable.filter((c) => !ignorableChars.includes(c)),
    ...contractions,
  ];

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
  const elem = (u) =>
    ignorableChars.includes(u)
      ? [0, 0, 0, 0]
      : [
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
  // With kn, ICU gives digit runs the primary of the root digit group, which
  // a tailoring such as POSIX can move every ASCII character above.
  const numericBase = make({ numeric: true, sensitivity: 'base' });
  const others = printable.filter(
    (c) => !/[0-9]/.test(c) && !ignorableChars.includes(c),
  );
  const numericFirst = others.every((c) => numericBase('0', c) < 0);
  if (
    !numericFirst &&
    others.some((c) => numericBase('0', c) < 0 !== primary.get(c) > digitZero)
  ) {
    throw new Error(`${tag}: numbers do not sort with the digit zero`);
  }
  return {
    tag,
    upperFirst: new Intl.Collator(tag).resolvedOptions().caseFirst === 'upper',
    numericFirst,
    ignorableChars: ignorableChars.join(''),
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
    if (code < 0x20 || code === 0x7f || table.ignorableChars.includes(ch)) {
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
      out.push(
        [table.numericFirst ? -1 : zero[0], zero[1], zero[tIndex]],
        [digits.length, 0, 0],
      );
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

function verify(table, tag, iterations = 60000) {
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
    for (let n = 0; n < iterations; n++) {
      const a = strings[Math.floor(random() * strings.length)];
      const b = strings[Math.floor(random() * strings.length)];
      if (sign(oracle(a, b)) !== sign(model(a, b))) {
        return `${oracleTag}: ${JSON.stringify(a)} vs ${JSON.stringify(b)} oracle=${sign(oracle(a, b))} model=${sign(model(a, b))}`;
      }
    }
  }
  return undefined;
}

// --- tag model (mirrors parseLocaleTag and lookupCollationLocale) ---------

const alpha = /^[a-z]+$/;
const digit = /^[0-9]+$/;

// parseTag splits a tag the way the Go resolver does: subtags are matched by
// shape in BCP 47 order, and anything that does not fit is skipped.
function parseTag(input) {
  const subtags = input
    .toLowerCase()
    .split(/[-_]/)
    .filter((s) => s !== '');
  const first = subtags[0];
  if (
    !first ||
    !alpha.test(first) ||
    !(first.length <= 3
      ? first.length >= 2
      : first.length >= 5 && first.length <= 8)
  ) {
    return undefined;
  }
  const tag = {
    language: first,
    script: '',
    region: '',
    posix: false,
    keys: {},
  };
  let i = 1;
  if (i < subtags.length && subtags[i].length === 4 && alpha.test(subtags[i])) {
    tag.script = subtags[i++];
  }
  if (
    i < subtags.length &&
    ((subtags[i].length === 2 && alpha.test(subtags[i])) ||
      (subtags[i].length === 3 && digit.test(subtags[i])))
  ) {
    tag.region = subtags[i++];
  }
  const variants = [];
  for (; i < subtags.length && subtags[i].length > 1; i++)
    variants.push(subtags[i]);
  const vaTypes = [];
  while (i < subtags.length) {
    const singleton = subtags[i++];
    if (singleton === 'x') break;
    const start = i;
    while (i < subtags.length && subtags[i].length > 1) i++;
    if (singleton !== 'u') continue;
    let key = '';
    let types = [];
    const flush = () => {
      if (!key) return;
      const type = types.length ? types.join('-') : 'true';
      if (key === 'va') vaTypes.push(type);
      if (!(key in tag.keys)) tag.keys[key] = type;
    };
    for (const subtag of subtags.slice(start, i)) {
      if (subtag.length === 2) {
        flush();
        key = subtag;
        types = [];
      } else if (key) {
        types.push(subtag);
      }
    }
    flush();
  }
  tag.posix = posixOf(applyVariantRules(tag, variants), vaTypes);
  return tag;
}

// posixOf reports whether a tag selects the POSIX order, which canonical
// -u-va-posix does. Canonicalization turns a posix variant into va-posix only
// when it is the only variant. With other variants, the first va type is
// kept; without them, every va-posix becomes the variant and any other va
// type overrides it.
function posixOf(variants, vaTypes) {
  if (vaTypes.length) {
    return variants.length
      ? vaTypes[0] === 'posix'
      : vaTypes.every((type) => type === 'posix');
  }
  return variants.length === 1 && variants[0] === 'posix';
}

function extensionSuffixes(co, posix) {
  const suffixes = [];
  if (co && posix) suffixes.push(`-u-co-${co}-va-posix`);
  if (posix) suffixes.push('-u-va-posix');
  if (co) suffixes.push(`-u-co-${co}`);
  suffixes.push('');
  return suffixes;
}

function chainOf({ language, script, region }) {
  // "-*" stands for any script without its own key: a language alias that
  // carries a script, such as "sh" for "sr-Latn", loses it to an explicit one.
  if (script) {
    return region
      ? [
          `${language}-${script}-${region}`,
          `${language}-${script}`,
          `${language}-*`,
          language,
        ]
      : [`${language}-${script}`, `${language}-*`, language];
  }
  return region ? [`${language}-${region}`, language] : [language];
}

const entries = new Map();
const rootOutcome = { table: 0, upperFirst: false };

function lookup(tag) {
  for (const suffix of extensionSuffixes(tag.keys.co, tag.posix)) {
    // "*" stands for every language, for collation types such as emoji that
    // replace any tailoring.
    for (const key of suffix ? [...chainOf(tag), '*'] : chainOf(tag)) {
      const entry = entries.get(key + suffix);
      if (entry) return entry;
    }
  }
  return rootOutcome;
}

function sameTable(a, b) {
  return (
    JSON.stringify([
      a.singles,
      a.contractions,
      a.numericFirst,
      a.ignorableChars,
    ]) ===
    JSON.stringify([
      b.singles,
      b.contractions,
      b.numericFirst,
      b.ignorableChars,
    ])
  );
}

const sameOutcome = (a, b) =>
  a.table === b.table && a.upperFirst === b.upperFirst;

// --- what Intl.Collator does -----------------------------------------------

// Only the collation-relevant identity of a tag: the locale Intl.Collator
// resolves to, which includes a supported co and va. An unsupported tag
// resolves to the host default, which the Go comparer replaces with root.
const resolvedCache = new Map();
function resolvedLocale(tag) {
  let resolved = resolvedCache.get(tag);
  if (resolved === undefined) {
    resolved = Intl.Collator.supportedLocalesOf([tag]).length
      ? new Intl.Collator(tag).resolvedOptions().locale
      : '';
    resolvedCache.set(tag, resolved);
  }
  return resolved;
}

// "en" uses the CLDR root collation; "und" would resolve to the host locale.
const root = derive('en');
const rootFailure = verify(root, 'en');
if (rootFailure) throw new Error(`root model mismatch: ${rootFailure}`);
if (root.contractions.length) throw new Error('root has ASCII contractions');
root.locales = ['root'];

const tables = [root];
const skipped = [];
const outcomeCache = new Map([['', rootOutcome]]);
function outcomeOf(resolved) {
  let outcome = outcomeCache.get(resolved);
  if (outcome) return outcome;
  outcome = rootOutcome;
  let table;
  let failure;
  try {
    table = derive(resolved);
  } catch (error) {
    failure = error.message;
  }
  if (table) {
    const index = tables.findIndex((t) => sameTable(t, table));
    // A table that matches a verified one needs only a shorter check.
    failure = verify(table, resolved, index < 0 ? 60000 : 10000);
    if (!failure) {
      if (index < 0) {
        table.locales = [resolved];
        outcome = { table: tables.length, upperFirst: table.upperFirst };
        tables.push(table);
      } else {
        tables[index].locales.push(resolved);
        outcome = { table: index, upperFirst: table.upperFirst };
      }
    }
  }
  if (failure) skipped.push(`${resolved} (${failure})`);
  if (outcome.table === 0 && !outcome.upperFirst) outcome = rootOutcome;
  outcomeCache.set(resolved, outcome);
  return outcome;
}

// --- enumeration -------------------------------------------------------------

const lower = 'abcdefghijklmnopqrstuvwxyz';
const languages = [];
for (const a of lower) {
  for (const b of lower) {
    languages.push(a + b);
    for (const c of lower) languages.push(a + b + c);
  }
}

// The CLDR 48 alias rules that involve variants (supplemental/aliases.json,
// languageAlias and variantAlias, without the grandfathered i- tags that
// Intl.Collator rejects). Canonicalization applies them before the POSIX
// variant is counted, so they can remove, replace, or add variants, change the
// language, or add a region.
const cldrVariantRules = [
  ['aa', 'saaho', 'ssy'],
  ['art', 'lojban', 'jbo'],
  ['cel', 'gaulish', 'xtg'],
  ['hy', 'arevmda', 'hyw'],
  ['no', 'bokmal', 'nb'],
  ['no', 'nynorsk', 'nn'],
  ['zh', 'guoyu', 'zh'],
  ['zh', 'hakka', 'hak'],
  ['zh', 'xiang', 'hsn'],
  ['und', 'hepburn-heploc', 'und-alalc97'],
  ['und', 'aaland', 'und-AX'],
  ['und', 'arevela', 'und'],
  ['und', 'arevmda', 'und'],
  ['und', 'bokmal', 'und'],
  ['und', 'hakka', 'und'],
  ['und', 'lojban', 'und'],
  ['und', 'nynorsk', 'und'],
  ['und', 'saaho', 'und'],
  ['und', 'xiang', 'und'],
  ['und', 'heploc', 'und-alalc97'],
  ['und', 'polytoni', 'und-polyton'],
];
// A rule for a language also applies to its aliases, such as "zho" for "zh".
const canonicalLanguage = new Map(
  languages.map((language) => [
    language,
    Intl.getCanonicalLocales(language)[0].split('-')[0].toLowerCase(),
  ]),
);
const variantRules = cldrVariantRules.flatMap(([language, variants, to]) => {
  const [replacement, ...rest] = to.toLowerCase().split('-');
  const rule = {
    variants: variants.split('-'),
    replacement: replacement === 'und' ? '' : replacement,
    region: rest.find((s) => s.length === 2 || /^[0-9]{3}$/.test(s)) ?? '',
    add: rest.filter((s) => s.length > 3),
  };
  if (language === 'und') return [{ ...rule, language: '' }];
  return languages
    .filter((spelling) => canonicalLanguage.get(spelling) === language)
    .map((spelling) => ({ ...rule, language: spelling }));
});

// applyVariantRules rewrites a tag's language, region, and variants with the
// rules until none applies, as canonicalization does. A rule adds its region
// only to a tag without one.
function applyVariantRules(tag, variants) {
  let current = variants;
  for (;;) {
    const rule = variantRules.find(
      (r) =>
        (r.language === '' || r.language === tag.language) &&
        r.variants.every((v) => current.includes(v)),
    );
    if (!rule) return current;
    current = [
      ...current.filter((v) => !rule.variants.includes(v)),
      ...rule.add.filter((v) => !current.includes(v)),
    ];
    if (rule.replacement) tag.language = rule.replacement;
    if (rule.region && !tag.region) tag.region = rule.region;
  }
}
const scriptNames = new Intl.DisplayNames('en', {
  type: 'script',
  fallback: 'none',
});
const scripts = [];
for (const a of lower.toUpperCase())
  for (const b of lower)
    for (const c of lower)
      for (const d of lower) {
        const script = a + b + c + d;
        if (scriptNames.of(script)) scripts.push(script);
      }
const regions = [];
for (const a of lower)
  for (const b of lower) regions.push((a + b).toUpperCase());
for (let i = 0; i < 1000; i++) regions.push(String(i).padStart(3, '0'));
// Every type a tag can request; "standard" and "search" cannot be requested.
const collations = Intl.supportedValuesOf('collation');

// Languages ICU has data for, or that canonicalize to one, such as "sh".
// Scripts and regions under any other language are unknown to ICU.
const known = languages.filter(
  (language) =>
    Intl.Collator.supportedLocalesOf([language]).length > 0 ||
    Intl.getCanonicalLocales(language)[0] !== language,
);

// Tags in order of specificity: a key's entry never affects a less specific
// tag, so most entries are final the first time they are set.
// Qzzz is not a registered script and stands for every unknown one.
const unknownScript = 'Qzzz';
if (scripts.includes(unknownScript))
  throw new Error(`${unknownScript} is registered`);
const tags = languages.map((language) => ({ text: language, level: 0 }));
for (const language of known) {
  tags.push({
    text: `${language}-${unknownScript}`,
    level: 0.5,
    wildcard: true,
  });
  for (const script of scripts) {
    const text = `${language}-${script}`;
    tags.push({ text, level: 1 });
    // A region under a script only matters when that script has data.
    if (resolvedLocale(text).includes(`-${script}`)) {
      for (const region of regions)
        tags.push({ text: `${text}-${region}`, level: 2 });
    }
  }
  for (const region of regions)
    tags.push({ text: `${language}-${region}`, level: 1 });
}
tags.sort((a, b) => a.level - b.level);

const extensions = [
  ...collations.map((co) => ({ co, posix: false })),
  { co: '', posix: true },
  ...collations.map((co) => ({ co, posix: true })),
];
const suffixOf = ({ co, posix }) => extensionSuffixes(co, posix)[0];

// Intl.Collator resolves the locale and then the extension keys, so a tag's
// order with an extension is that of its resolved locale with the extension.
const extended = (resolved, extension) =>
  resolved === '' ? '' : resolvedLocale(resolved + suffixOf(extension));

for (const tag of tags) {
  tag.resolved = resolvedLocale(tag.text);
  tag.parsed = parseTag(tag.text);
  tag.expected = outcomeOf(tag.resolved);
}
// Pairs whose order differs from the plain tag. Every other pair is checked
// only when its language has an entry for that extension.
const special = [];
const resolvedSet = [...new Set(tags.map((tag) => tag.resolved))].sort();
const specialExtensions = new Map();
for (const resolved of resolvedSet) {
  const plain = outcomeOf(resolved);
  const list = extensions.filter(
    (extension) =>
      !sameOutcome(outcomeOf(extended(resolved, extension)), plain),
  );
  if (list.length) specialExtensions.set(resolved, list);
}
for (const tag of tags) {
  for (const extension of specialExtensions.get(tag.resolved) ?? []) {
    // Check the assumption above directly for every pair that matters.
    const direct = outcomeOf(resolvedLocale(tag.text + suffixOf(extension)));
    special.push({
      text: tag.text + suffixOf(extension),
      parsed: parseTag(tag.text + suffixOf(extension)),
      level: tag.level,
      wildcard: tag.wildcard,
      expected: direct,
    });
  }
}

// --- emit entries --------------------------------------------------------------

// An extension that gives every tailored language the same order gets one
// "*" entry instead of one per language.
const tailoredLanguages = tags.filter(
  (tag) => tag.level === 0 && tag.expected !== rootOutcome,
);
for (const extension of extensions) {
  const outcomes = tailoredLanguages.map((tag) =>
    outcomeOf(resolvedLocale(tag.text + suffixOf(extension))),
  );
  if (
    outcomes.every((outcome) => sameOutcome(outcome, outcomes[0])) &&
    !tailoredLanguages.every((tag, i) => sameOutcome(tag.expected, outcomes[i]))
  ) {
    entries.set(`*${suffixOf(extension)}`, outcomes[0]);
  }
}

const keyOf = (parsed, wildcard) =>
  (wildcard ? `${parsed.language}-*` : chainOf(parsed)[0]) +
  extensionSuffixes(parsed.keys.co, parsed.posix)[0];
const languagesWithExtension = new Set();
function settle(list) {
  let changed = false;
  for (const tag of list) {
    if (!sameOutcome(lookup(tag.parsed), tag.expected)) {
      const key = keyOf(tag.parsed, tag.wildcard);
      entries.set(key, tag.expected);
      if (key.includes('-u-')) languagesWithExtension.add(tag.parsed.language);
      changed = true;
    }
  }
  return changed;
}
special.sort((a, b) => a.level - b.level);
for (let pass = 0; ; pass++) {
  if (pass > 10) throw new Error('lookup entries do not converge');
  let changed = settle(tags);
  changed = settle(special) || changed;
  // A plain pair can be shadowed by an extension entry of its language.
  const shadowed = [];
  for (const tag of tags) {
    if (!languagesWithExtension.has(tag.parsed.language)) continue;
    const listed = new Set(specialExtensions.get(tag.resolved) ?? []);
    for (const extension of extensions) {
      if (listed.has(extension)) continue;
      shadowed.push({
        parsed: parseTag(tag.text + suffixOf(extension)),
        wildcard: tag.wildcard,
        expected: tag.expected,
      });
    }
  }
  changed = settle(shadowed) || changed;
  if (!changed) break;
}

// --- verify the resolver beyond the enumeration -------------------------------

function checkTag(text) {
  // Intl.Collator rejects malformed tags; the Go comparer uses root for them.
  if (!isValidTag(text)) return;
  const parsed = parseTag(text);
  const expected = outcomeOf(resolvedLocale(text));
  const got = parsed ? lookup(parsed) : rootOutcome;
  if (!sameOutcome(got, expected)) {
    throw new Error(
      `${text}: resolver picks ${JSON.stringify(got)}, Intl.Collator uses ${resolvedLocale(text) || 'root'} ${JSON.stringify(expected)}`,
    );
  }
}
function isValidTag(text) {
  try {
    Intl.getCanonicalLocales(text);
    return true;
  } catch {
    return false;
  }
}
const random = mulberry32(0x10ca1e);
const pick = (list) => list[Math.floor(random() * list.length)];
// Unknown languages, invented subtags, and every extension on random tags.
for (let n = 0; n < 30000; n++) {
  const base = pick(tags).text;
  checkTag(base + suffixOf(pick(extensions)));
  checkTag(`${base}-Qzzz`);
  checkTag(`${pick(languages)}-${pick(scripts)}-${pick(regions)}`);
}
// Variants other than posix are truncated; these include CLDR variant aliases.
const tailored = [
  ...new Set([...entries.keys()].map((key) => key.split('-')[0])),
];
for (const language of tailored) {
  for (const variant of [
    'posix',
    'aaland',
    'arevela',
    'arevmda',
    'heploc',
    'polytoni',
    'fonipa',
    'valencia',
    '1901',
    '1996',
    'pinyin',
    'phonebk',
    'trad',
  ]) {
    for (const middle of ['', '-US', '-Latn', '-Cyrl-RS'])
      checkTag(`${language}${middle}-${variant}`);
  }
  for (const key of [
    'ka-shifted',
    'kr-latn-digit',
    'ks-level1',
    'kc',
    'va-other',
  ]) {
    checkTag(`${language}-u-${key}`);
  }
}
for (const tag of [
  'art-lojban',
  'zh-guoyu',
  'zh-hakka',
  'en-US-POSIX-u-co-phonebk',
  'en-US-u-kf-lower-va-posix',
]) {
  checkTag(tag);
}

// --- extension keys ------------------------------------------------------------

// The kf and kn types Intl.Collator honors, derived rather than read from the
// spec, since aliases such as kn-yes count too. Any other type leaves the
// locale default. A type has three to eight characters; a two-character
// subtag such as the "no" in kn-no starts another key.
const typeVocabulary = new Set([
  'true',
  'false',
  'yes',
  'no',
  'upper',
  'lower',
  'on',
  'off',
  'none',
  'foo',
  'posix',
  ...collations,
]);
{
  const random = mulberry32(0x7e5);
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
  for (let n = 0; n < 300; n++) {
    let type = '';
    for (let i = 3 + Math.floor(random() * 6); i > 0; i--)
      type += chars[Math.floor(random() * chars.length)];
    typeVocabulary.add(type);
  }
}
const numericTypes = new Set();
const caseFirstTypes = new Map();
for (const type of typeVocabulary) {
  if (type.length < 3 || !isValidTag(`en-u-kn-${type}`)) continue;
  if (new Intl.Collator(`en-u-kn-${type}`).resolvedOptions().numeric)
    numericTypes.add(type);
  if (
    new Intl.Collator(`en-u-kf-${type}`).resolvedOptions().caseFirst === 'upper'
  )
    caseFirstTypes.set(type, true);
  else if (
    new Intl.Collator(`da-u-kf-${type}`).resolvedOptions().caseFirst !== 'upper'
  )
    caseFirstTypes.set(type, false);
}

// Languages Intl.Collator has data for. A tag in any other language resolves
// to the default locale, and ResolveLocale then drops its extension keys too.
const supportedLanguages = new Set(
  languages.filter(
    (language) => Intl.Collator.supportedLocalesOf([language]).length > 0,
  ),
);
for (const tag of tags) {
  if ((tag.resolved !== '') !== supportedLanguages.has(tag.parsed.language)) {
    throw new Error(`${tag.text}: support differs from its language`);
  }
}

// effective mirrors NewLocaleComparer: the table, case order, and numeric
// collation a tag selects.
function effective(text) {
  const parsed = parseTag(text);
  if (!parsed || !supportedLanguages.has(parsed.language)) {
    return { table: 0, upperFirst: false, numeric: false };
  }
  const entry = lookup(parsed);
  const kf = caseFirstTypes.get(parsed.keys.kf);
  return {
    table: entry.table,
    upperFirst: kf ?? entry.upperFirst,
    numeric: numericTypes.has(parsed.keys.kn),
  };
}

// withoutCaseAndNumeric drops kf and kn from a canonical resolved locale, so
// the remainder names the collation data.
function withoutCaseAndNumeric(resolved) {
  const [base, extension] = resolved.split('-u-');
  if (extension === undefined) return resolved;
  if (resolved.includes('-t-') || resolved.includes('-x-')) {
    throw new Error(`unexpected extension in resolved locale ${resolved}`);
  }
  const kept = [];
  let skip = false;
  for (const subtag of extension.split('-')) {
    if (subtag.length === 2) skip = subtag === 'kf' || subtag === 'kn';
    if (!skip) kept.push(subtag);
  }
  return kept.length ? `${base}-u-${kept.join('-')}` : base;
}

// intlEffective is what Intl.Collator does with a tag it accepts.
// A tag in an unsupported language resolves to the host's default locale; the
// comparer uses root defaults instead, so only check that Intl.Collator
// dropped the tag's keys.
const hostOptions = new Intl.Collator().resolvedOptions();
function intlEffective(text) {
  const options = new Intl.Collator(text).resolvedOptions();
  if (Intl.Collator.supportedLocalesOf([text]).length === 0) {
    if (
      options.caseFirst !== hostOptions.caseFirst ||
      options.numeric !== hostOptions.numeric
    ) {
      throw new Error(`${text}: an unsupported locale kept its extension keys`);
    }
    return { table: 0, upperFirst: false, numeric: false };
  }
  const outcome = outcomeOf(withoutCaseAndNumeric(options.locale));
  return {
    table: outcome.table,
    upperFirst: options.caseFirst === 'upper',
    numeric: options.numeric,
  };
}

const sameEffective = (a, b) =>
  a.table === b.table &&
  a.upperFirst === b.upperFirst &&
  a.numeric === b.numeric;

// Every combination of the features that change the result, within a bound:
// bases that cover default case order, collation types, POSIX, aliases, and
// unknown or unmodeled locales; variants around posix; and every sequence of
// up to three kf, kn, co, and va keys, including repeats, aliases, empty
// types, multi-subtag types, and unsupported types. Random noise then adds
// what must not matter: attributes, other extensions, private use, and case.
const fuzzBases = [
  'en',
  'en-US',
  'en-Latn-US',
  'da',
  'es-MX',
  'sh',
  'sr-Latn-ME',
  'hu',
  'th',
  'xx',
  'fr-CA',
];
const fuzzVariants = [
  '',
  '-posix',
  '-fonipa',
  '-posix-fonipa',
  '-fonipa-posix',
  '-1901-posix',
];
const fuzzKeys = [
  ...['', '-true', '-false', '-yes', '-no', '-foo'].map((type) => `kn${type}`),
  ...['', '-upper', '-lower', '-false', '-yes', '-no', '-true', '-foo'].map(
    (type) => `kf${type}`,
  ),
  ...['', '-trad', '-emoji', '-foo', '-trad-foo'].map((type) => `co${type}`),
  ...['', '-posix', '-other', '-posix-foo', '-true'].map((type) => `va${type}`),
];
const fuzzExtensions = [''];
for (const a of fuzzKeys) {
  fuzzExtensions.push(`-u-${a}`);
  for (const b of fuzzKeys) {
    fuzzExtensions.push(`-u-${a}-${b}`);
    for (const c of fuzzKeys) fuzzExtensions.push(`-u-${a}-${b}-${c}`);
  }
}
const fuzzMismatches = [];
const fuzzSample = [];
const fuzzRandom = mulberry32(0xf022);
const noise = (text) => {
  let out = text;
  const roll = fuzzRandom();
  if (roll < 0.25) out = out.replace('-u-', '-u-attr-');
  else if (roll < 0.5) out = out.replace('-u-', '-t-de-u-');
  else if (roll < 0.75) out += '-x-kn-va-posix';
  else out = out.replace('-u-', '-a-bcd-u-');
  return [...out]
    .map((c) => (fuzzRandom() < 0.3 ? c.toUpperCase() : c))
    .join('');
};
let fuzzChecked = 0;
const checkEffective = (text) => {
  if (!isValidTag(text)) return;
  fuzzChecked++;
  const expected = intlEffective(text);
  const got = effective(text);
  if (!sameEffective(got, expected))
    fuzzMismatches.push({ text, got, expected });
  if (fuzzRandom() < 0.0005) fuzzSample.push(text);
};
for (const base of fuzzBases) {
  for (const variant of fuzzVariants) {
    for (const extension of fuzzExtensions) {
      const text = base + variant + extension;
      checkEffective(text);
      if (extension && fuzzRandom() < 0.05) checkEffective(noise(text));
    }
  }
}
// Every variant the rules mention, alone, with posix, and in pairs, under the
// languages the rules name (and their aliases), with and without a region.
const ruleVariants = [
  ...new Set([
    ...variantRules.flatMap((rule) => [...rule.variants, ...rule.add]),
    'fonipa',
  ]),
];
const variantForms = new Set();
for (const a of ruleVariants) {
  variantForms.add(`-${a}`).add(`-${a}-posix`).add(`-posix-${a}`);
  for (const b of ruleVariants) {
    if (a !== b) variantForms.add(`-${a}-${b}`).add(`-${a}-${b}-posix`);
  }
}
const variantBases = [
  ...new Set([
    'en',
    'en-US',
    'sv',
    'da',
    'xx',
    'sr-Latn',
    ...variantRules.flatMap((rule) =>
      rule.language ? [rule.language, `${rule.language}-US`] : [],
    ),
  ]),
];
for (const base of variantBases) {
  for (const variant of variantForms) {
    for (const extension of [
      '',
      '-u-va-posix',
      '-u-va-other',
      '-u-kn',
      '-u-kf-upper',
      '-u-co-trad',
    ]) {
      checkEffective(base + variant + extension);
    }
  }
}

if (fuzzMismatches.length) {
  if (process.env.FUZZ_DUMP) {
    await writeFile(
      process.env.FUZZ_DUMP,
      fuzzMismatches.map((m) => JSON.stringify(m)).join('\n'),
    );
  }
  for (const { text, got, expected } of fuzzMismatches.slice(0, 40)) {
    console.error(
      `${text}: resolver ${JSON.stringify(got)}, Intl.Collator ${JSON.stringify(expected)}`,
    );
  }
  throw new Error(
    `${fuzzMismatches.length} tags resolve differently from Intl.Collator`,
  );
}

// --- write -------------------------------------------------------------------

const goStrings = (list) =>
  list.length
    ? `[]string{${list.map((v) => JSON.stringify(v)).join(', ')}}`
    : 'nil';
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
  go += `\t{ // ${index === 0 ? 'root' : table.locales[0]}\n`;
  go += `\t\tsingles: ${hex(table.singles.flat())},\n`;
  if (table.numericFirst) go += `\t\tnumericFirst: true,\n`;
  if (table.ignorableChars)
    go += `\t\tignorableChars: ${JSON.stringify(table.ignorableChars)},\n`;
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

// collationLocales maps lowercase language[-script][-region] keys, optionally
// followed by -u-co-<type> and -va-posix, to a table. The resolver tries the
// tag's own key and then its truncations, so only keys whose order differs
// from their truncation are listed.
var collationLocales = map[string]collationLocale{
`;
for (const [key, { table, upperFirst }] of [...entries].sort(([a], [b]) =>
  a < b ? -1 : 1,
)) {
  go += `\t${JSON.stringify(key)}: {table: ${table}, upperFirst: ${upperFirst}},\n`;
}
go += `}

// collationLanguages lists, space-separated, the language subtags
// Intl.Collator has data for, including aliases such as "sh". A tag in any
// other language uses the default locale and ignores its extension keys.
const collationLanguages = " ${[...supportedLanguages].sort().join(' ')} "

// collationVariantRules are the CLDR alias rules that involve variants, for
// a language spelling or, with an empty language, any language. They apply
// in order until none matches.
var collationVariantRules = []collationVariantRule{
${variantRules
  .map(
    (rule) =>
      `\t{language: ${JSON.stringify(rule.language)}, variants: ${goStrings(rule.variants)}, replacement: ${JSON.stringify(rule.replacement)}, region: ${JSON.stringify(rule.region)}, add: ${goStrings(rule.add)}},\n`,
  )
  .join('')}}

// collationNumericTypes lists the kn types that enable numeric collation,
// including aliases. A key without a type has the type "true".
var collationNumericTypes = map[string]bool{
${[...numericTypes]
  .sort()
  .map((type) => `\t${JSON.stringify(type)}: true,\n`)
  .join('')}}

// collationCaseFirstTypes maps the kf types Intl.Collator honors to whether
// uppercase sorts first. Other types keep the locale default.
var collationCaseFirstTypes = map[string]bool{
${[...caseFirstTypes]
  .sort(([a], [b]) => (a < b ? -1 : 1))
  .map(([type, upper]) => `\t${JSON.stringify(type)}: ${upper},\n`)
  .join('')}}
`;
const skippedLocales = [
  ...new Set(skipped.map((s) => s.split(' (')[0])),
].sort();
if (skippedLocales.length) {
  go += `
// Locales whose ASCII order this model cannot reproduce and that therefore use
// the root order:
${skippedLocales.map((s) => `//   - ${s}`).join('\n')}
`;
}
const goFile = resolve(outDir, 'locale_collation_table.go');
await writeFile(goFile, go);
execFileSync('gofmt', ['-w', goFile]);

// Golden data for the Go test: per table, a corpus sorted by Intl.Collator
// with the locale's default and opposite case order, plus numeric collation
// for the root and one upper-first tailoring. Strings that tie are excluded so
// the expected order is unique.
const golden = [];
for (const [index, table] of tables.entries()) {
  const tag = index === 0 ? 'en' : table.locales[0];
  // The corpus is printable ASCII, so it has no ignorable characters.
  const strings = corpus(table, 0xbeef + index, 100);
  const suffixes = ['', 'kf-upper', 'kf-lower'];
  if (index === 0 || table.locales.includes('da')) suffixes.push('kn');
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

// Resolution data: the table Intl.Collator's order corresponds to for every
// tag that needed an entry, the tags it truncates to, and a sample of the
// enumeration with random extensions.
const resolution = new Map();
const addResolution = (text) => {
  if (!resolution.has(text)) resolution.set(text, intlEffective(text));
};
for (const tag of [...tags, ...special]) {
  if (entries.has(keyOf(tag.parsed, tag.wildcard))) addResolution(tag.text);
}
const sample = mulberry32(0x5a3e1e);
for (let n = 0; n < 1500; n++) {
  const tag = tags[Math.floor(sample() * tags.length)];
  addResolution(
    n % 3
      ? tag.text
      : tag.text +
          suffixOf(extensions[Math.floor(sample() * extensions.length)]),
  );
}

for (const text of fuzzSample) addResolution(text);

const goldenFile = resolve(outDir, 'testdata/locale_collation_golden.json');
await writeFile(
  goldenFile,
  `[\n${golden.map((entry) => JSON.stringify(entry)).join(',\n')}\n]\n`,
);
const resolutionFile = resolve(
  outDir,
  'testdata/locale_collation_resolution.json',
);
await writeFile(
  resolutionFile,
  `[\n${[...resolution]
    .sort(([a], [b]) => (a < b ? -1 : 1))
    .map(([locale, { table, upperFirst, numeric }]) =>
      JSON.stringify({ locale, table, upperFirst, numeric }),
    )
    .join(',\n')}\n]\n`,
);
execFileSync('pnpm', ['exec', 'rs', 'fmt', goldenFile, resolutionFile], {
  cwd: resolve(import.meta.dirname, '..'),
  stdio: 'ignore',
});

console.log(
  `tables: ${tables.length}, entries: ${entries.size}, enumerated tags: ${tags.length}, resolved locales: ${resolvedSet.length}, languages: ${supportedLanguages.size}, extension tags checked: ${fuzzChecked}`,
);
if (skipped.length) console.log(`skipped:\n  ${skipped.join('\n  ')}`);
