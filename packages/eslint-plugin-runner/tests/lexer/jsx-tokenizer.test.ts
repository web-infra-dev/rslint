/**
 * JSX tokenizer differential tests.
 *
 * espree is the oracle: every test runs the same source through the
 * runner's `tokenize` (with `jsx: true`) and through `espree.tokenize`,
 * then asserts the two token streams agree on type + value + range.
 *
 * Coverage targets:
 *   - basic open / close / self-closing / empty / fragment
 *   - attribute name-only / string value / expression value
 *   - JSXText with whitespace, multi-line
 *   - dotted member names (`<a.b.c>`)
 *   - namespaced names (`<svg:circle>`)
 *   - hyphenated attribute names (`data-foo`)
 *   - nested elements + expression containers
 *   - template literals inside JSX, JSX inside template `${…}`
 *   - mixed regex / division / JSX (`a < b`, `f(x) / 2`, `<a/>`)
 *   - TSX disambiguation cases (generic `<T>` vs JSX) — runner does
 *     NOT enter JSX for clear generic shapes
 *   - regression: non-JSX files unaffected
 */
import { describe, test, expect } from '@rstest/core';
import * as espree from 'espree';

import { tokenize } from '../../src/lexer/tokenizer.js';
import { buildLineStartOffsets } from '../../src/ast/normalize-ast.js';

interface SimpleToken {
  type: string;
  value: string;
  range: [number, number];
}

function runnerTokens(src: string, jsx = true): SimpleToken[] {
  const { tokens } = tokenize(src, buildLineStartOffsets(src), { jsx });
  return tokens.map((t) => ({
    type: t.type,
    value: t.value,
    range: t.range,
  }));
}

function espreeTokens(src: string, jsx = true): SimpleToken[] {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const tokens = (espree as any).tokenize(src, {
    ecmaVersion: 'latest',
    sourceType: 'module',
    ecmaFeatures: { jsx },
    range: true,
    loc: false,
  }) as Array<{ type: string; value: string; range: [number, number] }>;
  return tokens.map((t) => ({
    type: t.type,
    value: t.value,
    range: t.range,
  }));
}

/** Assert runner and espree agree on type+value+range for a given source. */
function diff(src: string, jsx = true) {
  const a = runnerTokens(src, jsx);
  const b = espreeTokens(src, jsx);
  expect(a).toEqual(b);
}

// ────────────────────────────────────────────────────────────────────
// Basic JSX shapes
// ────────────────────────────────────────────────────────────────────

describe('JSX tokenizer — basic elements', () => {
  test('empty self-closing: <a/>', () => diff('<a/>;'));
  test('self-closing with space: <a />', () => diff('<a />;'));
  test('empty open + close: <a></a>', () => diff('<a></a>;'));
  test('with text content: <a>hi</a>', () => diff('<a>hi</a>;'));
  test('fragment: <></>', () => diff('<></>;'));
  test('assigned: const e = <a/>;', () => diff('const e = <a/>;'));
  test('returned: () => <a/>', () => diff('() => <a/>;'));
});

describe('JSX tokenizer — attributes', () => {
  test('name-only attr: <a x/>', () => diff('<a x/>;'));
  test('string attr: <a x="y"/>', () => diff('<a x="y"/>;'));
  test("single-quote: <a x='y'/>", () => diff("<a x='y'/>;"));
  test('expr attr: <a x={1}/>', () => diff('<a x={1}/>;'));
  test('multiple attrs: <a x="y" z={1} w/>', () => diff('<a x="y" z={1} w/>;'));
});

// ────────────────────────────────────────────────────────────────────
// Backslash inside a JSX attribute string is an ORDINARY char — espree /
// acorn-jsx do NOT process C-style escapes there. The pre-fix scanner
// did `i += 2` on `\`, so a backslash right before the close quote
// (`value="C:\"`) consumed the close quote as an "escaped" quote and
// corrupted every following token. espree is the oracle for both cases.
// ────────────────────────────────────────────────────────────────────

