import { RuleTester } from '@typescript-eslint/rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unused-private-class-members', {
  valid: [
    'class Foo {}',
    `
class Foo {
  publicMember = 42;
}
    `,
    `
class Foo {
  public publicMember = 42;
}
    `,
    `
class Foo {
  protected publicMember = 42;
}
    `,
    `
class C {
  #usedInInnerClass;

  method(a) {
    return class {
      foo = a.#usedInInnerClass;
    };
  }
}
    `,
    `
class C {
  private accessor accessorMember = 42;

  method() {
    return this.accessorMember;
  }
}
    `,
    `
class C {
  private static staticMember = 42;

  static method() {
    return this.staticMember;
  }
}
    `,
    `
class Test1 {
  constructor(private parameterProperty: number) {}
  method() {
    return this.parameterProperty;
  }
}
    `,
    `
class Foo {
  private prop: number;

  method(thing: Foo) {
    return thing.prop;
  }
}
    `,
    `
class Foo {
  private static staticProp: number;

  method(thing: typeof Foo) {
    return thing.staticProp;
  }
}
    `,
    `
class Foo {
  private prop: number;

  method() {
    const self = this;
    return self.prop;
  }
}
    `,
    `
class Foo {
  #privateMember = 42;
  method() {
    return this.#privateMember;
  }
}
    `,
    `
class Foo {
  private privateMember = 42;
  method() {
    return this.privateMember;
  }
}
    `,
    `
class C {
  set #privateMember(value) {
    doSomething(value);
  }
  get #privateMember() {
    return something();
  }
  method() {
    this.#privateMember += 1;
  }
}
    `,
    `
class C {
  private set privateMember(value) {
    doSomething(value);
  }
  private get privateMember() {
    return something();
  }
  method() {
    this.privateMember += 1;
  }
}
    `,
    `
class Foo {
  #privateMember() {
    return 42;
  }
  anotherMethod() {
    return this.#privateMember();
  }
}
    `,
    `
class Foo {
  private privateMember() {
    return 42;
  }
  anotherMethod() {
    return this.privateMember();
  }
}
    `,
  ],
  invalid: [
    {
      code: `
class C {
  private accessor accessorMember = 42;
}
      `,
      errors: [
        {
          data: {
            classMemberName: 'accessorMember',
          },
          messageId: 'unusedPrivateClassMember',
        },
      ],
    },
    {
      code: `
class C {
  private static staticMember = 42;
}
      `,
      errors: [
        {
          data: {
            classMemberName: 'staticMember',
          },
          messageId: 'unusedPrivateClassMember',
        },
      ],
    },
    {
      code: `
class Test1 {
  constructor(private parameterProperty: number) {}
}
      `,
      errors: [
        {
          data: {
            classMemberName: 'parameterProperty',
          },
          messageId: 'unusedPrivateClassMember',
        },
      ],
    },
    {
      code: `
class C {
  private usedOutsideClass;
}

const instance = new C();
console.log(instance.usedOutsideClass);
      `,
      errors: [
        {
          data: {
            classMemberName: 'usedOutsideClass',
          },
          messageId: 'unusedPrivateClassMember',
        },
      ],
    },
    {
      code: `
class Foo {
  #privateMember = 5;
}
      `,
      errors: [
        {
          data: {
            classMemberName: '#privateMember',
          },
          messageId: 'unusedPrivateClassMember',
        },
      ],
    },
    {
      code: `
class Foo {
  private privateMember = 5;
}
      `,
      errors: [
        {
          data: {
            classMemberName: 'privateMember',
          },
          messageId: 'unusedPrivateClassMember',
        },
      ],
    },
    {
      code: `
class Foo {
  #privateMember = 5;
  method() {
    this.#privateMember = 42;
  }
}
      `,
      errors: [
        {
          data: {
            classMemberName: '#privateMember',
          },
          messageId: 'unusedPrivateClassMember',
        },
      ],
    },
    {
      code: `
class Foo {
  private privateMember = 5;
  method() {
    this.privateMember = 42;
  }
}
      `,
      errors: [
        {
          data: {
            classMemberName: 'privateMember',
          },
          messageId: 'unusedPrivateClassMember',
        },
      ],
    },
    {
      code: `
class C {
  #privateMember;

  foo() {
    this.#privateMember++;
  }
}
      `,
      errors: [
        {
          data: {
            classMemberName: '#privateMember',
          },
          messageId: 'unusedPrivateClassMember',
        },
      ],
    },
    {
      code: `
class Foo {
  #privateMember() {}
}
      `,
      errors: [
        {
          data: {
            classMemberName: '#privateMember',
          },
          messageId: 'unusedPrivateClassMember',
        },
      ],
    },
    {
      code: `
class Foo {
  private privateMember() {}
}
      `,
      errors: [
        {
          data: {
            classMemberName: 'privateMember',
          },
          messageId: 'unusedPrivateClassMember',
        },
      ],
    },
    {
      code: `
class Foo {
  set #privateMember(value) {}
}
      `,
      errors: [
        {
          data: {
            classMemberName: '#privateMember',
          },
          messageId: 'unusedPrivateClassMember',
        },
      ],
    },
    {
      code: `
class Foo {
  private set privateMember(value) {}
}
      `,
      errors: [
        {
          data: {
            classMemberName: 'privateMember',
          },
          messageId: 'unusedPrivateClassMember',
        },
      ],
    },
  ],
});
