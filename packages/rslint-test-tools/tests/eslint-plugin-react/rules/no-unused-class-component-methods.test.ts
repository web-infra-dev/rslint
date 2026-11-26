import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectUnusedWithClassError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedWithClass',
    message: 'Unused method or property "handleClick" of class "Foo"',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedWithClassError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedWithClass',
    message: 'Unused method or property "foo" of class "Foo"',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnusedWithClassError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unusedWithClass',
    message: 'Unused method or property "action" of class "Foo"',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-unused-class-component-methods', {
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

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        class SmockTestForTypeOfNullError extends React.Component {
          handleClick() {}
          foo;
          render() {
            let a;
            return <button disabled onClick={this.handleClick} foo={this.foo}>Text</button>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        var Foo = createReactClass({
          action() {},
          componentDidMount() {
            this.action();
          },
          render() {
            return null;
          },
        })
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        class Foo extends React.Component {
          action() {}
          componentDidMount() {
            const action = this.action;
            action();
          }
          render() {
            return null;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        class Foo extends React.Component {
          getValue() {}
          componentDidMount() {
            const action = this.getValue();
          }
          render() {
            return null;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        class Foo extends React.Component {
          renderContent() {}
          render() {
            return <div>{this.renderContent()}</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        class Foo extends React.Component {
          renderContent() {}
          render() {
            return (
              <div>
                <div>{this.renderContent()}</div>;
              </div>
            );
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        class Foo extends React.Component {
          property = {}
          render() {
            return <div property={this.property}>Example</div>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        class Foo extends React.Component {
          action = () => {}
          anotherAction = () => {
            this.action();
          }
          render() {
            return <button onClick={this.anotherAction}>Example</button>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        class Foo extends React.Component {
          action = () => {}
          anotherAction = () => this.action()
          render() {
            return <button onClick={this.anotherAction}>Example</button>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        class Foo extends React.Component {
          getValue = () => {}
          value = this.getValue()
          render() {
            return this.value;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        class Foo {
          action = () => {}
          anotherAction = () => this.action()
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        class Foo extends React.Component {
          action = async () => {}
          render() {
            return <button onClick={this.action}>Click</button>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        class Foo extends React.Component {
          async action() {
            console.log('error');
          }
          render() {
            return <button onClick={() => this.action()}>Click</button>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        class Foo extends React.Component {
          * action() {
            console.log('error');
          }
          render() {
            return <button onClick={() => this.action()}>Click</button>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        class Foo extends React.Component {
          async * action() {
            console.log('error');
          }
          render() {
            return <button onClick={() => this.action()}>Click</button>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        class Foo extends React.Component {
          action = function() {
            console.log('error');
          }
          render() {
            return <button onClick={() => this.action()}>Click</button>;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        class ClassAssignPropertyInMethodTest extends React.Component {
          constructor() {
            this.foo = 3;;
          }
          render() {
            return <SomeComponent foo={this.foo} />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        class ClassPropertyTest extends React.Component {
          foo;
          render() {
            return <SomeComponent foo={this.foo} />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        class ClassPropertyTest extends React.Component {
          foo = a;
          render() {
            return <SomeComponent foo={this.foo} />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        class Foo extends React.Component {
          ['foo'];
          render() {
            return <SomeComponent foo={this['foo']} />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        class ClassComputedTemplatePropertyTest extends React.Component {
          [\`foo\`] = a;
          render() {
            return <SomeComponent foo={this[\`foo\`]} />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        class ClassLiteralComputedMemberTest extends React.Component {
          ['foo']() {}
          render() {
            return <SomeComponent foo={this.foo} />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        class ClassComputedTemplateMemberTest extends React.Component {
          [\`foo\`]() {}
          render() {
            return <SomeComponent foo={this.foo} />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        class ClassUseAssignTest extends React.Component {
          foo() {}
          render() {
            this.foo;
            return <SomeComponent />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        class ClassUseAssignTest extends React.Component {
          foo() {}
          render() {
            const { foo } = this;
            return <SomeComponent />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
        class ClassUseDestructuringTest extends React.Component {
          ['foo']() {}
          render() {
            const { 'foo': bar } = this;
            return <SomeComponent />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        class ClassComputedMemberTest extends React.Component {
          [foo]() {}
          render() {
            return <SomeComponent />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
        class ClassWithLifecyleTest extends React.Component {
          constructor(props) {
            super(props);
          }
          static getDerivedStateFromProps() {}
          componentWillMount() {}
          UNSAFE_componentWillMount() {}
          componentDidMount() {}
          componentWillReceiveProps() {}
          UNSAFE_componentWillReceiveProps() {}
          shouldComponentUpdate() {}
          componentWillUpdate() {}
          UNSAFE_componentWillUpdate() {}
          static getSnapshotBeforeUpdate() {}
          componentDidUpdate() {}
          componentDidCatch() {}
          componentWillUnmount() {}
          getChildContext() {}
          render() {
            return <SomeComponent />;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
        var ClassWithLifecyleTest = createReactClass({
          mixins: [],
          constructor(props) {
          },
          getDefaultProps() {
            return {}
          },
          getInitialState: function() {
            return {x: 0};
          },
          componentWillMount() {},
          UNSAFE_componentWillMount() {},
          componentDidMount() {},
          componentWillReceiveProps() {},
          UNSAFE_componentWillReceiveProps() {},
          shouldComponentUpdate() {},
          componentWillUpdate() {},
          UNSAFE_componentWillUpdate() {},
          componentDidUpdate() {},
          componentDidCatch() {},
          componentWillUnmount() {},
          getChildContext() {},
          render() {
            return <SomeComponent />;
          },
        })
      `,
      languageOptions: upstreamModuleOptions,
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
      errors: [
        {
          messageId: 'unusedWithClass',
          message:
            'Unused method or property "getDerivedStateFromProps" of class "Foo"',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 35,
        },
      ],
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
      errors: [
        {
          messageId: 'unusedWithClass',
          message: 'Unused method or property "property" of class "Foo"',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 19,
        },
      ],
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
      errors: [
        {
          messageId: 'unused',
          message: 'Unused method or property "handleClick"',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 22,
        },
      ],
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
        {
          messageId: 'unusedWithClass',
          message: 'Unused method or property "handleScroll" of class "Foo"',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 23,
        },
        expectUnusedWithClassError1([4, 11, 4, 22]),
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
      errors: [expectUnusedWithClassError2([4, 18, 4, 21])],
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
      errors: [expectUnusedWithClassError2([3, 19, 3, 22])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        class Foo extends React.Component {
          handleClick() {}
          render() {
            return null;
          }
        }
      `,
      errors: [expectUnusedWithClassError1([3, 11, 3, 22])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        var Foo = createReactClass({
          a: 3,
          render() {
            return null;
          },
        })
      `,
      errors: [
        {
          messageId: 'unused',
          message: 'Unused method or property "a"',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 12,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        class Foo extends React.Component {
          handleClick = () => {}
          render() {
            return null;
          }
        }
      `,
      errors: [expectUnusedWithClassError1([3, 11, 3, 22])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        class Foo extends React.Component {
          action = async () => {}
          render() {
            return null;
          }
        }
      `,
      errors: [expectUnusedWithClassError3([3, 11, 3, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        class Foo extends React.Component {
          async action() {
            console.log('error');
          }
          render() {
            return null;
          }
        }
      `,
      errors: [expectUnusedWithClassError3([3, 17, 3, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        class Foo extends React.Component {
          * action() {
            console.log('error');
          }
          render() {
            return null;
          }
        }
      `,
      errors: [expectUnusedWithClassError3([3, 13, 3, 19])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        class Foo extends React.Component {
          async * action() {
            console.log('error');
          }
          render() {
            return null;
          }
        }
      `,
      errors: [expectUnusedWithClassError3([3, 19, 3, 25])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        class Foo extends React.Component {
          getInitialState() {}
          render() {
            return null;
          }
        }
      `,
      errors: [
        {
          messageId: 'unusedWithClass',
          message: 'Unused method or property "getInitialState" of class "Foo"',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 26,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        class Foo extends React.Component {
          action = function() {
            console.log('error');
          }
          render() {
            return null;
          }
        }
      `,
      errors: [expectUnusedWithClassError3([3, 11, 3, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
         class ClassAssignPropertyInMethodTest extends React.Component {
           constructor() {
             this.foo = 3;
           }
           render() {
             return <SomeComponent />;
           }
         }
       `,
      errors: [
        {
          messageId: 'unusedWithClass',
          message:
            'Unused method or property "foo" of class "ClassAssignPropertyInMethodTest"',
          line: 4,
          column: 19,
          endLine: 4,
          endColumn: 22,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
         class Foo extends React.Component {
           foo;
           render() {
             return <SomeComponent />;
           }
         }
       `,
      errors: [expectUnusedWithClassError2([3, 12, 3, 15])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
         class Foo extends React.Component {
           foo = a;
           render() {
             return <SomeComponent />;
           }
         }
       `,
      errors: [expectUnusedWithClassError2([3, 12, 3, 15])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
         class Foo extends React.Component {
           ['foo'];
           render() {
             return <SomeComponent />;
           }
         }
       `,
      errors: [expectUnusedWithClassError2([3, 13, 3, 18])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
         class Foo extends React.Component {
           ['foo'] = a;
           render() {
             return <SomeComponent />;
           }
         }
       `,
      errors: [expectUnusedWithClassError2([3, 13, 3, 18])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
         class Foo extends React.Component {
           foo = a;
           render() {
             return <SomeComponent foo={this[foo]} />;
           }
         }
       `,
      errors: [expectUnusedWithClassError2([3, 12, 3, 15])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
         class Foo extends React.Component {
           private foo;
           render() {
             return <SomeComponent />;
           }
         }
       `,
      errors: [expectUnusedWithClassError2([3, 20, 3, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
         class Foo extends React.Component {
           private foo() {}
           render() {
             return <SomeComponent />;
           }
         }
       `,
      errors: [expectUnusedWithClassError2([3, 20, 3, 23])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
         class Foo extends React.Component {
           private foo = 3;
           render() {
             return <SomeComponent />;
           }
         }
       `,
      errors: [expectUnusedWithClassError2([3, 20, 3, 23])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
