import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectUnsafeMethodError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unsafeMethod',
    message:
      'UNSAFE_componentWillMount is unsafe for use in async rendering. Update the component to use componentDidMount instead. See https://reactjs.org/blog/2018/03/27/update-on-async-rendering.html.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnsafeMethodError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unsafeMethod',
    message:
      'componentWillMount is unsafe for use in async rendering. Update the component to use componentDidMount instead. See https://reactjs.org/blog/2018/03/27/update-on-async-rendering.html.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnsafeMethodError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unsafeMethod',
    message:
      'UNSAFE_componentWillReceiveProps is unsafe for use in async rendering. Update the component to use getDerivedStateFromProps instead. See https://reactjs.org/blog/2018/03/27/update-on-async-rendering.html.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnsafeMethodError4(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unsafeMethod',
    message:
      'UNSAFE_componentWillUpdate is unsafe for use in async rendering. Update the component to use componentDidUpdate instead. See https://reactjs.org/blog/2018/03/27/update-on-async-rendering.html.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-unsafe', {
  valid: [
    // ---- Upstream valid: React.Component with safe lifecycle methods ----
    {
      code: `
        class Foo extends React.Component {
          componentDidUpdate() {}
          render() {}
        }
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Upstream valid: createReactClass with safe methods ----
    {
      code: `
        const Foo = createReactClass({
          componentDidUpdate: function() {},
          render: function() {}
        });
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Upstream valid: non-React class with deprecated names ----
    {
      code: `
        class Foo extends Bar {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Upstream valid: non-React class with UNSAFE_ names ----
    {
      code: `
        class Foo extends Bar {
          UNSAFE_componentWillMount() {}
          UNSAFE_componentWillReceiveProps() {}
          UNSAFE_componentWillUpdate() {}
        }
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Upstream valid: bar(...) is not createReactClass ----
    {
      code: `
        const Foo = bar({
          componentWillMount: function() {},
          componentWillReceiveProps: function() {},
          componentWillUpdate: function() {},
        });
      `,
      settings: { react: { version: '16.4.0' } },
    },
    {
      code: `
        const Foo = bar({
          UNSAFE_componentWillMount: function() {},
          UNSAFE_componentWillReceiveProps: function() {},
          UNSAFE_componentWillUpdate: function() {},
        });
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Upstream valid: deprecated names without checkAliases ----
    {
      code: `
        class Foo extends React.Component {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Upstream valid: React 16.2.0 — rule disabled ----
    {
      code: `
        class Foo extends React.Component {
          UNSAFE_componentWillMount() {}
          UNSAFE_componentWillReceiveProps() {}
          UNSAFE_componentWillUpdate() {}
        }
      `,
      settings: { react: { version: '16.2.0' } },
    },
    {
      code: `
        const Foo = createReactClass({
          componentWillMount: function() {},
          componentWillReceiveProps: function() {},
          componentWillUpdate: function() {},
        });
      `,
      settings: { react: { version: '16.4.0' } },
    },
    {
      code: `
        const Foo = createReactClass({
          UNSAFE_componentWillMount: function() {},
          UNSAFE_componentWillReceiveProps: function() {},
          UNSAFE_componentWillUpdate: function() {},
        });
      `,
      settings: { react: { version: '16.2.0' } },
    },
    // ---- Edge: empty class body — graceful degradation ----
    {
      code: `class Foo extends React.Component {}`,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Edge: createReactClass with spread — must not crash ----
    {
      code: `
        const Foo = createReactClass({
          ...mixins,
          render: function() {}
        });
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Edge: computed key matching unsafe name — never flags ----
    {
      code: `
        const k = 'UNSAFE_componentWillMount';
        class Foo extends React.Component {
          [k]() {}
        }
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Edge: StringLiteral key — upstream's getPropertyName returns
    // nameNode.name, undefined for Literal nodes, so this never flags. ----
    {
      code: `
        const Foo = createReactClass({
          'UNSAFE_componentWillMount': function() {},
          render: function() {},
        });
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Edge: plain class without extends — not a React component. ----
    {
      code: `
        class Foo {
          UNSAFE_componentWillMount() {}
        }
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Edge: custom pragma — React.Component is not detected. ----
    {
      code: `
        class Foo extends React.Component {
          UNSAFE_componentWillMount() {}
        }
      `,
      settings: { react: { pragma: 'Preact', version: '16.4.0' } },
    },
    // ---- Edge: TS NonNullExpression in extends — doesn't match ----
    {
      code: `
        class Foo extends React.Component! {
          UNSAFE_componentWillMount() {}
        }
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Edge: AsExpression in extends — doesn't match ----
    {
      code: `
        class Foo extends (React.Component as any) {
          UNSAFE_componentWillMount() {}
        }
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Edge: HOC return value in extends — doesn't match literal Component ----
    {
      code: `
        class Foo extends withRouter(Base) {
          UNSAFE_componentWillMount() {}
        }
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Edge: ComputedPropertyName with string literal — does NOT match ----
    {
      code: `
        const Foo = createReactClass({
          ['UNSAFE_componentWillMount']: function() {},
          render: function() {}
        });
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Edge: Object literal as JSX prop value — must not match ----
    {
      code: `
        const Foo = () => <div data-cfg={{ UNSAFE_componentWillMount: 1 }} />;
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Edge: TS abstract class without unsafe methods ----
    {
      code: `
        abstract class Foo extends React.Component {
          abstract render(): any;
        }
      `,
      settings: { react: { version: '16.4.0' } },
    },
    // ---- Edge: Class implements but no extends — not React component ----
    {
      code: `
        class Foo implements IFoo {
          UNSAFE_componentWillMount() {}
        }
      `,
      settings: { react: { version: '16.4.0' } },
    },
  ],
  invalid: [
    // Intentional correction: a literal superclass key identifies a React component.
    {
      code: `
        class Foo extends React['Component'] {
          UNSAFE_componentWillMount() {}
        }
      `,
      settings: { react: { version: '16.4.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 41])],
    },
    // ---- Upstream invalid #1: React.Component + checkAliases: true ----
    {
      code: `
        class Foo extends React.Component {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `,
      options: [{ checkAliases: true }],
      settings: { react: { version: '16.4.0' } },
      errors: [
        expectUnsafeMethodError2([3, 11, 3, 34]),
        {
          messageId: 'unsafeMethod',
          message:
            'componentWillReceiveProps is unsafe for use in async rendering. Update the component to use getDerivedStateFromProps instead. See https://reactjs.org/blog/2018/03/27/update-on-async-rendering.html.',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 41,
        },
        {
          messageId: 'unsafeMethod',
          message:
            'componentWillUpdate is unsafe for use in async rendering. Update the component to use componentDidUpdate instead. See https://reactjs.org/blog/2018/03/27/update-on-async-rendering.html.',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 35,
        },
      ],
    },
    // ---- Upstream invalid #2: React.Component + UNSAFE_ at 16.3.0 ----
    {
      code: `
        class Foo extends React.Component {
          UNSAFE_componentWillMount() {}
          UNSAFE_componentWillReceiveProps() {}
          UNSAFE_componentWillUpdate() {}
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [
        expectUnsafeMethodError1([3, 11, 3, 41]),
        expectUnsafeMethodError3([4, 11, 4, 48]),
        expectUnsafeMethodError4([5, 11, 5, 42]),
      ],
    },
    // ---- Upstream invalid #3: createReactClass + checkAliases: true ----
    {
      code: `
        const Foo = createReactClass({
          componentWillMount: function() {},
          componentWillReceiveProps: function() {},
          componentWillUpdate: function() {},
        });
      `,
      options: [{ checkAliases: true }],
      settings: { react: { version: '16.3.0' } },
      errors: [
        expectUnsafeMethodError2([3, 11, 3, 44]),
        {
          messageId: 'unsafeMethod',
          message:
            'componentWillReceiveProps is unsafe for use in async rendering. Update the component to use getDerivedStateFromProps instead. See https://reactjs.org/blog/2018/03/27/update-on-async-rendering.html.',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 51,
        },
        {
          messageId: 'unsafeMethod',
          message:
            'componentWillUpdate is unsafe for use in async rendering. Update the component to use componentDidUpdate instead. See https://reactjs.org/blog/2018/03/27/update-on-async-rendering.html.',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 45,
        },
      ],
    },
    // ---- Upstream invalid #4: createReactClass + UNSAFE_ at 16.3.0 ----
    {
      code: `
        const Foo = createReactClass({
          UNSAFE_componentWillMount: function() {},
          UNSAFE_componentWillReceiveProps: function() {},
          UNSAFE_componentWillUpdate: function() {},
        });
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [
        expectUnsafeMethodError1([3, 11, 3, 51]),
        expectUnsafeMethodError3([4, 11, 4, 58]),
        expectUnsafeMethodError4([5, 11, 5, 52]),
      ],
    },
    // ---- Edge: ClassExpression — same listener as ClassDeclaration ----
    {
      code: `
        const Foo = class extends React.Component {
          UNSAFE_componentWillMount() {}
        };
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 41])],
    },
    // ---- Edge: bare `extends Component` (no pragma qualifier) ----
    {
      code: `
        class Foo extends Component {
          UNSAFE_componentWillMount() {}
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 41])],
    },
    // ---- Edge: PureComponent extends — same flag set as Component ----
    {
      code: `
        class Foo extends React.PureComponent {
          UNSAFE_componentWillMount() {}
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 41])],
    },
    // ---- Edge: default version = "latest" → rule active ----
    {
      code: `
        class Foo extends React.Component {
          UNSAFE_componentWillMount() {}
        }
      `,
      errors: [expectUnsafeMethodError1([3, 11, 3, 41])],
    },
    // ---- Edge: nested classes — only inner is React component, only Inner reports ----
    {
      code: `
        class Outer {
          run() {
            return class Inner extends React.Component {
              UNSAFE_componentWillMount() {}
            };
          }
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([5, 15, 5, 45])],
    },
    // ---- Edge: class field with arrow function ----
    {
      code: `
        class Foo extends React.Component {
          UNSAFE_componentWillMount = () => {};
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 48])],
    },
    // ---- Edge: HOC-wrapped class expression ----
    {
      code: `
        const Foo = connect(state => state)(class extends React.Component {
          UNSAFE_componentWillMount() {}
        });
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 41])],
    },
    // ---- Edge: TypeScript generic class ----
    {
      code: `
        class Foo<P> extends React.Component<P> {
          UNSAFE_componentWillMount() {}
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 41])],
    },
    // ---- Edge: source-order vs alphabetical-order ----
    {
      code: `
        class Foo extends React.Component {
          UNSAFE_componentWillUpdate() {}
          UNSAFE_componentWillMount() {}
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [
        expectUnsafeMethodError4([3, 11, 3, 42]),
        expectUnsafeMethodError1([4, 11, 4, 41]),
      ],
    },
    // ---- Edge: React.memo wrapping ----
    {
      code: `
        const Foo = React.memo(class extends React.Component {
          UNSAFE_componentWillMount() {}
        });
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 41])],
    },
    // ---- Edge: Default export anonymous class ----
    {
      code: `
        export default class extends React.Component {
          UNSAFE_componentWillMount() {}
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 41])],
    },
    // ---- Edge: async unsafe method ----
    {
      code: `
        class Foo extends React.Component {
          async UNSAFE_componentWillMount() {}
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 47])],
    },
    // ---- Edge: setter named like unsafe lifecycle ----
    {
      code: `
        class Foo extends React.Component {
          set UNSAFE_componentWillMount(v) {}
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 46])],
    },
    // ---- Edge: TS abstract method ----
    {
      code: `
        abstract class Foo extends React.Component {
          abstract UNSAFE_componentWillMount(): void;
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 54])],
    },
    // ---- Edge: TS declare class ----
    {
      code: `
        declare class Foo extends React.Component {
          UNSAFE_componentWillMount(): void;
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 45])],
    },
    // ---- Edge: createReactClass with method shorthand ----
    {
      code: `
        const Foo = createReactClass({
          UNSAFE_componentWillMount() {},
          render() { return null; }
        });
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 41])],
    },
    // ---- Edge: Mixed UNSAFE_ + unprefixed names with checkAliases:true ----
    {
      code: `
        class Foo extends React.Component {
          componentWillMount() {}
          UNSAFE_componentWillReceiveProps() {}
        }
      `,
      options: [{ checkAliases: true }],
      settings: { react: { version: '16.3.0' } },
      errors: [
        expectUnsafeMethodError2([3, 11, 3, 34]),
        expectUnsafeMethodError3([4, 11, 4, 48]),
      ],
    },
    // ---- Edge: parens around extends + parens around pragma ----
    {
      code: `
        class Foo extends ((React)).Component {
          UNSAFE_componentWillMount() {}
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 41])],
    },
    // ---- Edge: ES5 component nested in ES6 component (only ES5 has unsafe) ----
    {
      code: `
        class Outer extends React.Component {
          render() {
            return createReactClass({
              UNSAFE_componentWillMount: function() {},
              render: function() { return null; }
            });
          }
        }
      `,
      settings: { react: { version: '16.3.0' } },
      errors: [expectUnsafeMethodError1([5, 15, 5, 55])],
    },
    // ---- Edge: ObjectExpression at non-first arg of createReactClass ----
    // upstream's `componentUtil.isES5Component` accepts any arg position;
    // verified empirically: `createReactClass(other, {...})` IS flagged.
    {
      code: `
        const Foo = createReactClass(other, {
          UNSAFE_componentWillMount: function() {},
        });
      `,
      settings: { react: { version: '16.4.0' } },
      errors: [expectUnsafeMethodError1([3, 11, 3, 51])],
    },
  ],
});
