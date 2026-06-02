import { describe, test, expect } from '@rstest/core';

import { buildDiagnostic } from '../../../src/eslint-plugin/linter/diagnostic-builder.js';
import { makeFixer } from '../../../src/eslint-plugin/linter/fixer.js';
import { lintFile } from '../../../src/eslint-plugin/linter/ecma-language-plugin.js';
import type { LoadedPlugins } from '../../../src/eslint-plugin/plugin/plugin-loader.js';

/**
 * `context.report()` argument validation parity with ESLint v10's
 * `computeMessageFromDescriptor` (lib/linter/file-report.js — strings
 * copied verbatim from the installed eslint@9.32.0, which is identical
 * to v10 for this function).
 *
 * Two divergences the runner used to have:
 *   1. message + messageId BOTH present → runner silently kept
 *      `message`; ESLint THROWS.
 *   2. messageId NOT in the `messages` map → runner fabricated a
 *      `(${messageId})` placeholder; ESLint THROWS.
 *
 * Blast-radius contract: these throws originate inside
 * `context.report()`, which runs inside the per-rule listener (or
 * `create()`) try/catch in `ecma-language-plugin.ts` /
 * `listener-merge.ts`. So a buggy rule's `report()` misuse must land as
 * a `ruleErrors` entry for that ONE rule — never crash the file or a
 * sibling rule. The last test exercises that end-to-end.
 */

const BOTH_MSG =
  'context.report() called with a message and a messageId. Please only pass one.';

// Verbatim ESLint strings (eslint@9.32.0 lib/linter/file-report.js,
// identical to v10). Each constant is annotated with the line it was
// copied from so the assertions below stay exact, not loose matches.
//
//   assertValidNodeInfo else-branch (file-report.js:175)
const NO_LOC_MSG =
  'Node must be provided when reporting error if location is not provided';
//   computeMessageFromDescriptor final throw (file-report.js:470)
const MISSING_MESSAGE_MSG =
  'Missing `message` property in report() call; add a message that describes the linting problem.';
//   validateSuggestions both-messageId-and-desc (file-report.js:418)
const SUGGEST_BOTH_MSG =
  "context.report() called with a suggest option that defines both a 'messageId' and an 'desc'. Please only pass one.";
//   validateSuggestions neither-desc-nor-messageId (file-report.js:423)
const SUGGEST_NEITHER_MSG =
  "context.report() called with a suggest option that doesn't have either a `desc` or `messageId`";

function unknownIdMsg(id: string, messages: Record<string, string>): string {
  // Reconstructed with ESLint's exact template so the assertion stays
  // exact (not a loose substring match) while remaining robust to the
  // `messages` object's serialization.
  return `context.report() called with a messageId of '${id}' which is not present in the 'messages' config: ${JSON.stringify(
    messages,
    null,
    2,
  )}`;
}

//   validateSuggestions unknown-suggestion-messageId (file-report.js:412)
function unknownSuggestIdMsg(
  id: string,
  messages: Record<string, string>,
): string {
  return `context.report() called with a suggest option with a messageId '${id}' which is not present in the 'messages' config: ${JSON.stringify(
    messages,
    null,
    2,
  )}`;
}

// A node carrying a `.range` so position computation in buildDiagnostic
// succeeds — this isolates the message-validation branch under test
// from the loc/node-range branches.
const NODE = { type: 'Identifier', range: [0, 3] as [number, number] } as never;

function build(descriptor: unknown, messages: Record<string, string>) {
  return buildDiagnostic({
    ruleName: 'test/rule',
    // descriptor is an `unknown` test input projected onto the typed
    // ReportDescriptor surface; buildDiagnostic validates at runtime.
    descriptor: descriptor as never,
    text: 'foo',
    messages,
    fixer: makeFixer(),
    collectFixes: false,
    suggestionsMode: 'off',
  });
}

