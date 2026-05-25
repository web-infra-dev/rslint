import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-useless-constructor', {
  valid: [
    'class A { }',
    'class A { constructor(){ doSomething(); } }',
    'class A extends B { constructor(){} }',
    "class A extends B { constructor(){ super('foo'); } }",
    'class A extends B { constructor(foo, bar){ super(foo, bar, 1); } }',
    'class A extends B { constructor(){ super(); doSomething(); } }',
    'class A extends B { constructor(...args){ super(...args); doSomething(); } }',
    'class A { dummyMethod(){ doSomething(); } }',
    'class A extends B.C { constructor() { super(foo); } }',
    'class A extends B.C { constructor([a, b, c]) { super(...arguments); } }',
    'class A extends B.C { constructor(a = f()) { super(...arguments); } }',
    'class A extends B { constructor(a, b, c) { super(a, b); } }',
    'class A extends B { constructor(foo, bar){ super(foo); } }',
    'class A extends B { constructor(test) { super(); } }',
    'class A extends B { constructor() { foo; } }',
    'class A extends B { constructor(foo, bar) { super(bar); } }',

    // TypeScript overload / declare / abstract constructors (no body)
    'declare class A { constructor(options: any); }',
    `
class A {
  constructor();
}
    `,
    `
abstract class A {
  constructor();
}
    `,

    // Parameter properties
    `
class A {
  constructor(private name: string) {}
}
    `,
    `
class A {
  constructor(public name: string) {}
}
    `,
    `
class A {
  constructor(protected name: string) {}
}
    `,

    // Access modifier on constructor
    `
class A {
  private constructor() {}
}
    `,
    `
class A {
  protected constructor() {}
}
    `,
    `
class A extends B {
  public constructor() {}
}
    `,
    `
class A extends B {
  public constructor() {
    super();
  }
}
    `,
    `
class A extends B {
  protected constructor(foo, bar) {
    super(bar);
  }
}
    `,
    `
class A extends B {
  private constructor(foo, bar) {
    super(bar);
  }
}
    `,
    `
class A extends B {
  public constructor(foo) {
    super(foo);
  }
}
    `,
    `
class A extends B {
  public constructor(foo) {}
}
    `,

    // Decorator on params
    `
class A extends Object {
  constructor(@Foo foo: string) {
    super(foo);
  }
}
    `,
    `
class A extends Object {
  constructor(foo: string, @Bar() bar) {
    super(foo, bar);
  }
}
    `,
  ],
  invalid: [
    {
      code: 'class A { constructor(){} }',
      errors: [
        {
          messageId: 'noUselessConstructor',
          line: 1,
          column: 11,
        },
      ],
    },
    {
      code: "class A { 'constructor'(){} }",
      errors: [
        {
          messageId: 'noUselessConstructor',
          line: 1,
          column: 11,
        },
      ],
    },
    {
      code: 'class A extends B { constructor() { super(); } }',
      errors: [
        {
          messageId: 'noUselessConstructor',
          line: 1,
          column: 21,
        },
      ],
    },
    {
      code: 'class A extends B { constructor(foo){ super(foo); } }',
      errors: [
        {
          messageId: 'noUselessConstructor',
          line: 1,
          column: 21,
        },
      ],
    },
    {
      code: 'class A extends B { constructor(foo, bar){ super(foo, bar); } }',
      errors: [
        {
          messageId: 'noUselessConstructor',
          line: 1,
          column: 21,
        },
      ],
    },
    {
      code: 'class A extends B { constructor(...args){ super(...args); } }',
      errors: [
        {
          messageId: 'noUselessConstructor',
          line: 1,
          column: 21,
        },
      ],
    },
    {
      code: 'class A extends B.C { constructor() { super(...arguments); } }',
      errors: [
        {
          messageId: 'noUselessConstructor',
          line: 1,
          column: 23,
        },
      ],
    },
    {
      code: 'class A extends B { constructor(a, b, ...c) { super(...arguments); } }',
      errors: [
        {
          messageId: 'noUselessConstructor',
          line: 1,
          column: 21,
        },
      ],
    },
    {
      code: 'class A extends B { constructor(a, b, ...c) { super(a, b, ...c); } }',
      errors: [
        {
          messageId: 'noUselessConstructor',
          line: 1,
          column: 21,
        },
      ],
    },

    // Public constructor without superClass
    {
      code: `
class A {
  public constructor() {}
}
      `,
      errors: [
        {
          messageId: 'noUselessConstructor',
          line: 3,
          column: 3,
        },
      ],
    },
  ],
});
