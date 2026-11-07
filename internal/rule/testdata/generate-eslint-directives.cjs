// Regenerates the complete pinned ESLint directive-engine suite without installing
// dependencies. See eslint-directives/README.md for provenance and adaptation.
const fs = require('node:fs'),
  assert = require('node:assert/strict'),
  Module = require('node:module'),
  path = require('node:path'),
  crypto = require('node:crypto');
const root = process.argv[2];
if (!root)
  throw Error(
    'usage: node generate-eslint-directives.cjs <downloaded-source-directory>',
  );
const revision = 'bdd4b1da21dca81d50933edef0a015216059d877';
const inputs = {
  engine: 'lib/linter/apply-disable-directives.js',
  tests: 'tests/lib/linter/apply-disable-directives.js',
  config: 'lib/config/config.js',
  language: 'lib/languages/js/index.js',
  fixer: 'lib/linter/source-code-fixer.js',
};
const sources = Object.fromEntries(
  Object.entries(inputs).map(([key, file]) => [
    key,
    fs.readFileSync(path.join(root, file), 'utf8'),
  ]),
);
assert.match(sources.language, /lineStart: 1,/);
assert.match(sources.language, /columnStart: 0,/);

const reference = path.resolve(process.argv[3] || '');
assert.equal(require(path.join(reference, 'package.json')).version, '10.14.0');
for (const key of ['engine', 'config', 'language', 'fixer'])
  assert.equal(
    fs.readFileSync(path.join(reference, inputs[key]), 'utf8'),
    sources[key],
  );
const { Config } = require(path.join(reference, inputs.config));
const apply = require(path.join(reference, inputs.engine)),
  stack = [],
  cases = [];