describe('buildDiagnostic — report() validation matches ESLint v10', () => {
  test('message + messageId BOTH present → throws ESLint string', () => {
    const messages = { error: 'Real message.' };
    expect(() =>
      build({ node: NODE, message: 'inline', messageId: 'error' }, messages),
    ).toThrow(BOTH_MSG);
    // Exact TypeError type + exact string.
    try {
      build({ node: NODE, message: 'inline', messageId: 'error' }, messages);
      throw new Error('expected build to throw');
    } catch (err) {
      expect(err).toBeInstanceOf(TypeError);
      expect((err as Error).message).toBe(BOTH_MSG);
    }
  });

  test('messageId not in messages map → throws ESLint string', () => {
    const messages = { error: 'Real message.' };
    const expected = unknownIdMsg('doesNotExist', messages);
    try {
      build({ node: NODE, messageId: 'doesNotExist' }, messages);
      throw new Error('expected build to throw');
    } catch (err) {
      expect(err).toBeInstanceOf(TypeError);
      expect((err as Error).message).toBe(expected);
    }
  });

  test('messageId not present in an EMPTY messages map → throws ESLint string', () => {
    const messages = {};
    const expected = unknownIdMsg('error', messages);
    try {
      build({ node: NODE, messageId: 'error' }, messages);
      throw new Error('expected build to throw');
    } catch (err) {
      expect(err).toBeInstanceOf(TypeError);
      expect((err as Error).message).toBe(expected);
    }
  });

  test('valid messageId → resolves to the mapped message (no throw)', () => {
    const messages = { error: 'Do not use `null`.' };
    const diag = build({ node: NODE, messageId: 'error' }, messages);
    expect(diag).not.toBeNull();
    expect(diag!.message).toBe('Do not use `null`.');
    expect(diag!.messageId).toBe('error');
  });

  test('valid messageId + data → interpolates the mapped template', () => {
    const messages = { named: "'{{ name }}' is banned." };
    const diag = build(
      { node: NODE, messageId: 'named', data: { name: 'foo' } },
      messages,
    );
    expect(diag).not.toBeNull();
    expect(diag!.message).toBe("'foo' is banned.");
  });

  test('plain message (no messageId) still works — no regression', () => {
    const diag = build({ node: NODE, message: 'plain message' }, {});
    expect(diag).not.toBeNull();
    expect(diag!.message).toBe('plain message');
  });
});

describe('#3 missing node-info throws ESLint string (was silently dropped)', () => {
  // ESLint's `assertValidNodeInfo` (file-report.js:169-178) requires
  // either a usable `node` location or a `loc`. The runner used to
  // `return null` for these, so `context.report()` produced no
  // diagnostic AND no ruleError — the report vanished silently.

  test('no node and no loc → throws NO_LOC_MSG', () => {
    try {
      build({ message: 'orphan report' }, {});
      throw new Error('expected build to throw');
    } catch (err) {
      expect(err).toBeInstanceOf(TypeError);
      expect((err as Error).message).toBe(NO_LOC_MSG);
    }
  });

  test('node with neither .range nor .loc → throws NO_LOC_MSG', () => {
    // A node object that carries no derivable position. ESLint reaches
    // the same dead-end (getLoc returns undefined); the runner surfaces
    // the validation message instead of dropping the report.
    const noLocNode = { type: 'Identifier' } as never;
    try {
      build({ node: noLocNode, message: 'no position' }, {});
      throw new Error('expected build to throw');
    } catch (err) {
      expect(err).toBeInstanceOf(TypeError);
      expect((err as Error).message).toBe(NO_LOC_MSG);
    }
  });

  test('node WITH .range still works — no regression', () => {
    const diag = build({ node: NODE, message: 'ok' }, {});
    expect(diag).not.toBeNull();
    expect(diag!.startPos).toBe(0);
    expect(diag!.endPos).toBe(3);
  });
});