describe('JSX tokenizer — backslash in attribute string (no escapes)', () => {
  // The trailing `\` sits immediately before the close quote — this is
  // the case the old `i += 2` mis-handled. Source chars (one literal
  // backslash): <input value="C:\" name="n" />;
  test('backslash before close quote: value="C:\\"', () =>
    diff('<input value="C:\\" name="n" />;'));

  // Control: backslash NOT before the close quote already matched espree
  // (both escape and ordinary interpretations end the string at the same
  // `"`). Source chars: <a b="\n" />;
  test('control — backslash mid-string: b="\\n"', () => diff('<a b="\\n" />;'));

  // Single-quoted variant of the close-quote case.
  test("backslash before single close quote: x='a\\'", () =>
    diff("<a x='a\\' y='b' />;"));

  // Unterminated attribute string ending in a stray backslash at EOF.
  // espree throws ("Unterminated string constant"), so it is not a
  // differential case; assert instead that the runner does NOT crash and
  // the final attribute-string token's range stays within EOF (the
  // Math.min(i, n) clamp). Source chars: <a b="x\
  test('EOF stray backslash does not overrun range: b="x\\', () => {
    const src = '<a b="x\\';
    const { tokens } = tokenize(src, buildLineStartOffsets(src), {
      jsx: true,
    });
    // No token may report a range end past the source length.
    for (const t of tokens) {
      expect(t.range[1]).toBeLessThanOrEqual(src.length);
    }
    // The attribute string is emitted as JSXText spanning the open quote
    // to EOF (`"x\`), not 1 byte past.
    const attrStr = tokens.find(
      (t) => t.type === 'JSXText' && t.value.startsWith('"'),
    );
    expect(attrStr).toBeDefined();
    expect(attrStr!.range).toEqual([5, src.length]);
    expect(attrStr!.value).toBe('"x\\');
  });
});

// ────────────────────────────────────────────────────────────────────
// Text content
// ────────────────────────────────────────────────────────────────────

describe('JSX tokenizer — text content', () => {
  test('plain text: <a>hello world</a>', () => diff('<a>hello world</a>;'));
  test('text with leading/trailing whitespace', () => diff('<a>  hi  </a>;'));
  test('text with newlines', () => diff('<a>\n  hi\n</a>;'));
  test('text adjacent to expression: <a>x{y}z</a>', () =>
    diff('<a>x{y}z</a>;'));
  test('text with entity-ish chars: <a>a & b</a>', () => diff('<a>a & b</a>;'));
});

// ────────────────────────────────────────────────────────────────────
// Dotted / namespaced names + hyphenated attrs
// ────────────────────────────────────────────────────────────────────

describe('JSX tokenizer — dotted / namespaced / hyphenated', () => {
  test('dotted name: <a.b.c/>', () => diff('<a.b.c/>;'));
  test('dotted in close: <a.b></a.b>', () => diff('<a.b></a.b>;'));
  test('namespaced: <svg:rect/>', () => diff('<svg:rect/>;'));
  test('hyphenated attr name: <a data-foo="y"/>', () =>
    diff('<a data-foo="y"/>;'));
  test('hyphenated tag name (not legal but lex test): <my-element/>', () =>
    diff('<my-element/>;'));
});

// ────────────────────────────────────────────────────────────────────
// Nesting + expression containers
// ────────────────────────────────────────────────────────────────────

describe('JSX tokenizer — nesting + expression containers', () => {
  test('nested elements: <a><b/></a>', () => diff('<a><b/></a>;'));
  test('deeply nested: <a><b><c/></b></a>', () => diff('<a><b><c/></b></a>;'));
  test('expr with call: <a>{f(1)}</a>', () => diff('<a>{f(1)}</a>;'));
  test('expr with object: <a>{ {x: 1} }</a>', () => diff('<a>{ {x: 1} }</a>;'));
  test('JSX inside expr: <a>{<b/>}</a>', () => diff('<a>{<b/>}</a>;'));
  test('JSX inside attr expr: <a x={<b/>}/>', () => diff('<a x={<b/>}/>;'));
});

// ────────────────────────────────────────────────────────────────────
// Sibling content after a child element closes
//
// A child element nests ON TOP of the parent's text frame; once the
// child self-closes / `>`-closes, the parent text frame must be the
// stack top again so the parent's remaining siblings (trailing text,
// inter-element whitespace, further child elements, expression
// containers) keep lexing as JSXText / JSX tokens. Pre-fix the parent
// text frame was discarded on the child OPEN, so after the child closed
// the JS scanner took over: the closing `</tag>` lexed as a phantom
// RegularExpression, sibling text as Identifier/Punctuator, and sibling
// whitespace JSXText was dropped. espree is the oracle for every case.
// ────────────────────────────────────────────────────────────────────

