import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-this-assignment', null as never, {
  valid: [
    { code: 'const {property} = this;' },
    { code: 'const property = this.property;' },
    { code: 'const [element] = this;' },
    { code: 'const element = this[0];' },
    { code: '([element] = this);' },
    { code: 'element = this[0];' },
    { code: 'property = this.property;' },
    { code: 'const [element] = [this];' },
    { code: '([element] = [this]);' },
    { code: 'const {property} = {property: this};' },
    { code: '({property} = {property: this});' },
    { code: 'const self = true && this;' },
    { code: 'const self = false || this;' },
    { code: 'const self = false ?? this;' },
    { code: 'foo.bar = this;' },
    { code: 'function foo(a = this) {}' },
    { code: 'function foo({a = this}) {}' },
    { code: 'function foo([a = this]) {}' },
    { code: 'class A {\n\tfoo = this;\n}' },
    { code: 'class A {\n\tstatic foo = this;\n}' },
  ],
  invalid: [
    {
      code: 'const foo = this;',
      errors: [
        {
          messageId: 'no-this-assignment',
          message: 'Do not assign `this` to `foo`.',
        },
      ],
    },
    {
      code: 'let foo;foo = this;',
      errors: [
        {
          messageId: 'no-this-assignment',
          message: 'Do not assign `this` to `foo`.',
        },
      ],
    },
    {
      code: 'var foo = bar, baz = this;',
      errors: [
        {
          messageId: 'no-this-assignment',
          message: 'Do not assign `this` to `baz`.',
        },
      ],
    },
  ],
});
