package member_ordering

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestMemberOrdering_GroupOrder tests basic member type group ordering
// across all four construct types (class, classExpression, interface, typeLiteral).
func TestMemberOrdering_GroupOrder(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		// --- Interface: default order ---
		{
			Code: `
interface Foo {
  [Z: string]: any;
  A: string;
  B: string;
  new ();
  G();
  H();
}
`,
		},
		// --- Interface: custom order ---
		{
			Code: `
interface Foo {
  [Z: string]: any;
  A: string;
  new ();
  G();
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"signature", "field", "constructor", "method"},
			},
		},
		// --- Interface: call-signature before field before method ---
		{
			Code: `
interface X {
  (): void;
  a: unknown;
  b(): void;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"call-signature", "field", "method"},
			},
		},
		// --- Type literal ---
		{
			Code: `
type Foo = {
  [Z: string]: any;
  A: string;
  (): void;
  C(): void;
};
`,
			Options: map[string]interface{}{
				"default": []interface{}{"signature", "field", "call-signature", "method"},
			},
		},
		// --- Class: full default order (static→instance→abstract fields, then constructor, then methods) ---
		{
			Code: `
class Foo {
  public static A: string;
  protected static B: string = "";
  private static C: string = "";
  public D: string = "";
  protected E: string = "";
  private F: string = "";
  constructor() {}
  public J() {}
  protected K() {}
  private L() {}
}
`,
		},
		// --- Class expression ---
		{
			Code: `
const foo = class {
  public static A: string;
  public B: string = "";
  constructor() {}
  public J() {}
};
`,
		},
		// --- Static block between fields and constructor ---
		{
			Code: `
class Foo {
  static A: string;
  static {}
  constructor() {}
  B() {}
}
`,
		},
		// --- Empty bodies ---
		{Code: `interface Empty {}`},
		{Code: `class Empty {}`},
		{Code: `type Empty = {};`},
		{Code: `const e = class {};`},
		// --- Single member ---
		{Code: `interface One { a: string; }`},
		{Code: `class One { a: string = ""; }`},
		// --- 'never' disables all checks ---
		{
			Code: `
interface Foo {
  G();
  A: string;
  [Z: string]: any;
}
`,
			Options: map[string]interface{}{"default": "never"},
		},
		// --- Unmatched member types in config: members not in config are not checked ---
		{
			Code: `
interface Foo {
  G();
  A: string;
  [Z: string]: any;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{
					"private-instance-method",
					"public-constructor",
					"protected-static-field",
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// --- Interface: field before signature (default) ---
		{
			Code: `
interface Foo {
  A: string;
  [Z: string]: any;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
		// --- Interface: method before field ---
		{
			Code: `
interface Foo {
  G();
  A: string;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
		// --- Interface: field before signature with custom order ---
		{
			Code: `
interface Foo {
  A: string;
  B: string;
  [Z: string]: any;
  G();
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"signature", "field", "method"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 5, Column: 3},
			},
		},
		// --- Class: method before field (default) ---
		{
			Code: `
class Foo {
  public G() {}
  public A: string = "";
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
		// --- Class: constructor before field ---
		{
			Code: `
class Foo {
  constructor() {}
  public A: string = "";
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
		// --- Class: method before constructor ---
		{
			Code: `
class Foo {
  public A: string = "";
  public G() {}
  constructor() {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 5, Column: 3},
			},
		},
		// --- Class expression: method before field ---
		{
			Code: `
const foo = class {
  public G() {}
  public A: string = "";
};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
		// --- Type literal: method before field ---
		{
			Code: `
type Foo = {
  G(): void;
  A: string;
};
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
		// --- call-signature wrong order ---
		{
			Code: `
interface X {
  a: unknown;
  (): void;
  b(): void;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"call-signature", "field", "method"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
		// --- Multiple group order errors ---
		{
			Code: `
interface Foo {
  G();
  A: string;
  [Z: string]: any;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"signature", "field", "method"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 5, Column: 3},
			},
		},
	})
}

// TestMemberOrdering_PerConstructConfig tests that each construct type
// can have its own independent configuration, with fallback to default.
func TestMemberOrdering_PerConstructConfig(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		// --- interfaces config overrides default ---
		{
			Code: `
interface Foo {
  G();
  H();
  new ();
  A: string;
  [Z: string]: any;
}
`,
			Options: map[string]interface{}{
				"default":    []interface{}{"signature", "field", "constructor", "method"},
				"interfaces": []interface{}{"method", "constructor", "field", "signature"},
			},
		},
		// --- interfaces: 'never' ignores interface ordering ---
		{
			Code: `
interface Foo {
  A: string;
  J();
  [Z: string]: any;
  new ();
}
`,
			Options: map[string]interface{}{"interfaces": "never"},
		},
		// --- classes config overrides default ---
		{
			Code: `
class Foo {
  constructor() {}
  public A: string = "";
}
`,
			Options: map[string]interface{}{
				"classes": []interface{}{"constructor", "field"},
			},
		},
		// --- classExpressions config separate from classes ---
		{
			Code: `
const foo = class {
  bar() {}
  x: string = "";
};
`,
			Options: map[string]interface{}{
				"classExpressions": []interface{}{"method", "field"},
			},
		},
		// --- typeLiterals config ---
		{
			Code: `
type T = {
  b(): void;
  a: string;
};
`,
			Options: map[string]interface{}{
				"typeLiterals": []interface{}{"method", "field"},
			},
		},
		// --- classExpressions uses default when not specified ---
		{
			Code: `
const foo = class {
  x: string = "";
  bar() {}
};
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// --- class uses default when classes not specified ---
		{
			Code: `
class Foo {
  bar() {}
  x: string = "";
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
	})
}

// TestMemberOrdering_AccessibilityAndScope tests accessibility modifiers
// (public/protected/private/#private) and scope (static/instance/abstract).
func TestMemberOrdering_AccessibilityAndScope(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		// --- #private fields come after public fields in default order ---
		{
			Code: `
class Foo {
  public A: string = "";
  #B: string = "";
  constructor() {}
}
`,
		},
		// --- Explicit accessibility ordering ---
		{
			Code: `
class Foo {
  public a: string = "";
  protected b: string = "";
  private c: string = "";
  #d: string = "";
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"public-field", "protected-field", "private-field", "#private-field"},
			},
		},
		// --- Static fields before instance fields ---
		{
			Code: `
class Foo {
  static a: string;
  b: string = "";
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"static-field", "instance-field"},
			},
		},
		// --- Abstract members in abstract class ---
		{
			Code: `
abstract class Foo {
  public static A: string;
  public B: string = "";
  public abstract C: string;
  constructor() {}
  public D() {}
  public abstract E(): void;
}
`,
		},
		// --- Specific scope ordering: static → instance → abstract ---
		{
			Code: `
abstract class Foo {
  static a: string;
  b: string = "";
  abstract c: string;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"static-field", "instance-field", "abstract-field"},
			},
		},
		// --- #private static field ---
		{
			Code: `
class Foo {
  public static A: string;
  static #B: string = "";
  constructor() {}
}
`,
		},
	}, []rule_tester.InvalidTestCase{
		// --- private field before public field ---
		{
			Code: `
class Foo {
  private a: string = "";
  public b: string = "";
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"public-field", "private-field"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
		// --- instance field before static field ---
		{
			Code: `
class Foo {
  b: string = "";
  static a: string;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"static-field", "instance-field"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
		// --- Abstract method before field (default order) ---
		{
			Code: `
abstract class Foo {
  abstract A(): void;
  B: string;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
	})
}

// TestMemberOrdering_Decorated tests decorated members (@Decorator).
func TestMemberOrdering_Decorated(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		// --- Decorated field before instance field ---
		{
			Code: `
class Foo {
  public static A: string;
  @Dec public B: string = "";
  public C: string = "";
  constructor() {}
}
`,
		},
		// --- Explicit decorated ordering ---
		{
			Code: `
class Foo {
  @Dec a: string = "";
  b: string = "";
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"decorated-field", "field"},
			},
		},
		// --- Decorated method ---
		{
			Code: `
class Foo {
  a: string = "";
  @Dec b() {}
  c() {}
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "decorated-method", "method"},
			},
		},
		// --- Decorated get/set ---
		{
			Code: `
class Foo {
  @Dec get a() { return 1; }
  get b() { return 2; }
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"decorated-get", "get"},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// --- Non-decorated field before decorated field ---
		{
			Code: `
class Foo {
  b: string = "";
  @Dec a: string = "";
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"decorated-field", "field"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
	})
}

// TestMemberOrdering_ReadonlyFallback tests that readonly-field and
// readonly-signature fall back to field and signature when not in config.
func TestMemberOrdering_ReadonlyFallback(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		// --- Explicit readonly-field ordering ---
		{
			Code: `
interface Foo {
  readonly A: string;
  B: string;
  C(): void;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"readonly-field", "field", "method"},
			},
		},
		// --- readonly-field falls back to field ---
		{
			Code: `
interface Foo {
  readonly A: string;
  B: string;
  C(): void;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
		},
		// --- Explicit readonly-signature ordering ---
		{
			Code: `
interface Foo {
  readonly [key: string]: any;
  [key: number]: string;
  A: string;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"readonly-signature", "signature", "field"},
			},
		},
		// --- readonly-signature falls back to signature ---
		{
			Code: `
interface Foo {
  readonly [key: string]: any;
  A: string;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"signature", "field"},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// --- readonly-field before non-readonly when only field in config ---
		{
			Code: `
interface Foo {
  C(): void;
  readonly A: string;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
	})
}

// TestMemberOrdering_OverloadSignatures tests that TypeScript method/constructor
// overload signatures (body-less declarations) are skipped.
func TestMemberOrdering_OverloadSignatures(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		// --- Method overloads are skipped ---
		{
			Code: `
class Foo {
  A: string = "";
  foo(): void;
  foo(a: string): void;
  foo(a?: string) {}
  bar() {}
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
		},
		// --- Constructor overloads are skipped ---
		{
			Code: `
class Foo {
  A: string = "";
  constructor(a: string);
  constructor(a: string, b: string);
  constructor(a: string, b?: string) {}
  bar() {}
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "constructor", "method"},
			},
		},
		// --- Decorated method with overloads ---
		{
			Code: `
class Foo {
  A: string = "";
  baz(): void;
  @Dec() baz() {}
  bar() {}
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
		},
		// --- Abstract methods are NOT overload signatures (they have no body but still count) ---
		{
			Code: `
abstract class Foo {
  A: string;
  abstract B(): void;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// --- Abstract method out of order is detected (not skipped) ---
		{
			Code: `
abstract class Foo {
  abstract B(): void;
  A: string;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
	})
}

// TestMemberOrdering_FunctionExpressionAsMethod tests that PropertyDeclarations
// initialized with FunctionExpression or ArrowFunction are classified as methods.
func TestMemberOrdering_FunctionExpressionAsMethod(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		// --- Arrow function property treated as method ---
		{
			Code: `
class Foo {
  A: string = "";
  B = () => {};
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
		},
		// --- Function expression property treated as method ---
		{
			Code: `
class Foo {
  A: string = "";
  B = function() {};
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// --- Arrow function property (method) before field ---
		{
			Code: `
class Foo {
  B = () => {};
  A: string = "";
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
	})
}

// TestMemberOrdering_GroupedMemberTypes tests when member types are grouped
// in arrays (same rank), e.g., [["get", "set"], "method"].
func TestMemberOrdering_GroupedMemberTypes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		// --- get/set grouped at same rank, before method ---
		{
			Code: `
class Foo {
  get A() { return ""; }
  set B(val: string) {}
  C() {}
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{
					[]interface{}{"get", "set"},
					"method",
				},
			},
		},
		// --- get/set interleaved (same rank) ---
		{
			Code: `
class Foo {
  set B(val: string) {}
  get A() { return ""; }
  C() {}
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{
					[]interface{}{"get", "set"},
					"method",
				},
			},
		},
		// --- field + readonly-field grouped ---
		{
			Code: `
interface Foo {
  readonly B: string;
  A: string;
  C(): void;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{
					[]interface{}{"field", "readonly-field"},
					"method",
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// --- method before grouped get/set ---
		{
			Code: `
class Foo {
  C() {}
  get A() { return ""; }
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{
					[]interface{}{"get", "set"},
					"method",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
	})
}

// TestMemberOrdering_AlphaSort tests all alphabetical/natural ordering modes.
func TestMemberOrdering_AlphaSort(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		// --- alphabetically (case-sensitive: uppercase before lowercase) ---
		{
			Code: `
interface Foo {
  A: string;
  B: string;
  a: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": "never",
					"order":       "alphabetically",
				},
			},
		},
		// --- alphabetically-case-insensitive ---
		{
			Code: `
interface Foo {
  a: string;
  B: string;
  c: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": "never",
					"order":       "alphabetically-case-insensitive",
				},
			},
		},
		// --- natural: numeric segments compared by value ---
		{
			Code: `
interface Foo {
  a1: string;
  a2: string;
  a10: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": "never",
					"order":       "natural",
				},
			},
		},
		// --- natural-case-insensitive ---
		{
			Code: `
interface Foo {
  a1: string;
  A2: string;
  a10: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": "never",
					"order":       "natural-case-insensitive",
				},
			},
		},
		// --- alphabetically within groups ---
		{
			Code: `
interface Foo {
  [Z: string]: any;
  a: string;
  b: string;
  c(): void;
  d(): void;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": []interface{}{"signature", "field", "method"},
					"order":       "alphabetically",
				},
			},
		},
		// --- as-written: no alpha enforcement ---
		{
			Code: `
interface Foo {
  b: string;
  a: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": "never",
					"order":       "as-written",
				},
			},
		},
		// --- Same name members: no error ---
		{
			Code: `
interface Foo {
  a: string;
  a: number;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": "never",
					"order":       "alphabetically",
				},
			},
		},
		// --- String literal property names (alphabetical by unquoted value) ---
		{
			Code: `
interface Foo {
  'a.b': string;
  'a.c': string;
  'b.a': string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": "never",
					"order":       "alphabetically",
				},
			},
		},
		// --- natural: a < a1 ---
		{
			Code: `
interface Foo {
  a: string;
  a1: string;
  b: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": "never",
					"order":       "natural",
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// --- alphabetically violation ---
		{
			Code: `
interface Foo {
  b: string;
  a: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": "never",
					"order":       "alphabetically",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectOrder", Line: 4, Column: 3},
			},
		},
		// --- case-insensitive violation (B > a case-insensitively) ---
		{
			Code: `
interface Foo {
  B: string;
  a: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": "never",
					"order":       "alphabetically-case-insensitive",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectOrder", Line: 4, Column: 3},
			},
		},
		// --- natural violation: a10 before a2 ---
		{
			Code: `
interface Foo {
  a1: string;
  a10: string;
  a2: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": "never",
					"order":       "natural",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectOrder", Line: 5, Column: 3},
			},
		},
		// --- alpha within group violation ---
		{
			Code: `
interface Foo {
  [Z: string]: any;
  b: string;
  a: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": []interface{}{"signature", "field", "method"},
					"order":       "alphabetically",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectOrder", Line: 5, Column: 3},
			},
		},
		// --- alpha violation in class ---
		{
			Code: `
class Foo {
  b: string = "";
  a: string = "";
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": "never",
					"order":       "alphabetically",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectOrder", Line: 4, Column: 3},
			},
		},
		// --- String literal property names ---
		{
			Code: `
interface Foo {
  'b.d': Foo;
  'b.c': Foo;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": "never",
					"order":       "alphabetically",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectOrder", Line: 4, Column: 3},
			},
		},
		// --- Group sort fails, but alpha sort within type groups still runs ---
		{
			Code: `
interface Foo {
  b(): void;
  a(): void;
  B: string;
  A: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": []interface{}{"field", "method"},
					"order":       "alphabetically",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				// group error: field B after method
				{MessageId: "incorrectGroupOrder", Line: 5, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 6, Column: 3},
				// alpha error within methods: a before b
				{MessageId: "incorrectOrder", Line: 4, Column: 3},
				// alpha error within fields: A before B
				{MessageId: "incorrectOrder", Line: 6, Column: 3},
			},
		},
	})
}

