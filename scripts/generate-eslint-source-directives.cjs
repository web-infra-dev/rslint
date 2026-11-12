const fs = require('node:fs'),
  path = require('node:path'),
  assert = require('node:assert/strict'),
  Module = require('node:module');
const upstream = path.resolve(process.argv[2]);
const ref = path.resolve(process.argv[3]);
assert.equal(require(ref + '/package.json').version, '10.14.0');
const root = path.dirname(require.resolve(ref + '/package.json'));
const req = Module.createRequire(root + '/package.json');
for (const file of [
  'lib/linter/apply-disable-directives.js',
  'lib/config/config.js',
  'lib/linter/source-code-fixer.js',
  'lib/languages/js/index.js',
])
  assert.equal(
    fs.readFileSync(path.join(root, file), 'utf8'),
    fs.readFileSync(path.join(upstream, file), 'utf8'),
  );
const espree = req('espree');
const code = fs.readFileSync(
  path.join(upstream, 'tests/lib/linter/linter.js'),
  'utf8',
);
const ast = espree.parse(code, {
  ecmaVersion: 'latest',
  sourceType: 'script',
  range: true,
  loc: true,
});
const wanted = new Set([
  'ruleFilter',
  '/*eslint-disable*/ and /*eslint-enable*/',
  '/*eslint-disable-line*/',
  '/*eslint-disable-next-line*/',
  'descriptions in directive comments',
  'reportUnusedDisableDirectives option',
  'getSuppressedMessages()',
  'when evaluating code with comments to disable and enable configurable rule as part of plugin',
  'when evaluating code with comments to enable and disable all reporting',
]);
const roots = [];
function visit(n) {
  if (!n || typeof n !== 'object') return;
  if (
    n.type === 'CallExpression' &&
    n.callee.name === 'describe' &&
    wanted.has(n.arguments[0]?.value)
  ) {
    roots.push(n);
    return;
  }
  for (const [key, v] of Object.entries(n)) {
    if (key === 'loc' || key === 'range') continue;
    if (Array.isArray(v)) v.forEach(visit);
    else if (v && typeof v === 'object') visit(v);
  }
}
visit(ast);
assert.equal(roots.length, wanted.size);
const stack = [],
  tests = [],
  skipped = [],
  calls = [];
let reset, active;
const describe = (name, fn) => {
  stack.push({ name, hooks: [] });
  try {
    fn();
  } finally {
    stack.pop();
  }
};
const beforeEach = (fn) => stack.at(-1).hooks.push(fn);
const it = (name, fn) => {
  const frame = new Error().stack
    .split('\n')
    .find((l) => l.includes('eslint-upstream-linter.js:'));
  const sourceLine =
    Number(frame.match(/eslint-upstream-linter.js:(\d+):/)[1]) - 3;
  const test = {
    name: stack
      .map((s) => s.name)
      .concat(name)
      .join(' > '),
    sourceLine,
    scopes: [...stack],
    reset,
    fn,
  };
  if (
    stack[0].name === 'descriptions in directive comments' &&
    !/eslint-(?:disable|enable)/.test(fn.toString())
  ) {
    skipped.push({
      name: test.name,
      sourceLine,
      reason:
        'Descriptions on globals/exported/inline-rule configuration are outside disable-directive scope.',
    });
    return;
  }
  tests.push(test);
};
const applyPath = root + '/lib/linter/apply-disable-directives.js';
const apply = require(applyPath);
const { Config } = require(root + '/lib/config/config.js');
require.cache[applyPath].exports = (options) => {
  const ignored = Object.keys(options.configuredRules ?? {}).filter(
    (ruleId) => {
      const severity = Config.getRuleNumericSeverity(
        options.configuredRules[ruleId],
      );
      return (
        severity !== 0 &&
        options.ruleFilter &&
        !options.ruleFilter({ ruleId, severity })
      );
    },
  );
  const before = options.problems.map((p) => ({
    ...JSON.parse(JSON.stringify(p)),
    fixturePos: options.sourceCode.getIndexFromLoc({
      line: p.line,
      column: p.column - 1,
    }),
  }));
  const result = apply(options);
  const ordinary = new Set(options.problems);
  let rslintSource = options.sourceCode.text;
  const offsets = options.sourceCode
    .getAllComments()
    .filter((c) => c.type === 'Block' || c.type === 'Line')
    .flatMap((c) => {
      const m =
        /^(\s*)eslint-(?:disable(?:-next-line|-line)?|enable)(?=\s|$)/u.exec(
          c.value,
        );
      return m ? [c.range[0] + 2 + m[1].length] : [];
    })
    .sort((a, b) => b - a);
  for (const offset of offsets)
    rslintSource =
      rslintSource.slice(0, offset) + 'rslint' + rslintSource.slice(offset + 6);
  const visible = result.filter((p) => !p.suppressions);
  const bom = options.sourceCode.hasBOM ? '\ufeff' : '';
  calls.push({
    rslintSource,
    fixed: fixer.applyFixes(bom + options.sourceCode.text, visible).output,
    rslintFixed: fixer.applyFixes(bom + rslintSource, visible).output,
    test: active,
    source: options.sourceCode.text,
    bom: options.sourceCode.hasBOM,
    problems: before,
    suppressed: options.problems.map((p) => !!p.suppressions),
    ignored,
    severity: options.reportUnusedDisableDirectives ?? 'off',
    disableFixes: options.disableFixes ?? false,
    diagnostics: result.filter((p) => !ordinary.has(p)),
  });
  return result;
};
const fixer = require(root + '/lib/linter/source-code-fixer.js');
const { Linter } = require(ref);
const assertions = {
  ...assert,
  isTrue: (x) => assert.strictEqual(x, true),
  isFalse: (x) => assert.strictEqual(x, false),
  isNull: (x) => assert.strictEqual(x, null),
  isUndefined: (x) => assert.strictEqual(x, undefined),
  throws: (fn, expected) =>
    assert.throws(
      fn,
      typeof expected === 'string'
        ? (e) => e.message.includes(expected)
        : expected,
    ),
};
for (const n of roots) {
  new Function(
    'Linter',
    'assert',
    'espree',
    'describe',
    'it',
    'beforeEach',
    'setReset',
    `let linter; const filename='filename.js'; setReset(()=>{linter=new Linter()});\n${'\n'.repeat(n.loc.start.line - 1)}${code.slice(...n.range)};\n//# sourceURL=eslint-upstream-linter.js`,
  )(Linter, assertions, espree, describe, it, beforeEach, (fn) => (reset = fn));
}
for (const [index, t] of tests.entries()) {
  active = { id: index, name: t.name, sourceLine: t.sourceLine };
  t.reset();
  for (const s of t.scopes) for (const hook of s.hooks) hook();
  try {
    t.fn();
  } catch (e) {
    console.error(active);
    throw e;
  }
}
assert.equal(tests.length, 486);
assert.equal(skipped.length, 6);
const noCalls = tests
  .map((t, id) => ({ id, name: t.name, sourceLine: t.sourceLine }))
  .filter((t) => !calls.some((c) => c.test.id === t.id));
