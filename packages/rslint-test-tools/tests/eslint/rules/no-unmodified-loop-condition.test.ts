import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unmodified-loop-condition', {
  valid: [
    // Basic modifications
    'var foo = 0; while (foo) { ++foo; }',
    'var foo = 0; while (foo) { foo += 1; }',
    'var foo = 0; while (foo < 10) { foo++; }',

    // Dynamic expressions in condition — skip check
    'while (ok(foo)) { }',
    'while (foo.ok) { }',
    'while (foo[0]) { }',

    // Comparison group: a < b is one group, a modified → OK
    'var a = 0, b = 10; while (a < b) { a++; }',

    // Modification via function call
    'var x = 0; function inc() { x++; } while (x < 10) { inc(); }',

    // Destructuring write
    'var x = 0; while (x < 10) { ({x} = {x: 1}); }',
    'var x = 0; while (x < 10) { [x] = [1]; }',

    // Arrow function in condition: not "dynamic"
    'var x = 0; while (x || (() => foo())) { x++; }',

    // Modification inside nested function DOES count (ESLint range-based)
    'var foo = 0; while (foo) { function f() { foo = 1; } }',
    'var foo = 0; while (foo) { var f = () => { foo = 1; }; }',

    // Modification in nested blocks (not function boundaries)
    'var x = 0; while (x < 10) { if (true) { x++; } }',
    'var x = 0; while (x < 10) { for (var i = 0; i < 1; i++) { x++; } }',

    // For-in/for-of as write target
    'var x = ""; while (x) { for (x in {a: 1}) {} }',

    // ConditionalExpression (ternary) as group
    'var a = 0, b = 0; while (a ? b : 0) { a++; }',
  ],
  invalid: [
    {
      code: 'var foo = 0; while (foo) { }',
      errors: [{ messageId: 'loopConditionNotModified' }],
    },
    // Both unmodified in comparison group
    {
      code: 'var a = 0, b = 0; while (a < b) { }',
      errors: [
        { messageId: 'loopConditionNotModified' },
        { messageId: 'loopConditionNotModified' },
      ],
    },
    // Variable shadowing
    {
      code: 'var foo = 0; while (foo) { let foo = 1; foo++; }',
      errors: [{ messageId: 'loopConditionNotModified' }],
    },
    // && — operands independent
    {
      code: 'var a = 0, b = 10; while (a && b) { a++; }',
      errors: [{ messageId: 'loopConditionNotModified' }],
    },
    // || partial — only a modified
    {
      code: 'var a = 0, b = 0; while (a || b) { a++; }',
      errors: [{ messageId: 'loopConditionNotModified' }],
    },
    // Function declared but NOT called in loop
    {
      code: 'var x = 0; function inc() { x++; } while (x < 10) { }',
      errors: [{ messageId: 'loopConditionNotModified' }],
    },
  ],
});