describe('JSX tokenizer — siblings after a child element', () => {
  // Depth 1: child then trailing text — the canonical repro.
  test('trailing text after self-closing child: <a><b/>x</a>', () =>
    diff('<a><b/>x</a>;'));
  test('leading text, child, trailing text: <a>x<b/>y</a>', () =>
    diff('<a>x<b/>y</a>;'));
  test('text between two element children: <a><b/>mid<c/></a>', () =>
    diff('<a><b/>mid<c/></a>;'));
  test('trailing text after non-self-closing child: <a><b>inner</b>after</a>', () =>
    diff('<a><b>inner</b>after</a>;'));
  test('alternating text and children: <div>one<i/>two<i/>three</div>', () =>
    diff('<div>one<i/>two<i/>three</div>;'));
  test('multi-char tag name child then text: <div><br/>text</div>', () =>
    diff('<div><br/>text</div>;'));
  test('punctuation-only trailing text: <p>Hello <b>world</b>!</p>', () =>
    diff('<p>Hello <b>world</b>!</p>;'));

  // Sibling whitespace must each be its own JSXText.
  test('inter-sibling newlines: <ul>\\n<li/>\\n<li/>\\n</ul>', () =>
    diff('<ul>\n<li/>\n<li/>\n</ul>;'));

  // Expression-container siblings around a child element.
  test('expression sibling after child: <a><b/>{x}</a>', () =>
    diff('<a><b/>{x}</a>;'));
  test('child after expression then trailing text: <a>{x}<b/>tail</a>', () =>
    diff('<a>{x}<b/>tail</a>;'));

  // Fragment child with a sibling.
  test('fragment child then sibling text: <a><>frag</>tail</a>', () =>
    diff('<a><>frag</>tail</a>;'));

  // Attribute on the child must not disturb the parent text frame.
  test('child with attribute then sibling text: <a><b id="x"/>z</a>', () =>
    diff('<a><b id="x"/>z</a>;'));

  // Depth 2/3: text resumes at EACH ancestor level as inner elements close.
  test('text resumes at two levels: <a><b><c/>m</b>n</a>', () =>
    diff('<a><b><c/>m</b>n</a>;'));
  test('text resumes at three levels: <a><b><c><d/>p</c>q</b>r</a>', () =>
    diff('<a><b><c><d/>p</c>q</b>r</a>;'));
});

// ────────────────────────────────────────────────────────────────────
// Template literals + JSX interaction
// ────────────────────────────────────────────────────────────────────

describe('JSX tokenizer — template literal interaction', () => {
  test('template inside JSX expr: <a>{`hi`}</a>', () => diff('<a>{`hi`}</a>;'));
  test('template with interpolation in JSX: <a>{`${x}`}</a>', () =>
    diff('<a>{`${x}`}</a>;'));
  test('JSX inside template expr: `${<a/>}`', () =>
    diff('const s = `${<a/>}`;'));
});

// ────────────────────────────────────────────────────────────────────
// Disambiguation — not-JSX cases
// ────────────────────────────────────────────────────────────────────

describe('JSX tokenizer — `<` is NOT JSX (regression guards)', () => {
  test('comparison: a < b', () => diff('const c = a < b;'));
  test('comparison chain: a < b && b < c', () =>
    diff('const c = a < b && b < c;'));
  test('shift: a << 1', () => diff('const c = a << 1;'));
});

describe('JSX tokenizer — TSX generic-vs-JSX heuristics', () => {
  // espree doesn't understand TypeScript syntax — `<T,>` and `<T
  // extends X>` throw "Unexpected token" in espree's jsx parser. So
  // we can't diff against espree for TSX disambiguation. Instead
  // assert the runner does NOT emit any JSXIdentifier / JSXText
  // tokens (i.e., the lookahead in `classifyJsxLAngle` correctly
  // rejected JSX entry and the `<` lexed as a plain Punctuator).
  const tsxGenericCases = [
    'const f = <T,>(x) => x;',
    'const f = <T extends X>(x) => x;',
    'const f = <T>(x) => x;',
  ];
  for (const src of tsxGenericCases) {
    test(`generic: ${src}`, () => {
      const out = runnerTokens(src, true);
      const jsxOnly = out.filter((t) =>
        ['JSXIdentifier', 'JSXText'].includes(t.type),
      );
      expect(jsxOnly).toEqual([]);
    });
  }
});

// ────────────────────────────────────────────────────────────────────
// Regression — non-JSX files unaffected
// ────────────────────────────────────────────────────────────────────