assert.equal(noCalls.length, 3);
// rslint can batch adjacent deletions that ESLint applies over multiple rounds.
// Accept the faster output only when a later call of the SAME upstream test
// independently produced that exact output. Individual fix ranges stay strict.
for (const [index, c] of calls.entries()) {
  const fixes = [
    ...c.diagnostics.map((d) => d.fix),
    ...c.problems.filter((p, i) => !c.suppressed[i]).map((p) => p.fix),
  ]
    .filter(Boolean)
    .sort((a, b) => a.range[0] - b.range[0] || a.range[1] - b.range[1]);
  if (!fixes.some((f, i) => i && fixes[i - 1].range[1] === f.range[0]))
    continue;
  if (
    fixes.some(
      (f, i) =>
        f.range[0] < 0 ||
        f.range[0] === f.range[1] ||
        (i && fixes[i - 1].range[1] > f.range[0]),
    )
  )
    continue;
  const batch = (source) => {
    for (const f of [...fixes].reverse())
      source = source.slice(0, f.range[0]) + f.text + source.slice(f.range[1]);
    return (c.bom ? '\ufeff' : '') + source;
  };
  const fixed = batch(c.source),
    rslintFixed = batch(c.rslintSource);
  if (fixed === c.fixed) continue;
  assert(
    calls
      .slice(index + 1)
      .some(
        (later) =>
          later.test.id === c.test.id &&
          later.fixed === fixed &&
          later.rslintFixed === rslintFixed,
      ),
    'batched output must match a later ESLint round: ' + c.test.name,
  );
  c.batchedFixed = fixed;
  c.rslintBatchedFixed = rslintFixed;
}
const unique = new Map();
const origins = [];
for (const c of calls) {
  const reports = c.problems.map((p, i) => ({
    ruleId: p.ruleId ?? '',
    pos: p.fixturePos,
    suppressed: c.suppressed[i],
    fix: p.fix,
  }));
  const key = JSON.stringify([
    c.source,
    reports,
    c.ignored,
    c.severity,
    c.disableFixes,
    c.bom,
  ]);
  if (!unique.has(key))
    unique.set(key, {
      name: c.test.name,
      sourceLine: c.test.sourceLine,
      source: c.source,
      rslintSource: c.rslintSource,
      bom: c.bom,
      reports,
      invalidRules: c.problems
        .filter((p) =>
          /^Definition for rule .* was not found\.$/.test(p.message),
        )
        .map((p) => p.ruleId),
      ignored: c.ignored,
      variants: [
        {
          severity: c.severity,
          disableFixes: c.disableFixes,
          diagnostics: c.diagnostics,
          fixed: c.fixed,
          rslintFixed: c.rslintFixed,
          batchedFixed: c.batchedFixed,
          rslintBatchedFixed: c.rslintBatchedFixed,
        },
      ],
    });
  origins.push({ test: c.test, caseIndex: [...unique.keys()].indexOf(key) });
}
const corpus = {
  manifest: {
    revision: 'bdd4b1da21dca81d50933edef0a015216059d877',
    sourceSHA256: require('node:crypto')
      .createHash('sha256')
      .update(code)
      .digest('hex'),
    upstreamTests: tests.length,
    importedCases: unique.size,
    variantCount: unique.size,
    exceptions: skipped,
    noEngineCalls: noCalls,
    origins,
  },
  cases: [...unique.values()],
};
fs.writeFileSync(
  path.join(
    __dirname,
    '../internal/rule/testdata/eslint-directives/source.json',
  ),
  JSON.stringify(corpus, null, 2) + '\n',
);
console.log('source corpus:', unique.size);
