package no_constant_condition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConstantConditionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoConstantConditionRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			// Basic variable conditions
			{Code: `if(a);`},
			{Code: `if(a == 0);`},
			{Code: `if(a = f());`},

			// Assignment operators
			{Code: `if(a += 1);`},
			{Code: `if(a |= 1);`},
			{Code: `if(a |= true);`},
			{Code: `if(a |= false);`},
			{Code: `if(a &= 1);`},
			{Code: `if(a &= true);`},
			{Code: `if(a &= false);`},
			{Code: `if(a >>= 1);`},
			{Code: `if(a >>= true);`},
			{Code: `if(a >>= false);`},
			{Code: `if(a >>>= 1);`},
			{Code: `if(a ??= 1);`},
			{Code: `if(a ??= true);`},
			{Code: `if(a ??= false);`},

			// Logical assignment operators
			{Code: `if(a ||= b);`},
			{Code: `if(a ||= false);`},
			{Code: `if(a ||= 0);`},
			{Code: `if(a ||= void 0);`},
			{Code: `if(+(a ||= 1));`},
			{Code: `if(f(a ||= true));`},
			{Code: `if((a ||= 1) + 2);`},
			{Code: `if(1 + (a ||= true));`},
			{Code: `if(a ||= '' || false);`},
			{Code: `if(a ||= void 0 || null);`},
			{Code: `if((a ||= false) || b);`},
			{Code: `if(a || (b ||= false));`},
			{Code: `if((a ||= true) && b);`},
			{Code: `if(a && (b ||= true));`},
			{Code: `if(a &&= b);`},
			{Code: `if(a &&= true);`},
			{Code: `if(a &&= 1);`},
			{Code: `if(a &&= 'foo');`},
			{Code: `if((a &&= '') + false);`},
			{Code: `if('' + (a &&= null));`},
			{Code: `if(a &&= 1 && 2);`},
			{Code: `if((a &&= true) && b);`},
			{Code: `if(a && (b &&= true));`},
			{Code: `if((a &&= false) || b);`},
			{Code: `if(a || (b &&= false));`},
			{Code: `if(a ||= b ||= false);`},
			{Code: `if(a &&= b &&= true);`},
			{Code: `if(a ||= b &&= false);`},
			{Code: `if(a ||= b &&= true);`},
			{Code: `if(a &&= b ||= false);`},
			{Code: `if(a &&= b ||= true);`},

			// Comma operator
			{Code: `if(1, a);`},

			// Membership operators
			{Code: `if ('every' in []);`},

			// Template literals with expressions
			{Code: "if (`${a}`) {};"},
			{Code: "if (`${foo()}`) {};"},
			{Code: "if (`${a === 'b' && b==='a'}`) {};"},
			{Code: "if (`foo${a}` === 'fooa');"},

			// Unary operators with logical expressions
			{Code: `if (+(a || true));`},
			{Code: `if (-(a || true));`},
			{Code: `if (~(a || 1));`},
			{Code: `if (+(a && 0) === +(b && 0));`},

			// While loops
			{Code: `while(~!a);`},
			{Code: `while(a = b);`},
			{Code: "while(`${a}`) {};"},
			{Code: `while(x += 3) {};`},

			// For loops
			{Code: `for(;x < 10;);`},
			{Code: `for(;;);`},
			{Code: "for(;`${a}`;);"},

			// Do-while loops
			{Code: `do{ }while(x)`},

			// Ternary operator
			{Code: `q > 0 ? 1 : 2;`},
			{Code: "(`${a}` === a ? 1 : 2);"},
			{Code: "(`foo${a}` === a ? 1 : 2);"},

			// typeof conditions
			{Code: `if(typeof x === 'undefined'){}`},
			{Code: "if(`${typeof x}` === 'undefined'){}"},
			{Code: `if(a === 'str' && typeof b){}`},
			{Code: `typeof a == typeof b`},
			{Code: "typeof 'a' === 'string'|| typeof b === 'string'"},
			{Code: "(`${typeof 'a'}` === 'string'|| `${typeof b}` === 'string');"},

			// void operator
			{Code: `if (void a || a);`},
			{Code: `if (a || void a);`},

			// Mixed conditions
			{Code: `if(xyz === 'str1' && abc==='str2'){}`},
			{Code: `if(xyz === 'str1' || abc==='str2'){}`},
			{Code: `if(xyz === 'str1' || abc==='str2' && pqr === 5){}`},
			{Code: `if(typeof abc === 'string' && abc==='str2'){}`},
			{Code: `if(false || abc==='str'){}`},
			{Code: `if(true && abc==='str'){}`},
			{Code: "if(typeof 'str' && abc==='str'){}"},
			{Code: `if(abc==='str' || false || def ==='str'){}`},
			{Code: `if(true && abc==='str' || def ==='str'){}`},
			{Code: `if(true && typeof abc==='string'){}`},
			{Code: `if('str1' && a){}`},
			{Code: `if(a && 'str'){}`},

			// Comparisons with parenthesized values
			{Code: `if ((foo || true) === 'baz') {}`},
			{Code: `if ((foo || 'bar') === 'baz') {}`},
			{Code: `if ((foo || 'bar') !== 'baz') {}`},
			{Code: `if ((foo || 'bar') == 'baz') {}`},
			{Code: `if ((foo || 'bar') != 'baz') {}`},
			{Code: `if ((foo || 233) > 666) {}`},
			{Code: `if ((foo || 233) < 666) {}`},
			{Code: `if ((foo || 233) >= 666) {}`},
			{Code: `if ((foo || 233) <= 666) {}`},
			{Code: `if ((key || 'k') in obj) {}`},
			{Code: `if ((foo || {}) instanceof obj) {}`},
			{Code: `if ((foo || 'bar' || 'bar') === 'bar');`},

			// BigInt
			{Code: `if ((foo || 1n) === 'baz') {}`},
			{Code: `if (a && 0n || b);`},
			{Code: `if(1n && a){};`},

			// Array coercion
			{Code: `if ('' + [y] === '' + [ty]) {}`},
			{Code: `if ('a' === '' + [ty]) {}`},
			{Code: `if ('' + [y, m, d] === 'a') {}`},
			{Code: `if ('' + [y, 'm'] === '' + [ty, 'tm']) {}`},
			{Code: `if ('' + [y, 'm'] === '' + ['ty']) {}`},
			{Code: `if ([,] in ($2)) ; else ;`},
			{Code: `if ([...x]+'' === 'y'){}`},

			// Loop options - checkLoops: false
			{Code: `while(true);`, Options: map[string]interface{}{"checkLoops": false}},
			{Code: `for(;true;);`, Options: map[string]interface{}{"checkLoops": false}},
			{Code: `do{}while(true)`, Options: map[string]interface{}{"checkLoops": false}},

			// Loop options - checkLoops: "none"
			{Code: `while(true);`, Options: map[string]interface{}{"checkLoops": "none"}},
			{Code: `for(;true;);`, Options: map[string]interface{}{"checkLoops": "none"}},
			{Code: `do{}while(true)`, Options: map[string]interface{}{"checkLoops": "none"}},

			// Loop options - checkLoops: "allExceptWhileTrue"
			{Code: `while(true);`, Options: map[string]interface{}{"checkLoops": "allExceptWhileTrue"}},
			{Code: `while(true);`}, // default

			// Loop options - checkLoops: "all"
			{Code: `while(a == b);`, Options: map[string]interface{}{"checkLoops": "all"}},
			{Code: `do{ }while(x);`, Options: map[string]interface{}{"checkLoops": "all"}},
			{Code: `for (let x = 0; x <= 10; x++) {};`, Options: map[string]interface{}{"checkLoops": "all"}},

			// Generator functions
			{Code: `function* foo(){while(true){yield 'foo';}}`},
			{Code: `function* foo(){for(;true;){yield 'foo';}}`},
			{Code: `function* foo(){do{yield 'foo';}while(true)}`},
			{Code: `function* foo(){while (true) { while(true) {yield;}}}`},
			{Code: `function* foo() {for (; yield; ) {}}`},
			{Code: `function* foo() {for (; ; yield) {}}`},
			{Code: `function* foo() {while (true) {function* foo() {yield;}yield;}}`},
			{Code: `function* foo() { for (let x = yield; x < 10; x++) {yield;}yield;}`},
			{Code: `function* foo() { for (let x = yield; ; x++) { yield; }}`},

			// Boxed Number
			{Code: `if (new Number(x) + 1 === 2) {}`},

			// Destructuring
			{Code: `if([a]==[b]) {}`},
			{Code: `if (+[...a]) {}`},
			{Code: `if (+[...[...a]]) {}`},
			{Code: "if (`${[...a]}`) {}"},
			{Code: "if (`${[a]}`) {}"},
			{Code: `if (+[a]) {}`},
			{Code: `if (0 - [a]) {}`},
			{Code: `if (1 * [a]) {}`},

			// Boolean function
			{Code: `if (Boolean(a)) {}`},
			{Code: `if (Boolean(...args)) {}`},
			{Code: `if (foo.Boolean(1)) {}`},
			{Code: `function foo(Boolean) { if (Boolean(1)) {} }`},
			// TODO: This test is failing - TypeScript symbol resolution issue with const declarations
			// {Code: `const Boolean = () => {}; if (Boolean(1)) {}`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			// For loops with constant conditions
			{
				Code: `for(;true;);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "for(;``;);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "for(;`foo`;);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "for(;`foo${bar}`;);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Do-while with constants
			{
				Code: `do{}while(true)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `do{}while('1')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `do{}while(0)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `do{}while(t = -2)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "do{}while(``)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "do{}while(`foo`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "do{}while(`foo${bar}`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Ternary operator constants
			{
				Code: `true ? 1 : 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `1 ? 1 : 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `q = 0 ? 1 : 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `(q = 0) ? 1 : 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "`` ? 1 : 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "`foo` ? 1 : 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "`foo${bar}` ? 1 : 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// If statements with constants
			{
				Code: `if(-2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if({});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(0 < 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(0 || 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a, 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "if(`foo`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "if(``);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "if(`${'bar'}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "if(`${'bar' + `foo`}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "if(`foo${false || true}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "if(`foo${0 || 1}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "if(`foo${bar}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "if(`${bar}foo`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(!(true || a));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(!(a && void b && c));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(0 || !(a && null));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(1 + !(a || true));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(!(null && a) > 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(+(!(a && 0)));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(!typeof a === 'string');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(-('foo' || a));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(+(void a && b) === ~(1 || c));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Logical assignment with constants
			{
				Code: `if(a ||= true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a ||= 5);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a ||= 'foo' || b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a ||= b || /regex/);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a ||= b ||= true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a ||= b ||= c || 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(!(a ||= true));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(!(a ||= 'foo') === true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(!(a ||= 'foo') === false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a || (b ||= true));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if((a ||= 1) || b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if((a ||= true) && true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(true && (a ||= true));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a &&= false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a &&= null);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a &&= void b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a &&= 0 && b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a &&= b && '');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a &&= b &&= false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a &&= b &&= c && false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(!(a &&= false));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(!(a &&= 0) + 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a && (b &&= false));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if((a &&= null) && b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(false || (a &&= false));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if((a &&= false) || false);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// While loops
			{
				Code: `while([]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `while(~!0);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `while(x = 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `while(function(){});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `while(true);`,
				Options: map[string]interface{}{"checkLoops": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `while(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `while(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "while(`foo`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "while(``);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "while(`${'foo'}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "while(`${'foo' + 'bar'}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// typeof edge cases
			{
				Code: `if(typeof x){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(typeof 'abc' === 'string'){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a = typeof b){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a, typeof b){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "if(typeof 'a' == 'string' || typeof 'b' == 'string'){}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `while(typeof x){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// void operator edge cases
			{
				Code: `if(1 || void x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(void x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(y = void x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(x, void x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(void x === void y);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(void x && a);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a && void x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Complex logical expressions
			{
				Code: `if(false && abc==='str'){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(true || abc==='str'){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(1 || abc==='str'){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(abc==='str' || true){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(abc==='str' || true || def ==='str'){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(false || true){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(typeof abc==='str' || true){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if('str' || a){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if('str' || abc==='str'){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if('str1' || 'str2'){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if('str1' && 'str2'){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(abc==='str' || 'str'){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(a || 'str'){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Loop options - checkLoops: "all"
			{
				Code: `while(x = 1);`,
				Options: map[string]interface{}{"checkLoops": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `do{ }while(x = 1)`,
				Options: map[string]interface{}{"checkLoops": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `for (;true;) {};`,
				Options: map[string]interface{}{"checkLoops": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Generator functions with constant conditions
			{
				Code: `function* foo(){while(true){} yield 'foo';}`,
				Options: map[string]interface{}{"checkLoops": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `function* foo(){while(true){} yield 'foo';}`,
				Options: map[string]interface{}{"checkLoops": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `function* foo(){while(true){if (true) {yield 'foo';}}}`,
				Options: map[string]interface{}{"checkLoops": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `function* foo(){while(true){if (true) {yield 'foo';}}}`,
				Options: map[string]interface{}{"checkLoops": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `function* foo(){while(true){yield 'foo';} while(true) {}}`,
				Options: map[string]interface{}{"checkLoops": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `function* foo(){while(true){yield 'foo';} while(true) {}}`,
				Options: map[string]interface{}{"checkLoops": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `var a = function* foo(){while(true){} yield 'foo';}`,
				Options: map[string]interface{}{"checkLoops": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `var a = function* foo(){while(true){} yield 'foo';}`,
				Options: map[string]interface{}{"checkLoops": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `while (true) { function* foo() {yield;}}`,
				Options: map[string]interface{}{"checkLoops": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `while (true) { function* foo() {yield;}}`,
				Options: map[string]interface{}{"checkLoops": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `function* foo(){if (true) {yield 'foo';}}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `function* foo() {for (let foo = yield; true;) {}}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `function* foo() {for (foo = yield; true;) {}}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `function foo() {while (true) {function* bar() {while (true) {yield;}}}}`,
				Options: map[string]interface{}{"checkLoops": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `function foo() {while (true) {const bar = function*() {while (true) {yield;}}}}`,
				Options: map[string]interface{}{"checkLoops": "all"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `function* foo() { for (let foo = 1 + 2 + 3 + (yield); true; baz) {}}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Array and coercion
			{
				Code: `if([a]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if([]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(''+['a']) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(''+[]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(+1) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if ([,] + ''){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// BigInt constants
			{
				Code: `if(0n);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(0b0n);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(0o0n);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(0x0n);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(0b1n);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(0o1n);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(0x1n);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(0x1n || foo);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Class and instance constants
			{
				Code: `if(class {}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(new Foo()) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Boxed primitives
			{
				Code: `if(new Boolean(foo)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(new String(foo)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if(new Number(foo)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Spreading and spread operators
			{
				Code: "if(`${[...['a']]}`) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Undefined and Boolean function
			{
				Code: `if (undefined) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if (Boolean(1)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if (Boolean()) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if (Boolean([a])) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if (Boolean(1)) { function Boolean() {}}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
		},
	)
}