describe('JSX tokenizer — non-JSX files unchanged', () => {
  test('plain JS still tokenizes correctly when jsx flag is true', () => {
    // No `<` in source — JSX path should never trigger.
    diff('const x = 1; const y = "two"; function f() { return x + 1; }');
  });
  test('jsx flag off skips entire JSX scan', () => {
    // Direct check that runnerTokens(src, false) doesn't emit JSXIdentifier.
    const src = '<a/>;';
    const out = runnerTokens(src, false);
    const hasJsx = out.some((t) =>
      ['JSXIdentifier', 'JSXText'].includes(t.type),
    );
    expect(hasJsx).toBe(false);
  });
});

// ────────────────────────────────────────────────────────────────────
// More comprehensive coverage — real-world JSX patterns
// ────────────────────────────────────────────────────────────────────

describe('JSX tokenizer — real-world patterns', () => {
  test('component (PascalCase): <Foo prop="x"/>', () =>
    diff('<Foo prop="x"/>;'));
  test('member-expr component: <Foo.Bar/>', () => diff('<Foo.Bar/>;'));
  test('spread attribute: <a {...props}/>', () => diff('<a {...props}/>;'));
  test('spread mixed: <a x={1} {...rest} y/>', () =>
    diff('<a x={1} {...rest} y/>;'));
  test('conditional render: cond && <a/>', () =>
    diff('const x = cond && <a/>;'));
  test('ternary render: cond ? <a/> : <b/>', () =>
    diff('const x = cond ? <a/> : <b/>;'));
  test('array of elements: [<a/>, <b/>]', () =>
    diff('const xs = [<a/>, <b/>];'));
  test('function return: function f() { return <a/>; }', () =>
    diff('function f() { return <a/>; }'));
  test('arrow body: () => <a>{x}</a>', () =>
    diff('const f = () => <a>{x}</a>;'));
  test('map render: items.map(x => <li>{x}</li>)', () =>
    diff('const xs = items.map(x => <li>{x}</li>);'));
});

describe('JSX tokenizer — comments and whitespace', () => {
  test('block comment in JSX expr: <a>{/* c */}</a>', () =>
    diff('<a>{/* c */}</a>;'));
  test('line comment in JSX expr: <a>{ // c \nx}</a>', () =>
    diff('<a>{// c\nx}</a>;'));
  test('multi-line JSX', () =>
    diff('<div\n  className="x"\n  data-id={1}\n>\n  hi\n</div>;'));
});

describe('JSX tokenizer — fragments', () => {
  test('empty fragment: <></>', () => diff('const e = <></>;'));
  test('fragment with content: <>hi</>', () => diff('<>hi</>;'));
  test('fragment with element: <><a/></>', () => diff('<><a/></>;'));
  test('nested fragments: <><></></>', () => diff('<><></></>;'));
});

describe('JSX tokenizer — boolean / null attribute values', () => {
  test('expr with bool: <a x={true}/>', () => diff('<a x={true}/>;'));
  test('expr with null: <a x={null}/>', () => diff('<a x={null}/>;'));
});

describe('JSX tokenizer — `<` after various prev tokens', () => {
  test('after open paren: f(<a/>)', () => diff('f(<a/>);'));
  test('after comma: f(1, <a/>)', () => diff('f(1, <a/>);'));
  test('after &&: x && <a/>', () => diff('const v = x && <a/>;'));
  test('after ||: x || <a/>', () => diff('const v = x || <a/>;'));
  test('after ??: x ?? <a/>', () => diff('const v = x ?? <a/>;'));
  test('after return: function f(){return <a/>;}', () =>
    diff('function f(){return <a/>;}'));
  test('after =>: x => <a/>', () => diff('const f = x => <a/>;'));
  test('after !: !<a/>', () => diff('const v = !<a/>;'));
  test('after typeof: typeof <a/>', () => diff('const v = typeof <a/>;'));
});

describe('JSX tokenizer — non-JSX `<` regression suite', () => {
  test('after Identifier: x < y', () => diff('const v = x < y;'));
  test('after number: 1 < 2', () => diff('const v = 1 < 2;'));
  test('after string: "a" < "b"', () => diff('const v = "a" < "b";'));
  test('after close paren: f() < 1', () => diff('const v = f() < 1;'));
  test('after close bracket: arr[0] < 1', () => diff('const v = arr[0] < 1;'));
});
