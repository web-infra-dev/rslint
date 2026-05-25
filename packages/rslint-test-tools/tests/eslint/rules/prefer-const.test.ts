import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-const', {
  valid: [
    // Already const
    'const x = 1;',
    // Reassignment operators
    'let x = 1; x = 2;',
    'let x = 1; x += 2;',
    'let x = 1; x++;',
    'let x = 1; ++x;',
    'let x = 1; x--;',
    'let x = 1; --x;',
    // Uninitialized, never assigned
    'let x;',
    'let x: number;',
    // Uninitialized, multiple assignments
    'let x; x = 0; x = 1;',
    // var (not let)
    'var x = 1;',
    // Reassigned in nested scopes
    'let x = 1; function f() { x = 2; }',
    'let x = 1; const f = () => { x = 2; };',
    'let x = 1; { x = 2; }',
    'let x = 1; if (true) { x = 2; }',
    // Reassigned via destructuring assignment
    'let x = 1; [x] = [2];',
    'let x = 1; ({x} = {x: 2});',
    'let a = 1; [{a}] = [{a: 2}];',
    // For loops
    'for (let i = 0; i < 10; i++) {}',
    'for (let i = 0; i < 10; i++) { i = 5; }',
    'for (let x = 10; x > 0; ) { break; }',
    'for (let x = 0, y = 10; x < y; ) { break; }',
    'for (let x in obj) { x = "modified"; }',
    'for (let x of arr) { x = "modified"; }',
    // destructuring: "all" — not all can be const
    {
      code: 'let {a, b} = {a: 1, b: 2}; b = 3;',
      options: { destructuring: 'all' },
    },
    {
      code: 'let [x, y] = [1, 2]; y = 3;',
      options: { destructuring: 'all' },
    },
    // ignoreReadBeforeAssign: true
    {
      code: 'let x; console.log(x); x = 0;',
      options: { ignoreReadBeforeAssign: true },
    },
    // destructuring: "all" — uninitialized, destructuring write, one has extra reassignment
    {
      code: 'let a: any, b: any; ({a, b} = ({} as any)); b = 1;',
      options: { destructuring: 'all' },
    },
    // destructuring: "all" — separate let statements, destructuring write, one reassigned
    {
      code: 'function f() { let a: any; let b: any; ({a, b} = ({} as any)); b = 1; void a; }',
      options: { destructuring: 'all' },
    },
    // Uninitialized, assigned inside nested block
    'let x: number; if (true) { x = 1; }',
    'let x: number; try { x = 1; } catch { x = 2; }',
    'let x: number; try { x = 1; } catch {}',
    'let x: number; { x = 1; }',
    'let x: number; for (let i = 0; i < 1; i++) { x = i; }',
    'let x: number; const fn = () => { x = 1; };',
    // Uninitialized, if without braces
    'function f() { let x: any; if (true) x = 0; return x; }',
    // Assignment in condition (not standalone)
    'function f() { let x: number; if (x = g()) { return x; } return 0; } function g(): number { return 1; }',
    'function f() { let x: string | null; while (x = g()) { void x; } } function g(): string | null { return null; }',
    'function f() { let x: number; for (; x = g(); ) { void x; } } function g(): number { return 0; }',
    // Chained assignment: inner write not standalone
    'let a = 0; let b: number; a = b = 1;',
    // Mixed member expression in destructuring assignment
    'function f() { let v: any; [({} as any).prop, v] = [1, 2]; return v; }',
    // Cross-declaration destructuring in different scope — unmergeable
    'function f() { let a: any; { let b: any; ({a, b} = ({} as any)); void b; } return a; }',
    // Cross-declaration array destructuring in different scope
    'function f() { let a: any; { let b: any; ([a, b] = ([] as any)); void b; } return a; }',
    // Cross-declaration with renamed property and member expression
    'let a: any; const b: any = {}; ({ a, c: (b as any).c } = ({} as any));',
  ],
  invalid: [
    // Simple let
    {
      code: 'let x = 1;',
      errors: [{ messageId: 'useConst' }],
    },
    {
      code: "let x = 'hello';",
      errors: [{ messageId: 'useConst' }],
    },
    // Object/array (property mutation ≠ reassignment)
    {
      code: 'let obj = {key: 0}; obj.key = 1;',
      errors: [{ messageId: 'useConst' }],
    },
    {
      code: 'let arr = [1, 2, 3];',
      errors: [{ messageId: 'useConst' }],
    },
    // Read only
    {
      code: 'let x = 1; console.log(x);',
      errors: [{ messageId: 'useConst' }],
    },
    // Multiple declarations, all never reassigned
    {
      code: 'let x = 1, y = 2;',
      errors: [{ messageId: 'useConst' }, { messageId: 'useConst' }],
    },
    // For-in/of without reassignment
    {
      code: 'for (let x in {a: 1}) { console.log(x); }',
      errors: [{ messageId: 'useConst' }],
    },
    {
      code: 'for (let x of [1, 2, 3]) { console.log(x); }',
      errors: [{ messageId: 'useConst' }],
    },
    // Function/arrow expression
    {
      code: 'let fn = function() {};',
      errors: [{ messageId: 'useConst' }],
    },
    {
      code: 'let fn = () => {};',
      errors: [{ messageId: 'useConst' }],
    },
    // Destructuring, none reassigned
    {
      code: 'let {a, b} = {a: 1, b: 2};',
      errors: [{ messageId: 'useConst' }, { messageId: 'useConst' }],
    },
    {
      code: 'let [x, y] = [1, 2];',
      errors: [{ messageId: 'useConst' }, { messageId: 'useConst' }],
    },
    // Uninitialized, standalone assignment
    {
      code: 'let x; x = 0;',
      errors: [{ messageId: 'useConst' }],
    },
    // Uninitialized, parenthesized assignment
    {
      code: 'let x: number; (x = 1);',
      errors: [{ messageId: 'useConst' }],
    },
    // Uninitialized, compound assignment
    {
      code: 'let x: any; x += 1;',
      errors: [{ messageId: 'useConst' }],
    },
    // Uninitialized, logical assignment
    {
      code: "let x: any; x ||= 'hi';",
      errors: [{ messageId: 'useConst' }],
    },
    // Uninitialized, array destructuring assignment
    {
      code: 'let x: number; [x] = [1];',
      errors: [{ messageId: 'useConst' }],
    },
    // Uninitialized, object shorthand destructuring assignment
    {
      code: 'let x: number; ({x} = {x: 1});',
      errors: [{ messageId: 'useConst' }],
    },
    // Uninitialized, object renamed destructuring assignment
    {
      code: 'let x: number; ({val: x} = {val: 1});',
      errors: [{ messageId: 'useConst' }],
    },
    // Uninitialized, array destructuring with default
    {
      code: 'let x: number; [x = 5] = [1];',
      errors: [{ messageId: 'useConst' }],
    },
    // Uninitialized, object renamed destructuring with default
    {
      code: 'let x: number; ({val: x = 5} = {val: 1});',
      errors: [{ messageId: 'useConst' }],
    },
    // Uninitialized, multiple via destructuring assignment (same declaration)
    {
      code: 'let a: number, b: number; [a, b] = [1, 2];',
      errors: [{ messageId: 'useConst' }, { messageId: 'useConst' }],
    },
    // destructuring options
    {
      code: 'let {a, b} = {a: 1, b: 2};',
      options: { destructuring: 'any' },
      errors: [{ messageId: 'useConst' }, { messageId: 'useConst' }],
    },
    {
      code: 'let {a, b} = {a: 1, b: 2};',
      options: { destructuring: 'all' },
      errors: [{ messageId: 'useConst' }, { messageId: 'useConst' }],
    },
    {
      code: 'let {a, b} = {a: 1, b: 2}; b = 3;',
      options: { destructuring: 'any' },
      errors: [{ messageId: 'useConst' }],
    },
    // Separate declarations
    {
      code: 'let {a} = {a: 1}; let {b} = {b: 2}; b = 1;',
      errors: [{ messageId: 'useConst' }],
    },
    // ignoreReadBeforeAssign: false
    {
      code: 'let x; x = 0;',
      options: { ignoreReadBeforeAssign: false },
      errors: [{ messageId: 'useConst' }],
    },
    // Cross-declaration destructuring in same scope — should report both
    {
      code: 'function f() { let a: any; let b: any; ({a, b} = ({} as any)); }',
      errors: [{ messageId: 'useConst' }, { messageId: 'useConst' }],
    },
    // Cross-declaration array destructuring in same scope
    {
      code: 'function f() { let a: any; let b: any; ([a, b] = ([] as any)); }',
      errors: [{ messageId: 'useConst' }, { messageId: 'useConst' }],
    },
    // Cross-declaration renamed object destructuring in same scope
    {
      code: 'function f() { let a: any; let b: any; ({x: a, y: b} = ({} as any)); }',
      errors: [{ messageId: 'useConst' }, { messageId: 'useConst' }],
    },
    // Cross-declaration in class static block
    {
      code: 'class C { static { let a: any; let b: any; ({a, b} = ({} as any)); } }',
      errors: [{ messageId: 'useConst' }, { messageId: 'useConst' }],
    },
    // destructuring: "all" — uninitialized, all targets have single write
    {
      code: 'let a: any, b: any; ({a, b} = ({} as any));',
      options: { destructuring: 'all' },
      errors: [{ messageId: 'useConst' }, { messageId: 'useConst' }],
    },
  ],
});