describe('#4 missing message throws ESLint string (was silently dropped)', () => {
  // ESLint's `computeMessageFromDescriptor` (file-report.js:469-471)
  // throws when neither `message` nor `messageId` is present. The runner
  // used to `return null`, dropping the report with no diagnostic/error.
  test('node present but no message and no messageId → throws MISSING_MESSAGE_MSG', () => {
    try {
      build({ node: NODE }, {});
      throw new Error('expected build to throw');
    } catch (err) {
      expect(err).toBeInstanceOf(TypeError);
      expect((err as Error).message).toBe(MISSING_MESSAGE_MSG);
    }
  });
});

describe('#2 suggest[] validation matches ESLint validateSuggestions', () => {
  // ESLint's `validateSuggestions` (file-report.js:398-434) runs before
  // any suggestion is built. The runner used to do zero validation and
  // fabricate `messages[messageId] ?? `(${messageId})`` for an unknown
  // id — shipping a placeholder suggestion that hid the mistake.
  const messages = { real: 'Real suggestion.' };

  test('(a) suggestion messageId not in messages → throws unknownSuggestIdMsg', () => {
    const expected = unknownSuggestIdMsg('ghost', messages);
    try {
      build(
        {
          node: NODE,
          message: 'm',
          suggest: [{ messageId: 'ghost', fix: () => null }],
        },
        messages,
      );
      throw new Error('expected build to throw');
    } catch (err) {
      expect(err).toBeInstanceOf(TypeError);
      expect((err as Error).message).toBe(expected);
    }
  });

  test('(b) suggestion with BOTH desc and messageId → throws SUGGEST_BOTH_MSG', () => {
    try {
      build(
        {
          node: NODE,
          message: 'm',
          suggest: [
            { messageId: 'real', desc: 'also a desc', fix: () => null },
          ],
        },
        messages,
      );
      throw new Error('expected build to throw');
    } catch (err) {
      expect(err).toBeInstanceOf(TypeError);
      expect((err as Error).message).toBe(SUGGEST_BOTH_MSG);
    }
  });

  test('(c) suggestion with NEITHER desc nor messageId → throws SUGGEST_NEITHER_MSG', () => {
    try {
      build(
        {
          node: NODE,
          message: 'm',
          suggest: [{ fix: () => null }],
        },
        messages,
      );
      throw new Error('expected build to throw');
    } catch (err) {
      expect(err).toBeInstanceOf(TypeError);
      expect((err as Error).message).toBe(SUGGEST_NEITHER_MSG);
    }
  });

  test('valid suggestion (known messageId, no desc) → no throw', () => {
    const diag = build(
      {
        node: NODE,
        message: 'm',
        suggest: [{ messageId: 'real', fix: () => null }],
      },
      messages,
    );
    expect(diag).not.toBeNull();
    expect(diag!.suggestions).toHaveLength(1);
    expect(diag!.suggestions![0].desc).toBe('Real suggestion.');
  });

  test('valid suggestion (desc only, no messageId) → no throw', () => {
    const diag = build(
      {
        node: NODE,
        message: 'm',
        suggest: [{ desc: 'inline desc', fix: () => null }],
      },
      {},
    );
    expect(diag).not.toBeNull();
    expect(diag!.suggestions).toHaveLength(1);
    expect(diag!.suggestions![0].desc).toBe('inline desc');
  });
});

