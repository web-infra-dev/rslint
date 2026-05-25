import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const patternIs = '^is[A-Z]([A-Za-z0-9]?)+';
const patternIsHas = '^(is|has)[A-Z]([A-Za-z0-9]?)+';

ruleTester.run('boolean-prop-naming', {} as never, {
  valid: [
    // Default (no options) → rule is a no-op.
    {
      code: `
        var Hello = createReactClass({
          propTypes: {isSomething: PropTypes.bool, hasValue: PropTypes.bool},
          render: function() { return <div />; }
        });
      `,
    },
    // Matching name with `is` pattern.
    {
      code: `
        class Hello extends React.Component {
          static propTypes = {isSomething: PropTypes.bool};
          render () { return <div />; }
        }
      `,
      options: [{ rule: patternIs }],
    },
    // React.FC<Props> with matching name.
    {
      code: `
        type Props = { isEnabled: boolean };
        const HelloNew: React.FC<Props> = (props) => { return <div /> };
      `,
      options: [{ rule: patternIs }],
    },
    // Wrapper not in propWrapperFunctions setting → silently skipped.
    {
      code: `
        function Card(props) {
          return <div>{props.showScore ? 'yeh' : 'no'}</div>;
        }
        Card.propTypes = wrap({ showScore: PropTypes.bool });
      `,
      options: [{ rule: patternIsHas }],
    },
    // Empty propTypes / empty type literal — must not crash.
    {
      code: `
        class Hello extends React.Component {
          static propTypes = {};
          render () { return <div />; }
        }
      `,
      options: [{ rule: patternIs }],
    },
    {
      code: `const Hello = (props: {}) => <div />;`,
      options: [{ rule: patternIs }],
    },
    // Non-component class with `propTypes` field — no report.
    {
      code: `
        class NotAComponent {
          propTypes = {something: PropTypes.bool};
        }
      `,
      options: [{ rule: patternIs }],
    },
    // TS expression wrappers on prop value (matching name still passes).
    {
      code: `
        class Hello extends React.Component {
          static propTypes = {isSomething: (PropTypes.bool as any)};
          render () { return <div />; }
        }
      `,
      options: [{ rule: patternIs }],
    },
    // Type alias indirection: A → B → object literal.
    {
      code: `
        type A = B;
        type B = { isEnabled: boolean };
        const Hello = (props: A) => <div />;
      `,
      options: [{ rule: patternIs }],
    },
    // Interface declaration merging — both members must match.
    {
      code: `
        interface Props { isFoo: boolean }
        interface Props { isBar: boolean }
        const Hello = (props: Props) => <div />;
      `,
      options: [{ rule: patternIs }],
    },
    // Three-level type composition (intersection × union × parens).
    {
      code: `
        type A = { isFoo: boolean };
        type B = { hasBar: boolean };
        type C = { isBaz: boolean };
        const Hello = (props: A & (B | C)) => <div />;
      `,
      options: [{ rule: patternIsHas }],
    },
    // Empty rule pattern → no-op.
    {
      code: `
        class Hello extends React.Component {
          static propTypes = {something: PropTypes.bool};
          render () { return <div />; }
        }
      `,
      options: [{ rule: '' }],
    },
    // Invalid regex degrades to no-op.
    {
      code: `
        class Hello extends React.Component {
          static propTypes = {something: PropTypes.bool};
          render () { return <div />; }
        }
      `,
      options: [{ rule: '[unclosed' }],
    },
    // Empty propTypeNames array → user-cleared list, no PropTypes match.
    {
      code: `
        class Hello extends React.Component {
          static propTypes = { something: PropTypes.bool };
          render () { return <div />; }
        }
      `,
      options: [{ rule: patternIs, propTypeNames: [] }],
    },
    // Interface heritage — matching members pass.
    {
      code: `
        interface Base { isVisible: boolean }
        interface Props extends Base { isReady: boolean }
        const Hello = (props: Props) => <div />;
      `,
      options: [{ rule: patternIs }],
    },
  ],
  invalid: [
    // createReactClass with non-matching prop.
    {
      code: `
        var Hello = createReactClass({
          propTypes: {something: PropTypes.bool},
          render: function() { return <div />; }
        });
      `,
      options: [{ rule: patternIs }],
      errors: [
        {
          message:
            "Prop name `something` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // Static class field.
    {
      code: `
        class Hello extends React.Component {
          static propTypes = {something: PropTypes.bool};
          render () { return <div />; }
        }
      `,
      options: [{ rule: patternIs }],
      errors: [
        {
          message:
            "Prop name `something` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // .propTypes assignment outside body.
    {
      code: `
        class Hello extends React.Component {
          render () { return <div />; }
        }
        Hello.propTypes = {something: PropTypes.bool}
      `,
      options: [{ rule: patternIs }],
      errors: [
        {
          message:
            "Prop name `something` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // ElementAccess assignment LHS.
    {
      code: `
        class Hello extends React.Component {
          render () { return <div />; }
        }
        Hello['propTypes'] = {something: PropTypes.bool};
      `,
      options: [{ rule: patternIs }],
      errors: [
        {
          message:
            "Prop name `something` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // Multi-segment receiver.
    {
      code: `
        class Hello extends React.Component {
          render () { return <div />; }
        }
        ns.Hello.propTypes = {something: PropTypes.bool};
      `,
      options: [{ rule: patternIs }],
      errors: [
        {
          message:
            "Prop name `something` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // TS expression wrappers report on mismatch.
    {
      code: `
        class Hello extends React.Component {
          render () { return <div />; }
        }
        Hello.propTypes = {something: (PropTypes.bool as any)};
      `,
      options: [{ rule: patternIs }],
      errors: [
        {
          message:
            "Prop name `something` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // Inline TS parameter type literal.
    {
      code: `const Hello = (props: {enabled:boolean}) => <div />;`,
      options: [{ rule: patternIsHas }],
      errors: [
        {
          message:
            "Prop name `enabled` doesn't match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // React.FC<Props> with type alias.
    {
      code: `
        type Props = { enabled: boolean }
        const HelloNew: React.FC<Props> = (props) => { return <div /> };
      `,
      options: [{ rule: patternIs }],
      errors: [
        {
          message:
            "Prop name `enabled` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // Intersection of type alias + inline literal.
    {
      code: `
        type Props = { enabled: boolean };
        const Hello = (props: Props & { semi: boolean }) => <div />;
      `,
      options: [{ rule: patternIsHas }],
      errors: [
        {
          message:
            "Prop name `enabled` doesn't match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
        },
        {
          message:
            "Prop name `semi` doesn't match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // Interface via type reference on first param.
    {
      code: `
        interface TestFNType { enabled: boolean }
        const HelloNew = (props: TestFNType) => { return <div /> };
      `,
      options: [{ rule: patternIs }],
      errors: [
        {
          message:
            "Prop name `enabled` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // QualifiedName type reference (`A.B`).
    {
      code: `
        type Inner = { enabled: boolean };
        const Hello = (props: ns.Inner) => <div />;
      `,
      options: [{ rule: patternIs }],
      errors: [
        {
          message:
            "Prop name `enabled` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // Type alias indirection.
    {
      code: `
        type A = B;
        type B = { enabled: boolean };
        const Hello = (props: A) => <div />;
      `,
      options: [{ rule: patternIs }],
      errors: [
        {
          message:
            "Prop name `enabled` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // Interface declaration merging.
    {
      code: `
        interface Props { enabled: boolean }
        interface Props { switched: boolean }
        const Hello = (props: Props) => <div />;
      `,
      options: [{ rule: patternIs }],
      errors: 2,
    },
    // HOC wrapping with import.
    {
      code: `
        import { memo } from 'react';
        const Hello = memo((props) => <div/>);
        Hello.propTypes = {something: PropTypes.bool};
      `,
      options: [{ rule: patternIs }],
      errors: [
        {
          message:
            "Prop name `something` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // Custom message with placeholders.
    {
      code: `
        class Hello extends React.Component {
          render () { return <div />; }
        }
        Hello.propTypes = {something: PropTypes.bool}
      `,
      options: [
        {
          rule: patternIs,
          message:
            'It is better if your prop ({{ propName }}) matches this pattern: ({{ pattern }})',
        },
      ],
      errors: [
        {
          message:
            'It is better if your prop (something) matches this pattern: (^is[A-Z]([A-Za-z0-9]?)+)',
        },
      ],
    },
    // Custom propTypeNames including bare identifiers.
    {
      code: `
        class Hello extends React.Component {
          static propTypes = {
            something: mutuallyExclusiveTrueProps,
            somethingElse: bool
          };
          render () { return <div />; }
        }
      `,
      options: [
        {
          propTypeNames: ['bool', 'mutuallyExclusiveTrueProps'],
          rule: patternIs,
        },
      ],
      errors: 2,
    },
    // propWrapperFunctions: bare-string entry.
    {
      code: `
        function Card(props) {
          return <div>{props.showScore ? 'yeh' : 'no'}</div>;
        }
        Card.propTypes = forbidExtraProps({
          showScore: PropTypes.bool
        });
      `,
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      options: [{ rule: patternIsHas }],
      errors: [
        {
          message:
            "Prop name `showScore` doesn't match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // propWrapperFunctions: {object, property} pair.
    {
      code: `
        function Card(props) { return <div /> }
        Card.propTypes = Object.assign({}, Card.propTypes, {
          showScore: PropTypes.bool
        });
      `,
      settings: {
        propWrapperFunctions: [{ object: 'Object', property: 'assign' }],
      },
      options: [{ rule: patternIsHas }],
      errors: [
        {
          message:
            "Prop name `showScore` doesn't match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // validateNested + PropTypes.shape recursion.
    {
      code: `
        class Hello extends React.Component {
          render() { return <div />; }
        }
        Hello.propTypes = {
          isSomething: PropTypes.bool.isRequired,
          nested: PropTypes.shape({ failingItIs: PropTypes.bool })
        };
      `,
      options: [{ rule: patternIs, validateNested: true }],
      errors: [
        {
          message:
            "Prop name `failingItIs` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // Both static propTypes + props: TypeLiteral.
    {
      code: `
        class Hello extends React.Component {
          static propTypes = { something: PropTypes.bool };
          props: { fooBar: boolean };
          render () { return <div />; }
        }
      `,
      options: [{ rule: patternIs }],
      errors: 2,
    },
    // Interface heritage `extends`.
    {
      code: `
        interface Base { hidden: boolean }
        interface Props extends Base { ready: boolean }
        const Hello = (props: Props) => <div />;
      `,
      options: [{ rule: patternIs }],
      errors: 2,
    },
    // Optional / readonly modifiers.
    {
      code: `
        type Props = { enabled?: boolean; readonly visible: boolean };
        const Hello = (props: Props) => <div/>;
      `,
      options: [{ rule: patternIs }],
      errors: 2,
    },
    // 3-segment QualifiedName.
    {
      code: `
        type C = { enabled: boolean };
        const Hello = (props: ns.sub.C) => <div />;
      `,
      options: [{ rule: patternIs }],
      errors: [
        {
          message:
            "Prop name `enabled` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
    // React.PropsWithChildren<Props>.
    {
      code: `
        type Props = { enabled: boolean };
        const Hello: React.PropsWithChildren<Props> = (props) => <div/>;
      `,
      options: [{ rule: patternIs }],
      errors: [
        {
          message:
            "Prop name `enabled` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
        },
      ],
    },
  ],
});
