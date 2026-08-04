package no_useless_assignment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUselessAssignmentUpstream migrates the full valid/invalid suite from
// upstream eslint/tests/lib/rules/no-useless-assignment.js 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific lock-in
// cases live in the no_useless_assignment_extras_test.go file.
func TestNoUselessAssignmentUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessAssignmentRule,
		[]rule_tester.ValidTestCase{
			{Code: `let v = 'used';
        console.log(v);
        v = 'used-2'
        console.log(v);`},
			{Code: `function foo() {
            let v = 'used';
            console.log(v);
            v = 'used-2';
            console.log(v);
        }`},
			{Code: `function foo() {
            let v = 'used';
            if (condition) {
                v = 'used-2';
                console.log(v);
                return
            }
            console.log(v);
        }`},
			{Code: `function foo() {
            let v = 'used';
            if (condition) {
                console.log(v);
            } else {
                v = 'used-2';
                console.log(v);
            }
        }`},
			{Code: `function foo() {
            let v = 'used';
            if (condition) {
                //
            } else {
                v = 'used-2';
            }
            console.log(v);
        }`},
			{Code: `var foo = function () {
            let v = 'used';
            console.log(v);
            v = 'used-2'
            console.log(v);
        }`},
			{Code: `var foo = () => {
            let v = 'used';
            console.log(v);
            v = 'used-2'
            console.log(v);
        }`},
			{Code: `class foo {
            static {
                let v = 'used';
                console.log(v);
                v = 'used-2'
                console.log(v);
            }
        }`},
			{Code: `function foo () {
            let v = 'used';
            for (let i = 0; i < 10; i++) {
                console.log(v);
                v = 'used in next iteration';
            }
        }`},
			{Code: `function foo () {
            let i = 0;
            i++;
            i++;
            console.log(i);
        }`},
			{Code: `export let foo = 'used';
        console.log(foo);
        foo = 'unused like but exported';`},
			{Code: `export function foo () {};
        console.log(foo);
        foo = 'unused like but exported';`},
			{Code: `export class foo {};
        console.log(foo);
        foo = 'unused like but exported';`},
			{Code: `export default function foo () {};
        console.log(foo);
        foo = 'unused like but exported';`},
			{Code: `export default class foo {};
        console.log(foo);
        foo = 'unused like but exported';`},
			{Code: `let foo = 'used';
        export { foo };
        console.log(foo);
        foo = 'unused like but exported';`},
			{Code: `function foo () {};
        export { foo };
        console.log(foo);
        foo = 'unused like but exported';`},
			{Code: `class foo {};
        export { foo };
        console.log(foo);
        foo = 'unused like but exported';`},
			{Code: `/* exported foo */
            let foo = 'used';
            console.log(foo);
            foo = 'unused like but exported with directive';`,
				// SKIP: rslint does not support ESLint's `/* exported */` directive comment.
				Skip: true},
			{Code: `/*eslint test/use-a:1*/
        let a = 'used';
        console.log(a);
        a = 'unused like but marked by markVariableAsUsed()';
        `,
				// SKIP: rslint does not support ESLint plugins calling markVariableAsUsed().
				Skip: true},
			{Code: `v = 'used';
        console.log(v);
        v = 'unused'`},
			{Code: `let v = 'used variable';`},
			{Code: `function foo() {
            return;

            const x = 1;
            if (y) {
                bar(x);
            }
        }`},
			{Code: `function foo() {
            const x = 1;
            console.log(x);
            return;

            x = 'Foo'
        }`},
			{Code: `function foo() {
            let a = 42;
            console.log(a);
            a++;
            console.log(a);
        }`},
			{Code: `function foo() {
            let a = 42;
            console.log(a);
            a--;
            console.log(a);
        }`},
			{Code: `function foo() {
            let a = 42;
            console.log(a);
            a = 10;
            a = a + 1;
            console.log(a);
        }`},
			{Code: `function foo() {
            let a = 42;
            console.log(a);
            a = 10;
            if (cond) {
                a = a + 1;
            } else {
                a = 2 + a;
            }
            console.log(a);
        }`},
			{Code: `function foo() {
            let a = 'used', b = 'used', c = 'used', d = 'used';
            console.log(a, b, c, d);
            ({ a, arr: [b, c, ...d] } = fn());
            console.log(a, b, c, d);
        }`},
			{Code: `function foo() {
            let a = 'used', b = 'used', c = 'used';
            console.log(a, b, c);
            ({ a = 'unused', foo: b, ...c } = fn());
            console.log(a, b, c);
        }`},
			{Code: `function foo() {
            let a = {};
            console.log(a);
            a.b = 'unused like, but maybe used in setter';
        }`},
			{Code: `function foo() {
            let a = { b: 42 };
            console.log(a);
            a.b++;
        }`},
			{Code: `function foo () {
            let v = 'used';
            console.log(v);
            function bar() {
                v = 'used in outer scope';
            }
            bar();
            console.log(v);
        }`},
			{Code: `function foo () {
            let v = 'used';
            console.log(v);
            setTimeout(() => console.log(v), 1);
            v = 'used in other scope';
        }`},
			{Code: `function foo () {
            let v = 'used';
            console.log(v);
            for (let i = 0; i < 10; i++) {
                if (condition) {
                    v = 'maybe used';
                    continue;
                }
                console.log(v);
            }
        }`},
			{Code: `/* globals foo */
        const bk = foo;
        foo = 42;
        try {
            // process
        } finally {
            foo = bk;
        }`},
			{Code: `
            const bk = console;
            console = { log () {} };
            try {
                // process
            } finally {
                console = bk;
            }`, Globals: map[string]bool{"console": false}},
			{Code: `let message = 'init';
        try {
            const result = call();
            message = result.message;
        } catch (e) {
            // ignore
        }
        console.log(message)`},
			{Code: `let message = 'init';
        try {
            message = call().message;
        } catch (e) {
            // ignore
        }
        console.log(message)`},
			{Code: `let v = 'init';
        try {
            v = callA();
            try {
                v = callB();
            } catch (e) {
                // ignore
            }
        } catch (e) {
            // ignore
        }
        console.log(v)`},
			{Code: `let v = 'init';
        try {
            try {
                v = callA();
            } catch (e) {
                // ignore
            }
        } catch (e) {
            // ignore
        }
        console.log(v)`},
			{Code: `let a;
        try {
            foo();
        } finally {
            a = 5;
        }
        console.log(a);`},
			{Code: `function* generator() {
            let done = false;
            try {
                yield 1;
                done = true;
            } catch {
                done = true;
            } finally {
                if (!done) {
                    console.log("done is false");
                }
            }
        }`},
			{Code: `function* generator() {
            let done = false;
            try {
                yield 1;
                done = true;
                yield 2;
            } finally {
                if (done) {
                    console.log("done is true");
                }
            }
        }`},
			{Code: `function* generator() {
            let done = false;
            try {
                yield 1;
            } catch {
                console.log(done);
            }
        }`},
			{Code: `function* generator() {
            let done = false;
            try {
                foo();
            } catch {
                yield 1;
                done = true;
            } finally {
                yield 2;
                if (!done) {
                    console.log(done);
                }
            }
        }`},
			{Code: `function foo() {
			let outcome = 'unknown';

			try {
				helper1();
				outcome = 'success';
			} catch (err) {
				helper2();
				outcome = 'exception'; 
			} finally {
				console.log(outcome);
			}
		}`},
			{Code: `function foo() {
			let outcome = 'unknown';

			try {
				new Foo();
				outcome = 'success';
			} catch (err) {
				new Bar();
				outcome = 'exception'; 
			} finally {
				console.log(outcome);
			}
		}`},
			{Code: `async function foo() {
			let outcome = 'unknown';

			try {
				await import("./foo.js");
				outcome = 'success';
			} catch (err) {
				await import("./bar.js");
				outcome = 'exception';
			} finally {
				console.log(outcome);
			}
		}`},
			{Code: `function foo() {
			let outcome = 'unknown';

			try {
				obj.foo;
				outcome = 'success';
			} catch (err) {
				obj.foo;
				outcome = 'exception';
			} finally {
				console.log(outcome);
			}
		}`},
			{Code: `function foo() {
			let outcome = 'unknown';

			try {
				helper1();
				outcome = 'success';
			} catch (err) {
			 	outcome = 'exception';
				helper2(); 
			} finally {
				console.log(outcome);
			}
		}`},
			{Code: `const obj = { a: 5 };
        const { a, b = a } = obj;
        console.log(b); // 5`},
			{Code: `const arr = [6];
        const [c, d = c] = arr;
        console.log(d); // 6`},
			{Code: `const obj = { a: 1 };
        let {
            a,
            b = (a = 2)
        } = obj;
        console.log(a, b);`},
			{Code: `let { a, b: {c = a} = {} } = obj;
        console.log(c);`},
			{Code: `function foo(){
            let bar;
            try {
                bar = 2;
                unsafeFn();
                return { error: undefined };
            } catch {
                return { bar }; 
            }
        }   
        function unsafeFn() {
            throw new Error();
        }`},
			{Code: `function foo(){
            let bar, baz;
            try {
                bar = 2;
                unsafeFn();
                return { error: undefined };
            } catch {
               baz = bar;
            }
            return baz;
        }   
        function unsafeFn() {
            throw new Error();
        }`},
			{Code: `function foo(){
            let bar;
            try {
                bar = 2;
                unsafeFn();
                bar = 4;
            } catch {
               // handle error
            }
            return bar;
        }   
        function unsafeFn() {
            throw new Error();
        }`},
			{Code: `/*eslint test/unknown-ref:1*/
        let a = "used";
		console.log(a);
		a = "unused";`,
				// SKIP: rslint does not support ESLint plugins injecting scope references.
				Skip: true},
			{Code: `/*eslint test/unknown-ref:1*/
		function foo() {
			let a = "used";
			console.log(a);
			a = "unused";
		}`,
				// SKIP: rslint does not support ESLint plugins injecting scope references.
				Skip: true},
			{Code: `/*eslint test/unknown-ref:1*/
		function foo() {
			let a = "used";
			if (condition) {
				a = "unused";
				return
			}
			console.log(a);
        }`,
				// SKIP: rslint does not support ESLint plugins injecting scope references.
				Skip: true},
			{Code: `
                function App() {
                    const A = "";
                    return <A/>;
                }
            `, Tsx: true},
			{Code: `
                function App() {
                    let A = "";
                    foo(A);
                    A = "A";
                    return <A/>;
                }
            `, Tsx: true},
			{Code: `
                function App() {
					let A = "a";
                    foo(A);
                    return <A/>;
                }
            `, Tsx: true},
			{Code: `function App() {
				let x = 0;
				foo(x);
				x = 1;
				return <A prop={x} />;
			}`, Tsx: true},
			{Code: `function App() {
				let x = "init";
				foo(x);
				x = "used";
				return <A>{x}</A>;
			}`, Tsx: true},
			{Code: `function App() {
				let props = { a: 1 };
				foo(props);
				props = { b: 2 };
				return <A {...props} />;
			}`, Tsx: true},
			{Code: `function App() {
				let NS = Lib;
				return <NS.Cmp />;
			}`, Tsx: true},
			{Code: `function App() {
				let a = 0;
				a++;
				return <A prop={a} />;
			}`, Tsx: true},
			{Code: `function App() {
				const obj = { a: 1 };
				const { a, b = a } = obj;
				return <A prop={b} />;
			}`, Tsx: true},
			{Code: `function App() {
				let { a, b: { c = a } = {} } = obj;
				return <A prop={c} />;
			}`, Tsx: true},
			{Code: `function App() {
				let x = "init";
				if (cond) {
					x = "used";
					return <A prop={x} />;
				}
				return <A prop={x} />;
			}`, Tsx: true},
			{Code: `function App() {
				let A;
				if (cond) {
				  A = Foo;
				} else {
				  A = Bar;
				}
				return <A />;
			}`, Tsx: true},
			{Code: `function App() {
				let m;
				try {
				  m = 2;
				  unsafeFn();
				  m = 4;
				} catch (e) {
				  // ignore
				}
				return <A prop={m} />;
			}`, Tsx: true},
			{Code: `function App() {
				const arr = [6];
				const [c, d = c] = arr;
				return <A prop={d} />;
			}`, Tsx: true},
			{Code: `function App() {
				const obj = { a: 1 };
				let {
				  a,
				  b = (a = 2)
				} = obj;
				return <A prop={a} />;
			}`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `let v = 'used';
            console.log(v);
            v = 'unused'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 3, Column: 13},
				},
			},
			{
				Code: `function foo() {
                let v = 'used';
                console.log(v);
                v = 'unused';
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 17},
				},
			},
			{
				Code: `function foo() {
                let v = 'used';
                if (condition) {
                    v = 'unused';
                    return
                }
                console.log(v);
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 21},
				},
			},
			{
				Code: `function foo() {
                let v = 'used';
                if (condition) {
                    console.log(v);
                } else {
                    v = 'unused';
                }
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 6, Column: 21},
				},
			},
			{
				Code: `var foo = function () {
                let v = 'used';
                console.log(v);
                v = 'unused'
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 17},
				},
			},
			{
				Code: `var foo = () => {
                let v = 'used';
                console.log(v);
                v = 'unused'
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 17},
				},
			},
			{
				Code: `class foo {
                static {
                    let v = 'used';
                    console.log(v);
                    v = 'unused'
                }
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 5, Column: 21},
				},
			},
			{
				Code: `function foo() {
                let v = 'unused';
                if (condition) {
                    v = 'used';
                    console.log(v);
                    return
                }
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 2, Column: 21},
				},
			},
			{
				Code: `function foo() {
                let v = 'used';
                console.log(v);
                v = 'unused';
                v = 'unused';
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 17},
					{MessageId: "unnecessaryAssignment", Line: 5, Column: 17},
				},
			},
			{
				Code: `function foo() {
                let v = 'used';
                console.log(v);
                v = 'unused';
                v = 'used';
                console.log(v);
                v = 'used';
                console.log(v);
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 17},
				},
			},
			{
				Code: `
            let v;
            v = 'unused';
            if (foo) {
                v = 'used';
            } else {
                v = 'used';
            }
            console.log(v);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 3, Column: 13},
				},
			},
			{
				Code: `function foo() {
                let v = 'used';
                console.log(v);
                v = 'unused';
                v = 'unused';
                v = 'used';
                console.log(v);
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 17},
					{MessageId: "unnecessaryAssignment", Line: 5, Column: 17},
				},
			},
			{
				Code: `function foo() {
                let v = 'unused';
                if (condition) {
                    if (condition2) {
                        v = 'used-2';
                    } else {
                        v = 'used-3';
                    }
                } else {
                    v = 'used-4';
                }
                console.log(v);
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 2, Column: 21},
				},
			},
			{
				Code: `function foo() {
                let v;
                if (condition) {
                    v = 'unused';
                } else {
                    //
                }
                if (condition2) {
                    v = 'used-1';
                } else {
                    v = 'used-2';
                }
                console.log(v);
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 21},
				},
			},
			{
				Code: `function foo() {
                let v = 'used';
                if (condition) {
                    v = 'unused';
                    v = 'unused';
                    v = 'used';
                }
                console.log(v);
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 21},
					{MessageId: "unnecessaryAssignment", Line: 5, Column: 21},
				},
			},
			{
				Code: `function foo() {
                let a = 42;
                console.log(a);
                a++;
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 17},
				},
			},
			{
				Code: `function foo() {
                let a = 42;
                console.log(a);
                a--;
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 17},
				},
			},
			{
				Code: `function foo() {
                let a = 'used', b = 'used', c = 'used', d = 'used';
                console.log(a, b, c, d);
                ({ a, arr: [b, c,, ...d] } = fn());
                console.log(c);
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 20},
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 29},
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 39},
				},
			},
			{
				Code: `function foo() {
                let a = 'used', b = 'used', c = 'used';
                console.log(a, b, c);
                ({ a = 'unused', foo: b, ...c } = fn());
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 20},
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 39},
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 45},
				},
			},
			{
				Code: `function foo () {
                let v = 'used';
                console.log(v);
                setTimeout(() => v = 42, 1);
                v = 'unused and variable is only updated in other scopes';
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 5, Column: 17},
				},
			},
			{
				Code: `function foo() {
                let v = 'used';
                if (condition) {
                    let v = 'used';
                    console.log(v);
                    v = 'unused';
                }
                console.log(v);
                v = 'unused';
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 6, Column: 21},
					{MessageId: "unnecessaryAssignment", Line: 9, Column: 17},
				},
			},
			{
				Code: `function foo() {
                let v = 'used';
                if (condition) {
                    console.log(v);
                    v = 'unused';
                } else {
                    v = 'unused';
                }
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 5, Column: 21},
					{MessageId: "unnecessaryAssignment", Line: 7, Column: 21},
				},
			},
			{
				Code: `function foo () {
                let v = 'used';
                console.log(v);
                v = 'unused';
                return;
                console.log(v);
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 17},
				},
			},
			{
				Code: `function foo () {
                let v = 'used';
                console.log(v);
                v = 'unused';
                throw new Error();
                console.log(v);
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 17},
				},
			},
			{
				Code: `function foo () {
                let v = 'used';
                console.log(v);
                for (let i = 0; i < 10; i++) {
                    v = 'unused';
                    continue;
                    console.log(v);
                }
            }
            function bar () {
                let v = 'used';
                console.log(v);
                for (let i = 0; i < 10; i++) {
                    v = 'unused';
                    break;
                    console.log(v);
                }
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 5, Column: 21},
					{MessageId: "unnecessaryAssignment", Line: 14, Column: 21},
				},
			},
			{
				Code: `function foo () {
                let v = 'used';
                console.log(v);
                for (let i = 0; i < 10; i++) {
                    if (condition) {
                        v = 'unused';
                        break;
                    }
                    console.log(v);
                }
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 6, Column: 25},
				},
			},
			{
				Code: `let message = 'unused';
            try {
                const result = call();
                message = result.message;
            } catch (e) {
                message = 'used';
            }
            console.log(message)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 5},
				},
			},
			{
				Code: `function* generator() {
                let done = false;
                yield 1;
                done = true;
                console.log(done);
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 2, Column: 21, EndLine: 2, EndColumn: 25},
				},
			},
			{
				Code: `function* generator() {
                let done = false;
                try {
                    yield 1;
                } finally {
                    done = true;
                    console.log(done);
                }
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 2, Column: 21, EndLine: 2, EndColumn: 25},
				},
			},
			{
				Code: `let message = 'unused';
            try {
                message = 'used';
                console.log(message)
            } catch (e) {
            }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 5},
				},
			},
			{
				Code: `let message = 'unused';
            try {
                message = call();
            } catch (e) {
                message = 'used';
            }
            console.log(message)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 5},
				},
			},
			{
				Code: `let v = 'unused';
            try {
                v = callA();
                try {
                    v = callB();
                } catch (e) {
                    // ignore
                }
            } catch (e) {
                v = 'used';
            }
            console.log(v)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 1, Column: 5},
				},
			},
			{
				Code: `function foo() {
				let outcome = 'unknown';

				try {
					bar();
				} catch (err) {
					new Baz();
					outcome = 'exception'; 
				} finally {
					return;
					console.log(outcome);
				}
			}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 2, Column: 9},
					{MessageId: "unnecessaryAssignment", Line: 8, Column: 6},
				},
			},
			{
				Code: `
            var x = 1; // used
            x = x + 1; // unused
            x = 5; // used
            f(x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 3, Column: 13},
				},
			},
			{
				Code: `
            var x = 1; // used
            x = // used
                x++; // unused
            f(x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 17},
				},
			},
			{
				Code: `const obj = { a: 1 };
            let {
                a,
                b = (a = 2)
            } = obj;
            a = 3
            console.log(a, b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 3, Column: 17},
					{MessageId: "unnecessaryAssignment", Line: 4, Column: 22},
				},
			},
			{
				Code: `function App() {
            let A = "unused";
            A = "used";
            return <A/>;
            }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 2, Column: 17, EndLine: 2, EndColumn: 18},
				},
			},
			{
				Code: `function App() {
            let A = "unused";
            A = "used";
            return <A></A>;
            }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 2, Column: 17, EndLine: 2, EndColumn: 18},
				},
			},
			{
				Code: `function App() {
            let A = "unused";
            A = "used";
            return <A.B />;
            }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 2, Column: 17, EndLine: 2, EndColumn: 18},
				},
			},
			{
				Code: `function App() {
            let x = "used";
            if (cond) {
              return <A prop={x} />;
            } else {
              x = "unused";
            }
            }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 6, Column: 15, EndLine: 6, EndColumn: 16},
				},
			},
			{
				Code: `function App() {
            let A;
            A = "unused";
            if (cond) {
              A = "used1";
            } else {
              A = "used2";
            }
            return <A/>;
            }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 3, Column: 13, EndLine: 3, EndColumn: 14},
				},
			},
			{
				Code: `function App() {
            let message = 'unused';
            try {
              const result = call();
              message = result.message;
            } catch (e) {
              message = 'used';
            }
            return <A prop={message} />;
            }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 2, Column: 17, EndLine: 2, EndColumn: 24},
				},
			},
			{
				Code: `function App() {
            let x = 1;
            x = x + 1;
            x = 5;
            return <A prop={x} />;
            }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 3, Column: 13, EndLine: 3, EndColumn: 14},
				},
			},
			{
				Code: `function App() {
            let x = 1;
            x = 2;
            return <A>{x}</A>;
            }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 2, Column: 17, EndLine: 2, EndColumn: 18},
				},
			},
			{
				Code: `function App() {
            let x = 0;
            x = 1;
            x = 2;
            return <A prop={x} />;
            }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssignment", Line: 2, Column: 17, EndLine: 2, EndColumn: 18},
					{MessageId: "unnecessaryAssignment", Line: 3, Column: 13, EndLine: 3, EndColumn: 14},
				},
			},
		},
	)
}