describe('#10 interpolate uses `in` (prototype chain), matching ESLint v10', () => {
  // ESLint v10 `lib/linter/interpolate.js:38` resolves a placeholder via
  // `term in data` — the prototype chain is included. The runner
  // previously used `Object.prototype.hasOwnProperty.call(data, key)`,
  // which would leave an inherited-key placeholder un-interpolated.
  test('own-key placeholder still interpolates normally — no regression', () => {
    const diag = build(
      { node: NODE, message: 'value is {{x}}', data: { x: 42 } as never },
      {},
    );
    expect(diag).not.toBeNull();
    expect(diag!.message).toBe('value is 42');
  });

  test('inherited-key placeholder resolves via `in` (v10 behavior)', () => {
    // `inherited` lives on the prototype, NOT as an own property.
    // hasOwnProperty would skip it (leaving "{{inherited}}"); `in`
    // resolves it. This pins the v10 deviation the fix introduced.
    const proto = { inherited: 'from-proto' };
    const data = Object.create(proto) as Record<string, string | number>;
    const diag = build(
      { node: NODE, message: 'got {{inherited}}', data: data as never },
      {},
    );
    expect(diag).not.toBeNull();
    expect(diag!.message).toBe('got from-proto');
  });
});

describe('report() misuse lands as a ruleError, not a file/sibling crash', () => {
  // The throw inside `context.report()` propagates out of the LISTENER
  // (`report` is called from a listener fn), and the listener-merge
  // visit wraps each listener in try/catch → `onListenerError` →
  // `result.ruleErrors`. So the misusing rule is isolated and a healthy
  // sibling rule on the SAME file still fires. This is the runner's
  // analogue of ESLint isolating a thrown rule.
  test('both message+messageId misuse → ruleErrors entry; sibling rule still fires', () => {
    interface RuleErr {
      rule: string;
      message: string;
    }
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/bad-report',
          {
            meta: { name: 'bad-report', messages: { real: 'Real msg.' } },
            create(ctx: { report: (d: unknown) => void }) {
              return {
                Identifier(node: { name: string }) {
                  if (node.name === 'flag') {
                    // Misuse: both message AND messageId → ESLint throws.
                    ctx.report({
                      node,
                      message: 'inline',
                      messageId: 'real',
                    });
                  }
                },
              };
            },
          },
        ],
        [
          'stub/healthy',
          {
            meta: { name: 'healthy' },
            create(ctx: { report: (d: unknown) => void }) {
              return {
                Identifier(node: { name: string }) {
                  if (node.name === 'flag') {
                    ctx.report({ node, message: 'healthy fired' });
                  }
                },
              };
            },
          },
        ],
      ]),
    };

    const result = lintFile(
      {
        filePath: 'misuse.ts',
        text: 'const flag = 1;',
        rules: {
          'stub/bad-report': { options: [] },
          'stub/healthy': { options: [] },
        },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    // The misuse must NOT crash the file: no parseError.
    expect(result.parseError).toBeUndefined();

    // The throw is attributed to the misusing rule via ruleErrors and
    // carries ESLint's exact validation message.
    const errs = (result.ruleErrors ?? []) as RuleErr[];
    const badEntry = errs.find((e) => e.rule === 'stub/bad-report');
    expect(badEntry).toBeDefined();
    expect(badEntry!.message).toContain(BOTH_MSG);

    // The misusing rule produced NO diagnostic (its report() threw
    // before a diagnostic was recorded).
    expect(
      result.diagnostics.some((d) => d.ruleName === 'stub/bad-report'),
    ).toBe(false);

    // The healthy sibling rule still ran and reported on the same file.
    const healthyDiag = result.diagnostics.find(
      (d) => d.ruleName === 'stub/healthy',
    );
    expect(healthyDiag).toBeDefined();
    expect(healthyDiag!.message).toBe('healthy fired');
  });

  test('unknown-messageId misuse → ruleErrors entry; sibling rule still fires', () => {
    interface RuleErr {
      rule: string;
      message: string;
    }
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/bad-id',
          {
            meta: { name: 'bad-id', messages: { real: 'Real msg.' } },
            create(ctx: { report: (d: unknown) => void }) {
              return {
                Identifier(node: { name: string }) {
                  if (node.name === 'flag') {
                    // Misuse: messageId not in the messages map.
                    ctx.report({ node, messageId: 'nope' });
                  }
                },
              };
            },
          },
        ],
        [
          'stub/healthy',
          {
            meta: { name: 'healthy' },
            create(ctx: { report: (d: unknown) => void }) {
              return {
                Identifier(node: { name: string }) {
                  if (node.name === 'flag') {
                    ctx.report({ node, message: 'healthy fired' });
                  }
                },
              };
            },
          },
        ],
      ]),
    };

    const result = lintFile(
      {
        filePath: 'misuse2.ts',
        text: 'const flag = 1;',
        rules: {
          'stub/bad-id': { options: [] },
          'stub/healthy': { options: [] },
        },
        collectFixes: false,
        suggestionsMode: 'off',
      },
      loaded,
    );

    expect(result.parseError).toBeUndefined();
    const errs = (result.ruleErrors ?? []) as RuleErr[];
    const badEntry = errs.find((e) => e.rule === 'stub/bad-id');
    expect(badEntry).toBeDefined();
    expect(badEntry!.message).toContain(
      "called with a messageId of 'nope' which is not present in the 'messages' config",
    );

    const healthyDiag = result.diagnostics.find(
      (d) => d.ruleName === 'stub/healthy',
    );
    expect(healthyDiag).toBeDefined();
  });
});