let testCount = 0;
const fixer = require(path.join(reference, inputs.fixer));
let sourceLine = 0;
const clone = (x) => JSON.parse(JSON.stringify(x));
function serialize(options) {
  const parents = [];
  const directives = (options.directives ?? []).map((d) => {
    let parent = -1;
    if (d.parentDirective) {
      parent = parents.indexOf(d.parentDirective);
      if (parent < 0) {
        parent = parents.length;
        parents.push(d.parentDirective);
      }
    }
    return { ...d, parentDirective: undefined, parent };
  });
  return clone({
    ...options,
    sourceCode: undefined,
    ruleFilter: undefined,
    directives,
    parents,
    text: options.sourceCode?.text,
    ignored: Object.keys(options.configuredRules ?? {}).filter(
      (ruleId) =>
        Config.getRuleNumericSeverity(options.configuredRules[ruleId]) !== 0 &&
        options.ruleFilter &&
        !options.ruleFilter({
          ruleId,
          severity: Config.getRuleNumericSeverity(
            options.configuredRules[ruleId],
          ),
        }),
    ),
  });
}
const describe = (name, fn) => {
  stack.push(name);
  try {
    fn();
  } finally {
    stack.pop();
  }
};
const it = (name, fn) => {
  testCount++;
  stack.push(name);
  const frame = new Error().stack
    .split('\n')
    .find((line) => line.includes('eslint-upstream-tests.js:'));
  assert(frame, 'missing upstream test location');
  sourceLine = Number(frame.match(/eslint-upstream-tests.js:(\d+):/)[1]) - 2;
  try {
    fn();
  } finally {
    stack.pop();
  }
};
const wrapped = (options) => {
  const entry = {
    name: stack.join(' > '),
    sourceLine,
    input: serialize(options),
  };
  cases.push(entry);
  try {
    const out = apply(options);
    entry.output = clone(out);
    entry.problemResults = options.problems.map((p) => ({
      suppressed: !!p.suppressions,
    }));
    return out;
  } catch (e) {
    entry.error = { name: e.name, message: e.message };
    throw e;
  }
};
const tests = sources.tests;
new Function(
  'require',
  'describe',
  'it',
  tests + '\n//# sourceURL=eslint-upstream-tests.js',
)(
  (id) => {
    if (id === 'chai')
      return {
        assert: {
          deepStrictEqual: assert.deepStrictEqual,
          throws: (fn, message) => assert.throws(fn, { message }),
        },
      };
    if (id.endsWith('apply-disable-directives')) return wrapped;
    if (id.endsWith('languages/js')) return { lineStart: 1, columnStart: 0 };
    throw Error(id);
  },
  describe,
  it,
);
assert.equal(testCount, 89, 'upstream test inventory changed');
assert.equal(
  cases.length,
  89,
  'every upstream test must invoke the engine exactly once',
);
function normalize(entry) {
  const input = entry.input,
    groups = [],
    byKey = new Map();
  for (const [index, d] of input.directives.entries()) {
    const p = input.parents[d.parent];
    const key = p ? JSON.stringify(p) + d.type : `anonymous:${index}`;
    let g = byKey.get(key);
    if (!g) {
      g = {
        type: d.type,
        line: d.line,
        column: d.column,
        parent: p,
        entries: [],
      };
      groups.push(g);
      byKey.set(key, g);
    }
    g.entries.push(d);
  }
  for (const g of groups) {
    const p = g.parent;
    const names = g.entries.map((d) => d.ruleId).filter((n) => n != null);
    let list = p?.value;
    if (!list || /^\s*eslint-/.test(list)) list = names.join(', ');
    let comment;
    if (input.text && list.trim()) {
      const offset = input.text.indexOf(list.trim(), p?.node.range[0] ?? 0);
      if (offset >= 0)
        for (const m of input.text.matchAll(/\/\*[\s\S]*?\*\/|\/\/[^\r\n]*/g)) {
          if (
            m.index <= offset &&
            offset < m.index + m[0].length &&
            m[0].includes('eslint-' + g.type)
          )
            comment = m[0];
        }
    }
    if (comment) {
      if (comment.startsWith('//')) comment = '/*' + comment.slice(2) + ' */';
    } else comment = `/* eslint-${g.type}${list ? ' ' + list : ''} */`;
    assert(!/[\r\n]/.test(comment));
    g.comment = comment;
  }
  const problems = input.problems.map((p, index) => ({ ...p, index }));
  const maxLine = Math.max(
    1,
    ...groups.map((g) => g.line + 2),
    ...problems.map((p) => p.line),
  );
  let text = '';
  const directives = [];
  for (let line = 1; line <= maxLine; line++) {
    const points = [
      ...new Set([
        ...groups.filter((g) => g.line === line).map((g) => g.column),
        ...problems.filter((p) => p.line === line).map((p) => p.column),
      ]),
    ].sort((a, b) => a - b);
    let lineText = '';
    for (const column of points) {
      for (const g of groups.filter(
        (g) => g.line === line && g.column === column,
      )) {
        const start = text.length + lineText.length;
        const col = lineText.length + 1;
        const parentDirective = {
          node: {
            range: [start, start + g.comment.length],
            loc: {
              start: { line, column: col - 1 },
              end: { line, column: col - 1 + g.comment.length },
            },
          },
          value: g.parent?.value,
          ruleIds:
            g.parent?.ruleIds ??
            g.entries.map((d) => d.ruleId).filter((n) => n != null),
        };
        // value contains only the rules list in production, even when a unit mock
        // used a whole directive string. It matters only for partial removals.
        const content = g.comment.slice(2, -2).trim();
        parentDirective.value = content
          .slice(('eslint-' + g.type).length)
          .split(/\s--+\s/u)[0]
          .trim();
        for (const d of g.entries)
          directives.push({
            ...d,
            ruleId: d.ruleId ?? null,
            line,
            column: col,
            parentDirective,
          });
        lineText += g.comment + ' ';
      }
      for (const p of problems.filter(
        (p) => p.line === line && p.column === column,
      )) {
        p.column = lineText.length + 1;
        p.pos = text.length + lineText.length;
      }
      lineText += ' ';
    }
    text += lineText + '\n';
  }
  const options = {
    ...input,
    directives,
    problems: problems.map(({ index, pos, ...p }) => p),
    sourceCode: { text, getRange: (n) => n.range, getLoc: (n) => n.loc },
  };
  const out = apply(options);
  const suppressed = options.problems.map((p) => !!p.suppressions);
  assert.deepStrictEqual(
    suppressed,
    entry.problemResults.map((p) => p.suppressed),
    entry.name + ' suppression changed during source layout',
  );
  const ordinary = new Set(options.problems);
  const unused = out.filter((p) => !ordinary.has(p));
  const originalUnused = entry.output.filter(
    (p) =>
      typeof p.message === 'string' && p.message.startsWith('Unused eslint-'),
  );
  const messageKeys = (messages) =>
    messages
      .map((m) => [m.message, m.severity])
      .sort((a, b) => JSON.stringify(a).localeCompare(JSON.stringify(b)));
  assert.deepStrictEqual(
    messageKeys(unused),
    messageKeys(originalUnused),
    entry.name + ' attribution changed during source layout',
  );
  const offsets = [
    ...new Set(
      directives.map((d) => {
        const start = d.parentDirective.node.range[0];
        const match = /^(?:\/\*|\/\/)(\s*)eslint-/u.exec(text.slice(start));
        assert(match, 'missing generated directive label');
        return start + 2 + match[1].length;
      }),
    ),
  ].sort((a, b) => b - a);
  let rslintSource = text;
  for (const offset of offsets)
    rslintSource =
      rslintSource.slice(0, offset) + 'rslint' + rslintSource.slice(offset + 6);
  const variants = [];
  for (const severity of ['off', 'warn', 'error'])
    for (const disableFixes of [false, true]) {
      const variantOptions = {
        ...options,
        problems: problems.map(({ index, pos, ...p }) => p),
        reportUnusedDisableDirectives: severity,
        disableFixes,
      };
      const result = apply(variantOptions);
      const ordinary = new Set(variantOptions.problems);
      const diagnostics = result.filter((p) => !ordinary.has(p));
      assert.deepStrictEqual(
        variantOptions.problems.map((p) => !!p.suppressions),
        suppressed,
        'severity/fix demand changed suppression',
      );
      variants.push({
        severity,
        disableFixes,
        diagnostics,
        fixed: fixer.applyFixes(text, diagnostics).output,
        rslintFixed: fixer.applyFixes(rslintSource, diagnostics).output,
      });
    }
  return {
    name: entry.name,
    sourceLine: entry.sourceLine,
    source: text,
    rslintSource,
    reports: problems.map((p, index) => ({
      ruleId: p.ruleId ?? '',
      pos: p.pos,
      suppressed: suppressed[index],
    })),
    variants,
  };
}

