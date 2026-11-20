import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectPatternMismatchError1(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'patternMismatch',
    "Prop name `something` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
    range,
  );
}

function expectPatternMismatchError2(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'patternMismatch',
    "Prop name `enabled` doesn't match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
    range,
  );
}

function expectPatternMismatchError3(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'patternMismatch',
    "Prop name `enabled` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
    range,
  );
}

function expectPatternMismatchError4(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'patternMismatch',
    "Prop name `semi` doesn't match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
    range,
  );
}

function expectPatternMismatchError5(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'patternMismatch',
    "Prop name `showScore` doesn't match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
    range,
  );
}

function expectPatternMismatchError6(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'patternMismatch',
    "Prop name `failingItIs` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

const patternIs = '^is[A-Z]([A-Za-z0-9]?)+';
const patternIsHas = '^(is|has)[A-Z]([A-Za-z0-9]?)+';

ruleTester.run('boolean-prop-naming', {
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
    // Interface heritage — matching members pass.
    {
      code: `
        interface Base { isVisible: boolean }
        interface Props extends Base { isReady: boolean }
        const Hello = (props: Props) => <div />;
      `,
      options: [{ rule: patternIs }],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {isSomething: PropTypes.bool},
          render: function() { return <div />; }
        });
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {isSomething: React.PropTypes.bool},
          render: function() { return <div />; }
        });
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        var Hello = React.createClass({
          propTypes: {isSomething: PropTypes.bool},
          render: function() { return <div />; }
        });
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      settings: { react: { createClass: 'createClass' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        var Hello = React.createClass({
          propTypes: {something: PropTypes.any},
          render: function() { return <div />; }
        });
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      settings: { react: { createClass: 'createClass' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render () { return <div />; }
        }
        Hello.propTypes = {isSomething: PropTypes.bool}
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render () { return <div />; }
        }
        Hello.propTypes = wrap({ a: PropTypes.bool })
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render () { return <div />; }
        }
        Hello.propTypes = {something: PropTypes.any}
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        class Hello extends Component {
          render () { return <div />; }
        }
        Hello.propTypes = {isSomething: PropTypes.bool}
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        const spreadProps = { aSpreadProp: PropTypes.string };
        class Hello extends React.Component {
          static propTypes = {isSomething: PropTypes.bool, ...spreadProps};
          render () { return <div />; }
        }
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        const spreadProps = { aSpreadProp: PropTypes.string };
        class Hello extends Component {
          render () { return <div />; }
        }
        Hello.propTypes = {isSomething: PropTypes.bool, ...spreadProps}
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {isSomething: React.PropTypes.bool};
          render () { return <div />; }
        }
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {something: PropTypes.any};
          render () { return <div />; }
        }
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props: {isSomething: boolean};
          render () { return <div />; }
        }
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props: {something: any};
          render () { return <div />; }
        }
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        var Hello = ({isSomething}) => { return <div /> }
        Hello.propTypes = {isSomething: PropTypes.bool};
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        type Props = {
          isSomething: boolean;
        };
        function Hello(props: Props): React.Element { return <div /> }
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {
            isSomething: PropTypes.mutuallyExclusiveTrueProps,
            something: PropTypes.bool
          };
          render () { return <div />; }
        }
      `,
      options: [
        {
          propTypeNames: ['mutuallyExclusiveTrueProps'],
          rule: '^is[A-Z]([A-Za-z0-9]?)+',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {
            isSomething: mutuallyExclusiveTrueProps,
            isSomethingElse: bool
          };
          render () { return <div />; }
        }
      `,
      options: [
        {
          propTypeNames: ['bool', 'mutuallyExclusiveTrueProps'],
          rule: '^is[A-Z]([A-Za-z0-9]?)+',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        var x = {a: 1}
        var y = {...x}
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        class Hello extends PureComponent {
          props: PropsType;
          render () { return <div /> }
        }
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
      function Card(props) {
        return <div>{props.showScore ? 'yeh' : 'no'}</div>;
      }
      Card.propTypes = merge({}, Card.propTypes, {
          showScore: PropTypes.bool
      });`,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {isSomething: PropTypes.bool.isRequired, hasValue: PropTypes.bool.isRequired},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {
            isSomething: PropTypes.bool.isRequired,
            hasValue: PropTypes.bool.isRequired
          };

          render() {
            return (
              <div />
            );
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return (
              <div />
            );
          }
        }

        Hello.propTypes = {
          isSomething: PropTypes.bool.isRequired,
          hasValue: PropTypes.bool.isRequired
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {something: PropTypes.shape({}).isRequired},
          render: function() { return <div />; }
        });
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        function SomeComponent({
            isSomething,
        }: {
            isSomething: boolean,
        }) {
            return (
                <span>{isSomething}</span>
            );
        }
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return (
              <div />
            );
          }
        }

        Hello.propTypes = {
          isSomething: PropTypes.bool.isRequired,
          nested: PropTypes.shape({
            isWorking: PropTypes.bool
          })
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return (
              <div />
            );
          }
        }

        Hello.propTypes = {
          isSomething: PropTypes.bool.isRequired,
          nested: PropTypes.shape({
            nested: PropTypes.shape({
              isWorking: PropTypes.bool
            })
          })
        };
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+', validateNested: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        type TestFNType = {
          isEnabled: boolean
        }
        const HelloNew = (props: TestFNType) => { return <div /> };
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        type Props = {
          isEnabled: boolean
        } & OtherProps
        const HelloNew = (props: Props) => { return <div /> };
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
        type Props = {
          isEnabled: boolean
        } & {
          hasLOL: boolean
        } & OtherProps
        const HelloNew = (props: Props) => { return <div /> };
      `,
      options: [{ rule: '(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        type Props = {
          isEnabled: boolean
        }

        const HelloNew: React.FC<Props> = (props) => { return <div /> };
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
        type Props = {
          isEnabled: boolean
        } & {
          hasLOL: boolean
        }

        const HelloNew: React.FC<Props> = (props) => { return <div /> };
      `,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
        type Props = {
          isEnabled: boolean
        } | {
          hasLOL: boolean
        }

        const HelloNew = (props: Props) => { return <div /> };
      `,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `
        type Props = {
          isEnabled: boolean
        } & ({
          hasLOL: boolean
        } | {
          isLOL: boolean
        })

        const HelloNew = (props: Props) => { return <div /> };
      `,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
      code: `
        export const DataRow = (props: { label: string; value: string; } & React.HTMLAttributes<HTMLDivElement>) => {
            const { label, value, ...otherProps } = props;
            return (
                <div {...otherProps}>
                    <span>{label}</span>
                    <span>{value}</span>
                </div>
            );
        };
      `,
      options: [
        {
          rule: '(^(is|has|should|without)[A-Z]([A-Za-z0-9]?)+|disabled|required|checked|defaultChecked)',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
      code: `
        // Strip @jsx comments, see https://github.com/microsoft/fluentui/issues/29126
        const resultCode = result.code
          .replace('/** @jsxRuntime automatic */', '')
          .replace('/** @jsxImportSource @fluentui/react-jsx-runtime */', '');
      `,
      languageOptions: upstreamModuleOptions,
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
      errors: [expectPatternMismatchError1([3, 23, 3, 48])],
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
      errors: [expectPatternMismatchError1([3, 31, 3, 56])],
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
      errors: [expectPatternMismatchError1([5, 28, 5, 53])],
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
      errors: [expectPatternMismatchError1([5, 31, 5, 56])],
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
      errors: [expectPatternMismatchError1([5, 31, 5, 56])],
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
      errors: [expectPatternMismatchError1([5, 28, 5, 62])],
    },
    // Inline TS parameter type literal.
    {
      code: `const Hello = (props: {enabled:boolean}) => <div />;`,
      options: [{ rule: patternIsHas }],
      errors: [expectPatternMismatchError2([1, 24, 1, 39])],
    },
    // React.FC<Props> with type alias.
    {
      code: `
        type Props = { enabled: boolean }
        const HelloNew: React.FC<Props> = (props) => { return <div /> };
      `,
      options: [{ rule: patternIs }],
      errors: [expectPatternMismatchError3([2, 24, 2, 40])],
    },
    // Intersection of type alias + inline literal.
    {
      code: `
        type Props = { enabled: boolean };
        const Hello = (props: Props & { semi: boolean }) => <div />;
      `,
      options: [{ rule: patternIsHas }],
      errors: [
        expectPatternMismatchError2([2, 24, 2, 40]),
        expectPatternMismatchError4([3, 41, 3, 54]),
      ],
    },
    // Interface via type reference on first param.
    {
      code: `
        interface TestFNType { enabled: boolean }
        const HelloNew = (props: TestFNType) => { return <div /> };
      `,
      options: [{ rule: patternIs }],
      errors: [expectPatternMismatchError3([2, 32, 2, 48])],
    },
    // Existing native false positive: a qualified reference is resolved by its
    // final name, so ns.Inner incorrectly selects the unrelated top-level Inner.
    // Preserve the baseline assertion until binding-aware type lookup is fixed.
    {
      code: `
        type Inner = { enabled: boolean };
        const Hello = (props: ns.Inner) => <div />;
      `,
      options: [{ rule: patternIs }],
      errors: [expectPatternMismatchError3([2, 24, 2, 40])],
    },
    // Type alias indirection.
    {
      code: `
        type A = B;
        type B = { enabled: boolean };
        const Hello = (props: A) => <div />;
      `,
      options: [{ rule: patternIs }],
      errors: [expectPatternMismatchError3([3, 20, 3, 36])],
    },
    // Interface declaration merging.
    {
      code: `
        interface Props { enabled: boolean }
        interface Props { switched: boolean }
        const Hello = (props: Props) => <div />;
      `,
      options: [{ rule: patternIs }],
      errors: [
        expectPatternMismatchError3([2, 27, 2, 43]),
        diagnostic(
          'patternMismatch',
          "Prop name `switched` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
          [3, 27, 3, 44],
        ),
      ],
    },
    // HOC wrapping with import.
    {
      code: `
        import { memo } from 'react';
        const Hello = memo((props) => <div/>);
        Hello.propTypes = {something: PropTypes.bool};
      `,
      options: [{ rule: patternIs }],
      errors: [expectPatternMismatchError1([4, 28, 4, 53])],
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
        diagnostic(
          'patternMismatch',
          'It is better if your prop (something) matches this pattern: (^is[A-Z]([A-Za-z0-9]?)+)',
          [5, 28, 5, 53],
        ),
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
      errors: [
        expectPatternMismatchError1([4, 13, 4, 50]),
        diagnostic(
          'patternMismatch',
          "Prop name `somethingElse` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
          [5, 13, 5, 32],
        ),
      ],
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
      errors: [expectPatternMismatchError5([6, 11, 6, 36])],
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
      errors: [expectPatternMismatchError5([4, 11, 4, 36])],
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
      errors: [expectPatternMismatchError6([7, 37, 7, 64])],
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
      errors: [
        expectPatternMismatchError1([3, 32, 3, 57]),
        diagnostic(
          'patternMismatch',
          "Prop name `fooBar` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
          [4, 20, 4, 35],
        ),
      ],
    },
    // Interface heritage `extends`.
    {
      code: `
        interface Base { hidden: boolean }
        interface Props extends Base { ready: boolean }
        const Hello = (props: Props) => <div />;
      `,
      options: [{ rule: patternIs }],
      errors: [
        diagnostic(
          'patternMismatch',
          "Prop name `hidden` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
          [2, 26, 2, 41],
        ),
        diagnostic(
          'patternMismatch',
          "Prop name `ready` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
          [3, 40, 3, 54],
        ),
      ],
    },
    // Optional / readonly modifiers.
    {
      code: `
        type Props = { enabled?: boolean; readonly visible: boolean };
        const Hello = (props: Props) => <div/>;
      `,
      options: [{ rule: patternIs }],
      errors: [
        expectPatternMismatchError3([2, 24, 2, 42]),
        diagnostic(
          'patternMismatch',
          "Prop name `visible` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
          [2, 43, 2, 68],
        ),
      ],
    },
    // The same existing name-based lookup issue applies to ns.sub.C.
    {
      code: `
        type C = { enabled: boolean };
        const Hello = (props: ns.sub.C) => <div />;
      `,
      options: [{ rule: patternIs }],
      errors: [expectPatternMismatchError3([2, 20, 2, 36])],
    },
    // React.PropsWithChildren<Props>.
    {
      code: `
        type Props = { enabled: boolean };
        const Hello: React.PropsWithChildren<Props> = (props) => <div/>;
      `,
      options: [{ rule: patternIs }],
      errors: [expectPatternMismatchError3([2, 24, 2, 40])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {something: React.PropTypes.bool},
          render: function() { return <div />; }
        });
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError1([3, 23, 3, 54])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        var Hello = React.createClass({
          propTypes: {something: PropTypes.bool},
          render: function() { return <div />; }
        });
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      settings: { react: { createClass: 'createClass' } },
      errors: [expectPatternMismatchError1([3, 23, 3, 48])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        class Hello extends Component {
          render () { return <div />; }
        }
        Hello.propTypes = {something: PropTypes.bool}
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError1([5, 28, 5, 53])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        const spreadProps = { aSpreadProp: PropTypes.string };
        class Hello extends Component {
          render () { return <div />; }
        }
        Hello.propTypes = {something: PropTypes.bool, ...spreadProps}
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError1([6, 28, 6, 53])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        const spreadProps = { aSpreadProp: PropTypes.string };
        class Hello extends React.Component {
          static propTypes = {something: PropTypes.bool, ...spreadProps};
          render () { return <div />; }
        }
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError1([4, 31, 4, 56])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        class Hello extends React.Component {
          props: {something: boolean};
          render () { return <div />; }
        }
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError1([3, 19, 3, 37])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        var Hello = ({something}) => { return <div /> }
        Hello.propTypes = {something: PropTypes.bool};
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError1([3, 28, 3, 53])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        type Props = {
          something: boolean;
        };
        function Hello(props: Props): React.Element { return <div /> }
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError1([3, 11, 3, 30])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {something: PropTypes.mutuallyExclusiveTrueProps};
          render () { return <div />; }
        }
      `,
      options: [
        {
          propTypeNames: ['bool', 'mutuallyExclusiveTrueProps'],
          rule: '^is[A-Z]([A-Za-z0-9]?)+',
        },
      ],
      errors: [expectPatternMismatchError1([3, 31, 3, 78])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {
            something: PropTypes.mutuallyExclusiveTrueProps,
            somethingElse: PropTypes.bool
          };
          render () { return <div />; }
        }
      `,
      options: [
        {
          propTypeNames: ['bool', 'mutuallyExclusiveTrueProps'],
          rule: '^is[A-Z]([A-Za-z0-9]?)+',
        },
      ],
      errors: [
        expectPatternMismatchError1([4, 13, 4, 60]),
        diagnostic(
          'patternMismatch',
          "Prop name `somethingElse` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
          [5, 13, 5, 42],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
      function Card(props) {
        return <div>{props.showScore ? 'yeh' : 'no'}</div>;
      }
      Card.propTypes = merge({}, Card.propTypes, {
          showScore: PropTypes.bool
      });`,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      settings: { propWrapperFunctions: ['merge'] },
      errors: [expectPatternMismatchError5([6, 11, 6, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
      function Card(props) {
        return <div>{props.showScore ? 'yeh' : 'no'}</div>;
      }
      Card.propTypes = Object.assign({}, Card.propTypes, {
          showScore: PropTypes.bool
      });`,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      settings: { propWrapperFunctions: ['Object.assign'] },
      errors: [expectPatternMismatchError5([6, 11, 6, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
      function Card(props) {
        return <div>{props.showScore ? 'yeh' : 'no'}</div>;
      }
      Card.propTypes = _.assign({}, Card.propTypes, {
          showScore: PropTypes.bool
      });`,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      settings: { propWrapperFunctions: ['_.assign'] },
      errors: [expectPatternMismatchError5([6, 11, 6, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
      function Card(props) {
        return <div>{props.showScore ? 'yeh' : 'no'}</div>;
      }
      Card.propTypes = forbidExtraProps({
          showScore: PropTypes.bool
      });`,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [expectPatternMismatchError5([6, 11, 6, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
      class Card extends React.Component {
        render() {
          return <div>{props.showScore ? 'yeh' : 'no'}</div>;
        }
      }
      Card.propTypes = forbidExtraProps({
          showScore: PropTypes.bool
      });`,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [expectPatternMismatchError5([8, 11, 8, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
      class Card extends React.Component {
        static propTypes = forbidExtraProps({
          showScore: PropTypes.bool
        });
        render() {
          return <div>{props.showScore ? 'yeh' : 'no'}</div>;
        }
      }`,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      settings: { propWrapperFunctions: ['forbidExtraProps'] },
      errors: [expectPatternMismatchError5([4, 11, 4, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render () { return <div />; }
        }
        Hello.propTypes = {something: PropTypes.bool}
      `,
      options: [
        {
          rule: '^is[A-Z]([A-Za-z0-9]?)+',
          message: "Boolean prop names must begin with either 'is' or 'has'",
        },
      ],
      errors: [
        diagnostic(
          'patternMismatch',
          "Boolean prop names must begin with either 'is' or 'has'",
          [5, 28, 5, 53],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `
        var Hello = createReactClass({
          propTypes: {something: PropTypes.bool.isRequired},
          render: function() { return <div />; }
        });
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError1([3, 23, 3, 59])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        class Hello extends React.Component {
          static propTypes = {
            something: PropTypes.bool.isRequired
          };

          render() {
            return (
              <div />
            );
          }
        }
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError1([4, 13, 4, 49])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return (
              <div />
            );
          }
        }

        Hello.propTypes = {
          something: PropTypes.bool.isRequired
        }
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError1([11, 11, 11, 47])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
        function SomeComponent({
            something,
        }: {
            something: boolean,
        }) {
            return (
                <span>{something}</span>
            );
        }
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError1([5, 13, 5, 32])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return (
              <div />
            );
          }
        }

        Hello.propTypes = {
          isSomething: PropTypes.bool.isRequired,
          nested: PropTypes.shape({
            failingItIs: PropTypes.bool
          })
        };
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+', validateNested: true }],
      errors: [expectPatternMismatchError6([13, 13, 13, 40])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render() {
            return (
              <div />
            );
          }
        }

        Hello.propTypes = {
          isSomething: PropTypes.bool.isRequired,
          nested: PropTypes.shape({
            nested: PropTypes.shape({
              failingItIs: PropTypes.bool
            })
          })
        };
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+', validateNested: true }],
      errors: [expectPatternMismatchError6([14, 15, 14, 42])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `
      import { bool } from 'prop-types';
        var Hello = createReactClass({
          propTypes: {something: bool},
          render: function() { return <div />; }
        });
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+', validateNested: true }],
      errors: [expectPatternMismatchError1([4, 23, 4, 38])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `
        type TestConstType = {
          enabled: boolean
        }
        const HelloNew = (props: TestConstType) => { return <div /> };
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError3([3, 11, 3, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `
        type TestFNType = {
          enabled: boolean
        }
        const HelloNew = (props: TestFNType) => { return <div /> };
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError3([3, 11, 3, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `
        type Props = {
          enabled: boolean
        } & OtherProps

        const HelloNew = (props: Props) => { return <div /> };
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError3([3, 11, 3, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `
        type Props = {
          enabled: boolean
        } & {
          hasLOL: boolean
        } & OtherProps

        const HelloNew = (props: Props) => { return <div /> };
      `,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError2([3, 11, 3, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `
        type Props = {
          enabled: boolean
        }

        const HelloNew: React.FC<Props> = (props) => { return <div /> };
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError3([3, 11, 3, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
        type Props = {
          enabled: boolean
        } & {
          hasLOL: boolean
        }

        const HelloNew: React.FC<Props> = (props) => { return <div /> };
      `,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError2([3, 11, 3, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
        type Props = {
          enabled: boolean
        } | {
          hasLOL: boolean
        }

        const HelloNew = (props: Props) => { return <div /> };
      `,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError2([3, 11, 3, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `
        type Props = {
          enabled: boolean
        } & ({
          hasLOL: boolean
        } | {
          lol: boolean
        })

        const HelloNew = (props: Props) => { return <div /> };
      `,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      errors: [
        expectPatternMismatchError2([3, 11, 3, 27]),
        diagnostic(
          'patternMismatch',
          "Prop name `lol` doesn't match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
          [7, 11, 7, 23],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
      code: `
        interface TestFNType {
          enabled: boolean
        }
        const HelloNew = (props: TestFNType) => { return <div /> };
      `,
      options: [{ rule: '^is[A-Z]([A-Za-z0-9]?)+' }],
      errors: [expectPatternMismatchError3([3, 11, 3, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `
        type Props = {
          enabled: boolean
        };
        type BaseProps = {
          semi: boolean
        };

        const Hello = (props: Props & BaseProps) => <div />;
      `,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      errors: [
        expectPatternMismatchError2([3, 11, 3, 27]),
        expectPatternMismatchError4([6, 11, 6, 24]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `
        type Props = {
          enabled: boolean
        };

        const Hello = (props: Props & {
          semi: boolean
        }) => <div />;
      `,
      options: [{ rule: '^(is|has)[A-Z]([A-Za-z0-9]?)+' }],
      errors: [
        expectPatternMismatchError2([3, 11, 3, 27]),
        expectPatternMismatchError4([7, 11, 7, 24]),
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