describe('#3 invalid esquery selector is isolated to its rule (lintFile never throws)', () => {
  test('bad selector → ruleError, sibling rule still fires, no throw', () => {
    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/bad-selector',
          {
            meta: { name: 'bad-selector' },
            create() {
              // Invalid esquery selector. `mergeListeners` calls
              // `esquery.parse` on it, which throws — pre-fix that escaped
              // out of `lintFile` (violating its never-throw contract).
              // Now it's skipped + recorded as a per-rule error.
              return { 'Foo > ': () => {} };
            },
          },
        ],
        [
          'stub/healthy',
          {
            meta: { name: 'healthy' },
            create(ctx: { report: (d: unknown) => void }) {
              return {
                Identifier(node: { name: string }) {
                  if (node.name === 'flag') {
                    ctx.report({ node, message: 'healthy fired' });
                  }
                },
              };
            },
          },
        ],
      ]),
    };

    let result!: ReturnType<typeof lintFile>;
    expect(() => {
      result = lintFile(
        {
          filePath: 'sel.ts',
          text: 'const flag = 1;',
          rules: {
            'stub/bad-selector': { options: [] },
            'stub/healthy': { options: [] },
          },
          collectFixes: false,
          suggestionsMode: 'off',
        },
        loaded,
      );
    }).not.toThrow();

    // The bad selector did not crash the file.
    expect(result.parseError).toBeUndefined();

    // It's attributed to its own rule via ruleErrors.
    const errs = (result.ruleErrors ?? []) as Array<{
      rule: string;
      message: string;
    }>;
    const badEntry = errs.find((e) => e.rule === 'stub/bad-selector');
    expect(badEntry).toBeDefined();
    expect(badEntry!.message).toMatch(/invalid selector/);

    // The sibling healthy rule still ran — ESLint-like isolation.
    expect(result.diagnostics.some((d) => d.ruleName === 'stub/healthy')).toBe(
      true,
    );
  });
});

// ─────────────────────────────────────────────────────────────────────
// End-to-end blast-radius for the #3 / #4 / #2 validations: a rule whose
// `report()` hits the bad form must land as a ruleErrors entry, a healthy
// sibling on the SAME file must still fire, lintFile must NOT throw, and
// the bad rule must produce NO diagnostic. Mirrors the message+messageId
// / unknown-messageId end-to-end tests above.
// ─────────────────────────────────────────────────────────────────────

interface RuleErr {
  rule: string;
  message: string;
}

/**
 * Run a single "bad" rule alongside a healthy sibling on `const flag = 1;`.
 * `badReport` is the descriptor (or thunk) the bad rule passes to
 * `ctx.report` when it visits the `flag` Identifier. `badMeta` supplies
 * the bad rule's `meta` (its `messages` map, etc).
 */
