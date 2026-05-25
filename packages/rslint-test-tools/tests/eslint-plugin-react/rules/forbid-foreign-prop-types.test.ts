import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('forbid-foreign-prop-types', {} as never, {
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
      errors: [{ messageId: 'forbiddenPropType' }],
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
      errors: [{ messageId: 'forbiddenPropType' }],
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
      errors: [{ messageId: 'forbiddenPropType' }],
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
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `
        class MyComponent extends React.Component {
          static fooBar = {
            baz: Qux.propTypes.baz
          };
        }
      `,
      errors: [{ messageId: 'forbiddenPropType' }],
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
      errors: [{ messageId: 'forbiddenPropType' }],
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
      errors: [{ messageId: 'forbiddenPropType' }],
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
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    // ---- Lock-in tests ----
    {
      code: `Foo?.propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `(Foo).propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `declare const Foo: any; Foo!.propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `Foo.propTypes.isRequired;`,
      errors: [{ messageId: 'forbiddenPropType' }],
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
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `Foo.notPropTypes = { x: Bar.propTypes.x };`,
      options: [{ allowInPropTypes: true }],
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `Foo['propTypes'] = { x: Bar.propTypes.x };`,
      options: [{ allowInPropTypes: true }],
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `({ propTypes } = SomeComponent);`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `({ propTypes: alias } = SomeComponent);`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `({ a: { propTypes } } = SomeComponent);`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `([{ propTypes }] = [SomeComponent]);`,
      errors: [{ messageId: 'forbiddenPropType' }],
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
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `var a = Foo.propTypes; var b = Bar["propTypes"];`,
      errors: [
        { messageId: 'forbiddenPropType' },
        { messageId: 'forbiddenPropType' },
      ],
    },
    // ---- JSX context ----
    {
      code: `var X = <div>{Bar.propTypes}</div>;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `var X = <Foo p={Bar.propTypes} />;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `var X = <Foo {...Bar.propTypes} />;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    // ---- TypeScript receiver / suffix wrappers ----
    {
      code: `(Foo as any).propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `(Foo satisfies any).propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `Foo.propTypes!;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `Foo.propTypes!.x;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `Foo?.['propTypes'];`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    // ---- Spread / array / object literal positions ----
    {
      code: `const x = {...Foo.propTypes};`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `const x = [Foo.propTypes];`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `const x = [...Foo.propTypes];`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    // ---- Function parameter / default contexts ----
    {
      code: `function f(x = Foo.propTypes) {}`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `function f({ propTypes }) {}`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `function f({ a: { propTypes } }) {}`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `function f({ propTypes = {} }) {}`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `var [{ propTypes }] = arr;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    // ---- Class / decorator positions ----
    {
      code: `class C extends Foo.propTypes {}`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `class C { [Foo.propTypes]() {} }`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `@Foo.propTypes class X {}`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `interface I extends Foo.propTypes {}`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `class C { static propTypes() { return Bar.propTypes; } }`,
      options: [{ allowInPropTypes: true }],
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `class C { static get propTypes() { return Bar.propTypes; } }`,
      options: [{ allowInPropTypes: true }],
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    // ---- Tagged template / call / new ----
    {
      code: 'Foo.propTypes`x`;',
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `Foo.propTypes();`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `new Foo.propTypes();`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `Foo.propTypes<any>();`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: '`${Foo.propTypes}`;',
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    // ---- Loops / control flow ----
    {
      code: `for (const k in Foo.propTypes) {}`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `for (const v of Foo.propTypes) {}`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `while (Foo.propTypes) {}`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `async function f() { return await Foo.propTypes; }`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `function* g() { yield Foo.propTypes; }`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    // ---- Destructuring assignment edge cases ----
    {
      code: `({ foo: a, propTypes } = SomeComponent);`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `({ propTypes, ...rest } = SomeComponent);`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `({ propTypes = {} } = SomeComponent);`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    // ---- Receiver-type variations ----
    {
      code: `class C { m() { return this.propTypes; } }`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `class C extends D { m() { return super.propTypes; } }`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `f().propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `(cond ? A : B).propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `(a, B).propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `arr[0].propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `import('x').propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    // ---- Chained .propTypes.propTypes (both fire) ----
    {
      code: `Foo.propTypes.propTypes;`,
      errors: [
        { messageId: 'forbiddenPropType' },
        { messageId: 'forbiddenPropType' },
      ],
    },
    // ---- Logical assignment LHS exclusions; RHS still fires ----
    {
      code: `Foo.propTypes ||= Bar.propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `Foo.propTypes &&= Bar.propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `Foo.propTypes ??= Bar.propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    // ---- Operator / unary positions ----
    {
      code: `delete Foo.propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `var x = typeof Foo.propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `void Foo.propTypes;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    // ---- Class expression / catch / fragment / spread call / member decorator ----
    {
      code: `const X = class extends Foo.propTypes {};`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `try {} catch ({ propTypes }) {}`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `var X = <>{Foo.propTypes}</>;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `f(...Foo.propTypes);`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `class C { @Foo.propTypes prop = 1; }`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `for (const { propTypes } of arr) {}`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
    {
      code: `var X = <NS.Foo p={Bar.propTypes} />;`,
      errors: [{ messageId: 'forbiddenPropType' }],
    },
  ],
});
