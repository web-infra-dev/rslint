import { RuleTester } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester();

ruleTester.run('no-this-alias', {
  valid: [
    'const self = foo(this);',
    {
      code: `
const { props, state } = this;
const { length } = this;
const { length, toString } = this;
const [foo] = this;
const [foo, bar] = this;
      `,
      options: [
        {
          allowDestructuring: true,
        },
      ],
    },
    {
      code: 'const self = this;',
      options: [
        {
          allowedNames: ['self'],
        },
      ],
    },
    // https://github.com/bradzacher/eslint-plugin-typescript/issues/281
    `
declare module 'foo' {
  declare const aVar: string;
}
    `,
  ],

  invalid: [
    {
      code: 'const self = this;',
      errors: [
        {
          messageId: 'thisAssignment',
          line: 1,
          column: 7,
        },
      ],
      options: [
        {
          allowDestructuring: true,
        },
      ],
    },
    {
      code: 'const self = this;',
      errors: [
        {
          messageId: 'thisAssignment',
          line: 1,
          column: 7,
        },
      ],
    },
    {
      code: `
let that;
that = this;
      `,
      errors: [
        {
          messageId: 'thisAssignment',
          line: 3,
          column: 1,
        },
      ],
    },
    {
      code: 'const { props, state } = this;',
      errors: [
        {
          messageId: 'thisDestructure',
          line: 1,
          column: 7,
        },
      ],
      options: [
        {
          allowDestructuring: false,
        },
      ],
    },
    {
      code: `
var unscoped = this;

function testFunction() {
  let inFunction = this;
}
const testLambda = () => {
  const inLambda = this;
};
      `,
      errors: [
        {
          messageId: 'thisAssignment',
          line: 2,
          column: 5,
        },
        {
          messageId: 'thisAssignment',
          line: 5,
          column: 7,
        },
        {
          messageId: 'thisAssignment',
          line: 8,
          column: 9,
        },
      ],
    },
    {
      code: `
class TestClass {
  constructor() {
    const inConstructor = this;
    const asThis: this = this;

    const asString = 'this';
    const asArray = [this];
    const asArrayString = ['this'];
  }

  public act(scope: this = this) {
    const inMemberFunction = this;
    const { act } = this;
    const { act, constructor } = this;
    const [foo] = this;
    const [foo, bar] = this;
  }
}
      `,
      errors: [
        {
          messageId: 'thisAssignment',
          line: 4,
          column: 11,
        },
        {
          messageId: 'thisAssignment',
          line: 5,
          column: 11,
        },
        {
          messageId: 'thisAssignment',
          line: 13,
          column: 11,
        },
        {
          messageId: 'thisDestructure',
          line: 14,
          column: 11,
        },
        {
          messageId: 'thisDestructure',
          line: 15,
          column: 11,
        },
        {
          messageId: 'thisDestructure',
          line: 16,
          column: 11,
        },
        {
          messageId: 'thisDestructure',
          line: 17,
          column: 11,
        },
      ],
      options: [
        {
          allowDestructuring: false,
        },
      ],
    },
  ],
});
