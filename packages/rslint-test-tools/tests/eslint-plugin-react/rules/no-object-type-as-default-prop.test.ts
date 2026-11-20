// Upstream: https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/no-object-type-as-default-prop.js
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
import { diagnostic } from '../expectations';

function expectForbiddenTypeDefaultParamError1(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'forbiddenTypeDefaultParam',
    'a has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.',
    range,
  );
}

function expectForbiddenTypeDefaultParamError2(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'forbiddenTypeDefaultParam',
    'b has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.',
    range,
  );
}

function expectForbiddenTypeDefaultParamError3(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'forbiddenTypeDefaultParam',
    'c has a/an regex literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of regex literal.',
    range,
  );
}

function expectForbiddenTypeDefaultParamError4(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'forbiddenTypeDefaultParam',
    'd has a/an arrow function as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of arrow function.',
    range,
  );
}

function expectForbiddenTypeDefaultParamError5(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'forbiddenTypeDefaultParam',
    'e has a/an function expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of function expression.',
    range,
  );
}

function expectForbiddenTypeDefaultParamError6(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'forbiddenTypeDefaultParam',
    'f has a/an class expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of class expression.',
    range,
  );
}

function expectForbiddenTypeDefaultParamError7(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'forbiddenTypeDefaultParam',
    'g has a/an construction expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of construction expression.',
    range,
  );
}

function expectForbiddenTypeDefaultParamError8(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'forbiddenTypeDefaultParam',
    'h has a/an JSX element as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of JSX element.',
    range,
  );
}

function expectForbiddenTypeDefaultParamError9(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'forbiddenTypeDefaultParam',
    'i has a/an Symbol literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of Symbol literal.',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);
const expectedViolations = [
  expectForbiddenTypeDefaultParamError1([3, 11, 3, 17]),
  expectForbiddenTypeDefaultParamError2([4, 11, 4, 29]),
  expectForbiddenTypeDefaultParamError3([5, 11, 5, 23]),
  expectForbiddenTypeDefaultParamError4([6, 11, 6, 23]),
  expectForbiddenTypeDefaultParamError5([7, 11, 7, 28]),
  expectForbiddenTypeDefaultParamError6([8, 11, 8, 23]),
  expectForbiddenTypeDefaultParamError7([9, 11, 9, 26]),
  expectForbiddenTypeDefaultParamError8([10, 11, 10, 24]),
  expectForbiddenTypeDefaultParamError9([11, 11, 11, 28]),
];

ruleTester.run('no-object-type-as-default-prop', {
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
      errors: [
        expectForbiddenTypeDefaultParamError1([3, 11, 3, 17]),
        expectForbiddenTypeDefaultParamError2([4, 11, 4, 29]),
        expectForbiddenTypeDefaultParamError3([5, 11, 5, 23]),
        expectForbiddenTypeDefaultParamError4([6, 11, 6, 23]),
        expectForbiddenTypeDefaultParamError5([7, 11, 7, 28]),
        expectForbiddenTypeDefaultParamError6([8, 11, 8, 23]),
        expectForbiddenTypeDefaultParamError7([9, 11, 9, 26]),
        expectForbiddenTypeDefaultParamError8([10, 11, 10, 24]),
        expectForbiddenTypeDefaultParamError9([11, 11, 11, 28]),
      ],
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
      errors: [
        expectForbiddenTypeDefaultParamError1([3, 11, 3, 17]),
        expectForbiddenTypeDefaultParamError2([4, 11, 4, 29]),
        expectForbiddenTypeDefaultParamError3([5, 11, 5, 23]),
        expectForbiddenTypeDefaultParamError4([6, 11, 6, 23]),
        expectForbiddenTypeDefaultParamError5([7, 11, 7, 28]),
        expectForbiddenTypeDefaultParamError6([8, 11, 8, 23]),
        expectForbiddenTypeDefaultParamError7([9, 11, 9, 26]),
        expectForbiddenTypeDefaultParamError8([10, 11, 10, 24]),
        expectForbiddenTypeDefaultParamError9([11, 11, 11, 28]),
      ],
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
      errors: [
        expectForbiddenTypeDefaultParamError1([3, 11, 3, 17]),
        expectForbiddenTypeDefaultParamError2([4, 11, 4, 29]),
        expectForbiddenTypeDefaultParamError3([5, 11, 5, 23]),
        expectForbiddenTypeDefaultParamError4([6, 11, 6, 23]),
        expectForbiddenTypeDefaultParamError5([7, 11, 7, 28]),
        expectForbiddenTypeDefaultParamError6([8, 11, 8, 23]),
        expectForbiddenTypeDefaultParamError7([9, 11, 9, 26]),
        expectForbiddenTypeDefaultParamError8([10, 11, 10, 24]),
        expectForbiddenTypeDefaultParamError9([11, 11, 11, 28]),
      ],
    },
  ],
});
