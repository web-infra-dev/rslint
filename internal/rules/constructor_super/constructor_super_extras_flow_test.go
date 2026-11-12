package constructor_super

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

// Regression expectations follow execution semantics; differences from ESLint
// v10.8.1 are recorded in constructor_super.md. Inputs have inverse controls
// for returns, bindings, exception timing, loop targets and short-circuiting.
func TestConstructorSuperExecutionRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConstructorSuperRule,
		[]rule_tester.ValidTestCase{
			// Class-definition effects and safe conversion controls.
			{Code: "class A extends B { constructor() { try { class C { value = fail(); static value = 0; } return {}; } catch { return; } } }"},
			{Code: "class A extends B { constructor() { try { const C = class { static value = 0; }; return {}; } catch { return; } } }"},
			{Code: "class A extends B { constructor() { try { class C { static {} static value; static method() { fail(); } } return {}; } catch { return; } } }"},
			{Code: "class A extends B { constructor() { try { class C { static { try { throw 0; } catch {} } } return {}; } catch { return; } } }"},
			{Code: "class A extends B { constructor() { try { class C { static value = () => fail(); } return {}; } catch { return; } } }"},
			{Code: "class A extends B { constructor() { try { class C { static { throw 0; } } } catch { super(); } } }"},
			{Code: "class A extends B { constructor() { try { class C { static value = fail(); } return {}; } catch { super(); } } }"},
			{Code: "class A extends B { constructor() { try { `${(super(), 0)}`; } catch { super(); } } }"},
			{Code: "class A extends B { constructor() { try { ({[(super(), 0)]: 0}); } catch { super(); } } }"},
			{Code: "class A extends B { constructor() { try { ({[0]: super()}); } catch { super(); } } }"},
			{Code: "class A extends B { constructor() { try { `${0}${true}${null}${void 0}${1n}`; return {}; } catch { return; } } }"},
			{Code: "class A extends B { constructor() { try { class C { static value = 0; [super()]() {} } } catch { return {}; } } }"},
			{Code: "class A extends B { constructor() { return super(); } }"},
			{Code: "class A extends B { constructor() { return (super()); } }"},
			{Code: "class A extends B { constructor() { return condition ? super() : {}; } }"},
			{Code: "class A extends B { constructor() { throw super(); } }"},
			{Code: "class A extends B { constructor() { try { let x; return {}; } catch { return; } } }"},
			{Code: "class A extends B { constructor() { try { const x = 0; return {}; } catch { return; } } }"},
			{Code: "class A extends B { constructor() { try { const {x} = {x:0}; return {}; } catch { return; } } }"},
			{Code: "class A extends B { constructor() { try { const {x:{y}} = {x:{y:0}}; return {}; } catch { return; } } }"},
			{Code: "class A extends B { constructor() { try {return {};} finally {super();} } }"},
			{Code: "class A extends B { constructor() { try {throw 0;} finally {return {};} } }"},
			{Code: "class A extends B { constructor() { try {return {};} finally {return {};} } }"},
			{Code: "class A extends B { constructor() { try {return;} finally {super();} } }"},
			{Code: "class A extends B { constructor() { while(true && true){return {};} } }"},
			{Code: "class A extends B { constructor() { while(false || true){return {};} } }"},
			{Code: "class A extends B { constructor() { while(true || sideEffect()){return {};} } }"},
			{Code: "class A extends B { constructor() { while((false)){super();}super(); } }"},
			{Code: "class A extends B { constructor() { do{super();}while(false); } }"},
			{Code: "class A extends B { constructor() { outer:inner:do{super();continue outer;}while(false); } }"},
			{Code: "class A extends B { constructor() { outer:inner:while(true){super();break outer;} } }"},
			{Code: "class A extends B { constructor() { for(const x of (super(),xs)){} } }"},
			{Code: "class A extends B { constructor() { const {x=super()}={x:void 0}; } }"},
			{Code: "class A extends B { constructor() { const {x=super()}={x:(super(),1)}; } }"},
			{Code: "class A extends B { constructor() { const {\"16\":x=super()}={0x10:void 0}; } }"},
			{Code: "class A extends B { constructor() { const {[\"x\"]:x=super()}={[\"x\"]:void 0}; } }"},
			{Code: "class A extends B { constructor() { try{super();}catch{super();} } }"},
			{Code: "class A extends B { constructor() { const obj={get x(){return this.x;}};super(); } }"},
			{Code: "class A extends B { constructor() { const fn=()=>super();super(); } }"},
			{Code: "class A extends B { constructor() { super();try{return;}finally{} } }"},
			{Code: "class A extends B { constructor() { super();try{let x;return;}catch{super();} } }"},
			{Code: "class A extends B { constructor() { super();try{return {};}finally{return;} } }"},
			{Code: "class A extends B { constructor() { try{const {x}=null;return;}catch{return {};} } }"},
			{Code: "class A extends B { constructor() { try{throw 0;}catch{super();} } }"},
			{Code: "class A extends B { constructor() { while(true){try{super();break;}finally{}} } }"},
			{Code: "class A extends B { constructor() { while(true){try{break;}finally{super();}} } }"},
			{Code: "class A extends B { constructor() { do{try{return {};}finally{break;}}while(false);super(); } }"},
			{Code: "class A extends B { constructor() { null?.(super());super(); } }"},
			{Code: "class A extends B { constructor() { (null?.foo)(super()); } }"},
			{Code: "class A extends B { constructor() { (foo?.bar)(super()); } }"},
			{Code: "class A extends B { constructor() { try{const x=0;}catch{return;}return {}; } }"},
			{Code: "class A extends B { constructor() { let obj={};({x:obj[super()]=super()}={x:1}); } }"},
			{Code: "class A extends B { constructor() { const {[\"__proto__\"]:x=super()}={[\"__proto__\"]:void 0}; } }"},
			{Code: "class A extends B { static x=maybeThrow(); constructor(){try{super();}catch{super();}} }"},
		},
		[]rule_tester.InvalidTestCase{
			// Escaping static exceptions and conversions after a successful call.
			{Code: "class A extends B { constructor() { try { class C { static value = fail(); } return {}; } catch {} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 101},
			}},
			{Code: "class A extends B { constructor() { try { const C = class { static value = fail(); }; return {}; } catch {} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 110},
			}},
			{Code: "class A extends B { constructor() { try { class C { static { fail(); } } return {}; } catch {} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 97},
			}},
			{Code: "class A extends B { constructor() { try { class C { static { throw 0; } } return {}; } catch {} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 98},
			}},
			{Code: "class A extends B { constructor() { try { class C { static { try { throw 0; } finally {} } } return {}; } catch {} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 117},
			}},
			{Code: "class A extends B { constructor() { super(); try { class C { static value = fail(); } } catch { super(); } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 97, EndLine: 1, EndColumn: 104},
			}},
			{Code: "class A extends B { constructor() { try { class C { static value = fail(); [super()]() {} } } catch { super(); } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 103, EndLine: 1, EndColumn: 110},
			}},
			{Code: "class B { toString() { throw new Error(\"coercion failed\"); } } class A extends B { constructor() { try { `${super()}`; } catch { super(); } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 130, EndLine: 1, EndColumn: 137},
			}},
			{Code: "class A extends B { constructor() { try { ({[super()]: 0}); } catch { super(); } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 71, EndLine: 1, EndColumn: 78},
			}},
			{Code: "class A extends B { constructor() { try { ({[super()]() {}}); } catch { super(); } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 73, EndLine: 1, EndColumn: 80},
			}},
			{Code: "class A extends B { constructor() { try { class C { [super()]() {} } } catch { super(); } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 80, EndLine: 1, EndColumn: 87},
			}},
			{Code: "class A extends B { constructor() { try { `${super()}${0}`; } catch { super(); } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 71, EndLine: 1, EndColumn: 78},
			}},
			{Code: "class A extends B { constructor() { try { tag`${super()}`; } catch { super(); } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 70, EndLine: 1, EndColumn: 77},
			}},
			{Code: "class A extends B { constructor() { for(const {x=super()} of []){break;} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 75},
			}},
			{Code: "class A extends B { constructor() { for(const {[super()]:x} of []){break;} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 77},
			}},
			{Code: "class A extends B { constructor() { for(const {x=super()} of xs){} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 69},
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 50, EndLine: 1, EndColumn: 57},
			}},
			{Code: "class A extends B { constructor() { let x;for({x=super()} of xs){break;} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 75},
			}},
			{Code: "class A extends B { constructor() { let x;for([x=super()] of xs){break;} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 75},
			}},
			{Code: "class A extends B { constructor() { for(obj[super()] of []){break;} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 70},
			}},
			{Code: "class A extends B { constructor() { const {x=super()}={x:1}; } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 63},
			}},
			{Code: "class A extends B { constructor() { try{const {x}=null;return {};}catch{return;} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 83},
			}},
			{Code: "class A extends B { constructor() { try{const [x]=null;return {};}catch{return;} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 83},
			}},
			{Code: "class A extends B { constructor() { try{const {x}={get x(){throw 0;}};return {};}catch{return;} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 98},
			}},
			{Code: "class A extends B { constructor() { try{missing=0;return {};}catch{return;} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 78},
			}},
			{Code: "class A extends B { constructor() { switch(0){default:super();break;case super():break;} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 55, EndLine: 1, EndColumn: 62},
			}},
			{Code: "class A extends B { constructor() { switch(0){case super():break;default:super();break;} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 74, EndLine: 1, EndColumn: 81},
			}},
			{Code: "class A extends B { constructor() { outer:inner:while(true){super();continue outer;} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 61, EndLine: 1, EndColumn: 68},
			}},
			{Code: "class A extends B { constructor() { outer:inner:do{super();continue outer;}while(true); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 52, EndLine: 1, EndColumn: 59},
			}},
			{Code: "class A extends B { constructor() { foo?.(super()); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 54},
			}},
			{Code: "class A extends B { constructor() { foo?.[super()]; } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 54},
			}},
			{Code: "class A extends B { constructor() { foo?.(super()).bar(super()); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 67},
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 56, EndLine: 1, EndColumn: 63},
			}},
			{Code: "class A extends B { constructor() { super();foo?.(super()); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 51, EndLine: 1, EndColumn: 58},
			}},
			{Code: "class A extends B { constructor() { super();return super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 52, EndLine: 1, EndColumn: 59},
			}},
			{Code: "class A extends B { constructor() { try{return {};}finally{return;} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 70},
			}},
			{Code: "class A extends B { constructor() { try{return foo();}catch{return;} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 71},
			}},
			{Code: "class A extends B { constructor() { let obj={};[obj[super()]=super()]=[]; } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 62, EndLine: 1, EndColumn: 69},
			}},
			{Code: "class A extends B { constructor() { let obj={};({x:obj[super()]=super()}={x:void 0}); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 65, EndLine: 1, EndColumn: 72},
			}},
			{Code: "class A extends B { constructor(x=super()) {} }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 46},
			}},
			{Code: "class A extends B { constructor(x=super()) {super();} }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 45, EndLine: 1, EndColumn: 52},
			}},
			{Code: "class A extends B { x=maybeThrow(); constructor(){try{super();}catch{super();}} }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 70, EndLine: 1, EndColumn: 77},
			}},
			{Code: "class A extends B { #method(){} constructor(){try{super();}catch{super();}} }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 66, EndLine: 1, EndColumn: 73},
			}},
			{Code: "class A extends B {\r\n  constructor(){\r\n    /*😀*/ super();\r\n    return super();\r\n  }\r\n}", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 4, Column: 12, EndLine: 4, EndColumn: 19},
			}},
		},
	)
}
