import { RuleTester } from '../rule-tester';

// Upstream: https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/state-in-constructor.js
// The Go upstream suite also asserts message IDs, full ranges, and absence of edits.
const ruleTester = new RuleTester();
ruleTester.run('state-in-constructor', {} as never, {
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
      errors: [
        {
          messageId: 'stateInitClassProp',
          message: 'State initialization should be in a class property',
          line: 5,
          column: 13,
          endLine: 5,
          endColumn: 36,
        },
      ],
      output: null,
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
      errors: [
        {
          messageId: 'stateInitClassProp',
          message: 'State initialization should be in a class property',
          line: 5,
          column: 13,
          endLine: 5,
          endColumn: 36,
        },
      ],
      output: null,
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
      errors: [
        {
          messageId: 'stateInitConstructor',
          message: 'State initialization should be in a constructor',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 29,
        },
      ],
      output: null,
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
      errors: [
        {
          messageId: 'stateInitConstructor',
          message: 'State initialization should be in a constructor',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 29,
        },
      ],
      output: null,
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
      errors: [
        {
          messageId: 'stateInitConstructor',
          message: 'State initialization should be in a constructor',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 29,
        },
      ],
      output: null,
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
      errors: [
        {
          messageId: 'stateInitConstructor',
          message: 'State initialization should be in a constructor',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 29,
        },
      ],
      output: null,
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
      errors: [
        {
          messageId: 'stateInitClassProp',
          message: 'State initialization should be in a class property',
          line: 5,
          column: 13,
          endLine: 5,
          endColumn: 36,
        },
      ],
      output: null,
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
      errors: [
        {
          messageId: 'stateInitClassProp',
          message: 'State initialization should be in a class property',
          line: 6,
          column: 15,
          endLine: 6,
          endColumn: 38,
        },
      ],
      output: null,
    },
  ],
});