const normalized = cases.filter((c) => !c.error).map(normalize);
const exceptions = cases
  .filter((c) => c.error)
  .map((c) => ({
    name: c.name,
    sourceLine: c.sourceLine,
    error: c.error,
    reason:
      'Invalid internal directive type is not a source-comment input. The original TypeError assertion runs in the importer; Go parser rejection is covered by TestMatchDirective.',
  }));
assert.equal(normalized.length, 88);
assert.equal(exceptions.length, 1);
const manifest = {
  revision,
  upstreamTests: testCount,
  importedCases: normalized.length,
  variantCount: normalized.reduce((n, c) => n + c.variants.length, 0),
  sources: Object.entries(inputs).map(([key, file]) => ({
    file,
    url: `https://github.com/eslint/eslint/blob/${revision}/${file}`,
    sha256: crypto.createHash('sha256').update(sources[key]).digest('hex'),
  })),
  exceptions,
};
const directory = path.join(__dirname, 'eslint-directives');
fs.writeFileSync(
  path.join(directory, 'upstream.json'),
  JSON.stringify({ manifest, cases: normalized }, null, 2) + '\n',
);
fs.writeFileSync(
  path.join(directory, 'original.json'),
  JSON.stringify(cases, null, 2) + '\n',
);
console.log(
  JSON.stringify({
    upstreamTests: testCount,
    importedCases: normalized.length,
    variants: manifest.variantCount,
    exceptions: exceptions.length,
  }),
);
