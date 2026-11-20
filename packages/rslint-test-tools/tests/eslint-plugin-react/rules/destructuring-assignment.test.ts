import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectUseDestructAssignmentError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'useDestructAssignment',
    message: 'Must use destructuring props assignment',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectDestructureInSignatureError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'destructureInSignature',
    message: 'Must destructure props in the function signature.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('destructuring-assignment', {
  valid: [
    {
      code: `
        const MyComponent = ({ id, className }) => (
          <div id={id} className={className} />
        );
      `,
    },
    {
      code: `
        const MyComponent = (props) => {
          const { id, className } = props;
          return <div id={id} className={className} />
        };
      `,
      options: ['always'],
    },
    {
      code: `
        const Foo = class extends React.PureComponent {
          render() {
            return <div>{this.props.foo}</div>;
          }
        };
      `,
      options: ['never'],
    },
    {
      code: `
        class Foo extends React.Component {
          bar = this.props.bar
        }
      `,
      options: ['always', { ignoreClassFields: true }],
    },
    {
      code: `
        function Foo(props) {
          const {a} = props;
          return <Goo {...props}>{a}</Goo>;
        }
      `,
      options: ['always', { destructureInSignature: 'always' }],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        export const revisionStates2 = {
            [A.b]: props => {
              return props.editor !== null
                ? 'xyz'
                : 'abc'
            },
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        export function hof(namespace) {
          const initialState = {
            bounds: null,
            search: false,
          };
          return (props) => {
            const {x, y} = props
            if (y) {
              return <span>{y}</span>;
            }
            return <span>{x}</span>
          };
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        export function hof(namespace) {
          const initialState = {
            bounds: null,
            search: false,
          };

          return (state = initialState, action) => {
            if (action.type === 'ABC') {
              return {...state, bounds: stuff ? action.x : null};
            }

            if (action.namespace !== namespace) {
              return state;
            }

            return null
          };
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        const MyComponent = ({ id, className }) => (
          <div id={id} className={className} />
        );
      `,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        const MyComponent = (props) => {
          const { id, className } = props;
          return <div id={id} className={className} />
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        const MyComponent = (props) => (
          <div id={id} props={props} />
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        const MyComponent = (props) => (
          <div id={id} props={props} />
        );
      `,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        const MyComponent = (props, { color }) => (
          <div id={id} props={props} color={color} />
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        const MyComponent = (props, { color }) => (
          <div id={id} props={props} color={color} />
        );
      `,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        class Foo extends React.Component {
          doStuff() {}
          render() {
            return <div>{this.props.foo}</div>;
          }
        }
      `,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        const Foo = class extends React.PureComponent {
          render() {
            const { foo } = this.props;
            return <div>{foo}</div>;
          }
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        const Foo = class extends React.PureComponent {
          render() {
            const { foo } = this.props;
            return <div>{foo}</div>;
          }
        };
      `,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        const MyComponent = (props) => {
          const { h, i } = hi;
          return <div id={props.id} className={props.className} />
        };
      `,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        const Foo = class extends React.PureComponent {
          constructor() {
            this.state = {};
            this.state.foo = 'bar';
          }
        };
      `,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        const div = styled.div\`
          & .button {
            border-radius: \${props => props.borderRadius}px;
          }
        \`
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        export default (context: $Context) => ({
          foo: context.bar
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        class Foo {
          bar(context) {
            return context.baz;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        class Foo {
          bar(props) {
            return props.baz;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        class Input extends React.Component {
          id = \`\${this.props.name}\`;
          render() {
            return <div />;
          }
        }
      `,
      options: ['always', { ignoreClassFields: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        function Foo({ context }) {
          const d = context.describe();
          return <div>{d}</div>;
        }
      `,
      options: ['always'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        const obj = {
          foo(arg) {
            const a = arg.func();
            return null;
          },
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        const columns = [
          {
            render: (val) => {
              if (val.url) {
                return (
                  <a href={val.url}>
                    {val.test}
                  </a>
                );
              }
              return null;
            },
          },
        ];
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        const columns = [
          {
            render: val => <span>{val}</span>,
          },
          {
            someRenderFunc: function(val) {
              if (val.url) {
                return (
                  <a href={val.url}>
                    {val.test}
                  </a>
                );
              }
              return null;
            },
          },
        ];
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        export default (fileName) => {
          const match = fileName.match(/some expression/);
          if (match) {
            return fn;
          }
          return null;
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        class C extends React.Component {
          componentDidMount() {
            const { forwardRef } = this.props;

            this.ref.current.focus();

            if (typeof forwardRef === 'function') {
              forwardRef(this.ref);
            }
          }
          render() {
            return <div />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        function Foo(props) {
          const {a} = props;
          return <Goo f={() => props}>{a}</Goo>;
        }
      `,
      options: ['always', { destructureInSignature: 'always' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        import { useContext } from 'react';

        const MyComponent = (props) => {
          const {foo} = useContext(aContext);
          return <div>{foo}</div>
        };
      `,
      options: ['always'],
      settings: { react: { version: '16.9.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
        import { useContext } from 'react';

        const MyComponent = (props) => {
          const foo = useContext(aContext);
          return <div>{foo.test}</div>
        };
      `,
      options: ['never'],
      settings: { react: { version: '16.9.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        import { useContext } from 'react';

        const MyComponent = (props) => {
          const foo = useContext(aContext);
          return <div>{foo.test}</div>
        };
      `,
      options: ['always'],
      settings: { react: { version: '16.9.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
        const MyComponent = (props) => {
          const foo = useContext(aContext);
          return <div>{foo.test}</div>
        };
      `,
      options: ['always'],
      settings: { react: { version: '16.8.999' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
        const MyComponent = (props) => {
          const {foo} = useContext(aContext);
          return <div>{foo}</div>
        };
      `,
      options: ['never'],
      settings: { react: { version: '16.8.999' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `
        const MyComponent = (props) => {
          const {foo} = useContext(aContext);
          return <div>{foo}</div>
        };
      `,
      options: ['always'],
      settings: { react: { version: '16.8.999' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
      code: `
        const MyComponent = (props) => {
          const foo = useContext(aContext);
          return <div>{foo.test}</div>
        };
      `,
      options: ['never'],
      settings: { react: { version: '16.8.999' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
      code: `
        import { useContext } from 'react';

        const MyComponent = (props) => {
          const foo = useContext(aContext);
          return <div>{foo?.test}</div>
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `
        const MyComponent = (props) => {
          return (<div id={props.id} />)
        };
      `,
      errors: [expectUseDestructAssignmentError([3, 28, 3, 36])],
    },
    {
      code: `
        const MyComponent = ({ id, className }) => (
          <div id={id} className={className} />
        );
      `,
      options: ['never'],
      errors: [
        {
          messageId: 'noDestructPropsInSFCArg',
          message:
            'Must never use destructuring props assignment in SFC argument',
          line: 2,
          column: 29,
          endLine: 4,
          endColumn: 10,
        },
      ],
    },
    {
      code: `
        const Foo = class extends React.PureComponent {
          render() {
            return <div>{this.props.foo}</div>;
          }
        };
      `,
      errors: [expectUseDestructAssignmentError([4, 26, 4, 40])],
    },
    {
      code: `
        const Foo = class extends React.PureComponent {
          render() {
            return <div>{this.state.foo}</div>;
          }
        };
      `,
      errors: [
        {
          messageId: 'useDestructAssignment',
          message: 'Must use destructuring state assignment',
          line: 4,
          column: 26,
          endLine: 4,
          endColumn: 40,
        },
      ],
    },
    {
      code: `
        const Foo = class extends React.PureComponent {
          render() {
            const { foo } = this.props;
            return <div>{foo}</div>;
          }
        };
      `,
      options: ['never'],
      errors: [
        {
          messageId: 'noDestructAssignment',
          message: 'Must never use destructuring props assignment',
          line: 4,
          column: 19,
          endLine: 4,
          endColumn: 39,
        },
      ],
    },
    {
      code: `
        function Foo(props) {
          const {a} = props;
          return <p>{a}</p>;
        }
      `,
      options: ['always', { destructureInSignature: 'always' }],
      errors: [expectDestructureInSignatureError([3, 17, 3, 28])],
      output: `
        function Foo({a}) {
\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20
          return <p>{a}</p>;
        }
      `,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        const MyComponent = (props, { color }) => (
          <div id={props.id} className={props.className} />
        );
      `,
      options: ['never'],
      errors: [
        {
          messageId: 'noDestructContextInSFCArg',
          message:
            'Must never use destructuring context assignment in SFC argument',
          line: 2,
          column: 29,
          endLine: 4,
          endColumn: 10,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        const Foo = class extends React.PureComponent {
          render() {
            return <div>{this.context.foo}</div>;
          }
        };
      `,
      errors: [
        {
          messageId: 'useDestructAssignment',
          message: 'Must use destructuring context assignment',
          line: 4,
          column: 26,
          endLine: 4,
          endColumn: 42,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        class Foo extends React.Component {
          render() { return this.foo(); }
          foo() {
            return this.props.children;
          }
        }
      `,
      errors: [expectUseDestructAssignmentError([5, 20, 5, 39])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <Text>{this.props.foo}</Text>;
          }
        });
      `,
      errors: [expectUseDestructAssignmentError([4, 27, 4, 41])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        module.exports = {
          Foo(props) {
            return <p>{props.a}</p>;
          }
        }
      `,
      errors: [expectUseDestructAssignmentError([4, 24, 4, 31])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        export default function Foo(props) {
          return <p>{props.a}</p>;
        }
      `,
      errors: [expectUseDestructAssignmentError([3, 22, 3, 29])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        function hof() {
          return (props) => <p>{props.a}</p>;
        }
      `,
      errors: [expectUseDestructAssignmentError([3, 33, 3, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        const Foo = class extends React.PureComponent {
          render() {
            const foo = this.props.foo;
            return <div>{foo}</div>;
          }
        };
        `,
      errors: [expectUseDestructAssignmentError([4, 25, 4, 39])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
        const MyComponent = (props) => {
          const { id, className } = props;
          return <div id={id} className={className} />
        };
      `,
      options: ['never'],
      errors: [
        {
          messageId: 'noDestructAssignment',
          message: 'Must never use destructuring props assignment',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 42,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        const Foo = class extends React.PureComponent {
          render() {
            const { foo } = this.state;
            return <div>{foo}</div>;
          }
        };
      `,
      options: ['never'],
      errors: [
        {
          messageId: 'noDestructAssignment',
          message: 'Must never use destructuring state assignment',
          line: 4,
          column: 19,
          endLine: 4,
          endColumn: 39,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        const columns = [
          {
            CustomComponentName: function(props) {
              if (props.url) {
                return (
                  <a href={props.url}>
                    {props.test}
                  </a>
                );
              }
              return null;
            },
          },
        ];
      `,
      errors: [
        expectUseDestructAssignmentError([5, 19, 5, 28]),
        expectUseDestructAssignmentError([7, 28, 7, 37]),
        expectUseDestructAssignmentError([8, 22, 8, 32]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        function Foo(props, context) {
          const d = context.describe();
          return <div>{d}</div>;
        }
      `,
      options: ['always'],
      errors: [
        {
          messageId: 'useDestructAssignment',
          message: 'Must use destructuring context assignment',
          line: 3,
          column: 21,
          endLine: 3,
          endColumn: 37,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
        export default (props) => {
          const match = props.str.match(/some expression/);
          if (match) {
            return <span>jsx</span>;
          }
          return null;
        };
      `,
      options: ['always'],
      errors: [expectUseDestructAssignmentError([3, 25, 3, 34])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        import React from 'react';

        const TestComp = (props) => {
          props.onClick3102();

          return (
            <div
              onClick={(evt) => {
                if (props.onClick3102) {
                  props.onClick3102(evt);
                }
              }}
            >
              <div />
            </div>
          );
        };
      `,
      errors: [
        expectUseDestructAssignmentError([5, 11, 5, 28]),
        expectUseDestructAssignmentError([10, 21, 10, 38]),
        expectUseDestructAssignmentError([11, 19, 11, 36]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
        export const revisionStates2 = {
            [A.b]: props => {
              return props.editor !== null
                ? <span>{props.editor}</span>
                : null
            },
        };
      `,
      errors: [
        expectUseDestructAssignmentError([4, 22, 4, 34]),
        expectUseDestructAssignmentError([5, 26, 5, 38]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        export function hof(namespace) {
          const initialState = {
            bounds: null,
            search: false,
          };
          return (props) => {
            if (props.y) {
              return <span>{props.y}</span>;
            }
            return <span>{props.x}</span>
          };
        }
      `,
      errors: [
        expectUseDestructAssignmentError([8, 17, 8, 24]),
        expectUseDestructAssignmentError([9, 29, 9, 36]),
        expectUseDestructAssignmentError([11, 27, 11, 34]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
          function Foo(props) {
            const {a} = props;
            return <p>{a}</p>;
          }
        `,
      options: ['always', { destructureInSignature: 'always' }],
      errors: [expectDestructureInSignatureError([3, 19, 3, 30])],
      output: `
          function Foo({a}) {
\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20
            return <p>{a}</p>;
          }
        `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `
          function Foo(props: FooProps) {
            const {a} = props;
            return <p>{a}</p>;
          }
        `,
      options: ['always', { destructureInSignature: 'always' }],
      errors: [expectDestructureInSignatureError([3, 19, 3, 30])],
      output: `
          function Foo({a}: FooProps) {
\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20
            return <p>{a}</p>;
          }
        `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        type Props = { text: string };
        export const MyComponent: React.FC<Props> = (props) => {
          type MyType = typeof props.text;
          return <div>{props.text as MyType}</div>;
        };
      `,
      options: ['always', { destructureInSignature: 'always' }],
      errors: [
        expectUseDestructAssignmentError([4, 32, 4, 42]),
        expectUseDestructAssignmentError([5, 24, 5, 34]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
        type Props = { text: string };
        export const MyOtherComponent: React.FC<Props> = (props) => {
          const { text } = props;
          type MyType = typeof props.text;
          return <div>{text as MyType}</div>;
        };
      `,
      options: ['always', { destructureInSignature: 'always' }],
      errors: [expectUseDestructAssignmentError([5, 32, 5, 42])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
        function C(props: Props) {
          void props.a
          typeof props.b
          return <div />
        }
      `,
      options: ['always'],
      errors: [
        expectUseDestructAssignmentError([3, 16, 3, 23]),
        expectUseDestructAssignmentError([4, 18, 4, 25]),
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