// TestMemberOrdering_Optionality tests optionalityOrder (required-first / optional-first).
func TestMemberOrdering_Optionality(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		// --- required-first: required then optional ---
		{
			Code: `
interface Foo {
  a: string;
  b: string;
  c?: string;
  d?: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "required-first",
				},
			},
		},
		// --- optional-first: optional then required ---
		{
			Code: `
interface Foo {
  a?: string;
  b?: string;
  c: string;
  d: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "optional-first",
				},
			},
		},
		// --- All same optionality: no error ---
		{
			Code: `
interface X {
  b?: string;
  c?: string;
  d?: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "required-first",
					"order":            "as-written",
				},
			},
		},
		// --- Static blocks and index signatures are always non-optional ---
		{
			Code: `
class X {
  a: string;
  b: string;
  static {}
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "required-first",
					"order":            "alphabetically",
				},
			},
		},
		// --- Index signature + optionality ---
		{
			Code: `
class X {
  a: string;
  [i: number]: string;
  b?: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "required-first",
					"order":            "alphabetically",
				},
			},
		},
		// --- Call/construct signatures are always non-optional ---
		{
			Code: `
interface X {
  a: string;
  (a: number): string;
  new (i: number): string;
  b?: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "required-first",
					"order":            "alphabetically",
				},
			},
		},
		// --- required-first + alphabetically within partitions ---
		{
			Code: `
interface X {
  c: string;
  d: string;
  a?: string;
  b?: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "required-first",
					"order":            "alphabetically",
				},
			},
		},
		// --- optional-first + computed property optional ---
		{
			Code: `
class X {
  ['c']?: string;
  a: string;
  b: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "optional-first",
					"order":            "alphabetically",
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// --- required-first: optional before required ---
		{
			Code: `
interface Foo {
  a?: string;
  b: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "required-first",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectRequiredMembersOrder", Line: 3, Column: 3},
			},
		},
		// --- optional-first: required before optional ---
		{
			Code: `
interface Foo {
  a: string;
  b?: string;
  c: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "optional-first",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectRequiredMembersOrder", Line: 3, Column: 3},
			},
		},
		// --- required-first + alpha: alpha error within optional partition ---
		{
			Code: `
interface X {
  m: string;
  d?: string;
  b?: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "required-first",
					"order":            "alphabetically",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectOrder", Line: 5, Column: 3},
			},
		},
		// --- required-first with memberTypes: optionality violation on boundary ---
		{
			Code: `
interface X {
  a: string;
  b?: string;
  c: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      []interface{}{"call-signature", "field", "method"},
					"optionalityOrder": "required-first",
					"order":            "as-written",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectRequiredMembersOrder", Line: 4, Column: 3},
			},
		},
		// --- optional-first: static block treated as non-optional causing violation ---
		{
			Code: `
class X {
  a?: string;
  static {}
  b?: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "required-first",
					"order":            "as-written",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectRequiredMembersOrder", Line: 3, Column: 3},
			},
		},
		// --- optional-first: multiple switches → report first switch ---
		{
			Code: `
class Test {
  a?: string;
  b?: string;
  f: string;
  c?: string;
  d?: string;
  g: string;
  h: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "optional-first",
					"order":            "as-written",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectRequiredMembersOrder", Line: 5, Column: 3},
			},
		},
		// --- optional-first: first member is required → violation ---
		{
			Code: `
class Test {
  a: string;
  b: string;
  f?: string;
  c?: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "optional-first",
					"order":            "as-written",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectRequiredMembersOrder", Line: 3, Column: 3},
			},
		},
		// --- required-first + natural-case-insensitive: alpha error ---
		{
			Code: `
class X {
  b: string;
  a: string;
}
`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes":      "never",
					"optionalityOrder": "required-first",
					"order":            "natural-case-insensitive",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectOrder", Line: 4, Column: 3},
			},
		},
	})
}

// TestMemberOrdering_GetterSetter tests getter/setter specific behavior.
func TestMemberOrdering_GetterSetter(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		// --- Getters before setters before methods ---
		{
			Code: `
class Foo {
  get a() { return 1; }
  get b() { return 2; }
  set c(v: number) {}
  d() {}
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"get", "set", "method"},
			},
		},
		// --- Interface with getter/setter signatures ---
		{
			Code: `
interface Foo {
  get a(): number;
  set b(v: number);
  c(): void;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"get", "set", "method"},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// --- Setter before getter ---
		{
			Code: `
class Foo {
  set b(v: number) {}
  get a() { return 1; }
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"get", "set"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
	})
}

