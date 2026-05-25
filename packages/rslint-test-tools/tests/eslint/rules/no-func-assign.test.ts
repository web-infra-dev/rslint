import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-func-assign', {
  valid: [
    'function foo() { var foo = bar; }',
    'function foo(foo) { foo = bar; }',
    'function foo() { var foo; foo = bar; }',
    'var foo = () => {}; foo = bar;',
    'var foo = function() {}; foo = bar;',
    'var foo = function() { foo = bar; };',
    "import bar from 'bar'; function foo() { var foo = bar; }",
    'function foo() { let foo = 1; foo = 2; }',
    'function foo() { const foo = 1; }',
    'function foo() { function foo() {} }',
    'function foo() {} try {} catch(foo) { foo = 1; }',
  ],
  invalid: [
    {
      code: 'function foo() {}; foo = bar;',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: 'function foo() { foo = bar; }',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: 'foo = bar; function foo() { };',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: '[foo] = bar; function foo() { };',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: '({x: foo = 0} = bar); function foo() { };',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: 'function foo() { [foo] = bar; }',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: '(function() { ({x: foo = 0} = bar); function foo() { }; })();',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: 'var a = function foo() { foo = 123; };',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: 'function foo() {} ({foo} = bar);',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: 'function foo() {} for (foo in bar) {}',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: 'function foo() {} for (foo of bar) {}',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: 'function foo() {} ({...foo} = bar);',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: 'function foo() {} [...foo] = bar;',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: 'function foo() {} --foo;',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: 'function foo() {} foo--;',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: 'function foo() { function foo() {} foo = 1; }',
      errors: [{ messageId: 'isAFunction' }],
    },
    {
      code: 'function foo() {} try { foo = 1; } catch(foo) { foo = 2; }',
      errors: [{ messageId: 'isAFunction' }],
    },
  ],
});
