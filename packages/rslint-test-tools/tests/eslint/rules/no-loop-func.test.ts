import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-loop-func', {
  valid: [
    "string = 'function a() {}';",
    'for (var i=0; i<l; i++) { } var a = function() { i; };',
    'for (var i=0, a=function() { i; }; i<l; i++) { }',
    'for (var x in xs.filter(function(x) { return x != upper; })) { }',
    'for (var x of xs.filter(function(x) { return x != upper; })) { }',
    'for (var i=0; i<l; i++) { (function() {}) }',
    'for (var i in {}) { (function() {}) }',
    'for (var i of {}) { (function() {}) }',
    'for (let i=0; i<l; i++) { (function() { i; }) }',
    'for (let i in {}) { i = 7; (function() { i; }) }',
    'for (const i of {}) { (function() { i; }) }',
    'for (using i of foo) { (function() { i; }) }',
    'for (await using i of foo) { (function() { i; }) }',
    'for (var i = 0; i < 10; ++i) { using foo = bar(i); (function() { foo; }) }',
    'for (var i = 0; i < 10; ++i) { await using foo = bar(i); (function() { foo; }) }',
    'for (let i = 0; i < 10; ++i) { for (let x in xs.filter(x => x != i)) {  } }',
    'let a = 0; for (let i=0; i<l; i++) { (function() { a; }); }',
    'let a = 0; for (let i in {}) { (function() { a; }); }',
    'let a = 0; for (let i of {}) { (function() { a; }); }',
    'let a = 0; for (let i=0; i<l; i++) { (function() { (function() { a; }); }); }',
    'let a = 0; for (let i in {}) { function foo() { (function() { a; }); } }',
    'let a = 0; for (let i of {}) { (() => { (function() { a; }); }); }',
    'var a = 0; for (let i=0; i<l; i++) { (function() { a; }); }',
    'var a = 0; for (let i in {}) { (function() { a; }); }',
    'var a = 0; for (let i of {}) { (function() { a; }); }',
    `let result = {};
for (const score in scores) {
  const letters = scores[score];
  letters.split('').forEach(letter => {
    result[letter] = score;
  });
}
result.__default = 6;`,
    `while (true) {
    (function() { a; });
}
let a;`,
    'while(i) { (function() { i; }) }',
    'do { (function() { i; }) } while (i)',
    'var i; while(i) { (function() { i; }) }',
    'var i; do { (function() { i; }) } while (i)',
    'for (var i=0; i<l; i++) { (function() { undeclared; }) }',
    'for (let i=0; i<l; i++) { (function() { undeclared; }) }',
    'for (var i in {}) { i = 7; (function() { undeclared; }) }',
    'for (let i in {}) { i = 7; (function() { undeclared; }) }',
    'for (const i of {}) { (function() { undeclared; }) }',
    'for (let i = 0; i < 10; ++i) { for (let x in xs.filter(x => x != undeclared)) {  } }',
    `let current = getStart();
while (current) {
(() => {
    current;
    current.a;
    current.b;
    current.c;
    current.d;
})();

current = current.upper;
}`,
    'for (var i=0; (function() { i; })(), i<l; i++) { }',
    'for (var i=0; i<l; (function() { i; })(), i++) { }',
    'for (var i = 0; i < 10; ++i) { (()=>{ i;})() }',
    'for (var i = 0; i < 10; ++i) { (function a(){i;})() }',
    `var arr = [];

for (var i = 0; i < 5; i++) {
    arr.push((f => f)((() => i)()));
}`,
    `var arr = [];

for (var i = 0; i < 5; i++) {
    arr.push((() => {
        return (() => i)();
    })());
}`,
    `const foo = bar;

for (var i = 0; i < 5; i++) {
    arr.push(() => foo);
}

foo = baz;`,
    `using foo = bar;

for (var i = 0; i < 5; i++) {
    arr.push(() => foo);
}

foo = baz;`,
    `await using foo = bar;

for (var i = 0; i < 5; i++) {
    arr.push(() => foo);
}

foo = baz;`,
    `for (let i = 0; i < 10; i++) {
  function foo() {
    console.log('A');
  }
}`,
    `let someArray: MyType[] = [];
for (let i = 0; i < 10; i += 1) {
  someArray = someArray.filter((item: MyType) => !!item);
}`,
    `type MyType = 1;
let someArray: MyType[] = [];
for (let i = 0; i < 10; i += 1) {
  someArray = someArray.filter((item: MyType) => !!item);
}`,
    `for (var i = 0; i < 10; i++) {
  const process = (item: UnconfiguredGlobalType) => {
    return item.id;
  };
}`,
    `for (var i = 0; i < 10; i++) {
  const process = (configItem: ConfiguredType, unconfigItem: UnconfiguredType) => {
    return {
      config: configItem.value,
      unconfig: unconfigItem.value
    };
  };
}`,

    // Destructuring: const iterator bindings are fresh per iteration.
    'for (const {a} of arr) { (function() { a; }) }',
    'for (const [a] of arr) { (function() { a; }) }',
    'for (const [[a]] of arr) { (function() { a; }) }',
    'for (const {a: {b}} of arr) { (function() { b; }) }',
    'for (const {a = 10} of arr) { (function() { a; }) }',

    // Labeled loop — break label is not a variable reference.
    'outer: for (var i = 0; i < l; i++) { (function() { break outer; }); }',

    // Nested loops closing over only fresh-let bindings.
    'for (let i = 0; i < l; i++) { for (let j = 0; j < i; j++) { (function() { j; }) } }',

    // A FunctionExpression inside a method inside a loop — the method is its
    // own scope boundary so the function is NOT "inside the loop".
    'for (var i = 0; i < l; i++) { const o = { m() { var f = function() { x; }; } }; }',
    'for (var i = 0; i < l; i++) { class C { m() { var f = function() { x; }; } } }',

    // Method referencing only a const outer binding.
    'const k = 10; for (var i = 0; i < l; i++) { class C { m() { return k; } } }',

    // Computed key captures loop var; body doesn't — no through ref in body.
    'for (var i = 0; i < l; i++) { var o = { [i]: function() {} }; }',

    // Class-expression name as through ref with no writes — safe.
    'for (var i = 0; i < l; i++) { const C = class Foo { m() { return Foo; } }; }',

    // for-await-of with const iterator binding — fresh per iteration.
    'async function f(xs) { for await (const x of xs) { (function() { x; }) } }',

    // Parameter default captures a fresh let binding (safe).
    'for (let i = 0; i < l; i++) { let j = i; (function(x = j) { x; }); }',

    // ForStatement with no init, loop var declared outside, not modified.
    'var i; for (; i < l; ) { (function() { i; }) }',

    // Import binding — read-only, safe.
    'import { foo } from "./mod"; for (var i = 0; i < l; i++) { (function() { foo; }) }',

    // Enum reference (const-like).
    'enum Color { Red } for (var i = 0; i < l; i++) { (function() { Color.Red; }) }',

    // Non-loop-enclosing outer function: inner function safe (loop variable is fresh let).
    'function outer() { for (var i = 0; i < l; i++) { let j = i; (function() { j; }); } }',
  ],
  invalid: [
    {
      code: 'for (var i=0; i<l; i++) { (function() { i; }) }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var i=0; i<l; i++) { for (var j=0; j<m; j++) { (function() { i+j; }) } }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var i in {}) { (function() { i; }) }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var i of {}) { (function() { i; }) }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var i=0; i < l; i++) { (() => { i; }) }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var i=0; i < l; i++) { var a = function() { i; } }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var i=0; i < l; i++) { function a() { i; }; a(); }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'let a; for (let i=0; i<l; i++) { a = 1; (function() { a; });}',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'let a; for (let i in {}) { (function() { a; }); a = 1; }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'let a; for (let i of {}) { (function() { a; }); } a = 1;',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'let a; for (let i=0; i<l; i++) { (function() { (function() { a; }); }); a = 1; }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'let a; for (let i in {}) { a = 1; function foo() { (function() { a; }); } }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'let a; for (let i of {}) { (() => { (function() { a; }); }); } a = 1;',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var i = 0; i < 10; ++i) { for (let x in xs.filter(x => x != i)) {  } }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (let x of xs) { let a; for (let y of ys) { a = 1; (function() { a; }); } }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var x of xs) { for (let y of ys) { (function() { x; }); } }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var x of xs) { (function() { x; }); }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'var a; for (let x of xs) { a = 1; (function() { a; }); }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'var a; for (let x of xs) { (function() { a; }); a = 1; }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'let a; function foo() { a = 10; } for (let x of xs) { (function() { a; }); } foo();',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'let a; function foo() { a = 10; for (let x of xs) { (function() { a; }); } } foo();',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'let a; for (var i=0; i<l; i++) { (function* (){i;})() }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'let a; for (var i=0; i<l; i++) { (async function (){i;})() }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `let current = getStart();
const arr = [];
while (current) {
    (function f() {
        current;
        arr.push(f);
    })();

    current = current.upper;
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `var arr = [];

