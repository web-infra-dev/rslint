import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unused-private-class-members', {
  valid: [
    'class Foo {}',
    'class Foo {\n    publicMember = 42;\n}',
    'class Foo {\n    #usedMember = 42;\n    method() {\n        return this.#usedMember;\n    }\n}',
    'class Foo {\n    #usedMember = 42;\n    anotherMember = this.#usedMember;\n}',
    'class Foo {\n    #usedMember = 42;\n    foo() {\n        anotherMember = this.#usedMember;\n    }\n}',
    'class C {\n    #usedMember;\n\n    foo() {\n        bar(this.#usedMember += 1);\n    }\n}',
    'class Foo {\n    #usedMember = 42;\n    method() {\n        return someGlobalMethod(this.#usedMember);\n    }\n}',
    'class C {\n    #usedInOuterClass;\n\n    foo() {\n        return class {};\n    }\n\n    bar() {\n        return this.#usedInOuterClass;\n    }\n}',
    'class Foo {\n    #usedInForInLoop;\n    method() {\n        for (const bar in this.#usedInForInLoop) {\n\n        }\n    }\n}',
    'class Foo {\n    #usedInForOfLoop;\n    method() {\n        for (const bar of this.#usedInForOfLoop) {\n\n        }\n    }\n}',
    'class Foo {\n    #usedInAssignmentPattern;\n    method() {\n        [bar = 1] = this.#usedInAssignmentPattern;\n    }\n}',
    'class Foo {\n    #usedInArrayPattern;\n    method() {\n        [bar] = this.#usedInArrayPattern;\n    }\n}',
    'class Foo {\n    #usedInAssignmentPattern;\n    method() {\n        [bar] = this.#usedInAssignmentPattern;\n    }\n}',
    'class C {\n    #usedInObjectAssignment;\n\n    method() {\n        ({ [this.#usedInObjectAssignment]: a } = foo);\n    }\n}',
    'class C {\n    set #accessorWithSetterFirst(value) {\n        doSomething(value);\n    }\n    get #accessorWithSetterFirst() {\n        return something();\n    }\n    method() {\n        this.#accessorWithSetterFirst += 1;\n    }\n}',
    'class Foo {\n    set #accessorUsedInMemberAccess(value) {}\n\n    method(a) {\n        [this.#accessorUsedInMemberAccess] = a;\n    }\n}',
    'class C {\n    get #accessorWithGetterFirst() {\n        return something();\n    }\n    set #accessorWithGetterFirst(value) {\n        doSomething(value);\n    }\n    method() {\n        this.#accessorWithGetterFirst += 1;\n    }\n}',
    'class C {\n    #usedInInnerClass;\n\n    method(a) {\n        return class {\n            foo = a.#usedInInnerClass;\n        }\n    }\n}',
    'class Foo {\n    #usedMethod() {\n        return 42;\n    }\n    anotherMethod() {\n        return this.#usedMethod();\n    }\n}',
    'class C {\n    set #x(value) {\n        doSomething(value);\n    }\n\n    foo() {\n        this.#x = 1;\n    }\n}',
  ],
  invalid: [
    {
      code: 'class Foo {\n    #unusedMember = 5;\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class Foo {\n    /** docs */\n    #unusedMember = 1;\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class Foo {\n    // remove me\n    #unusedMember = 1;\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class Foo {\n    /* remove */ #unusedMember = 1;\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'class Foo {\n    /* keep */ #unusedMember = 1; foo = 1\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 16,
          endLine: 2,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'class C {\n    #unused1; /* keep */ foo;\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'class C {\n    bar; #unused2; // keep\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class C {\n    // comment\n    #unused; foo;\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'class C {\n    // comment\n    #unused; /*\n    */ foo;\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'class Foo {\n    #unusedMember = 1; // trailing\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class Foo {\n    foo = 1; /*\n    */ #unusedMember = 1;\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'class Foo {\n    foo = 1; // keep this\n    #unusedMember = 1;\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class First {}\nclass Second {\n    #unusedMemberInSecondClass = 5;\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'class First {\n    #unusedMemberInFirstClass = 5;\n}\nclass Second {}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'class First {\n    #firstUnusedMemberInSameClass = 5;\n    #secondUnusedMemberInSameClass = 5;\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 34,
        },
        {
          messageId: 'unusedPrivateClassMember',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'class Foo {\n    #usedOnlyInWrite = 5;\n    method() {\n        this.#usedOnlyInWrite = 42;\n    }\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'class Foo {\n    #usedOnlyInWriteStatement = 5;\n    method() {\n        this.#usedOnlyInWriteStatement += 42;\n    }\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'class C {\n    #usedOnlyInIncrement;\n\n    foo() {\n        this.#usedOnlyInIncrement++;\n    }\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'class C {\n    #unusedInOuterClass;\n\n    foo() {\n        return class {\n            #unusedInOuterClass;\n\n            bar() {\n                return this.#unusedInOuterClass;\n            }\n        };\n    }\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'class C {\n    #unusedOnlyInSecondNestedClass;\n\n    foo() {\n        return class {\n            #unusedOnlyInSecondNestedClass;\n\n            bar() {\n                return this.#unusedOnlyInSecondNestedClass;\n            }\n        };\n    }\n\n    baz() {\n        return this.#unusedOnlyInSecondNestedClass;\n    }\n\n    bar() {\n        return class {\n            #unusedOnlyInSecondNestedClass;\n        }\n    }\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 20,
          column: 13,
          endLine: 20,
          endColumn: 43,
        },
      ],
    },
    {
      code: 'class Foo {\n    #unusedMethod() {}\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class Foo {\n    #unusedMethod() {}\n    #usedMethod() {\n        return 42;\n    }\n    publicMethod() {\n        return this.#usedMethod();\n    }\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class Foo {\n    set #unusedSetter(value) {}\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'class Foo {\n    get #unusedAccessor() {\n        return 1;\n    }\n    set #unusedAccessor(value) {}\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'class Foo {\n    #unusedForInLoop;\n    method() {\n        for (this.#unusedForInLoop in bar) {\n\n        }\n    }\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'class Foo {\n    #unusedForOfLoop;\n    method() {\n        for (this.#unusedForOfLoop of bar) {\n\n        }\n    }\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'class Foo {\n    #unusedInDestructuring;\n    method() {\n        ({ x: this.#unusedInDestructuring } = bar);\n    }\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'class Foo {\n    #unusedInRestPattern;\n    method() {\n        [...this.#unusedInRestPattern] = bar;\n    }\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'class Foo {\n    #unusedInAssignmentPattern;\n    method() {\n        [this.#unusedInAssignmentPattern = 1] = bar;\n    }\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'class Foo {\n    #unusedInAssignmentPattern;\n    method() {\n        [this.#unusedInAssignmentPattern] = bar;\n    }\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'class Foo {\n    foo = 1\n    #unusedMethod() {}\n    [0]() {}\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class Foo {\n    foo = 1\n    #unusedMethod() {}\n    *generator() {}\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class Foo {\n    foo = 1\n    #unusedMethod() {}\n    in = 2\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class Foo {\n    foo = 1\n    #unused\n    instanceof() {}\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'class C {\n    foo = () => {}\n    #unused\n    [bar]\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'class C {\n    foo\n    #unused\n    [bar]\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'class Foo {\n    foo = 1\n    /** docs */\n    #unusedMethod() {}\n    [0]() {}\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class Foo {\n    // keep\n\n    /** remove */\n    #unusedMember = 1;\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 5,
          column: 5,
          endLine: 5,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class Foo {\n    // keep one\n    // keep two\n\n    /** remove */\n    #unusedMember = 1;\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 6,
          column: 5,
          endLine: 6,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class Foo {\n    // maybe unrelated\n\n    #unusedMember = 1;\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class C {\n    #usedOnlyInTheSecondInnerClass;\n\n    method(a) {\n        return class {\n            #usedOnlyInTheSecondInnerClass;\n\n            method2(b) {\n                foo = b.#usedOnlyInTheSecondInnerClass;\n            }\n\n            method3(b) {\n                foo = b.#usedOnlyInTheSecondInnerClass;\n            }\n        }\n    }\n}',
      errors: [
        {
          messageId: 'unusedPrivateClassMember',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 35,
        },
      ],
    },
  ],
});
