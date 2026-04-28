import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unused-class-component-methods', {} as never, {
  valid: [
    // ---- Handler referenced in JSX event ----
    {
      code: `
        class Foo extends React.Component {
          handleClick() {}
          render() {
            return <button onClick={this.handleClick}>Text</button>;
          }
        }
      `,
    },
    // ---- createReactClass handler referenced in JSX event ----
    {
      code: `
        var Foo = createReactClass({
          handleClick() {},
          render() {
            return <button onClick={this.handleClick}>Text</button>;
          },
        })
      `,
    },
    // ---- Method called from lifecycle method ----
    {
      code: `
        class Foo extends React.Component {
          action() {}
          componentDidMount() {
            this.action();
          }
          render() {
            return null;
          }
        }
      `,
    },
    // ---- Class-field arrow handler ----
    {
      code: `
        class Foo extends React.Component {
          handleClick = () => {}
          render() {
            return <button onClick={this.handleClick}>Button</button>;
          }
        }
      `,
    },
    // ---- this.X = ... in constructor (defined + used) ----
    {
      code: `
        class ClassAssignPropertyInMethodTest extends React.Component {
          constructor() {
            this.foo = 3;
          }
          render() {
            return <SomeComponent foo={this.foo} />;
          }
        }
      `,
    },
    // ---- Computed string-literal key referenced via element access ----
    {
      code: `
        class Foo extends React.Component {
          ['foo'] = a;
          render() {
            return <SomeComponent foo={this['foo']} />;
          }
        }
      `,
    },
    // ---- state class field is a lifecycle name (never reported) ----
    {
      code: `
        class ClassComputedTemplatePropertyTest extends React.Component {
          state = {}
          render() {
            return <div />;
          }
        }
      `,
    },
    // ---- Destructuring from `this` ----
    {
      code: `
        class ClassUseDestructuringTest extends React.Component {
          foo() {}
          render() {
            const { foo } = this;
            return <SomeComponent />;
          }
        }
      `,
    },
    // ---- Canonical lifecycle methods are always ignored ----
    {
      code: `
        class ClassWithLifecyleTest extends React.Component {
          constructor(props) {
            super(props);
          }
          static getDerivedStateFromProps() {}
          componentDidMount() {}
          shouldComponentUpdate() {}
          componentDidUpdate() {}
          componentWillUnmount() {}
          render() {
            return <SomeComponent />;
          }
        }
      `,
    },
    // ---- createReactClass ES5 lifecycle names ----
    {
      code: `
        var ClassWithLifecyleTest = createReactClass({
          mixins: [],
          getDefaultProps() { return {} },
          getInitialState: function() { return {x: 0}; },
          render() { return <SomeComponent />; },
        })
      `,
    },
  ],
  invalid: [
    // ---- Non-standard lifecycle method on class ----
    {
      code: `
        class Foo extends React.Component {
          getDerivedStateFromProps() {}
          render() {
            return <div>Example</div>;
          }
        }
      `,
      errors: [{ messageId: 'unusedWithClass' }],
    },
    // ---- Unused class field ----
    {
      code: `
        class Foo extends React.Component {
          property = {}
          render() {
            return <div>Example</div>;
          }
        }
      `,
      errors: [{ messageId: 'unusedWithClass' }],
    },
    // ---- Unused method on createReactClass (no className) ----
    {
      code: `
        var Foo = createReactClass({
          handleClick() {},
          render() {
            return null;
          },
        })
      `,
      errors: [{ messageId: 'unused' }],
    },
    // ---- Multiple unused methods reported in source order ----
    {
      code: `
        class Foo extends React.Component {
          handleScroll() {}
          handleClick() {}
          render() {
            return null;
          }
        }
      `,
      errors: [
        { messageId: 'unusedWithClass' },
        { messageId: 'unusedWithClass' },
      ],
    },
    // ---- Unused `this.foo = …` in constructor ----
    {
      code: `
        class Foo extends React.Component {
          constructor() {
            this.foo = 3;
          }
          render() {
            return <SomeComponent />;
          }
        }
      `,
      errors: [{ messageId: 'unusedWithClass' }],
    },
    // ---- TS private method ----
    {
      code: `
        class Foo extends React.Component {
          private foo() {}
          render() {
            return <SomeComponent />;
          }
        }
      `,
      errors: [{ messageId: 'unusedWithClass' }],
    },
  ],
});
