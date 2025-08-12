package member_ordering

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestMemberOrderingRule(t *testing.T) {
	validTests := []rule_tester.ValidTestCase{
		// Basic valid cases
		{
			Code: `
interface Foo {
  [Z: string]: any;
  A: string;
  B: string;
  C: string;
  D: string;
  E: string;
  F: string;
  new ();
  G();
  H();
  I();
  J();
  K();
  L();
}`,
		},
		{
			Code: `
interface Foo {
  A: string;
  J();
  K();
  D: string;
  E: string;
  F: string;
  new ();
  G();
  H();
  [Z: string]: any;
  B: string;
  C: string;
  I();
  L();
}`,
			Options: map[string]interface{}{
				"default": "never",
			},
		},
		{
			Code: `
class Foo {
  [Z: string]: any;
  public static A: string;
  protected static B: string = '';
  private static C: string = '';
  static #C: string = '';
  public D: string = '';
  protected E: string = '';
  private F: string = '';
  #F: string = '';
  constructor() {}
  public static G() {}
  protected static H() {}
  private static I() {}
  static #I() {}
  public J() {}
  protected K() {}
  private L() {}
  #L() {}
}`,
		},
		{
			Code: `
interface Foo {
  [Z: string]: any;
  A: string;
  B: string;
  C: string;
  D: string;
  E: string;
  F: string;
  new ();
  G();
  H();
  I();
  J();
  K();
  L();
}`,
			Options: map[string]interface{}{
				"default": []interface{}{"signature", "field", "constructor", "method"},
			},
		},
		{
			// grouped member types
			Code: `
class Foo {
  A: string;
  constructor() {}
  get B() {}
  set B() {}
  get C() {}
  set C() {}
  D(): void;
}`,
			Options: map[string]interface{}{
				"default": []interface{}{
					"field",
					"constructor",
					[]interface{}{"get", "set"},
					"method",
				},
			},
		},
		{
			// optionality order - required first
			Code: `
interface Foo {
  a: string;
  b: string;
  c?: string;
  d?: string;
}`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"optionalityOrder": "required-first",
				},
			},
		},
		{
			// optionality order - optional first
			Code: `
interface Foo {
  a?: string;
  b?: string;
  c: string;
  d: string;
}`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"optionalityOrder": "optional-first",
				},
			},
		},
		{
			// decorated members
			Code: `
class Foo {
  @Dec() B: string;
  @Dec() A: string;
  constructor() {}
  D: string;
  C: string;
  E(): void;
  F(): void;
}`,
			Options: map[string]interface{}{
				"default": []interface{}{"decorated-field", "field"},
			},
		},
		{
			// private identifier members
			Code: `
class Foo {
  imPublic() {}
  #imPrivate() {}
}`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": []interface{}{"public-method", "#private-method"},
					"order":       "alphabetically-case-insensitive",
				},
			},
		},
		{
			// readonly fields
			Code: `
class Foo {
  readonly B: string;
  readonly A: string;
  constructor() {}
  D: string;
  C: string;
  E(): void;
  F(): void;
}`,
			Options: map[string]interface{}{
				"default": []interface{}{"readonly-field", "field"},
			},
		},
		{
			// static initialization blocks
			Code: `
class Foo {
  static {}
  m() {}
  f = 1;
}`,
			Options: map[string]interface{}{
				"default": []interface{}{"static-initialization", "method", "field"},
			},
		},
		{
			// accessor properties
			Code: `
class Foo {
  accessor bar;
  baz() {}
}`,
			Options: map[string]interface{}{
				"default": []interface{}{"accessor", "method"},
			},
		},
		{
			// interface with get/set
			Code: `
interface Foo {
  get x(): number;
  y(): void;
}`,
			Options: map[string]interface{}{
				"default": []interface{}{"get", "method"},
			},
		},
		{
			// type literals
			Code: `
type Foo = {
  [Z: string]: any;
  A: string;
  B: string;
  C: string;
  D: string;
  E: string;
  F: string;
  new ();
  G();
  H();
  I();
  J();
  K();
  L();
};`,
		},
		{
			// readonly signatures
			Code: `
interface Foo {
  readonly [i: string]: string;
  readonly A: string;
  [i: number]: string;
  B: string;
}`,
			Options: map[string]interface{}{
				"default": []interface{}{
					"readonly-signature",
					"readonly-field",
					"signature",
					"field",
				},
			},
		},
		{
			// grouped member types - gets and sets in correct order
			Code: `
class Foo {
  A: string;
  constructor() {}
  get B() {}
  get C() {}
  set B() {}
  set C() {}
  D(): void;
}`,
			Options: map[string]interface{}{
				"default": []interface{}{
					"field",
					"constructor",
					[]interface{}{"get"},
					[]interface{}{"set"},
					"method",
				},
			},
		},
	}

	invalidTests := []rule_tester.InvalidTestCase{
		{
			// incorrect order - constructor before methods
			Code: `
interface Foo {
  [Z: string]: any;
  A: string;
  B: string;
  C: string;
  D: string;
  E: string;
  F: string;
  G();
  H();
  I();
  J();
  K();
  L();
  new ();
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "incorrectGroupOrder",
					Line:      16,
					Column:    3,
				},
			},
		},
		{
			// incorrect order - methods before fields
			Code: `
interface Foo {
  A: string;
  B: string;
  C: string;
  D: string;
  E: string;
  F: string;
  G();
  H();
  I();
  J();
  K();
  L();
  new ();
  [Z: string]: any;
}`,
			Options: map[string]interface{}{
				"default": []interface{}{"signature", "method", "constructor", "field"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 9, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 10, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 11, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 12, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 13, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 14, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 15, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 16, Column: 3},
			},
		},
		{
			// class - public static methods after instance methods
			Code: `
class Foo {
  [Z: string]: any;
  public static A: string = '';
  protected static B: string = '';
  private static C: string = '';
  static #C: string = '';
  public D: string = '';
  protected E: string = '';
  private F: string = '';
  #F: string = '';
  constructor() {}
  public J() {}
  protected K() {}
  private L() {}
  #L() {}
  public static G() {}
  protected static H() {}
  private static I() {}
  static #I() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 17, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 18, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 19, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 20, Column: 3},
			},
		},
		{
			// alphabetical order - incorrect
			Code: `
class Foo {
  static C: boolean;
  [B: string]: any;
  private A() {}
}`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"order": "alphabetically",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
		{
			// alphabetical order within groups
			Code: `
interface Foo {
  B: string;
  A: string;
  C: string;
}`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"order": "alphabetically",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectOrder", Line: 4, Column: 3},
			},
		},
		{
			// optionality order - incorrect
			Code: `
interface Foo {
  a: string;
  b?: string;
  c: string;
  d?: string;
}`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"optionalityOrder": "required-first",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectRequiredMembersOrder", Line: 4, Column: 3},
			},
		},
		{
			// private identifier members - alphabetical order incorrect within group
			Code: `
class Foo {
  z() {}
  a() {}
}`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": []interface{}{"public-method"},
					"order":       "alphabetically-case-insensitive",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectOrder", Line: 4, Column: 3},
			},
		},
		{
			// readonly fields - alphabetical order incorrect
			Code: `
class Foo {
  readonly B: string;
  readonly A: string;
  constructor() {}
  D: string;
  C: string;
  E(): void;
  F(): void;
}`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"memberTypes": []interface{}{"readonly-field", "constructor", "field", "method"},
					"order":       "alphabetically",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectOrder", Line: 4, Column: 3},
				{MessageId: "incorrectOrder", Line: 7, Column: 3},
			},
		},
		{
			// class expressions
			Code: `
const foo = class Foo {
  public static A: string = '';
  protected static B: string = '';
  private static C: string = '';
  public D: string = '';
  protected E: string = '';
  private F: string = '';
  constructor() {}
  public J() {}
  protected K() {}
  private L() {}
  public static G() {}
  protected static H() {}
  private static I() {}
};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 13, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 14, Column: 3},
				{MessageId: "incorrectGroupOrder", Line: 15, Column: 3},
			},
		},
		{
			// abstract members
			Code: `
abstract class Foo {
  abstract A(): void;
  B: string;
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 4, Column: 3},
			},
		},
		{
			// complex member type ordering
			Code: `
class Foo {
  public J() {}
  public static G() {}
  public D: string = '';
  public static A: string = '';
  private L() {}
  constructor() {}
  protected K() {}
  protected static H() {}
  private static I() {}
  protected static B: string = '';
  private static C: string = '';
  protected E: string = '';
  private F: string = '';
}`,
			Options: map[string]interface{}{
				"default": []interface{}{
					"public-method",
					"public-field",
					"constructor",
					"method",
					"field",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectGroupOrder", Line: 8, Column: 3},
			},
		},
		{
			// readonly signatures
			Code: `
interface Foo {
  readonly [i: string]: string;
  readonly A: string;
  [i: number]: string;
  B: string;
}`,
			Options: map[string]interface{}{
				"default": []interface{}{
					"readonly-signature",
					"readonly-field",
					"signature",
					"field",
				},
			},
		},
		{
			// case insensitive alphabetical ordering
			Code: `
class Foo {
  b: string;
  A: string;
  C: string;
}`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"order": "alphabetically-case-insensitive",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectOrder", Line: 4, Column: 3},
			},
		},
		{
			// natural ordering
			Code: `
class Foo {
  a10: string;
  a2: string;
  a1: string;
}`,
			Options: map[string]interface{}{
				"default": map[string]interface{}{
					"order": "natural",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "incorrectOrder", Line: 4, Column: 3},
				{MessageId: "incorrectOrder", Line: 5, Column: 3},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, validTests, invalidTests)
}
