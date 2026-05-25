package member_ordering

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestEdgeCaseProbes covers edge cases not in the main test matrix:
// SemicolonClassElement, function-type-annotation vs method,
// computed/numeric property names, #private alpha sort, sibling constructs,
// per-construct configs in same file, decorated readonly fields, etc.
func TestEdgeCaseProbes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule,
		[]rule_tester.ValidTestCase{
			// --- SemicolonClassElement between members ---
			{
				Code: `
class Foo {
  a: string = "";
  ;
  constructor() {}
  ;
  b() {}
}
`,
			},
			// --- Function type annotation is a field, NOT a method ---
			{
				Code: `
class Foo {
  a: string = "";
  b: () => void;
  constructor() {}
}
`,
			},
			// --- Interface: function type property is a field ---
			{
				Code: `
interface Foo {
  a: string;
  b: () => void;
  c(): void;
}
`,
				Options: map[string]interface{}{
					"default": []interface{}{"field", "method"},
				},
			},
			// --- Computed property name (dynamic key) treated as field ---
			{
				Code: `
const sym = Symbol();
class Foo {
  a: string = "";
  [sym]: string;
  constructor() {}
}
`,
			},
			// --- Numeric property names (natural order) ---
			{
				Code: `
interface Foo {
  1: string;
  2: string;
  10: string;
}
`,
				Options: map[string]interface{}{
					"default": map[string]interface{}{
						"memberTypes": "never",
						"order":       "natural",
					},
				},
			},
			// --- #private in alphabetical sorting (# stripped, sorted as bare name) ---
			{
				Code: `
class Foo {
  #a: string = "";
  #b: string = "";
  #c: string = "";
}
`,
				Options: map[string]interface{}{
					"default": map[string]interface{}{
						"memberTypes": "never",
						"order":       "alphabetically",
					},
				},
			},
			// --- Multiple sibling constructs: each independently valid ---
			{
				Code: `
interface A {
  x: string;
  y(): void;
}
interface B {
  a: string;
  b(): void;
}
`,
				Options: map[string]interface{}{
					"default": []interface{}{"field", "method"},
				},
			},
			// --- classExpressions and classes separate configs in same file ---
			{
				Code: `
class C {
  constructor() {}
  a: string = "";
}
const E = class {
  a: string = "";
  constructor() {}
};
`,
				Options: map[string]interface{}{
					"classes":          []interface{}{"constructor", "field"},
					"classExpressions": []interface{}{"field", "constructor"},
				},
			},
			// --- Optional method in class ---
			{
				Code: `
class Foo {
  a: string = "";
  b?: string;
}
`,
				Options: map[string]interface{}{
					"default": map[string]interface{}{
						"memberTypes":      "never",
						"optionalityOrder": "required-first",
					},
				},
			},
			// --- Getter/setter with same property name ---
			{
				Code: `
class Foo {
  get x() { return 1; }
  set x(v: number) {}
}
`,
				Options: map[string]interface{}{
					"default": []interface{}{
						[]interface{}{"get", "set"},
					},
				},
			},
			// --- Abstract accessor ---
			{
				Code: `
abstract class Foo {
  abstract accessor x: number;
  abstract accessor y: number;
  abstract z(): void;
}
`,
				Options: map[string]interface{}{
					"default": []interface{}{"accessor", "method"},
				},
			},
			// --- Decorated readonly field falls back to decorated-field ---
			{
				Code: `
class Foo {
  @Dec readonly a: string = "";
  b: string = "";
}
`,
				Options: map[string]interface{}{
					"default": []interface{}{"decorated-field", "field"},
				},
			},
			// --- Public static readonly field with full specificity ---
			{
				Code: `
class Foo {
  public static readonly A: string = "";
  public static B: string;
  public readonly C: string = "";
  public D: string = "";
}
`,
				Options: map[string]interface{}{
					"default": []interface{}{
						"public-static-readonly-field",
						"public-static-field",
						"public-instance-readonly-field",
						"public-instance-field",
					},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// --- #private alpha out of order ---
			{
				Code: `
class Foo {
  #b: string = "";
  #a: string = "";
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
			// --- Sibling constructs: second interface wrong order ---
			{
				Code: `
interface A {
  x: string;
  y(): void;
}
interface B {
  b(): void;
  a: string;
}
`,
				Options: map[string]interface{}{
					"default": []interface{}{"field", "method"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectGroupOrder", Line: 8, Column: 3},
				},
			},
			// --- Function type annotation IS field: method before it → error ---
			{
				Code: `
interface Foo {
  c(): void;
  b: () => void;
}
`,
				Options: map[string]interface{}{
					"default": []interface{}{"field", "method"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
				},
			},
			// --- Numeric property names out of natural order ---
			{
				Code: `
interface Foo {
  1: string;
  10: string;
  2: string;
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
			// --- Decorated readonly before non-decorated readonly → error when config reverses ---
			{
				Code: `
class Foo {
  @Dec readonly a: string = "";
  readonly b: string = "";
}
`,
				Options: map[string]interface{}{
					"default": []interface{}{"readonly-field", "decorated-readonly-field"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
				},
			},
		},
	)
}
