import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string, filename = 'file.mjs') => ({ code, filename });
const invalid = (
  code: string,
  messageId: 'member-access' | 'complex-constructor',
  filename = 'file.mjs',
) => ({
  code,
  filename,
  errors: [
    {
      messageId,
      message:
        messageId === 'member-access'
          ? 'Do not access members directly from a `new` expression.'
          : 'Do not use a complex expression as a constructor.',
    },
  ],
});

ruleTester.run('no-unreadable-new-expression', null as never, {
  valid: [
    valid('const foo = new Foo();'),
    valid('const foo = new Foo;'),
    valid('const foo = new Foo(bar);'),
    valid('const foo = new Foo(...bar);'),
    valid('const foo = new Foo.Bar;'),
    valid('const bar = new foo.Bar();'),
    valid('const bar = new foo.bar.Baz();'),
    valid(
      'const formatter = new Intl.ListFormat("en-US", {type: "disjunction"});',
    ),
    valid('const foo = Foo().Bar;'),
    valid('const foo = Foo().Bar();'),
    valid('const foo = Foo.Bar();'),
    valid('const foo = new Foo(); foo.getBar();'),
    valid('const foo = new Foo(); const Bar = foo.Bar;'),
    valid('const {Bar} = foo; const bar = new Bar();'),
    valid('const Bar = foo.Bar; const bar = new Bar();'),
    valid('const bar = Foo ? new Foo() : foo.Bar;'),
    valid('const foo = new Foo<Type>();', 'file.ts'),
    valid('const foo = new Foo<Type>;', 'file.ts'),
  ],
  invalid: [
    invalid('const bar = new Foo().getBar();', 'member-access'),
    invalid('const bar = (new Foo()).getBar();', 'member-access'),
    invalid('const Bar = new Foo().Bar;', 'member-access'),
    invalid('const Bar = (new Foo()).Bar;', 'member-access'),
    invalid('const Bar = (new Foo).Bar;', 'member-access'),
    invalid('const Bar = new Foo()[Bar];', 'member-access'),
    invalid('const Bar = (new Foo())[Bar];', 'member-access'),
    invalid('const bar = (new Foo)?.getBar();', 'member-access'),
    invalid('const baz = new Foo().bar.baz;', 'member-access'),
    invalid('new Foo().bar`x`;', 'member-access'),
    invalid('const Bar = new (Foo().Bar);', 'complex-constructor'),
    invalid('const bar = new foo[Bar]();', 'complex-constructor'),
    invalid('const bar = (new foo).Bar();', 'member-access'),
    invalid('const bar = new foo().Bar();', 'member-access'),
    invalid('const bar = new (foo().Bar)();', 'complex-constructor'),
    invalid('const bar = new (foo())();', 'complex-constructor'),
    invalid('const bar = new (foo ? Foo : Bar)();', 'complex-constructor'),
    invalid('const bar = new class {}();', 'complex-constructor'),
    invalid('const timestamp = new Date().getTime();', 'member-access'),
    invalid(
      'const formatted = new Intl.ListFormat("en-US", {type: "disjunction"}).format(words);',
      'member-access',
    ),
    invalid(
      'const foo = new (Foo as typeof Bar)();',
      'complex-constructor',
      'file.ts',
    ),
    invalid('const bar = (new Foo<Type>()).bar;', 'member-access', 'file.ts'),
  ],
});