// TestMemberOrdering_Nested tests that nested class/interface/type definitions
// are independently validated.
func TestMemberOrdering_Nested(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		// --- Inner class valid even though outer has different order ---
		{
			Code: `
class Outer {
  a: string = "";
  constructor() {}
  b() {
    const inner = class {
      x: string = "";
      y() {}
    };
  }
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "constructor", "method"},
			},
		},
		// --- Type literal inside interface ---
		{
			Code: `
interface Outer {
  [key: string]: any;
  a: string;
  b: {
    [key: string]: any;
    x: string;
    y(): void;
  };
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"signature", "field", "method"},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// --- Inner class has wrong order (detected independently) ---
		{
			Code: `
class Outer {
  a: string = "";
  b() {
    const inner = class {
      y() {}
      x: string = "";
    };
  }
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 7, Column: 7},
			},
		},
		// --- Inner type literal has wrong order ---
		{
			Code: `
interface Outer {
  [key: string]: any;
  a: {
    y(): void;
    x: string;
  };
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "method"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 6, Column: 5},
			},
		},
	})
}

// TestMemberOrdering_AutoAccessor tests the ES2022 auto accessor syntax.
func TestMemberOrdering_AutoAccessor(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		// --- Auto accessor before method ---
		{
			Code: `
class Foo {
  accessor a = 1;
  b() {}
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"accessor", "method"},
			},
		},
		// --- Accessor ordering with scope ---
		{
			Code: `
class Foo {
  static accessor a = 1;
  accessor b = 2;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"static-accessor", "instance-accessor"},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// --- Method before accessor ---
		{
			Code: `
class Foo {
  b() {}
  accessor a = 1;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"accessor", "method"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
	})
}

// TestMemberOrdering_ConstructSignature tests construct signatures (new ()).
func TestMemberOrdering_ConstructSignature(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		// --- construct signature as constructor type ---
		{
			Code: `
interface Foo {
  [Z: string]: any;
  A: string;
  new (): Foo;
  B(): void;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"signature", "field", "constructor", "method"},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// --- construct signature in wrong position ---
		{
			Code: `
interface Foo {
  B(): void;
  new (): Foo;
  A: string;
}
`,
			Options: map[string]interface{}{
				"default": []interface{}{"field", "constructor", "method"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 5, Column: 3},
			},
		},
	})
}
