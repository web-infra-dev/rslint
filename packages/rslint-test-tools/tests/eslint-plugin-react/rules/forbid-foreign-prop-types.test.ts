import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectForbiddenPropTypeError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'forbiddenPropType',
    message:
      'Using propTypes from another component is not safe because they may be removed in production builds',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('forbid-foreign-prop-types', {
  valid: [
    // ---- Upstream valid cases ----
    {
      code: `import { propTypes } from "SomeComponent";`,
    },
    {
      code: `import { propTypes as someComponentPropTypes } from "SomeComponent";`,
    },
    {
      code: `const foo = propTypes`,
    },
    {
      code: `foo(propTypes)`,
    },
    {
      code: `foo + propTypes`,
    },
    {
      code: `const foo = [propTypes]`,
    },
    {
      code: `const foo = { propTypes }`,
    },
    {
      code: `Foo.propTypes = propTypes`,
    },
    {
      code: `Foo["propTypes"] = propTypes`,
    },
    {
      code: `const propTypes = "bar"; Foo[propTypes];`,
    },
    {
      code: `
        const Message = (props) => (<div>{props.message}</div>);
        Message.propTypes = {
          message: PropTypes.string
        };
        const Hello = (props) => (<Message>Hello {props.name}</Message>);
        Hello.propTypes = {
          name: Message.propTypes.message
        };
      `,
      options: [{ allowInPropTypes: true }],
    },
    {
      code: `
        class MyComponent extends React.Component {
          static propTypes = {
            baz: Qux.propTypes.baz
          };
        }
      `,
      options: [{ allowInPropTypes: true }],
    },
    // ---- Universal edge shapes locked-in ----
    {
      code: 'Foo[`propTypes`]',
    },
    {
      code: `class C { #propTypes = 1; m() { return this.#propTypes } }`,
    },
    {
      code: `Foo[0]`,
    },
    {
      code: `var { "propTypes": x } = SomeComponent;`,
    },
    {
      code: `var { ...propTypes } = SomeComponent;`,
    },
    {
      code: `Foo.propTypes += {}`,
    },
    {
      code: `(Foo.propTypes) = {};`,
    },
    {
      code: `
        Bar.propTypes = {
          x: () => Foo.propTypes
        };
      `,
      options: [{ allowInPropTypes: true }],
    },
    {
      code: `
        class C {
          static propTypes = {
            x: () => Foo.propTypes
          };
        }
      `,
      options: [{ allowInPropTypes: true }],
    },
    {
      code: `var { ["propTypes"]: x } = SomeComponent;`,
    },
    {
      code: `var { foo: propTypes } = SomeComponent;`,
    },
    // ---- TS type-level positions (do NOT fire) ----
    {
      code: `type X = typeof Foo.propTypes;`,
    },
    {
      code: `type X = { p: typeof Foo.propTypes };`,
    },
    // ---- allowInPropTypes: true + deeply nested foreign access ----
    {
      code: `
        function withFoo(C) {
          C.propTypes = { ...Inner.propTypes };
          return C;
        }
      `,
      options: [{ allowInPropTypes: true }],
    },
    {
      code: `
        Foo.propTypes = {
          x: function() {
            return { y: function() { return Bar.propTypes; } };
          }
        };
      `,
      options: [{ allowInPropTypes: true }],
    },
    {
      code: `Foo.propTypes = (cond ? Bar.propTypes : Baz.propTypes) || {};`,
      options: [{ allowInPropTypes: true }],
    },
    {
      code: `
        class C {
          static propTypes = { ...Other.propTypes, x: cond ? Inner.propTypes.y : null };
        }
      `,
      options: [{ allowInPropTypes: true }],
    },
    // ---- Regular OLE must NOT fire ----
    {
      code: `const x = { propTypes: 1, foo: 2 };`,
    },
    {
      code: `const x = { y: { propTypes: 1 } };`,
    },
    // ---- Bracket access with non-string literal arguments (NOT match) ----
    {
      code: `Foo[1_000];`,
    },
    {
      code: `Foo[1n];`,
    },
    // ---- Computed template-literal destructuring key (NOT match) ----
    {
      code: 'var { [`propTypes`]: x } = SomeComponent;',
    },
    // ---- Function rest param destructuring (NOT match) ----
    {
      code: `function f({ ...propTypes }) {}`,
    },
    // ---- Case sensitivity ----
    {
      code: `Foo.PropTypes;`,
    },
    {
      code: `Foo.proptypes;`,
    },
    {
      code: `Foo["PropTypes"];`,
    },
    {
      code: `var { PropTypes } = SomeComponent;`,
    },
    // ---- Other React static fields (NOT match) ----
    {
      code: `Foo.defaultProps;`,
    },
    {
      code: `Foo.contextTypes;`,
    },
    // ---- JSX dotted tag (NOT match) ----
    // tsgo represents `<Foo.Bar />` with PropertyAccessExpression while
    // ESTree uses a distinct JSXMemberExpression. The `isJsxTagName`
    // skip aligns the byte-for-byte output.
    {
      code: `var X = <Foo.propTypes />;`,
    },
    {
      code: `var X = <Foo.propTypes></Foo.propTypes>;`,
    },
    {
      code: `var X = <Foo.propTypes.Bar />;`,
    },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `
        var Foo = createReactClass({
          propTypes: Bar.propTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
      errors: [expectForbiddenPropTypeError([3, 26, 3, 35])],
    },
    {
      code: `
        var Foo = createReactClass({
          propTypes: Bar["propTypes"],
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
      errors: [expectForbiddenPropTypeError([3, 26, 3, 37])],
    },
    {
      code: `
        var { propTypes } = SomeComponent
        var Foo = createReactClass({
          propTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
      errors: [expectForbiddenPropTypeError([2, 15, 2, 24])],
    },
    {
      code: `
        var { propTypes: things, ...foo } = SomeComponent
        var Foo = createReactClass({
          propTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
      errors: [expectForbiddenPropTypeError([2, 15, 2, 32])],
    },
    {
      code: `
        class MyComponent extends React.Component {
          static fooBar = {
            baz: Qux.propTypes.baz
          };
        }
      `,
      errors: [expectForbiddenPropTypeError([4, 22, 4, 31])],
    },
    {
      code: `
        var { propTypes: typesOfProps } = SomeComponent
        var Foo = createReactClass({
          propTypes: typesOfProps,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
      errors: [expectForbiddenPropTypeError([2, 15, 2, 38])],
    },
    {
      code: `
        const Message = (props) => (<div>{props.message}</div>);
        Message.propTypes = {
          message: PropTypes.string
        };
        const Hello = (props) => (<Message>Hello {props.name}</Message>);
        Hello.propTypes = {
          name: Message.propTypes.message
        };
      `,
      options: [{ allowInPropTypes: false }],
      errors: [expectForbiddenPropTypeError([8, 25, 8, 34])],
    },
    {
      code: `
        class MyComponent extends React.Component {
          static propTypes = {
            baz: Qux.propTypes.baz
          };
        }
      `,
      options: [{ allowInPropTypes: false }],
      errors: [expectForbiddenPropTypeError([4, 22, 4, 31])],
    },
    // ---- Lock-in tests ----
    {
      code: `Foo?.propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 6, 1, 15])],
    },
    {
      code: `(Foo).propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 7, 1, 16])],
    },
    {
      code: `declare const Foo: any; Foo!.propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 30, 1, 39])],
    },
    {
      code: `Foo.propTypes.isRequired;`,
      errors: [expectForbiddenPropTypeError([1, 5, 1, 14])],
    },
    {
      code: `
        class MyComponent extends React.Component {
          static fooBar = {
            baz: Qux.propTypes.baz
          };
        }
      `,
      options: [{ allowInPropTypes: true }],
      errors: [expectForbiddenPropTypeError([4, 22, 4, 31])],
    },
    {
      code: `Foo.notPropTypes = { x: Bar.propTypes.x };`,
      options: [{ allowInPropTypes: true }],
      errors: [expectForbiddenPropTypeError([1, 29, 1, 38])],
    },
    {
      code: `Foo['propTypes'] = { x: Bar.propTypes.x };`,
      options: [{ allowInPropTypes: true }],
      errors: [expectForbiddenPropTypeError([1, 29, 1, 38])],
    },
    {
      code: `({ propTypes } = SomeComponent);`,
      errors: [expectForbiddenPropTypeError([1, 4, 1, 13])],
    },
    {
      code: `({ propTypes: alias } = SomeComponent);`,
      errors: [expectForbiddenPropTypeError([1, 4, 1, 20])],
    },
    {
      code: `({ a: { propTypes } } = SomeComponent);`,
      errors: [expectForbiddenPropTypeError([1, 9, 1, 18])],
    },
    {
      code: `([{ propTypes }] = [SomeComponent]);`,
      errors: [expectForbiddenPropTypeError([1, 5, 1, 14])],
    },
    {
      code: `
        class C {
          static block = (() => {
            return Bar.propTypes;
          })();
        }
      `,
      options: [{ allowInPropTypes: true }],
      errors: [expectForbiddenPropTypeError([4, 24, 4, 33])],
    },
    {
      code: `var a = Foo.propTypes; var b = Bar["propTypes"];`,
      errors: [
        expectForbiddenPropTypeError([1, 13, 1, 22]),
        expectForbiddenPropTypeError([1, 36, 1, 47]),
      ],
    },
    // ---- JSX context ----
    {
      code: `var X = <div>{Bar.propTypes}</div>;`,
      errors: [expectForbiddenPropTypeError([1, 19, 1, 28])],
    },
    {
      code: `var X = <Foo p={Bar.propTypes} />;`,
      errors: [expectForbiddenPropTypeError([1, 21, 1, 30])],
    },
    {
      code: `var X = <Foo {...Bar.propTypes} />;`,
      errors: [expectForbiddenPropTypeError([1, 22, 1, 31])],
    },
    // ---- TypeScript receiver / suffix wrappers ----
    {
      code: `(Foo as any).propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 14, 1, 23])],
    },
    {
      code: `(Foo satisfies any).propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 21, 1, 30])],
    },
    {
      code: `Foo.propTypes!;`,
      errors: [expectForbiddenPropTypeError([1, 5, 1, 14])],
    },
    {
      code: `Foo.propTypes!.x;`,
      errors: [expectForbiddenPropTypeError([1, 5, 1, 14])],
    },
    {
      code: `Foo?.['propTypes'];`,
      errors: [expectForbiddenPropTypeError([1, 7, 1, 18])],
    },
    // ---- Spread / array / object literal positions ----
    {
      code: `const x = {...Foo.propTypes};`,
      errors: [expectForbiddenPropTypeError([1, 19, 1, 28])],
    },
    {
      code: `const x = [Foo.propTypes];`,
      errors: [expectForbiddenPropTypeError([1, 16, 1, 25])],
    },
    {
      code: `const x = [...Foo.propTypes];`,
      errors: [expectForbiddenPropTypeError([1, 19, 1, 28])],
    },
    // ---- Function parameter / default contexts ----
    {
      code: `function f(x = Foo.propTypes) {}`,
      errors: [expectForbiddenPropTypeError([1, 20, 1, 29])],
    },
    {
      code: `function f({ propTypes }) {}`,
      errors: [expectForbiddenPropTypeError([1, 14, 1, 23])],
    },
    {
      code: `function f({ a: { propTypes } }) {}`,
      errors: [expectForbiddenPropTypeError([1, 19, 1, 28])],
    },
    {
      code: `function f({ propTypes = {} }) {}`,
      errors: [expectForbiddenPropTypeError([1, 14, 1, 28])],
    },
    {
      code: `var [{ propTypes }] = arr;`,
      errors: [expectForbiddenPropTypeError([1, 8, 1, 17])],
    },
    // ---- Class / decorator positions ----
    {
      code: `class C extends Foo.propTypes {}`,
      errors: [expectForbiddenPropTypeError([1, 21, 1, 30])],
    },
    {
      code: `class C { [Foo.propTypes]() {} }`,
      errors: [expectForbiddenPropTypeError([1, 16, 1, 25])],
    },
    {
      code: `@Foo.propTypes class X {}`,
      errors: [expectForbiddenPropTypeError([1, 6, 1, 15])],
    },
    {
      code: `interface I extends Foo.propTypes {}`,
      errors: [expectForbiddenPropTypeError([1, 25, 1, 34])],
    },
    {
      code: `class C { static propTypes() { return Bar.propTypes; } }`,
      options: [{ allowInPropTypes: true }],
      errors: [expectForbiddenPropTypeError([1, 43, 1, 52])],
    },
    {
      code: `class C { static get propTypes() { return Bar.propTypes; } }`,
      options: [{ allowInPropTypes: true }],
      errors: [expectForbiddenPropTypeError([1, 47, 1, 56])],
    },
    // ---- Tagged template / call / new ----
    {
      code: 'Foo.propTypes`x`;',
      errors: [expectForbiddenPropTypeError([1, 5, 1, 14])],
    },
    {
      code: `Foo.propTypes();`,
      errors: [expectForbiddenPropTypeError([1, 5, 1, 14])],
    },
    {
      code: `new Foo.propTypes();`,
      errors: [expectForbiddenPropTypeError([1, 9, 1, 18])],
    },
    {
      code: `Foo.propTypes<any>();`,
      errors: [expectForbiddenPropTypeError([1, 5, 1, 14])],
    },
    {
      code: '`${Foo.propTypes}`;',
      errors: [expectForbiddenPropTypeError([1, 8, 1, 17])],
    },
    // ---- Loops / control flow ----
    {
      code: `for (const k in Foo.propTypes) {}`,
      errors: [expectForbiddenPropTypeError([1, 21, 1, 30])],
    },
    {
      code: `for (const v of Foo.propTypes) {}`,
      errors: [expectForbiddenPropTypeError([1, 21, 1, 30])],
    },
    {
      code: `while (Foo.propTypes) {}`,
      errors: [expectForbiddenPropTypeError([1, 12, 1, 21])],
    },
    {
      code: `async function f() { return await Foo.propTypes; }`,
      errors: [expectForbiddenPropTypeError([1, 39, 1, 48])],
    },
    {
      code: `function* g() { yield Foo.propTypes; }`,
      errors: [expectForbiddenPropTypeError([1, 27, 1, 36])],
    },
    // ---- Destructuring assignment edge cases ----
    {
      code: `({ foo: a, propTypes } = SomeComponent);`,
      errors: [expectForbiddenPropTypeError([1, 12, 1, 21])],
    },
    {
      code: `({ propTypes, ...rest } = SomeComponent);`,
      errors: [expectForbiddenPropTypeError([1, 4, 1, 13])],
    },
    {
      code: `({ propTypes = {} } = SomeComponent);`,
      errors: [expectForbiddenPropTypeError([1, 4, 1, 18])],
    },
    // ---- Receiver-type variations ----
    {
      code: `class C { m() { return this.propTypes; } }`,
      errors: [expectForbiddenPropTypeError([1, 29, 1, 38])],
    },
    {
      code: `class C extends D { m() { return super.propTypes; } }`,
      errors: [expectForbiddenPropTypeError([1, 40, 1, 49])],
    },
    {
      code: `f().propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 5, 1, 14])],
    },
    {
      code: `(cond ? A : B).propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 16, 1, 25])],
    },
    {
      code: `(a, B).propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 8, 1, 17])],
    },
    {
      code: `arr[0].propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 8, 1, 17])],
    },
    {
      code: `import('x').propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 13, 1, 22])],
    },
    // ---- Chained .propTypes.propTypes (both fire) ----
    {
      code: `Foo.propTypes.propTypes;`,
      errors: [
        expectForbiddenPropTypeError([1, 5, 1, 14]),
        expectForbiddenPropTypeError([1, 15, 1, 24]),
      ],
    },
    // ---- Logical assignment LHS exclusions; RHS still fires ----
    {
      code: `Foo.propTypes ||= Bar.propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 23, 1, 32])],
    },
    {
      code: `Foo.propTypes &&= Bar.propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 23, 1, 32])],
    },
    {
      code: `Foo.propTypes ??= Bar.propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 23, 1, 32])],
    },
    // ---- Operator / unary positions ----
    {
      code: `delete Foo.propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 12, 1, 21])],
    },
    {
      code: `var x = typeof Foo.propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 20, 1, 29])],
    },
    {
      code: `void Foo.propTypes;`,
      errors: [expectForbiddenPropTypeError([1, 10, 1, 19])],
    },
    // ---- Class expression / catch / fragment / spread call / member decorator ----
    {
      code: `const X = class extends Foo.propTypes {};`,
      errors: [expectForbiddenPropTypeError([1, 29, 1, 38])],
    },
    {
      code: `try {} catch ({ propTypes }) {}`,
      errors: [expectForbiddenPropTypeError([1, 17, 1, 26])],
    },
    {
      code: `var X = <>{Foo.propTypes}</>;`,
      errors: [expectForbiddenPropTypeError([1, 16, 1, 25])],
    },
    {
      code: `f(...Foo.propTypes);`,
      errors: [expectForbiddenPropTypeError([1, 10, 1, 19])],
    },
    {
      code: `class C { @Foo.propTypes prop = 1; }`,
      errors: [expectForbiddenPropTypeError([1, 16, 1, 25])],
    },
    {
      code: `for (const { propTypes } of arr) {}`,
      errors: [expectForbiddenPropTypeError([1, 14, 1, 23])],
    },
    {
      code: `var X = <NS.Foo p={Bar.propTypes} />;`,
      errors: [expectForbiddenPropTypeError([1, 24, 1, 33])],
    },
  ],
});