function runBadVsHealthy(
  badRuleName: string,
  badMeta: Record<string, unknown>,
  badReport: (node: unknown) => unknown,
): ReturnType<typeof lintFile> {
  const loaded: LoadedPlugins = {
    plugins: [],
    rules: new Map<string, unknown>([
      [
        badRuleName,
        {
          meta: badMeta,
          create(ctx: { report: (d: unknown) => void }) {
            return {
              Identifier(node: { name: string }) {
                if (node.name === 'flag') {
                  ctx.report(badReport(node));
                }
              },
            };
          },
        },
      ],
      [
        'stub/healthy',
        {
          meta: { name: 'healthy' },
          create(ctx: { report: (d: unknown) => void }) {
            return {
              Identifier(node: { name: string }) {
                if (node.name === 'flag') {
                  ctx.report({ node, message: 'healthy fired' });
                }
              },
            };
          },
        },
      ],
    ]),
  };

  return lintFile(
    {
      filePath: 'e2e.ts',
      text: 'const flag = 1;',
      rules: {
        [badRuleName]: { options: [] },
        'stub/healthy': { options: [] },
      },
      collectFixes: false,
      suggestionsMode: 'off',
    },
    loaded,
  );
}

function assertIsolated(
  result: ReturnType<typeof lintFile>,
  badRuleName: string,
  expectedMessage: string,
): void {
  // No file-level crash.
  expect(result.parseError).toBeUndefined();
  // Attributed to the bad rule, with the verbatim ESLint string.
  const errs = (result.ruleErrors ?? []) as RuleErr[];
  const badEntry = errs.find((e) => e.rule === badRuleName);
  expect(badEntry).toBeDefined();
  expect(badEntry!.message).toContain(expectedMessage);
  // Bad rule produced no diagnostic.
  expect(result.diagnostics.some((d) => d.ruleName === badRuleName)).toBe(
    false,
  );
  // Healthy sibling still fired on the same file.
  const healthyDiag = result.diagnostics.find(
    (d) => d.ruleName === 'stub/healthy',
  );
  expect(healthyDiag).toBeDefined();
  expect(healthyDiag!.message).toBe('healthy fired');
}

describe('#3/#4/#2 misuse lands as a ruleError, not a file/sibling crash', () => {
  test('#3 missing node-info (loc-less node) → ruleError; sibling still fires', () => {
    let result!: ReturnType<typeof lintFile>;
    expect(() => {
      result = runBadVsHealthy(
        'stub/bad-node-info',
        { name: 'bad-node-info' },
        // Report a node with no .range and no .loc → buildDiagnostic throws.
        () => ({ node: { type: 'Identifier' }, message: 'x' }),
      );
    }).not.toThrow();
    assertIsolated(result, 'stub/bad-node-info', NO_LOC_MSG);
  });

  test('#4 missing message → ruleError; sibling still fires', () => {
    let result!: ReturnType<typeof lintFile>;
    expect(() => {
      result = runBadVsHealthy(
        'stub/bad-no-message',
        { name: 'bad-no-message' },
        // Report with a valid node but neither message nor messageId.
        (node) => ({ node }),
      );
    }).not.toThrow();
    assertIsolated(result, 'stub/bad-no-message', MISSING_MESSAGE_MSG);
  });

  test('#2 suggestion with neither desc nor messageId → ruleError; sibling still fires', () => {
    let result!: ReturnType<typeof lintFile>;
    expect(() => {
      result = runBadVsHealthy(
        'stub/bad-suggest',
        { name: 'bad-suggest', messages: { real: 'Real.' } },
        (node) => ({
          node,
          message: 'm',
          suggest: [{ fix: () => null }],
        }),
      );
    }).not.toThrow();
    assertIsolated(result, 'stub/bad-suggest', SUGGEST_NEITHER_MSG);
  });
});
