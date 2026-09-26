// Upstream: https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/no-object-type-as-default-prop.js
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const expectedViolations = [
  {
    messageId: 'forbiddenTypeDefaultParam',
    message:
      'a has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.',
    line: 3,
    column: 11,
    endLine: 3,
    endColumn: 17,
  },
  {
    messageId: 'forbiddenTypeDefaultParam',
    message:
      'b has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.',
    line: 4,
    column: 11,
    endLine: 4,
    endColumn: 29,
  },
  {
    messageId: 'forbiddenTypeDefaultParam',
    message:
      'c has a/an regex literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of regex literal.',
    line: 5,
    column: 11,
    endLine: 5,
    endColumn: 23,
  },
  {
    messageId: 'forbiddenTypeDefaultParam',
    message:
      'd has a/an arrow function as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of arrow function.',
    line: 6,
    column: 11,
    endLine: 6,
    endColumn: 23,
  },
  {
    messageId: 'forbiddenTypeDefaultParam',
    message:
      'e has a/an function expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of function expression.',
    line: 7,
    column: 11,
    endLine: 7,
    endColumn: 28,
  },
  {
    messageId: 'forbiddenTypeDefaultParam',
    message:
      'f has a/an class expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of class expression.',
    line: 8,
    column: 11,
    endLine: 8,
    endColumn: 23,
  },
  {
    messageId: 'forbiddenTypeDefaultParam',
    message:
      'g has a/an construction expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of construction expression.',
    line: 9,
    column: 11,
    endLine: 9,
    endColumn: 26,
  },
  {
    messageId: 'forbiddenTypeDefaultParam',
    message:
      'h has a/an JSX element as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of JSX element.',
    line: 10,
    column: 11,
    endLine: 10,
    endColumn: 24,
  },
  {
    messageId: 'forbiddenTypeDefaultParam',
    message:
      'i has a/an Symbol literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of Symbol literal.',
    line: 11,
    column: 11,
    endLine: 11,
    endColumn: 28,
  },
];

ruleTester.run('no-object-type-as-default-prop', {} as never, {
  valid: [
    {
      code: `
      function Foo({
        bar = emptyFunction,
      }) {
        return null;
      }
    `,
    },
    {
      code: `
      function Foo({
        bar = emptyFunction,
        ...rest
      }) {
        return null;
      }
    `,
    },
    {
      code: `
      function Foo({
        bar = 1,
        baz = 'hello',
      }) {
        return null;
      }
    `,
    },
    {
      code: `
      function Foo(props) {
        return null;
      }
    `,
    },
    {
      code: `
      function Foo(props) {
        return null;
      }

      Foo.defaultProps = {
        bar: () => {}
      }
    `,
    },
    {
      code: `
      const Foo = () => {
        return null;
      };
    `,
    },
    {
      code: `
      const Foo = ({bar = 1}) => {
        return null;
      };
    `,
    },
    {
      code: `
      const Foo = ({bar = 1}, context) => {
        return null;
      };
    `,
    },
    {
      code: `
      export default function NotAComponent({foo = {}}) {}
    `,
    },
    // Exact upstream documentation examples omit component returns, so even
    // their invalid illustrations produce no diagnostics in v7.37.5.
    {
      code: `const emptyArray = [];

function Component({
  items = emptyArray,
}) {}`,
    },
    {
      code: `function Component({
  items = [],
}) {}`,
    },
    {
      code: `const Component = ({
  items = {},
}) => {}`,
    },
    {
      code: `const Component = ({
  items = () => {},
}) => {}`,
    },
    {
      code: `const emptyArray = [];

function Component({
  items = emptyArray,
}) {}`,
    },
    {
      code: `const emptyObject = {};
const Component = ({
  items = emptyObject,
}) => {}`,
    },
    {
      code: `const noopFunc = () => {};
const Component = ({
  items = noopFunc,
}) => {}`,
    },
    {
      code: `// primitives are all compared by value, so are safe to be inlined
function Component({
  num = 3,
  str = 'foo',
  bool = true,
}) {}`,
    },
  ],
  invalid: [
    {
      code: `
        function Foo({
          a = {},
          b = ['one', 'two'],
          c = /regex/i,
          d = () => {},
          e = function() {},
          f = class {},
          g = new Thing(),
          h = <Thing />,
          i = Symbol('foo')
        }) {
          return null;
        }
      `,
      errors: expectedViolations,
    },
    {
      code: `
        const Foo = ({
          a = {},
          b = ['one', 'two'],
          c = /regex/i,
          d = () => {},
          e = function() {},
          f = class {},
          g = new Thing(),
          h = <Thing />,
          i = Symbol('foo')
        }) => {
          return null;
        }
      `,
      errors: expectedViolations,
    },
    {
      code: `
        const Foo = ({
          a = {},
          b = ['one', 'two'],
          c = /regex/i,
          d = () => {},
          e = function() {},
          f = class {},
          g = new Thing(),
          h = <Thing />,
          i = Symbol('foo')
        }, context) => {
          return null;
        }
      `,
      errors: expectedViolations,
    },
  ],
});
