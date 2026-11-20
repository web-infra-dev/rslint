import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

// Upstream: https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/state-in-constructor.js
// The Go upstream suite also asserts message IDs, full ranges, and absence of edits.
function expectStateInitClassPropError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'stateInitClassProp',
    message: 'State initialization should be in a class property',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectStateInitConstructorError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'stateInitConstructor',
    message: 'State initialization should be in a constructor',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);
ruleTester.run('state-in-constructor', {
  valid: [
    {
      code: `
        class Foo extends React.Component {
          render() {
            return <div>Foo</div>
          }
        }
      `,
    },
    {
      code: `
        class Foo extends React.Component {
          render() {
            return <div>Foo</div>
          }
        }
      `,
      options: ['never'],
    },
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
    },
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      options: ['always'],
    },
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
    },
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.baz = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
    },
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.baz = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      options: ['never'],
    },
    {
      code: `
        class Foo extends React.Component {
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
    },
    {
      code: `
        class Foo extends React.Component {
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      options: ['never'],
    },
    {
      code: `
        const Foo = () => <div>Foo</div>
      `,
    },
    {
      code: `
        const Foo = () => <div>Foo</div>
      `,
      options: ['never'],
    },
    {
      code: `
        function Foo () {
          return <div>Foo</div>
        }
      `,
    },
    {
      code: `
        function Foo () {
          return <div>Foo</div>
        }
      `,
      options: ['never'],
    },
    {
      code: `
        class Foo extends React.Component {
          state = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      options: ['never'],
    },
    {
      code: `
        class Foo extends React.Component {
          state = { bar: 0 }
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      options: ['never'],
    },
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.baz = { bar: 0 }
          }
          state = { baz: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      options: ['never'],
    },
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            if (foobar) {
              this.state = { bar: 0 }
            }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
    },
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            foobar = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
    },
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            foobar = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      options: ['never'],
    },
  ],
  invalid: [
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      options: ['never'],
      errors: [expectStateInitClassPropError([5, 13, 5, 36])],
    },
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      options: ['never'],
      errors: [expectStateInitClassPropError([5, 13, 5, 36])],
    },
    {
      code: `
        class Foo extends React.Component {
          state = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      errors: [expectStateInitConstructorError([3, 11, 3, 29])],
    },
    {
      code: `
        class Foo extends React.Component {
          state = { bar: 0 }
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      errors: [expectStateInitConstructorError([3, 11, 3, 29])],
    },
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.baz = { bar: 0 }
          }
          state = { baz: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      errors: [expectStateInitConstructorError([7, 11, 7, 29])],
    },
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          state = { baz: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      errors: [expectStateInitConstructorError([7, 11, 7, 29])],
    },
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          state = { baz: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      options: ['never'],
      errors: [expectStateInitClassPropError([5, 13, 5, 36])],
    },
    {
      code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            if (foobar) {
              this.state = { bar: 0 }
            }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
      options: ['never'],
      errors: [expectStateInitClassPropError([6, 15, 6, 38])],
    },
  ],
});
