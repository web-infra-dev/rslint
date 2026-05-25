// TestBraceStyleExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk / GitHub-issue
// shape it covers, so future refactors can't silently regress them without
// breaking a named lock-in.
//
// Dimension 4 walk for @stylistic/brace-style:
//
//   - Receiver / expression wrappers — N/A. The rule fires on block-bearing
//     statement nodes (BlockStatement, ClassBody, SwitchStatement, etc.) and
//     never inspects member-access receivers or type-wrapper expressions.
//   - Access / key forms — covered for class members (identifier / computed
//     / private / string-literal method names) and for object literal
//     methods.
//   - Declaration / container forms — covered: function decl / function expr
//     / arrow / method / class field arrow; constructor / getter / setter
//     bodies; class decl vs class expr; async / generator variants.
//   - Nesting / traversal boundaries — covered: class-in-class,
//     function-in-function, namespace-in-namespace, switch-in-switch,
//     try-in-try.
//   - Graceful degradation — covered: empty `{}` (various contexts), empty
//     switch, empty static block, empty namespace, empty try/catch/finally.
//
// Branch walk for upstream's `brace-style.ts`:
//
//   - `validateCurlyPair` 4-branch matrix — six minimum-input cases (one
//     per messageId, plus negative coverage on the `allowSingleLine`
//     escape).
//   - `validateCurlyBeforeKeyword` 2-branch matrix — nextLineClose and
//     sameLineClose each isolated.
//   - `STATEMENT_LIST_PARENTS` skip set: a Block whose parent is each of
//     `Program`, `Block`, `StaticBlock`, `SwitchCase` (CaseClause +
//     DefaultClause in tsgo) is locked in as valid under EVERY style.
//   - `allowSingleLine` exception path: each of the three exception-eligible
//     checks (sameLineOpen, blockSameLine, singleLineClose) has a same-line
//     test that PASSES with allowSingleLine and FAILS without it.
//   - Comment-blocks-fix path: `nextLineOpen` and `nextLineClose` each
//     locked in with a case where a comment in the trivia suppresses the
//     autofix (Output is the input).
//
// tsgo AST quirks specifically locked in:
//
//   - ClassDeclaration / ClassExpression with heritage clauses, type
//     parameters, decorators (single & multiple), abstract / declare
//     modifiers — for each the previous token before `{` resolves
//     correctly (heritage-clause name, type-parameter `>`, modifier
//     keyword, or class name).
//   - Static block is treated as its own container (not a nested Block) —
//     the Block listener skips the static body to avoid double-reporting.
//   - CatchClause without parameter (ES2019 `catch {}`) — close-before-keyword
//     still fires when finally follows.
//   - Module / namespace bodies (ModuleBlock) — both `module "..."` (string
//     literal name) and `namespace X` (identifier name) variants.
//   - WithStatement body Block — parses under strict mode, listener fires.
//
// Real-user shapes (sampled from common TS/JS codebases & ESLint issue tracker):
//
//   - React-style class component with handler arrow fields.
//   - IIFE patterns `(function() {})()` and `(() => {})()`.
//   - `export default class {}` and `export default function() {}` shapes.
//   - Function expression / arrow as object property value.
//   - Deeply chained `else if` (5+ levels).
//   - Nested try-catch with arrow inside catch.
//   - TS function overloads (only impl has body).
package brace_style_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/brace_style"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestBraceStyleExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&brace_style.BraceStyleRule,
		[]rule_tester.ValidTestCase{
			// ============================================================
			// Dimension 4: declaration / container forms
			// ============================================================

			// ---- Dimension 4: function expression body ----
			{Code: "const f = function () {\n  return;\n};"},

			// ---- Dimension 4: named function expression body ----
			{Code: "const f = function foo() {\n  return;\n};"},

			// ---- Dimension 4: arrow function block body ----
			{Code: "const f = () => {\n  return;\n};"},

			// ---- Dimension 4: arrow function expression body — no Block, listener never fires ----
			{Code: "const f = () => x;"},

			// ---- Dimension 4: async function body ----
			{Code: "async function f() {\n  return;\n}"},

			// ---- Dimension 4: async arrow body ----
			{Code: "const f = async () => {\n  return;\n};"},

			// ---- Dimension 4: generator function body ----
			{Code: "function* g() {\n  yield 1;\n}"},

			// ---- Dimension 4: async generator body ----
			{Code: "async function* g() {\n  yield 1;\n}"},

			// ---- Dimension 4: method body (object literal shorthand) ----
			{Code: "const o = {\n  m() {\n    return;\n  },\n};"},

			// ---- Dimension 4: computed method name ----
			{Code: "const o = {\n  [Symbol.iterator]() {\n    return;\n  },\n};"},

			// ---- Dimension 4: getter / setter bodies ----
			{Code: "const o = {\n  get x() {\n    return 1;\n  },\n  set x(v) {\n    this._x = v;\n  },\n};"},

			// ---- Dimension 4: constructor body ----
			{Code: "class C {\n  constructor() {\n    this.x = 1;\n  }\n}"},

			// ---- Dimension 4: class method (regular) ----
			{Code: "class C {\n  m() {\n    return;\n  }\n}"},

			// ---- Dimension 4: class async method ----
			{Code: "class C {\n  async m() {\n    return;\n  }\n}"},

			// ---- Dimension 4: class generator method ----
			{Code: "class C {\n  *m() {\n    yield 1;\n  }\n}"},

			// ---- Dimension 4: class async generator method ----
			{Code: "class C {\n  async *m() {\n    yield 1;\n  }\n}"},

			// ---- Dimension 4: class private method ----
			{Code: "class C {\n  #m() {\n    return;\n  }\n}"},

			// ---- Dimension 4: class computed-key method ----
			{Code: "class C {\n  [Symbol.iterator]() {\n    return;\n  }\n}"},

			// ---- Dimension 4: class string-literal-key method ----
			{Code: "class C {\n  \"m\"() {\n    return;\n  }\n}"},

			// ---- Dimension 4: class field arrow initializer ----
			{Code: "class C {\n  handler = (a) => {\n    return a;\n  };\n}"},

			// ---- Dimension 4: class field arrow with expression body — no inner Block ----
			{Code: "class C {\n  handler = (a) => a;\n}"},

			// ============================================================
			// tsgo AST quirks — class declaration / expression shapes
			// ============================================================

			// ---- tsgo lock-in: class with type parameters ----
			{Code: "class C<T> {\n  x: T;\n}"},

			// ---- tsgo lock-in: class with type parameter constraints ----
			{Code: "class C<T extends string> {\n  x: T;\n}"},

			// ---- tsgo lock-in: class with multiple type parameters ----
			{Code: "class C<T, U, V> {\n  x: T;\n}"},

			// ---- tsgo lock-in: class with extends ----
			{Code: "class C extends Base {\n  x = 1;\n}"},

			// ---- tsgo lock-in: class with extends and implements ----
			{Code: "class C extends Base implements I1, I2 {\n  x = 1;\n}"},

			// ---- tsgo lock-in: class with extends + generics ----
			{Code: "class C<T> extends Base<T> implements I<T> {\n  x: T;\n}"},

			// ---- tsgo lock-in: class with single decorator ----
			{Code: "@dec\nclass C {\n  x = 1;\n}"},

			// ---- tsgo lock-in: class with decorator with args ----
			{Code: "@dec(arg, arg2)\nclass C {\n  x = 1;\n}"},

			// ---- tsgo lock-in: class with multiple decorators ----
			{Code: "@a\n@b()\nclass C {\n  x = 1;\n}"},

			// ---- tsgo lock-in: class with decorator + allman ----
			// Locks in that scanner finds `Foo` as previous token, not `@dec`.
			{Code: "@dec\nclass Foo\n{\n}", Options: optsAllman()},

			// ---- tsgo lock-in: abstract class ----
			{Code: "abstract class C {\n  abstract m(): void;\n}"},

			// ---- tsgo lock-in: declare class ----
			{Code: "declare class C {\n  x: number;\n}"},

			// ---- tsgo lock-in: named class expression ----
			{Code: "const C = class Foo {\n  x = 1;\n};"},

			// ---- tsgo lock-in: anonymous class expression ----
			{Code: "const C = class {\n  x = 1;\n};"},

			// ---- tsgo lock-in: class expression with extends ----
			{Code: "const C = class extends Base {\n  x = 1;\n};"},

			// ============================================================
			// Branch lock-ins for STATEMENT_LIST_PARENTS skip set
			// ============================================================

			// ---- Locks in shouldSkipBlock(parent=Program): standalone top-level block ----
			{Code: "{ x; }"},
			{Code: "{ x; }", Options: optsStroustrup()},
			{Code: "{ x; }", Options: optsAllman()},

			// ---- Locks in shouldSkipBlock(parent=Block): nested standalone block ----
			{Code: "function f() {\n  { x; }\n}"},
			{Code: "function f() {\n  { x; }\n}", Options: optsStroustrup()},
			{Code: "function f()\n{\n  { x; }\n}", Options: optsAllman()},

			// ---- Locks in shouldSkipBlock(parent=CaseClause): block in case ----
			{Code: "switch (x) {\n  case 1: { x; }\n}"},
			{Code: "switch (x) {\n  case 1: { x; }\n}", Options: optsStroustrup()},

			// ---- Locks in shouldSkipBlock(parent=DefaultClause): block in default ----
			{Code: "switch (x) {\n  default: { x; }\n}"},
			{Code: "switch (x) {\n  default: { x; }\n}", Options: optsStroustrup()},

			// ---- Locks in shouldSkipBlock(parent=ClassStaticBlockDeclaration): inner Block of static block ----
			// The static-block listener handles this; the inner Block listener
			// must skip it to avoid double-reporting.
			{Code: "class C {\n  static {\n    foo;\n  }\n}"},

			// ============================================================
			// Branch lock-ins for the 6 messageIds (isolation cases)
			// ============================================================

			// (See invalid section for the failing companions.)
			// ---- valid baseline for allowSingleLine on each style ----
			{Code: "function f() { return; }", Options: opts1tbsSingle()},
			{Code: "function f() {}", Options: opts1tbsSingle()},
			{Code: "function f() { return; }", Options: optsStroustrupSingle()},
			{Code: "function f() {}", Options: optsAllmanSingle()},

			// ============================================================
			// rslint-specific lock-ins
			// ============================================================

			// ---- TS namespace body (KindModuleBlock) — basic ----
			{Code: "namespace Foo {\n  export const x = 1;\n}"},

			// ---- TS nested namespace ----
			{Code: "namespace A.B {\n  export const x = 1;\n}"},

			// ---- TS nested namespace declarations ----
			{Code: "namespace Outer {\n  namespace Inner {\n    export const x = 1;\n  }\n}"},

			// ---- TS declare module (ambient) ----
			{Code: "declare module 'foo' {\n  export const x: number;\n}"},

			// ---- TS declare global ----
			{Code: "declare global {\n  interface Window { x: number; }\n}"},

			// ---- Catch without parameter (ES2019) ----
			{Code: "try {\n  a();\n} catch {\n  b();\n}"},

			// ---- Catch without parameter + finally — locks in close-before-finally still fires ----
			{Code: "try {\n  a();\n} catch {\n  b();\n} finally {\n  c();\n}"},

			// ---- Try-only-finally (no catch) ----
			{Code: "try {\n  a();\n} finally {\n  c();\n}"},

			// ---- Stroustrup try-only-finally ----
			{Code: "try {\n  a();\n}\nfinally {\n  c();\n}", Options: optsStroustrup()},

			// ---- Allman try-only-finally ----
			{Code: "try\n{\n  a();\n}\nfinally\n{\n  c();\n}", Options: optsAllman()},

			// ---- Do-while block ----
			{Code: "do {\n  bar();\n} while (true);"},

			// ---- For-of block ----
			{Code: "for (const x of arr) {\n  bar();\n}"},

			// ---- Labeled statement containing a block — parent != STATEMENT_LIST_PARENT ----
			// The Block parent is `LabeledStatement`; listener DOES fire.
			{Code: "loop: {\n  break loop;\n}"},

			// ============================================================
			// Dimension 4: nesting / traversal boundaries
			// ============================================================

			// ---- Class in class — both bodies checked independently ----
			{Code: "class Outer {\n  inner = class Inner {\n    x = 1;\n  };\n}"},

			// ---- Function in function ----
			{Code: "function outer() {\n  function inner() {\n    return;\n  }\n  return inner;\n}"},

			// ---- Switch in switch ----
			{Code: "switch (x) {\n  case 1:\n    switch (y) {\n      case 2: break;\n    }\n    break;\n}"},

			// ---- Try in try ----
			{Code: "try {\n  try {\n    a();\n  } catch (e) {\n    b();\n  }\n} catch (e) {\n  c();\n}"},

			// ---- Try in catch ----
			{Code: "try {\n  a();\n} catch (e) {\n  try {\n    b();\n  } catch (e2) {\n    c();\n  }\n}"},

			// ---- 5-level else-if chain ----
			{Code: "if (a) {\n  x;\n} else if (b) {\n  y;\n} else if (c) {\n  z;\n} else if (d) {\n  w;\n} else if (e) {\n  v;\n} else {\n  u;\n}"},

			// ---- 5-level else-if stroustrup ----
			{Code: "if (a) {\n  x;\n}\nelse if (b) {\n  y;\n}\nelse if (c) {\n  z;\n}\nelse if (d) {\n  w;\n}\nelse if (e) {\n  v;\n}\nelse {\n  u;\n}", Options: optsStroustrup()},

			// ============================================================
			// Real-user code shapes
			// ============================================================

			// ---- Real-user: React-style class component with arrow fields ----
			{Code: "class Component {\n  handleClick = () => {\n    this.setState({});\n  };\n  render() {\n    return null;\n  }\n}"},

			// ---- Real-user: IIFE with function expression ----
			{Code: "(function() {\n  console.log('init');\n})();"},

			// ---- Real-user: IIFE with arrow ----
			{Code: "(() => {\n  console.log('init');\n})();"},

			// ---- Real-user: IIFE single-line with allowSingleLine ----
			{Code: "(function() { return 1; })();", Options: opts1tbsSingle()},

			// ---- Real-user: export default function ----
			{Code: "export default function foo() {\n  return 1;\n}"},

			// ---- Real-user: anonymous export default function ----
			{Code: "export default function() {\n  return 1;\n}"},

			// ---- Real-user: export default class ----
			{Code: "export default class Foo {\n  x = 1;\n}"},

			// ---- Real-user: anonymous export default class ----
			{Code: "export default class {\n  x = 1;\n}"},

			// ---- Real-user: nested try-catch with arrow ----
			{Code: "try {\n  const f = () => {\n    throw new Error();\n  };\n  f();\n} catch (e) {\n  console.error(e);\n}"},

			// ---- Real-user: arrow inside async function ----
			{Code: "async function f() {\n  const cb = () => {\n    return 1;\n  };\n  return cb();\n}"},

			// ---- Real-user: TS function overloads (only impl has body) ----
			{Code: "function f(x: string): void;\nfunction f(x: number): void;\nfunction f(x: any): void {\n  console.log(x);\n}"},

			// ============================================================
			// TS-only "shape" non-targets: the rule MUST NOT fire on these
			// (locks in that I'm not over-firing on body-like constructs that
			// upstream's ESTree listeners never touch)
			// ============================================================

			// ---- TS interface declaration body — not handled by upstream ----
			// Same style violations of `1tbs` placement on interface braces
			// should NOT fire. Upstream has no TSInterfaceBody listener.
			{Code: "interface I\n{\n  x: number;\n}"},
			{Code: "interface I\n{\n  x: number;\n}", Options: optsStroustrup()},

			// ---- TS enum declaration body — not handled ----
			{Code: "enum E\n{\n  A,\n  B,\n}"},
			{Code: "enum E { A, B }"},

			// ---- TS const enum body — not handled ----
			{Code: "const enum E\n{\n  A,\n  B,\n}"},

			// ---- TS type literal — not handled ----
			{Code: "type T = {\n  x: number;\n};"},
			{Code: "type T = { x: number };"},

			// ---- TS mapped type with braces — not handled ----
			{Code: "type T<X> = {\n  [K in keyof X]: X[K];\n};"},

			// ---- Object literal in expression position — not handled ----
			// (Despite using `{}`, an ObjectLiteralExpression is NOT a Block.)
			{Code: "const o = { x: 1, y: 2 };"},
			{Code: "const o =\n{\n  x: 1,\n};"},

			// ---- JSX expression container — not a Block ----
			{Code: "const el = <div>{value}</div>;", FileName: "react.tsx"},
			{Code: "const el = <div>{ value }</div>;", FileName: "react.tsx"},

			// ---- Destructuring pattern uses `{}` — not a Block ----
			{Code: "const { a, b } = o;"},
			{Code: "function f({ x }) { return x; }", Options: opts1tbsSingle()},

			// ============================================================
			// More TS-specific class shapes (locking in tsgo behavior)
			// ============================================================

			// ---- Class with access modifiers ----
			{Code: "class C {\n  public x = 1;\n  private y = 2;\n  protected z = 3;\n}"},

			// ---- Class with readonly fields ----
			{Code: "class C {\n  readonly x = 1;\n}"},

			// ---- Class with optional method ----
			{Code: "class C {\n  m?(): void {\n    return;\n  }\n}"},

			// ---- Class with definite assignment field ----
			{Code: "class C {\n  x!: number;\n}"},

			// ---- Class with override modifier ----
			{Code: "class Sub extends Base {\n  override m() {\n    return;\n  }\n}"},

			// ---- Class with decorated method ----
			{Code: "class C {\n  @dec m() {\n    return;\n  }\n}"},

			// ---- Class with decorated field arrow ----
			{Code: "class C {\n  @dec handler = (a) => {\n    return a;\n  };\n}"},

			// ---- Class with multiple decorated members ----
			{Code: "class C {\n  @a x = 1;\n  @b y = 2;\n  @c m() {\n    return;\n  }\n}"},

			// ---- Class with parameter properties ----
			{Code: "class C {\n  constructor(public x: number, private y: string) {\n    return;\n  }\n}"},

			// ---- Class with all-static members ----
			{Code: "class Utils {\n  static a = 1;\n  static b = 2;\n  static m() {\n    return;\n  }\n}"},

			// ---- Class with mixed static/instance + static block ----
			{Code: "class C {\n  static x = 1;\n  static {\n    C.x = 2;\n  }\n  m() {\n    return this;\n  }\n}"},

			// ============================================================
			// More function/control-flow forms
			// ============================================================

			// ---- TS function with type predicates ----
			{Code: "function isString(x: unknown): x is string {\n  return typeof x === 'string';\n}"},

			// ---- TS function with generic constraint ----
			{Code: "function f<T extends string>(x: T): T {\n  return x;\n}"},

			// ---- Async arrow returning a Promise ----
			{Code: "const f: () => Promise<void> = async () => {\n  await Promise.resolve();\n};"},

			// ---- Higher-order function ----
			{Code: "function curry(f: Function): Function {\n  return function (x: unknown) {\n    return function (y: unknown) {\n      return f(x, y);\n    };\n  };\n}"},

			// ---- Method chained body ----
			{Code: "arr.map((x) => {\n  return x * 2;\n}).filter((x) => {\n  return x > 0;\n});"},

			// ---- For-await-of with try inside ----
			{Code: "async function f() {\n  for await (const x of stream) {\n    try {\n      await process(x);\n    } catch (e) {\n      console.error(e);\n    }\n  }\n}"},

			// ============================================================
			// More if-else chain / switch shapes
			// ============================================================

			// ---- if-else single-line on one branch, multi on another ----
			{Code: "if (a) { foo(); } else {\n  bar();\n  baz();\n}", Options: opts1tbsSingle()},

			// ---- switch with multiple case statements grouping ----
			{Code: "switch (x) {\n  case 1:\n  case 2:\n  case 3:\n    foo();\n    break;\n}"},

			// ---- switch with break in case body ----
			{Code: "switch (x) {\n  case 1: {\n    foo();\n    break;\n  }\n  case 2: {\n    bar();\n    break;\n  }\n}"},

			// ---- switch with throw / return as case body ----
			{Code: "switch (x) {\n  case 1: return 1;\n  case 2: throw new Error();\n  default: return 0;\n}"},

			// ---- nested switch with shared `default` ----
			{Code: "switch (x) {\n  case 1: {\n    switch (y) {\n      case 'a': foo(); break;\n      default: bar();\n    }\n    break;\n  }\n}"},

			// ============================================================
			// More namespace / module shapes
			// ============================================================

			// ---- namespace with method-like signature ----
			{Code: "namespace Foo {\n  export function bar() {\n    return 1;\n  }\n}"},

			// ---- namespace with nested class ----
			{Code: "namespace Foo {\n  export class Bar {\n    x = 1;\n  }\n}"},

			// ---- export = ... pattern ----
			{Code: "namespace Foo {\n  export const x = 1;\n}\nexport = Foo;"},

			// ---- ambient module with declarations ----
			{Code: "declare module 'pkg' {\n  export function fn(): void;\n  export class C {\n    x: number;\n  }\n}"},

			// ============================================================
			// Comments at unusual positions
			// ============================================================

			// ---- Comment after `)` of function head ----
			{Code: "function f() /* c */ {\n  return;\n}"},

			// ---- Comment between `}` and `else` — same-line, no diagnostic ----
			{Code: "if (a) {\n  b;\n} /* c */ else {\n  d;\n}"},

			// ---- Block comment inside block body ----
			{Code: "function f() {\n  /* doc */\n  return;\n}"},

			// ---- Leading block comment on class ----
			{Code: "/** doc */\nclass C {\n  x = 1;\n}"},

			// ---- Inline comment after `{` ----
			{Code: "function f() { /* short */\n  return;\n}"},

			// ---- Real-user: class with overload signatures ----
			{Code: "class C {\n  m(x: string): void;\n  m(x: number): void;\n  m(x: any): void {\n    console.log(x);\n  }\n}"},

			// ---- Real-user: abstract method (no body) coexisting with concrete methods ----
			{Code: "abstract class C {\n  abstract a(): void;\n  b() {\n    this.a();\n  }\n}"},

			// ============================================================
			// Fix idempotence — running the fix once should converge
			// ============================================================

			// The invalid section below covers the fixed outputs; the cases
			// here are the EXPECTED OUTPUT of those fixes — they must pass
			// the rule as valid after one fix pass. This is the idempotence
			// contract: rule(fix(code)) should produce zero new diagnostics.
			{Code: "function foo() {\n return; \n}"},
			{Code: "function foo() { \n return; \n}"},
			{Code: "if (a) {\n  b;\n} else {\n  c;\n}"},
			{Code: "if (a) {\n  b;\n}\nelse {\n  c;\n}", Options: optsStroustrup()},
			{Code: "if (a)\n{\n  b;\n}\nelse\n{\n  c;\n}", Options: optsAllman()},
		},
		[]rule_tester.InvalidTestCase{
			// ============================================================
			// Branch isolation: each of the 6 messageIds in its own case
			// ============================================================

			// ---- Locks in upstream validateCurlyPair: nextLineOpen ----
			{
				Code:   "function f()\n{\n  return;\n}",
				Output: []string{"function f() {\n  return;\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 1},
				},
			},

			// ---- Locks in upstream validateCurlyPair: sameLineOpen ----
			{
				Code:    "function f() {\n  return;\n}",
				Output:  []string{"function f() \n{\n  return;\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 14},
				},
			},

			// ---- Locks in upstream validateCurlyPair: blockSameLine ----
			{
				Code:   "function f() { foo;\n}",
				Output: []string{"function f() {\n foo;\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 14},
				},
			},

			// ---- Locks in upstream validateCurlyPair: singleLineClose ----
			{
				Code:   "function f() {\n  foo; }",
				Output: []string{"function f() {\n  foo; \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 2, Column: 8},
				},
			},

			// ---- Locks in upstream validateCurlyBeforeKeyword: nextLineClose ----
			{
				Code:   "if (a) {\n  b;\n}\nelse {\n  c;\n}",
				Output: []string{"if (a) {\n  b;\n} else {\n  c;\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 3, Column: 1},
				},
			},

			// ---- Locks in upstream validateCurlyBeforeKeyword: sameLineClose ----
			{
				Code:    "if (a) {\n  b;\n} else {\n  c;\n}",
				Output:  []string{"if (a) {\n  b;\n}\n else {\n  c;\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineClose", Line: 3, Column: 1},
				},
			},

			// ============================================================
			// Locks in commentsExistBetween fix-suppression branches
			// ============================================================

			// ---- nextLineClose: comment between } and keyword → fix suppressed ----
			// Diagnostic still fires; no Output → no fix applied.
			{
				Code: "if (a) {\n  b;\n} /* comment */\nelse {\n  c;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 3, Column: 1},
				},
			},

			// ---- nextLineOpen: line comment between paren and { → fix suppressed ----
			// (Already covered in upstream tests as `if (foo) // comment \n{...}`,
			// duplicated here with try/catch to lock in the suppression for
			// validateCurlyBeforeKeyword's sibling, validateCurlyPair.)
			{
				Code: "function f() // line comment\n{\n  return;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 1},
				},
			},

			// ---- nextLineClose with try/finally and a block-comment in between ----
			{
				Code: "try {\n  a();\n}\n/* note */\nfinally {\n  b();\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 3, Column: 1},
				},
			},

			// ============================================================
			// tsgo AST quirk locks — class declaration / expression shapes
			// ============================================================

			// ---- tsgo: class with type parameters, allman ----
			{
				Code:    "class C<T> {\n  x: T;\n}",
				Output:  []string{"class C<T> \n{\n  x: T;\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 12},
				},
			},

			// ---- tsgo: class with extends clause, allman ----
			// Locks in that scanner correctly identifies `Base` as previous
			// token before `{` (not `extends`).
			{
				Code:    "class C extends Base {\n  x = 1;\n}",
				Output:  []string{"class C extends Base \n{\n  x = 1;\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 22},
				},
			},

			// ---- tsgo: class with extends + implements, default ----
			// Multi-line heritage shouldn't fire nextLineOpen if `{` is on
			// same line as the last heritage identifier.
			{
				Code: "class C\n  extends Base\n  implements I1\n{\n  x = 1;\n}",
				Output: []string{
					"class C\n  extends Base\n  implements I1 {\n  x = 1;\n}",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 4, Column: 1},
				},
			},

			// ---- tsgo: decorated class, allman should still fire ----
			{
				Code:    "@dec\nclass C {\n  x = 1;\n}",
				Output:  []string{"@dec\nclass C \n{\n  x = 1;\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 2, Column: 9},
				},
			},

			// ---- tsgo: anonymous class expr with name ----
			{
				Code:   "const C = class\n{};",
				Output: []string{"const C = class {};"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 1},
				},
			},

			// ---- tsgo: class expr with extends, allman ----
			{
				Code:    "const C = class extends Base {\n  x = 1;\n};",
				Output:  []string{"const C = class extends Base \n{\n  x = 1;\n};"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 30},
				},
			},

			// ============================================================
			// Real-user fix scenarios
			// ============================================================

			// ---- Real-user: function expression inside object property value ----
			// Locks in that `tokenBefore` for `{` is `)`, not `:` of property.
			{
				Code:   "const o = { fn: function() { return; \n} };",
				Output: []string{"const o = { fn: function() {\n return; \n} };"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 28},
				},
			},

			// ---- Real-user: arrow function body single-line ----
			{
				Code:   "const f = () => { return; }",
				Output: []string{"const f = () => {\n return; \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 17},
					{MessageId: "singleLineClose", Line: 1, Column: 27},
				},
			},

			// ---- Real-user: class method body single-line ----
			{
				Code:   "class C {\n  m() { return; }\n}",
				Output: []string{"class C {\n  m() {\n return; \n}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 2, Column: 7},
					{MessageId: "singleLineClose", Line: 2, Column: 17},
				},
			},

			// ---- Real-user: class private method body single-line ----
			{
				Code:   "class C {\n  #m() { return; }\n}",
				Output: []string{"class C {\n  #m() {\n return; \n}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 2, Column: 8},
					{MessageId: "singleLineClose", Line: 2, Column: 18},
				},
			},

			// ---- Real-user: class async method ----
			{
				Code:   "class C {\n  async m() { return; }\n}",
				Output: []string{"class C {\n  async m() {\n return; \n}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 2, Column: 13},
					{MessageId: "singleLineClose", Line: 2, Column: 23},
				},
			},

			// ---- Real-user: class getter ----
			{
				Code:   "class C {\n  get x() { return 1; }\n}",
				Output: []string{"class C {\n  get x() {\n return 1; \n}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 2, Column: 11},
					{MessageId: "singleLineClose", Line: 2, Column: 23},
				},
			},

			// ---- Real-user: constructor body ----
			{
				Code:   "class C {\n  constructor() { this.x = 1; }\n}",
				Output: []string{"class C {\n  constructor() {\n this.x = 1; \n}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 2, Column: 17},
					{MessageId: "singleLineClose", Line: 2, Column: 31},
				},
			},

			// ---- Real-user: catch without parameter, stroustrup ----
			// Locks in that `catch {}` (no param) is detected and close-before-catch fires.
			{
				Code:    "try {\n  a();\n} catch {\n  b();\n}",
				Output:  []string{"try {\n  a();\n}\n catch {\n  b();\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineClose", Line: 3, Column: 1},
				},
			},

			// ---- Real-user: catch without parameter + finally, both close-before checks fire ----
			{
				Code:    "try {\n  a();\n} catch {\n  b();\n} finally {\n  c();\n}",
				Output:  []string{"try {\n  a();\n}\n catch {\n  b();\n}\n finally {\n  c();\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineClose", Line: 3, Column: 1},
					{MessageId: "sameLineClose", Line: 5, Column: 1},
				},
			},

			// ---- Real-user: try-only-finally — close-before-finally fires ----
			// No catch in between; close-before-keyword still triggers on tryBlock's }.
			{
				Code:    "try {\n  a();\n}\nfinally {\n  c();\n}",
				Output:  []string{"try {\n  a();\n} finally {\n  c();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 3, Column: 1},
				},
			},

			// ---- Real-user: deeply chained else-if with one violation in the chain ----
			{
				Code:   "if (a) {\n  x;\n} else if (b) {\n  y;\n} else if (c) {\n  z;\n}\nelse {\n  w;\n}",
				Output: []string{"if (a) {\n  x;\n} else if (b) {\n  y;\n} else if (c) {\n  z;\n} else {\n  w;\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 7, Column: 1},
				},
			},

			// ---- Real-user: nested try-catch — only inner has a violation ----
			{
				Code:   "try {\n  try {\n    a();\n  }\n  catch (e) {\n    b();\n  }\n} catch (e) {\n  c();\n}",
				Output: []string{"try {\n  try {\n    a();\n  } catch (e) {\n    b();\n  }\n} catch (e) {\n  c();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 4, Column: 3},
				},
			},

			// ---- Real-user: class-in-class — only inner one violates ----
			{
				Code:   "class Outer {\n  inner = class Inner\n  {\n  };\n}",
				Output: []string{"class Outer {\n  inner = class Inner {\n  };\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 3, Column: 3},
				},
			},

			// ---- Real-user: namespace with a single same-line empty ----
			{
				Code:    "namespace Foo { }",
				Output:  []string{"namespace Foo \n{ }"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 15},
				},
			},

			// ---- Real-user: nested namespace, only outer violates ----
			{
				Code:   "namespace Outer\n{\n  namespace Inner {\n    export const x = 1;\n  }\n}",
				Output: []string{"namespace Outer {\n  namespace Inner {\n    export const x = 1;\n  }\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 1},
				},
			},

			// ---- Real-user: declare module ambient with bad braces ----
			{
				Code:   "declare module 'foo'\n{\n  export const x: number;\n}",
				Output: []string{"declare module 'foo' {\n  export const x: number;\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 1},
				},
			},

			// ---- Real-user: SwitchCase with same-line close ----
			{
				Code:   "switch (x) {\n  case 1: foo(); }",
				Output: []string{"switch (x) {\n  case 1: foo(); \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 2, Column: 18},
				},
			},

			// ---- Real-user: switch in switch — inner has a violation ----
			{
				Code:   "switch (x) {\n  case 1:\n    switch (y) {\n      case 2: break; }\n    break;\n}",
				Output: []string{"switch (x) {\n  case 1:\n    switch (y) {\n      case 2: break; \n}\n    break;\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 4, Column: 22},
				},
			},

			// ============================================================
			// Multi-violation source-order locks (cross-listener ordering)
			// ============================================================

			// ---- Locks ordering: seven reports in source order across multiple listeners ----
			// All reports come from different listener firings (Block enter,
			// Block exit, close-before-keyword trigger). At the same `}`
			// column the close-before-keyword report fires BEFORE the block's
			// own `singleLineClose` — matches ESLint's parent-listener-first
			// emission order (IfStatement listener fires before its
			// consequent BlockStatement listener).
			{
				Code:    "if (a) { x; } else { y; }",
				Output:  []string{"if (a) \n{\n x; \n}\n else \n{\n y; \n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 8},
					{MessageId: "blockSameLine", Line: 1, Column: 8},
					{MessageId: "sameLineClose", Line: 1, Column: 13},
					{MessageId: "singleLineClose", Line: 1, Column: 13},
					{MessageId: "sameLineOpen", Line: 1, Column: 20},
					{MessageId: "blockSameLine", Line: 1, Column: 20},
					{MessageId: "singleLineClose", Line: 1, Column: 25},
				},
			},

			// ---- Locks ordering: try-block close-before-catch + catch-body close-before-finally ----
			// Two close-before-keyword reports in a single TryStatement.
			{
				Code:    "try {\n  a();\n} catch (e) {\n  b();\n} finally {\n  c();\n}",
				Output:  []string{"try {\n  a();\n}\n catch (e) {\n  b();\n}\n finally {\n  c();\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineClose", Line: 3, Column: 1},
					{MessageId: "sameLineClose", Line: 5, Column: 1},
				},
			},

			// ---- Locks ordering: class body singleLineClose comes from KindClassDeclaration exit, not KindBlock ----
			{
				Code:   "class Foo {\n  m() {\n}}",
				Output: []string{"class Foo {\n  m() {\n}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleLineClose", Line: 3, Column: 2},
				},
			},

			// ============================================================
			// More tsgo-quirk-driven invalid cases
			// ============================================================

			// ---- Class with decorator + bad single-line allman ----
			{
				Code:    "@dec\nclass C { x = 1; }",
				Output:  []string{"@dec\nclass C \n{\n x = 1; \n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 2, Column: 9},
					{MessageId: "blockSameLine", Line: 2, Column: 9},
					{MessageId: "singleLineClose", Line: 2, Column: 18},
				},
			},

			// ---- Abstract class with same-line body, allman ----
			{
				Code:    "abstract class C { x: number; }",
				Output:  []string{"abstract class C \n{\n x: number; \n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 18},
					{MessageId: "blockSameLine", Line: 1, Column: 18},
					{MessageId: "singleLineClose", Line: 1, Column: 31},
				},
			},

			// ---- Class with `override` method body bad ----
			{
				Code:   "class S extends B {\n  override m() { return; }\n}",
				Output: []string{"class S extends B {\n  override m() {\n return; \n}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 2, Column: 16},
					{MessageId: "singleLineClose", Line: 2, Column: 26},
				},
			},

			// ---- Class field arrow with bad braces ----
			{
				Code:   "class C {\n  handler = (a) => { return a; };\n}",
				Output: []string{"class C {\n  handler = (a) => {\n return a; \n};\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 2, Column: 20},
					{MessageId: "singleLineClose", Line: 2, Column: 32},
				},
			},

			// ---- Parameter property constructor with bad braces ----
			{
				Code:   "class C {\n  constructor(public x: number) { this.y = x; }\n}",
				Output: []string{"class C {\n  constructor(public x: number) {\n this.y = x; \n}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 2, Column: 33},
					{MessageId: "singleLineClose", Line: 2, Column: 47},
				},
			},

			// ============================================================
			// More multi-violation source-order locks
			// ============================================================

			// ---- if-else-if chain with violation at each level, default 1tbs ----
			// The chain is: if {} else if {}\nelse if {}\nelse {}.
			// The 2 `\nelse` transitions BOTH violate 1tbs's nextLineClose.
			{
				Code:   "if (a) {\n  x;\n}\nelse if (b) {\n  y;\n}\nelse {\n  z;\n}",
				Output: []string{"if (a) {\n  x;\n} else if (b) {\n  y;\n} else {\n  z;\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineClose", Line: 3, Column: 1},
					{MessageId: "nextLineClose", Line: 6, Column: 1},
				},
			},

			// ---- if-else chain in stroustrup with violations at each level ----
			{
				Code:    "if (a) {\n  x;\n} else if (b) {\n  y;\n} else {\n  z;\n}",
				Output:  []string{"if (a) {\n  x;\n}\n else if (b) {\n  y;\n}\n else {\n  z;\n}"},
				Options: optsStroustrup(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineClose", Line: 3, Column: 1},
					{MessageId: "sameLineClose", Line: 5, Column: 1},
				},
			},

			// ---- try-catch-finally with all three blocks single-line wrong ----
			{
				Code:   "try { a(); } catch (e) { b(); } finally { c(); }",
				Output: []string{"try {\n a(); \n} catch (e) {\n b(); \n} finally {\n c(); \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 5},
					{MessageId: "singleLineClose", Line: 1, Column: 12},
					{MessageId: "blockSameLine", Line: 1, Column: 24},
					{MessageId: "singleLineClose", Line: 1, Column: 31},
					{MessageId: "blockSameLine", Line: 1, Column: 41},
					{MessageId: "singleLineClose", Line: 1, Column: 48},
				},
			},

			// ---- Nested class — outer + inner BOTH violate (allman) ----
			{
				Code:    "class O {\n  inner = class I {\n    x = 1;\n  };\n}",
				Output:  []string{"class O \n{\n  inner = class I \n{\n    x = 1;\n  };\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 9},
					{MessageId: "sameLineOpen", Line: 2, Column: 19},
				},
			},

			// ---- Switch with violation on body + violation in case-internal block ----
			// Important: case-internal block is SKIPPED by STATEMENT_LIST_PARENTS,
			// so only the OUTER switch's braces should violate.
			{
				Code:   "switch (x) { case 1: {\n    foo();\n  }\n  break;\n}",
				Output: []string{"switch (x) {\n case 1: {\n    foo();\n  }\n  break;\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 12},
				},
			},

			// ---- Real-user: IIFE with bad braces ----
			{
				Code:   "(function() { return 1; })();",
				Output: []string{"(function() {\n return 1; \n})();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 13},
					{MessageId: "singleLineClose", Line: 1, Column: 25},
				},
			},

			// ---- Real-user: arrow IIFE with bad braces ----
			{
				Code:   "(() => { return 1; })();",
				Output: []string{"(() => {\n return 1; \n})();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 8},
					{MessageId: "singleLineClose", Line: 1, Column: 20},
				},
			},

			// ---- Real-user: export default function with bad braces ----
			{
				Code:   "export default function foo() { return 1; }",
				Output: []string{"export default function foo() {\n return 1; \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 31},
					{MessageId: "singleLineClose", Line: 1, Column: 43},
				},
			},

			// ---- Real-user: export default class with bad braces ----
			{
				Code:    "export default class\n{\n  x = 1;\n}",
				Output:  []string{"export default class {\n  x = 1;\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nextLineOpen", Line: 2, Column: 1},
				},
			},

			// ---- Generic class with bad braces, allman ----
			{
				Code:    "class C<T extends string> { x: T; }",
				Output:  []string{"class C<T extends string> \n{\n x: T; \n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 27},
					{MessageId: "blockSameLine", Line: 1, Column: 27},
					{MessageId: "singleLineClose", Line: 1, Column: 35},
				},
			},

			// ---- Generic function with bad braces ----
			{
				Code:   "function f<T>(x: T): T { return x; }",
				Output: []string{"function f<T>(x: T): T {\n return x; \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 24},
					{MessageId: "singleLineClose", Line: 1, Column: 36},
				},
			},

			// ---- Async generator method with bad braces ----
			{
				Code:   "class C {\n  async *m() { yield 1; }\n}",
				Output: []string{"class C {\n  async *m() {\n yield 1; \n}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 2, Column: 14},
					{MessageId: "singleLineClose", Line: 2, Column: 25},
				},
			},

			// ---- TS function with type predicate, bad single-line ----
			{
				Code:   "function isStr(x: unknown): x is string { return typeof x === 'string'; }",
				Output: []string{"function isStr(x: unknown): x is string {\n return typeof x === 'string'; \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 41},
					{MessageId: "singleLineClose", Line: 1, Column: 73},
				},
			},

			// ---- Decorated method with same-line body ----
			{
				Code:   "class C {\n  @dec m() { return 1; }\n}",
				Output: []string{"class C {\n  @dec m() {\n return 1; \n}\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 2, Column: 12},
					{MessageId: "singleLineClose", Line: 2, Column: 24},
				},
			},

			// ---- For-of body with bad braces ----
			{
				Code:   "for (const x of arr) { foo(x); }",
				Output: []string{"for (const x of arr) {\n foo(x); \n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 22},
					{MessageId: "singleLineClose", Line: 1, Column: 32},
				},
			},

			// ---- While body with bad braces, allman ----
			{
				Code:    "while (x) {\n  foo();\n}",
				Output:  []string{"while (x) \n{\n  foo();\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 11},
				},
			},

			// ---- Do-while body with bad braces, allman ----
			{
				Code:    "do {\n  foo();\n} while (x);",
				Output:  []string{"do \n{\n  foo();\n} while (x);"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 4},
				},
			},

			// ---- Labeled statement block with bad braces, allman ----
			{
				Code:    "loop: {\n  break loop;\n}",
				Output:  []string{"loop: \n{\n  break loop;\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 7},
				},
			},

			// ---- TS namespace with bad single-line body, stroustrup ----
			{
				Code:    "namespace Foo {\n  export const x = 1;\n}",
				Output:  []string{"namespace Foo \n{\n  export const x = 1;\n}"},
				Options: optsAllman(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sameLineOpen", Line: 1, Column: 15},
				},
			},

			// ---- Multiple namespace decls, only one violates ----
			{
				Code:   "namespace A { export const x = 1; }\nnamespace B {\n  export const y = 2;\n}",
				Output: []string{"namespace A {\n export const x = 1; \n}\nnamespace B {\n  export const y = 2;\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "blockSameLine", Line: 1, Column: 13},
					{MessageId: "singleLineClose", Line: 1, Column: 35},
				},
			},
		},
	)
}