for (var i = 0; i < 5; i++) {
    (function fun () {
        if (arr.includes(fun)) return i;
        else arr.push(fun);
    })();
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `let current = getStart();
const arr = [];
while (current) {
    const p = (async () => {
        await someDelay();
        current;
    })();

    arr.push(p);
    current = current.upper;
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `var arr = [];

for (var i = 0; i < 5; i++) {
    arr.push((f => f)(
        () => i
    ));
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `var arr = [];

for (var i = 0; i < 5; i++) {
    arr.push((() => {
        return () => i;
    })());
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `var arr = [];

for (var i = 0; i < 5; i++) {
    arr.push((() => {
        return () => { return i };
    })());
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `var arr = [];

for (var i = 0; i < 5; i++) {
    arr.push((() => {
        return () => {
            return () => i
        };
    })());
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `var arr = [];

for (var i = 0; i < 5; i++) {
    arr.push((() => {
        return () =>
            (() => i)();
    })());
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `var arr = [];

for (var i = 0; i < 5; i ++) {
    (() => {
        arr.push((async () => {
            await 1;
            return i;
        })());
    })();
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `var arr = [];

for (var i = 0; i < 5; i ++) {
    (() => {
        (function f() {
            if (!arr.includes(f)) {
                arr.push(f);
            }
            return i;
        })();
    })();

}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `var arr1 = [], arr2 = [];

for (var [i, j] of ["a", "b", "c"].entries()) {
    (() => {
        arr1.push((() => i)());
        arr2.push(() => j);
    })();
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `var arr = [];

for (var i = 0; i < 5; i ++) {
    ((f) => {
        arr.push(f);
    })(() => {
        return (() => i)();
    });

}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `for (var i = 0; i < 5; i++) {
    (async () => {
        () => i;
    })();
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `for (var i = 0; i < 10; i++) {
    items.push({
        id: i,
        name: "Item " + i
    });

    const process = function (callback){
        callback({ id: i, name: "Item " + i });
    };
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `for (var i = 0; i < 10; i++) {
  function foo() {
    console.log(i);
  }
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `for (var i = 0; i < 10; i++) {
  const handler = (event: Event) => {
    console.log(i);
  };
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `interface Item {
  id: number;
  name: string;
}

const items: Item[] = [];
for (var i = 0; i < 10; i++) {
  items.push({
    id: i,
    name: "Item " + i
  });

  const process = function(callback: (item: Item) => void): void {
    callback({ id: i, name: "Item " + i });
  };
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `type Processor<T> = (item: T) => void;

for (var i = 0; i < 10; i++) {
  const processor: Processor<number> = (item) => {
    return item + i;
  };
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: `for (var i = 0; i < 10; i++) {
  const process = (item: UnconfiguredGlobalType) => {
    console.log(i, item.value);
  };
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },

    // Destructuring bindings in for-in/of iterate on each step.
    {
      code: 'for (var {a} of arr) { (function() { a; }) }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var [[a]] of arr) { (function() { a; }) }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var {a = 10} of arr) { (function() { a; }) }',
      errors: [{ messageId: 'unsafeRefs' }],
    },

    // Destructuring assignment in the loop body is a write to the target.
    {
      code: 'var a; for (var i = 0; i < l; i++) { [a] = [i]; (function() { a; }) }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'var a; for (var i = 0; i < l; i++) { ({a} = {a: i}); (function() { a; }) }',
      errors: [{ messageId: 'unsafeRefs' }],
    },

    // Async function declaration nested in a loop.
    {
      code: 'for (var i = 0; i < l; i++) { async function f() { i; } }',
      errors: [{ messageId: 'unsafeRefs' }],
    },

    // The outer function leaks through-refs from inner nested functions.
    {
      code: 'for (var i = 0; i < l; i++) { function foo() { function bar() { i; } bar(); } }',
      errors: [{ messageId: 'unsafeRefs' }],
    },

    // Class/object methods, getters, setters, constructors inside a loop.
    {
      code: 'for (var i = 0; i < l; i++) { const o = { m() { return i; } }; }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var i = 0; i < l; i++) { class C { m() { return i; } } }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var i = 0; i < l; i++) { class C { get x() { return i; } } }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var i = 0; i < l; i++) { class C { set x(v) { this.a = i + v; } } }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var i = 0; i < l; i++) { class C { constructor() { this.a = i; } } }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var i = 0; i < l; i++) { const o = { async m() { return i; } }; }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
    {
      code: 'for (var i = 0; i < l; i++) { const o = { *m() { yield i; } }; }',
      errors: [{ messageId: 'unsafeRefs' }],
    },

    // A function inside a static block inside a loop.
    {
      code: `for (var i = 0; i < l; i++) {
    class C {
        static {
            var f = function() { i; };
        }
    }
}`,
      errors: [{ messageId: 'unsafeRefs' }],
    },

    // Parameter default value captures the loop var.
    {
      code: 'for (var i = 0; i < l; i++) { (function(x = i) { x; }); }',
      errors: [{ messageId: 'unsafeRefs' }],
    },

    // Computed key + body both reference the loop var.
    {
      code: 'for (var i = 0; i < l; i++) { var o = { [i]: function() { return i; } }; }',
      errors: [{ messageId: 'unsafeRefs' }],
    },

    // Outer binding reassigned in loop — method reads it unsafely.
    {
      code: 'var Foo; for (var i = 0; i < l; i++) { Foo = i; const C = class { m() { return Foo; } }; }',
      errors: [{ messageId: 'unsafeRefs' }],
    },

    // for-await-of with `var` iterator — re-bound on each iteration.
    {
      code: 'async function f(xs) { for await (var x of xs) { (function() { x; }) } }',
      errors: [{ messageId: 'unsafeRefs' }],
    },

    // Outer FD is flagged because a nested FE leaks an unsafe ref through it.
    {
      code: 'for (var i = 0; i < l; i++) { function outer() { (function() { i; }); } }',
      errors: [{ messageId: 'unsafeRefs' }],
    },

    // FD directly in loop body referencing the loop var.
    {
      code: 'for (var i = 0; i < l; i++) { function foo() { return i; } foo(); }',
      errors: [{ messageId: 'unsafeRefs' }],
    },

    // Two FD/FE reports in the same loop body (one for foo, one for g).
    {
      code: 'for (var i = 0; i < l; i++) { function foo() { return i; } foo = null; var g = function() { foo; }; }',
      errors: [{ messageId: 'unsafeRefs' }, { messageId: 'unsafeRefs' }],
    },

    // Multiple distinct unsafe refs listed in source order.
    {
      code: 'var a, b; for (var i = 0; i < l; i++) { a = i; b = i; (function() { a + b; }) }',
      errors: [{ messageId: 'unsafeRefs' }],
    },
  ],
});
