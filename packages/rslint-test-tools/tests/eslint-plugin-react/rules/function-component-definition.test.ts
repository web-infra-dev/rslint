import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('function-component-definition', {} as never, {
  valid: [
    {
      code: `
        class Hello extends React.Component {
          render() { return <div>Hello {this.props.name}</div> }
        }
      `,
      options: [{ namedComponents: 'arrow-function' }],
    },
    {
      code: `
        class Hello extends React.Component {
          render() { return <div>Hello {this.props.name}</div> }
        }
      `,
      options: [{ namedComponents: 'function-declaration' }],
    },
    {
      code: `
        class Hello extends React.Component {
          render() { return <div>Hello {this.props.name}</div> }
        }
      `,
      options: [{ namedComponents: 'function-expression' }],
    },
    {
      code: 'var Hello = (props) => { return <div/> }',
      options: [{ namedComponents: 'arrow-function' }],
    },
    {
      code: 'const Hello = (props) => { return <div/> }',
      options: [{ namedComponents: 'arrow-function' }],
    },
    {
      code: 'function Hello(props) { return <div/> }',
      options: [{ namedComponents: 'function-declaration' }],
    },
    {
      code: 'var Hello = function(props) { return <div/> }',
      options: [{ namedComponents: 'function-expression' }],
    },
    {
      code: 'const Hello = function(props) { return <div/> }',
      options: [{ namedComponents: 'function-expression' }],
    },
    {
      code: 'function Hello() { return function() { return <div/> } }',
      options: [{ unnamedComponents: 'function-expression' }],
    },
    {
      code: 'function Hello() { return () => { return <div/> }}',
      options: [{ unnamedComponents: 'arrow-function' }],
    },
    {
      code: 'var Foo = React.memo(function Foo() { return <p/> })',
      options: [{ namedComponents: 'function-declaration' }],
    },
    {
      code: 'const Foo = React.memo(function Foo() { return <p/> })',
      options: [{ namedComponents: 'function-declaration' }],
    },
    {
      // shouldn't trigger this rule since functions stating with a lowercase
      // letter are not considered components
      code: `
        const selectAvatarByUserId = (state, id) => {
          const user = selectUserById(state, id)
          return null
        }
      `,
      options: [{ namedComponents: 'function-declaration' }],
    },
    {
      // shouldn't trigger this rule since functions stating with a lowercase
      // letter are not considered components
      code: `
        function ensureValidSourceType(sourceType: string) {
          switch (sourceType) {
            case 'ALBUM':
            case 'PLAYLIST':
              return sourceType;
            default:
              return null;
          }
        }
      `,
      options: [{ namedComponents: 'arrow-function' }],
    },
    {
      code: 'function Hello(props: Test) { return <p/> }',
      options: [{ namedComponents: 'function-declaration' }],
    },
    {
      code: 'var Hello = function(props: Test) { return <p/> }',
      options: [{ namedComponents: 'function-expression' }],
    },
    {
      code: 'var Hello = (props: Test) => { return <p/> }',
      options: [{ namedComponents: 'arrow-function' }],
    },
    {
      code: 'var Hello: React.FC<Test> = function(props) { return <p/> }',
      options: [{ namedComponents: 'function-expression' }],
    },
    {
      code: 'var Hello: React.FC<Test> = (props) => { return <p/> }',
      options: [{ namedComponents: 'arrow-function' }],
    },
    {
      code: 'function Hello<Test>(props: Props<Test>) { return <p/> }',
      options: [{ namedComponents: 'function-declaration' }],
    },
    {
      code: 'function Hello<Test extends {}>(props: Props<Test>) { return <p/> }',
      options: [{ namedComponents: 'function-declaration' }],
    },
    {
      code: 'var Hello = function<Test>(props: Props<Test>) { return <p/> }',
      options: [{ namedComponents: 'function-expression' }],
    },
    {
      code: 'var Hello = function<Test extends {}>(props: Props<Test>) { return <p/> }',
      options: [{ namedComponents: 'function-expression' }],
    },
    {
      code: 'var Hello = <Test extends {}>(props: Props<Test>) => { return <p/> }',
      options: [{ namedComponents: 'arrow-function' }],
    },
    {
      code: 'function wrapper() { return function<Test>(props: Props<Test>) { return <p/> } } ',
      options: [{ unnamedComponents: 'function-expression' }],
    },
    {
      code: 'function wrapper() { return function<Test extends {}>(props: Props<Test>) { return <p/> } } ',
      options: [{ unnamedComponents: 'function-expression' }],
    },
    {
      code: 'function wrapper() { return<Test extends {}>(props: Props<Test>) => { return <p/> } } ',
      options: [{ unnamedComponents: 'arrow-function' }],
    },
    {
      code: 'var Hello = function(props): ReactNode { return <p/> }',
      options: [{ namedComponents: 'function-expression' }],
    },
    {
      code: 'var Hello = (props): ReactNode => { return <p/> }',
      options: [{ namedComponents: 'arrow-function' }],
    },
    {
      code: 'function wrapper() { return function(props): ReactNode { return <p/> } }',
      options: [{ unnamedComponents: 'function-expression' }],
    },
    {
      code: 'function wrapper() { return (props): ReactNode => { return <p/> } }',
      options: [{ unnamedComponents: 'arrow-function' }],
    },
    {
      code: 'function Hello(props): ReactNode { return <p/> }',
      options: [{ namedComponents: 'function-declaration' }],
    },
    // https://github.com/jsx-eslint/eslint-plugin-react/issues/2765
    {
      code: `
        const obj = {
          serialize: (el) => {
            return <p/>
          }
        };
      `,
      options: [{ namedComponents: 'function-declaration' }],
    },
    {
      code: `
        const obj = {
          serialize: (el) => {
            return <p/>
          }
        }
      `,
      options: [{ namedComponents: 'arrow-function' }],
    },
    {
      code: `
        const obj = {
          serialize: (el) => {
            return <p/>
          }
        }
      `,
      options: [{ namedComponents: 'function-expression' }],
    },
    {
      code: `
        const obj = {
          serialize: function (el) {
            return <p/>
          }
        }
      `,
      options: [{ namedComponents: 'function-declaration' }],
    },
    {
      code: `
        const obj = {
          serialize: function (el) {
            return <p/>
          }
        };
      `,
      options: [{ namedComponents: 'arrow-function' }],
    },
    {
      code: `
        const obj = {
          serialize: function (el) {
            return <p/>
          }
        };
      `,
      options: [{ namedComponents: 'function-expression' }],
    },
    {
      code: `
        const obj = {
          serialize(el) {
            return <p/>
          }
        };
      `,
      options: [{ namedComponents: 'function-declaration' }],
    },
    {
      code: `
        const obj = {
          serialize(el) {
            return <p/>
          }
        };
      `,
      options: [{ namedComponents: 'arrow-function' }],
    },
    {
      code: `
        const obj = {
          serialize(el) {
            return <p/>
          }
        };
      `,
      options: [{ namedComponents: 'function-expression' }],
    },
    {
      code: `
        const obj = {
          serialize(el) {
            return <p/>
          }
        };
      `,
      options: [{ unnamedComponents: 'arrow-function' }],
    },
    {
      code: `
        const obj = {
          serialize(el) {
            return <p/>
          }
        };
      `,
      options: [{ unnamedComponents: 'function-expression' }],
    },
    {
      code: `
        const obj = {
          serialize: (el) => {
            return <p/>
          }
        };
      `,
      options: [{ unnamedComponents: 'arrow-function' }],
    },
    {
      code: `
        const obj = {
          serialize: (el) => {
            return <p/>
          }
        };
      `,
      options: [{ unnamedComponents: 'function-expression' }],
    },
    {
      code: `
        const obj = {
          serialize: function (el) {
            return <p/>
          }
        };
      `,
      options: [{ unnamedComponents: 'arrow-function' }],
    },
    {
      code: `
        const obj = {
          serialize: function (el) {
            return <p/>
          }
        };
      `,
      options: [{ unnamedComponents: 'function-expression' }],
    },

    {
      code: 'function Hello(props) { return <div/> }',
      options: [
        { namedComponents: ['function-declaration', 'function-expression'] },
      ],
    },
    {
      code: 'var Hello = function(props) { return <div/> }',
      options: [
        { namedComponents: ['function-declaration', 'function-expression'] },
      ],
    },
    {
      code: 'var Foo = React.memo(function Foo() { return <p/> })',
      options: [
        { namedComponents: ['function-declaration', 'function-expression'] },
      ],
    },
    {
      code: 'function Hello(props: Test) { return <p/> }',
      options: [
        { namedComponents: ['function-declaration', 'function-expression'] },
      ],
    },
    {
      code: 'var Hello = function(props: Test) { return <p/> }',
      options: [
        { namedComponents: ['function-expression', 'function-expression'] },
      ],
    },
    {
      code: 'var Hello = (props: Test) => { return <p/> }',
      options: [{ namedComponents: ['arrow-function', 'function-expression'] }],
    },
    {
      code: `
        function wrap(Component) {
          return function(props) {
            return <div><Component {...props}/></div>;
          };
        }
      `,
      options: [
        { unnamedComponents: ['arrow-function', 'function-expression'] },
      ],
    },
    {
      code: `
        function wrap(Component) {
          return (props) => {
            return <div><Component {...props}/></div>;
          };
        }
      `,
      options: [
        { unnamedComponents: ['arrow-function', 'function-expression'] },
      ],
    },
    {
      // should not report non-jsx components
      code: `
        export default (key, subTree = {}) => {
          return (state) => {
            const dataInStore = getFromDataModel(key)(state);
            const fullPaths = dataInStore.map((item, index) => {
              return [key, index];
            });

            return {
              key,
              paths: fullPaths.map((p) => [p[1]]),
              fullPaths,
              subTree: Object.keys(subTree).length ? subTree : null,
            }
          };
        }
      `,
    },
    {
      // should not report non-jsx components
      code: `
        function mapStateToProps() {
          const internItems = makeInternArray();
          const internClassList = makeInternArray();

          return (state, props) => {
            const { store, bucket, singleCharacter } = props;

            return {
              store: null,
              destinyVersion: store.destinyVersion,
              storeId: store.id,
            }
          }
        }
      `,
    },
  ],
  invalid: [
    {
      code: `
        function Hello(props) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'arrow-function' }],
      errors: [{ messageId: 'arrow-function' }],
    },
    {
      code: `
        var Hello = function(props) {
          return <div/>;
        };
      `,
      options: [{ namedComponents: 'arrow-function' }],
      errors: [{ messageId: 'arrow-function' }],
    },
    {
      code: `
        var Hello = (props) => {
          return <div/>;
        };
      `,
      options: [{ namedComponents: 'function-declaration' }],
      errors: [{ messageId: 'function-declaration' }],
    },
    {
      code: `
        var Hello = function(props) {
          return <div/>;
        };
      `,
      options: [{ namedComponents: 'function-declaration' }],
      errors: [{ messageId: 'function-declaration' }],
    },
    {
      code: `
        var Hello = (props) => {
          return <div/>;
        };
      `,
      options: [{ namedComponents: 'function-expression' }],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        let Hello = (props) => {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'function-expression' }],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        let Hello;
        Hello = (props) => {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'function-expression' }],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        let Hello = (props) => {
          return <div/>;
        }
        Hello = function(props) {
          return <span/>;
        }
      `,
      options: [{ namedComponents: 'function-expression' }],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        function Hello(props) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'function-expression' }],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        function wrap(Component) {
          return function(props) {
            return <div><Component {...props}/></div>;
          };
        }
      `,
      errors: [{ messageId: 'arrow-function' }],
      options: [{ unnamedComponents: 'arrow-function' }],
    },
    {
      code: `
        function wrap(Component) {
          return (props) => {
            return <div><Component {...props}/></div>;
          };
        }
      `,
      errors: [{ messageId: 'function-expression' }],
      options: [{ unnamedComponents: 'function-expression' }],
    },
    {
      code: `
        var Hello = (props: Test) => {
          return <div/>;
        };
      `,
      options: [{ namedComponents: 'function-declaration' }],
      errors: [{ messageId: 'function-declaration' }],
    },
    {
      code: `
        var Hello = function(props: Test) {
          return <div/>;
        };
      `,
      options: [{ namedComponents: 'function-declaration' }],
      errors: [{ messageId: 'function-declaration' }],
    },
    {
      code: `
        function Hello(props: Test) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'arrow-function' }],
      errors: [{ messageId: 'arrow-function' }],
    },
    {
      code: `
        var Hello = function(props: Test) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'arrow-function' }],
      errors: [{ messageId: 'arrow-function' }],
    },
    {
      code: `
        function Hello(props: Test) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'function-expression' }],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        function Hello(props: Test) {
          return React.createElement('div');
        }
      `,
      options: [{ namedComponents: 'function-expression' }],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        import * as React from 'react';
        function Hello(props: Test) {
          return React.createElement('div');
        }
      `,
      options: [{ namedComponents: 'function-expression' }],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        export function Hello(props: Test) {
          return React.createElement('div');
        }
      `,
      options: [{ namedComponents: 'function-expression' }],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        function Hello(props) {
          return React.createElement('div');
        }
        export default Hello;
      `,
      options: [{ namedComponents: 'function-expression' }],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        var Hello = (props: Test) => {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'function-expression' }],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        var Hello: React.FC<Test> = (props) => {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'function-expression' }],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        var Hello: React.FC<Test> = function(props) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'arrow-function' }],
      errors: [{ messageId: 'arrow-function' }],
    },
    {
      code: `
        var Hello: React.FC<Test> = function(props) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'function-declaration' }],
      errors: [{ messageId: 'function-declaration' }],
    },
    {
      code: `
        var Hello: React.FC<Test> = (props) => {
          return <div/>;
        };
      `,
      options: [{ namedComponents: 'function-declaration' }],
      errors: [{ messageId: 'function-declaration' }],
    },
    {
      code: `
        function Hello<Test extends {}>(props: Test) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'arrow-function' }],
      errors: [{ messageId: 'arrow-function' }],
    },
    {
      code: `
        function Hello<Test>(props: Test) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'arrow-function' }],
      errors: [{ messageId: 'arrow-function' }],
    },
    {
      code: `
        function Hello<Test extends {}>(props: Test) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'function-expression' }],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        var Hello = function<Test extends {}>(props: Test) {
          return <div/>;
        };
      `,
      options: [{ namedComponents: 'function-declaration' }],
      errors: [{ messageId: 'function-declaration' }],
    },
    {
      code: `
        var Hello = <Test extends {}>(props: Test) => {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'function-expression' }],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        var Hello = function<Test extends {}>(props: Test) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'arrow-function' }],
      errors: [{ messageId: 'arrow-function' }],
    },
    {
      code: `
        var Hello = function<Test extends {}>(props: Test) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'function-declaration' }],
      errors: [{ messageId: 'function-declaration' }],
    },
    {
      code: `
        function wrap(Component) {
          return function<Test extends {}>(props) {
            return <div><Component {...props}/></div>
          }
        }
      `,
      errors: [{ messageId: 'arrow-function' }],
      options: [{ unnamedComponents: 'arrow-function' }],
    },
    {
      code: `
        function wrap(Component) {
          return function<Test>(props) {
            return <div><Component {...props}/></div>
          }
        }
      `,
      errors: [{ messageId: 'arrow-function' }],
      options: [{ unnamedComponents: 'arrow-function' }],
    },
    {
      code: `
        function wrap(Component) {
          return <Test extends {}>(props) => {
            return <div><Component {...props}/></div>
          }
        }
      `,
      errors: [{ messageId: 'function-expression' }],
      options: [{ unnamedComponents: 'function-expression' }],
    },
    {
      code: `
        function wrap(Component) {
          return function(props): ReactNode {
            return <div><Component {...props}/></div>
          }
        }
      `,
      errors: [{ messageId: 'arrow-function' }],
      options: [{ unnamedComponents: 'arrow-function' }],
    },
    {
      code: `
        function wrap(Component) {
          return (props): ReactNode => {
            return <div><Component {...props}/></div>
          }
        }
      `,
      errors: [{ messageId: 'function-expression' }],
      options: [{ unnamedComponents: 'function-expression' }],
    },
    {
      code: `
        export function Hello(props) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'arrow-function' }],
      errors: [{ messageId: 'arrow-function' }],
    },
    {
      code: `
        export var Hello = function(props) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'arrow-function' }],
      errors: [{ messageId: 'arrow-function' }],
    },
    {
      code: `
        export var Hello = (props) => {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'function-declaration' }],
      errors: [{ messageId: 'function-declaration' }],
    },
    {
      code: `
        export default function Hello(props) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: 'arrow-function' }],
      errors: [{ messageId: 'arrow-function' }],
    },
    {
      code: `
        module.exports = function Hello(props) {
          return <div/>;
        }
      `,
      options: [{ unnamedComponents: 'arrow-function' }],
      errors: [{ messageId: 'arrow-function' }],
    },
    {
      code: `
        function Hello(props) {
          return <div/>;
        }
      `,
      options: [{ namedComponents: ['arrow-function', 'function-expression'] }],
      errors: [{ messageId: 'arrow-function' }],
    },
    {
      code: `
        var Hello = (props) => {
          return <div/>;
        };
      `,
      options: [
        { namedComponents: ['function-declaration', 'function-expression'] },
      ],
      errors: [{ messageId: 'function-declaration' }],
    },
    {
      code: `
        var Hello = (props) => {
          return <div/>;
        };
      `,
      options: [
        { namedComponents: ['function-expression', 'function-declaration'] },
      ],
      errors: [{ messageId: 'function-expression' }],
    },
    {
      code: `
        const genX = (symbol) => \`the symbol is \${symbol}\`;

        const IndexPage = () => {
          return (
            <div>
              Hello World.{genX('$')}
            </div>
          )
        }

        export default IndexPage;
      `,
      output: `
        const genX = (symbol) => \`the symbol is \${symbol}\`;

        function IndexPage() {
          return (
            <div>
              Hello World.{genX('$')}
            </div>
          )
        }

        export default IndexPage;
      `,
      options: [{ namedComponents: ['function-declaration'] }],
      errors: [{ messageId: 'function-declaration' }],
    },
  ],
});
