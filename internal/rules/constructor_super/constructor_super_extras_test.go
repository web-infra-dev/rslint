package constructor_super

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConstructorSuperExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConstructorSuperRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Non-derived classes
			{Code: `class A { }`},
			{Code: `class A { constructor() { } }`},

			// Derived classes with proper super() calls
			{Code: `class A extends B { }`},
			{Code: `class A extends B { constructor() { super(); } }`},

			// Classes extending null without constructor
			{Code: `class A extends null { }`},

			// Derived classes with super() in all paths
			{Code: `class A extends B { constructor() { if (true) super(); else super(); } }`},
			{Code: `class A extends B { constructor() { a ? super() : super(); } }`},

			// Early return substitutes for super()
			{Code: `class A extends B { constructor() { if (true) return a; super(); } }`},

			// Switch statements with super in all branches
			{Code: `class A extends B { constructor() { switch (a) { case 0: super(); break; default: super(); break; } } }`},

			// Complex expressions as base class
			{Code: `class A extends (B = C) { constructor() { super(); } }`},
			{Code: `class A extends (B || C) { constructor() { super(); } }`},
			{Code: `class A extends (a ? B : C) { constructor() { super(); } }`},

			// Nested classes (outer has super)
			{Code: `class A extends B { constructor() { super(); class C extends D { constructor() { super(); } } } }`},
			{Code: `class A extends B { constructor() { super(); class C { } } }`},

			// No super needed - no extends
			{Code: `class A { constructor() { class B extends C { constructor() { super(); } } } }`},

			// Throw statements
			{Code: `class A extends B { constructor() { throw new Error(); } }`},

			// If-else with super in both branches followed by other statements
			// (regression: trailing statements must not be required to also call super())
			{Code: `class A extends B { constructor() { if (a) { super(); } else { super(); } b(); } }`},
			{Code: `class A extends B { constructor() { if (a) { c(); super(); d(); } else { super(); } e(); } }`},

			// Switch with super in all cases followed by other statements
			{Code: `class A extends B { constructor() { switch (a) { case 0: super(); break; default: super(); break; } b(); } }`},

			// Always-truthy loops that call super() before the only exit
			{Code: `class A extends B { constructor() { while (true) { super(); break; } } }`},
			{Code: `class A extends B { constructor() { for (;;) { super(); break; } } }`},
			{Code: `class A extends B { constructor() { label: while (true) { super(); break label; } } }`},

			// do-while always runs its body at least once, regardless of the
			// condition, but only counts as a clean single call if it breaks
			// out immediately afterward (otherwise it's a duplicate - see
			// invalid cases below)
			{Code: `class A extends B { constructor() { do { super(); break; } while (true); } }`},
			{Code: `class A extends B { constructor() { do { super(); break; } while (false); } }`},

			// try/finally: finally always runs, so super() there satisfies the
			// requirement regardless of what the try/catch do
			{Code: `class A extends B { constructor() { try { super(); } finally {} } }`},
			{Code: `class A extends B { constructor() { try { a(); } catch (e) { } finally { super(); } } }`},

			// try without catch: an exception before super() propagates out of
			// the constructor, which is an acceptable terminating path
			{Code: `class A extends B { constructor() { try { super(); } finally { b(); } } }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Bare returns do not substitute for super(), even in every branch.
			{Code: `class A extends B { constructor() { if (true) return; super(); } }`, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 65},
			}},
			{Code: `class A extends B { constructor() { if (a) return; if (b) return; super(); } }`, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 77},
			}},
			{Code: `class A extends B { constructor() { if (a) { return; } else { return; } } }`, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 74},
			}},
			// Non-derived classes calling super()
			{
				Code: `class A { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 27},
				},
			},

			// Classes extending null calling super()
			{
				Code: `class A extends null { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 40},
				},
			},

			// Classes extending literals calling super()
			{
				Code: `class A extends 100 { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 39},
				},
			},
			{
				Code: `class A extends "test" { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 42},
				},
			},

			// Derived classes missing super() entirely
			{
				Code: `class A extends B { constructor() { } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingAll", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A extends B { constructor() { a; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingAll", Line: 1, Column: 21},
				},
			},

			// Nested classes where outer lacks super()
			{
				Code: `class A extends B { constructor() { class C extends D { constructor() { super(); } } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingAll", Line: 1, Column: 21},
				},
			},

			// Missing super() in some code paths
			{
				Code: `class A extends B { constructor() { if (a) super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A extends B { constructor() { a && super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A extends B { constructor() { if (a) return; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 53},
				},
			},
			{
				Code: `class A extends B { constructor() { switch (a) { case 0: super(); break; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A extends B { constructor() { switch (a) { case 0: super(); break; case 1: break; default: super(); break; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},

			// Duplicate super() calls
			{
				Code: `class A extends B { constructor() { super(); super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 46},
				},
			},
			{
				Code: `class A extends B { constructor() { super(); if (a) super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 53},
				},
			},
			{
				Code: `class A extends B { constructor() { if (a) super(); else super(); super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 67},
				},
			},
			// Calls in later branch statements are each duplicates once an
			// earlier branch has guaranteed the first super() call.
			{
				Code: `class A extends B { constructor() { if (a) { super(); } else { super(); } if (b) { super(); } else { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate"},
					{MessageId: "duplicate"},
				},
			},
			// The loop call is a duplicate both because super() was already
			// called and because the loop can revisit it; report it only once.
			{
				Code: `class A extends B { constructor() { super(); while (true) { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate"},
				},
			},

			// super() inside a loop without a guaranteed break right after it
			// can be revisited on a later iteration - a duplicate, regardless
			// of whether the loop's condition provably allows only one pass
			{
				Code: `class A extends B { constructor() { while (true) { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 52},
				},
			},
			{
				Code: `class A extends B { constructor() { while (true) { super(); continue; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 52},
				},
			},
			{
				Code: `class A extends B { constructor() { do { super(); } while (true); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 42},
				},
			},
			{
				Code: `class A extends B { constructor() { do { super(); } while (false); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 42},
				},
			},
			{
				Code: `class A extends B { constructor() { for (;;) { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 48},
				},
			},

			// finally always runs after the try block, so a second super()
			// call there duplicates the try block's call
			{
				Code: `class A extends B { constructor() { try { super(); } finally { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 64},
				},
			},
			// a trailing super() after a try that already satisfied the
			// requirement is a duplicate too
			{
				Code: `class A extends B { constructor() { try { super(); } finally { } super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 66},
				},
			},

			// Invalid extends expressions with super() call
			{
				Code: `class A extends (B = 5) { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 43},
				},
			},
			{
				Code: `class A extends (B += C) { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 44},
				},
			},
			{
				Code: `class A extends (B -= C) { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 44},
				},
			},
			{
				Code: `class A extends (B *= C) { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 44},
				},
			},
			{
				Code: `class A extends (B /= C) { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 44},
				},
			},

			// If-else without all paths having super
			{
				Code: `class A extends B { constructor() { if (a) { super(); } else { } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},
			{
				Code: `class A extends B { constructor() { if (a) { } else { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},

			// Ternary expression without super in one branch
			{
				Code: `class A extends B { constructor() { if (a) super(); else if (b) super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},

			// Switch without default
			{
				Code: `class A extends B { constructor() { switch (a) { case 0: super(); break; case 1: super(); break; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},

			// Multiple super calls in switch
			{
				Code: `class A extends B { constructor() { switch (a) { case 0: super(); break; default: super(); break; } super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "duplicate", Line: 1, Column: 101},
				},
			},

			// Loops that may exit without ever calling super()
			{
				Code: `class A extends B { constructor() { while (true) { if (a) break; super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 78},
					{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 66, EndLine: 1, EndColumn: 73},
				},
			},
			{
				Code: `class A extends B { constructor() { for (var x of y) { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 68},
					{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 56, EndLine: 1, EndColumn: 63},
				},
			},
			{
				// ESLint v10.8.1 reports only missingAll here; keep the approved
				// missingSome + duplicate behavior described in the rule documentation.
				Code: `class A extends B { constructor() { for (var i = 0; i < 10; i++) { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 80},
					{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 68, EndLine: 1, EndColumn: 75},
				},
			},
			{
				Code: `class A extends B { constructor() { while (a) { super(); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 61},
					{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 49, EndLine: 1, EndColumn: 56},
				},
			},

			// try/catch without finally: catch must independently call super() too
			{
				Code: `class A extends B { constructor() { try { super(); } catch (e) {} } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingSome", Line: 1, Column: 21},
				},
			},

			// Combination of errors
			{
				Code: `class A extends B { constructor() { super(); } } class C extends null { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "badSuper", Line: 1, Column: 89},
				},
			},
		},
	)
}

// Additional expression, catch/finally, scope and nested-loop regressions,
// checked against ESLint v10.8.1 with complete diagnostic ranges.
func TestConstructorSuperRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConstructorSuperRule,
		[]rule_tester.ValidTestCase{
			{Code: "class A extends function(){}{constructor(){super();}}"},
			{Code: "class A extends this{constructor(){super();}}"},
			{Code: "class A extends new C{constructor(){super();}}"},
			{Code: "class A extends (0, B){constructor(){super();}}"},
			{Code: "class A extends (B || 0){constructor(){super();}}"},
			{Code: "class A extends (0 || B){constructor(){super();}}"},
			{Code: "class A extends (B ?? 0){constructor(){super();}}"},
			{Code: "class A extends (0 && B){constructor(){super();}}"},
			{Code: "class A extends (a ? 0 : B){constructor(){super();}}"},
			{Code: "class A extends (B ||= 0){constructor(){super();}}"},
			{Code: "class A extends (B ??= 0){constructor(){super();}}"},
			{Code: "class A extends obj?.(){constructor(){super();}}"},
			{Code: "class A extends B{constructor(){(super());}}"},
			{Code: "class A extends B{constructor(){foo(super());}}"},
			{Code: "class A extends B{constructor(){a?super():super();}}"},
			{Code: "class A extends B{constructor(){if(a)return {};else super();}}"},
			{Code: "class A extends B{constructor(){throw a;super();}}"},
			{Code: "class A extends B{constructor(){while(a){if(b){super();return;}}super();}}"},
			{Code: "class A extends B{constructor(){for(const x of xs){if(a){super();return;}}super();}}"},
			{Code: "class A extends B{constructor(){switch(a){case 0:case 1:super();break;default:return {};}}}"},
			{Code: "class A extends B{constructor(){try{return;}finally{super();}}}"},
			{Code: "class A extends B{constructor(){try{throw 0;}finally{super();}}}"},
			{Code: "class A extends B{constructor(){class C extends (super(), B){};}}"},
			{Code: "class A extends B{constructor(){const obj={ [super()]: 1 };}}"},
			{Code: "class A extends B{constructor(){const obj={ [super()](){} };}}"},
			{Code: "class A extends B{constructor(){class C { [super()](){} };}}"},
			{Code: "class A extends B{constructor(){const arrow=()=>super();super();}}"},
			{Code: "class A extends B{constructor(){(super());}}"},
			{Code: "class A extends B{constructor(){foo(super());}}"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "class A extends undefined{constructor(){super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 41, EndLine: 1, EndColumn: 48},
			}},
			{Code: "class A extends true{constructor(){super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 36, EndLine: 1, EndColumn: 43},
			}},
			{Code: "class A extends 123n{constructor(){super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 36, EndLine: 1, EndColumn: 43},
			}},
			{Code: "class A extends /a/{constructor(){super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 35, EndLine: 1, EndColumn: 42},
			}},
			{Code: "class A extends `x`{constructor(){super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 35, EndLine: 1, EndColumn: 42},
			}},
			{Code: "class A extends (B, 0){constructor(){super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 38, EndLine: 1, EndColumn: 45},
			}},
			{Code: "class A extends (B && 0){constructor(){super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 40, EndLine: 1, EndColumn: 47},
			}},
			{Code: "class A extends (B &&= 0){constructor(){super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 41, EndLine: 1, EndColumn: 48},
			}},
			{Code: "class A extends B{constructor(){super(super());}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 33, EndLine: 1, EndColumn: 47},
			}},
			{Code: "class A extends B{constructor(){return super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 40, EndLine: 1, EndColumn: 47},
			}},
			{Code: "class A extends B{constructor(){super(),super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 41, EndLine: 1, EndColumn: 48},
			}},
			{Code: "class A extends B{constructor(){super()&&super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 42, EndLine: 1, EndColumn: 49},
			}},
			{Code: "class A extends B{constructor(){a?super():0;super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 45, EndLine: 1, EndColumn: 52},
			}},
			{Code: "class A extends B{constructor(){if(a){super();if(b)super();}else super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 52, EndLine: 1, EndColumn: 59},
			}},
			{Code: "class A extends B{constructor(){if(a)return;else super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 19, EndLine: 1, EndColumn: 59},
			}},
			{Code: "class A extends B{constructor(){return;super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 49},
			}},
			{Code: "class A extends B{constructor(){if(a)throw 0;}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 47},
			}},
			{Code: "class A extends B{constructor(){if(a)return {};}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 19, EndLine: 1, EndColumn: 49},
			}},
			{Code: "class A extends B{constructor(){while(a){if(b){super();break;}}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 19, EndLine: 1, EndColumn: 65},
			}},
			{Code: "class A extends B{constructor(){do{super();continue;}while(a);}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 36, EndLine: 1, EndColumn: 43},
			}},
			{Code: "class A extends B{constructor(){do{if(a)break;super();}while(b);}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 19, EndLine: 1, EndColumn: 66},
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 47, EndLine: 1, EndColumn: 54},
			}},
			{Code: "class A extends B{constructor(){outer:while(a){super();break outer;}super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 69, EndLine: 1, EndColumn: 76},
			}},
			{Code: "class A extends B{constructor(){outer:while(a){while(b){super();continue outer;}}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 19, EndLine: 1, EndColumn: 83},
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 57, EndLine: 1, EndColumn: 64},
			}},
			{Code: "class A extends B{constructor(){switch(a){case 0:super();case 1:break;default:super();}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 19, EndLine: 1, EndColumn: 89},
			}},
			{Code: "class A extends B{constructor(){switch(a){default:super();case 1:super();}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 66, EndLine: 1, EndColumn: 73},
			}},
			{Code: "class A extends B{constructor(){try{super();}catch(e){super();}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 55, EndLine: 1, EndColumn: 62},
			}},
			{Code: "class A extends B{constructor(){try{super();foo();}catch(e){super();}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 61, EndLine: 1, EndColumn: 68},
			}},
			{Code: "class A extends B{constructor(){try{foo();}finally{if(a)return;}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 66},
			}},
			{Code: "class A extends B{constructor(){try{}catch(e){super();}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 19, EndLine: 1, EndColumn: 57},
			}},
			{Code: "class A extends B{constructor(){try{let x;}catch(e){super();}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 19, EndLine: 1, EndColumn: 63},
			}},
			{Code: "class A extends B{constructor(){try{let x=0;}catch(e){super();}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 19, EndLine: 1, EndColumn: 65},
			}},
			{Code: "class A extends B{constructor(){try{if(a)super();}catch(e){}finally{super();}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 69, EndLine: 1, EndColumn: 76},
			}},
			{Code: "class A extends B{constructor(){try{try{super();}catch(e){}}catch(e){super();}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 19, EndLine: 1, EndColumn: 80},
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 70, EndLine: 1, EndColumn: 77},
			}},
			{Code: "class A extends B{constructor(){try{try{super();}finally{foo();}}catch(e){super();}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 75, EndLine: 1, EndColumn: 82},
			}},
			{Code: "class A extends B{constructor(){super(super());}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 33, EndLine: 1, EndColumn: 47},
			}},
			{Code: "class A extends B{constructor(){return super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 40, EndLine: 1, EndColumn: 47},
			}},
			{Code: "class A extends B{constructor(){super(),super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 41, EndLine: 1, EndColumn: 48},
			}},
			{Code: "class A extends null{constructor(){while(a)super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 22, EndLine: 1, EndColumn: 53},
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 44, EndLine: 1, EndColumn: 51},
			}},
			{Code: "class A extends null{constructor(){do{super();}while(a);}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 39, EndLine: 1, EndColumn: 46},
			}},
			{Code: "class A extends null{constructor(){super();while(a)super();}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 36, EndLine: 1, EndColumn: 43},
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 52, EndLine: 1, EndColumn: 59},
			}},
			{Code: "class A extends null{constructor(){try{if(a)return {};else return;}finally{super();}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 76, EndLine: 1, EndColumn: 83},
			}},
			{Code: "class A extends B{constructor(){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){while(a){super();}}}}}}}}}}}}}}}}}}}}}}}}}}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 19, EndLine: 1, EndColumn: 282},
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 249, EndLine: 1, EndColumn: 256},
			}},
		},
	)
}
