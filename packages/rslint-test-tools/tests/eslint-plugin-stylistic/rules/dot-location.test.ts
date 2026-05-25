import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('dot-location', null as never, {
  valid: [
    // default (object)
    { code: 'obj.prop' },
    { code: 'obj.\nprop' },
    { code: '(obj).\nprop' },
    { code: "obj\n['prop']" },
    { code: "obj['prop']" },

    // explicit "object"
    { code: 'obj.\nprop', options: ['object'] },
    { code: 'obj . prop', options: ['object'] },
    { code: 'obj /* a */ . prop', options: ['object'] },
    { code: 'f(a\n).prop', options: ['object'] },
    { code: '(obj).prop', options: ['object'] },
    { code: '(\nobj\n).\nprop', options: ['object'] },

    // explicit "property"
    { code: 'obj\n.prop', options: ['property'] },
    { code: '(obj)\n.prop', options: ['property'] },
    { code: 'obj . prop', options: ['property'] },
    { code: 'obj . /* a */ prop', options: ['property'] },

    // bracket access skipped
    { code: 'obj\n[prop]', options: ['object'] },
    { code: 'obj\n[prop]', options: ['property'] },

    // optional chaining
    { code: 'obj?.\nprop', options: ['object'] },
    { code: 'obj\n?.prop', options: ['property'] },

    // private property
    { code: 'class C { #a; foo() { this.\n#a; } }', options: ['object'] },
    { code: 'class C { #a; foo() { this\n.#a; } }', options: ['property'] },

    // MetaProperty
    { code: 'import.meta' },

    // TS-only valid: TSQualifiedName / TSImportType
    { code: "type Foo = import('foo').\nProp", options: ['object'] },
    { code: "type Foo = import('foo')\n.Prop", options: ['property'] },
    { code: 'type Foo = Obj.\nProp', options: ['object'] },
    { code: 'type Foo = Obj\n.Prop', options: ['property'] },

    // JSX member expression (filename defaults to virtual.tsx)
    { code: 'const _ = <Form.\nInput />', options: ['object'] },
    { code: 'const _ = <Form\n.Input />', options: ['property'] },
  ],
  invalid: [
    {
      code: 'obj\n.property',
      output: 'obj.\nproperty',
      options: ['object'],
      errors: [
        {
          messageId: 'expectedDotAfterObject',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 2,
        },
      ],
    },
    {
      code: 'obj.\nproperty',
      output: 'obj\n.property',
      options: ['property'],
      errors: [
        {
          messageId: 'expectedDotBeforeProperty',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: '(obj).\nproperty',
      output: '(obj)\n.property',
      options: ['property'],
      errors: [
        {
          messageId: 'expectedDotBeforeProperty',
          line: 1,
          column: 6,
        },
      ],
    },
    {
      code: '5\n.toExponential()',
      output: '5 .\ntoExponential()',
      options: ['object'],
      errors: [
        {
          messageId: 'expectedDotAfterObject',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: '5_000\n.toExponential()',
      output: '5_000 .\ntoExponential()',
      options: ['object'],
      errors: [
        {
          messageId: 'expectedDotAfterObject',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: '0b1010_1010\n.toExponential()',
      output: '0b1010_1010.\ntoExponential()',
      options: ['object'],
      errors: [
        {
          messageId: 'expectedDotAfterObject',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: 'obj\n?.prop',
      output: 'obj?.\nprop',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject' }],
    },
    {
      code: 'obj?.\nprop',
      output: 'obj\n?.prop',
      options: ['property'],
      errors: [{ messageId: 'expectedDotBeforeProperty' }],
    },
    {
      code: 'class C { #a; foo() { this\n.#a; } }',
      output: 'class C { #a; foo() { this.\n#a; } }',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject' }],
    },
    {
      code: 'class C { #a; foo() { this.\n#a; } }',
      output: 'class C { #a; foo() { this\n.#a; } }',
      options: ['property'],
      errors: [{ messageId: 'expectedDotBeforeProperty' }],
    },
    // TSImportType
    {
      code: "type Foo = import('foo')\n.Prop",
      output: "type Foo = import('foo').\nProp",
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject' }],
    },
    {
      code: "type Foo = import('foo').\nProp",
      output: "type Foo = import('foo')\n.Prop",
      options: ['property'],
      errors: [{ messageId: 'expectedDotBeforeProperty' }],
    },
    // TSQualifiedName
    {
      code: 'type Foo = Obj\n.Prop',
      output: 'type Foo = Obj.\nProp',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject' }],
    },
    {
      code: 'type Foo = Obj.\nProp',
      output: 'type Foo = Obj\n.Prop',
      options: ['property'],
      errors: [{ messageId: 'expectedDotBeforeProperty' }],
    },
    // JSX member expression
    {
      code: 'const _ = <Form\n.Input />',
      output: 'const _ = <Form.\nInput />',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject' }],
    },
    {
      code: 'const _ = <Form.\nInput />',
      output: 'const _ = <Form\n.Input />',
      options: ['property'],
      errors: [{ messageId: 'expectedDotBeforeProperty' }],
    },
  ],
});
