import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string, filename = 'file.js') => ({ code, filename });
const invalid = (code: string, output: string, filename = 'file.js') => ({
  code,
  filename,
  output,
  errors: [
    {
      messageId: 'empty-brace-spaces',
      message: 'Do not add spaces between braces.',
    },
  ],
});

ruleTester.run('empty-brace-spaces', null as never, {
  valid: [
    valid(''),
    valid('class A {}'),
    valid('function f() {}'),
    valid('const o = {};'),
    valid('class A { /* comment */ }'),
    valid('function f() {\n\t// comment\n}'),
    valid('const o = { /* comment */ };'),
    valid('class A {static { /* comment */ }}'),
    valid('if (foo) {}'),
    valid('try {} catch {}'),
    valid('foo = {unicorn};'),
    valid('class A {baz() {}}'),
    // Not a BlockStatement, ClassBody, StaticBlock, or ObjectExpression.
    valid('switch (foo) { }'),
    valid('const { } = foo;'),
    valid('import { } from "foo";'),
    valid('const x: { } = {};', 'file.ts'),
    valid('type T = { };', 'file.ts'),
    valid('interface I { }', 'file.ts'),
    valid('namespace N { }', 'file.ts'),
    valid('enum E { }', 'file.ts'),
  ],
  invalid: [
    invalid('class A { }', 'class A {}'),
    invalid('class A {\n}', 'class A {}'),
    invalid('const o = { };', 'const o = {};'),
    invalid('foo = function () { };', 'foo = function () {};'),
    invalid('if (foo) { }', 'if (foo) {}'),
    invalid('if (foo) {} else { }', 'if (foo) {} else {}'),
    invalid('for (;;) { }', 'for (;;) {}'),
    invalid('while (foo) { }', 'while (foo) {}'),
    invalid('do { } while (foo)', 'do {} while (foo)'),
    invalid('try { } catch {}', 'try {} catch {}'),
    invalid('foo = () => { }', 'foo = () => {}'),
    invalid('class Foo {bar() { }}', 'class Foo {bar() {}}'),
    invalid('class A {static { }}', 'class A {static {}}'),
    invalid('class A { }', 'class A {}', 'file.ts'),
    invalid('abstract class A {\n}', 'abstract class A {}', 'file.ts'),
    invalid('const A = class { }', 'const A = class {}', 'file.ts'),
    invalid('class A<T> { }', 'class A<T> {}', 'file.ts'),
    invalid('f({ });', 'f({});'),
    invalid('const f = ({a}) => { }', 'const f = ({a}) => {}', 'file.ts'),
    // A brace pair in the class header must not be mistaken for the body's.
    invalid(
      'class A extends mixin({}) { }',
      'class A extends mixin({}) {}',
      'file.ts',
    ),
    invalid('class A<T extends {}> { }', 'class A<T extends {}> {}', 'file.ts'),
  ],
});
