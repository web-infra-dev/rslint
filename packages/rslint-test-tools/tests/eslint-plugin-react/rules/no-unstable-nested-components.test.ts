import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// Message constants mirror eslint-plugin-react v7.37.x byte-for-byte,
// including the typographic apostrophe (U+2019) and double quotation
// marks (U+201C / U+201D) around the parent name.
const ERROR_MESSAGE =
  'Do not define components during render. React will see a new component type on every render and destroy the entire subtree’s DOM nodes and state (https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, move this component definition out of the parent component “ParentComponent” and pass data as props.';
const ERROR_MESSAGE_WITHOUT_NAME =
  'Do not define components during render. React will see a new component type on every render and destroy the entire subtree’s DOM nodes and state (https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, move this component definition out of the parent component and pass data as props.';
const ERROR_MESSAGE_COMPONENT_AS_PROPS =
  ERROR_MESSAGE +
  ' If you want to allow component creation in props, set allowAsProps option to true.';

ruleTester.run('no-unstable-nested-components', {} as never, {
  valid: [
    {
      code: `
        function ParentComponent() {
          return (
            <div>
              <OutsideDefinedFunctionComponent />
            </div>
          );
        }
      `,
    },
    {
      code: `
        function ParentComponent() {
          return (
            <SomeComponent
              footer={<OutsideDefinedComponent />}
              header={<div />}
              />
          );
        }
      `,
    },
    {
      code: `
        function ParentComponent() {
          return (
            <RenderPropComponent>
              {() => <div />}
            </RenderPropComponent>
          );
        }
      `,
    },
    {
      code: `
        function ParentComponent() {
          return (
            <RenderPropComponent children={() => <div />} />
          );
        }
      `,
    },
    {
      code: `
        function ParentComponent(props) {
          return (
            <ul>
              {props.items.map(item => (
                <li key={item.id}>
                  {item.name}
                </li>
              ))}
            </ul>
          );
        }
      `,
    },
    {
      code: `
        function ParentComponent() {
          return (
            <ComponentWithProps footer={() => <div />} />
          );
        }
      `,
      options: [{ allowAsProps: true }],
    },
    {
      code: `
        function ParentComponent() {
          return (
            <ComponentForProps renderFooter={() => <div />} />
          );
        }
      `,
    },
    {
      code: `
        function ParentComponent() {
          return <SomeComponent renderers={{ notComponent: () => null }} />;
        }
      `,
    },
    {
      code: `
        function ParentComponent() {
          return (
            <SomeComponent renderers={{ Header: () => <div /> }} />
          )
        }
      `,
    },
    // ---- settings.react.pragma override — `Preact.createElement` is
    // recognized as a JSX-returning call when pragma is configured. The
    // outer ParentComponent has no nested component definition here, so
    // the rule stays silent; the assertion is "no false-positive on the
    // outer createElement call" given the custom pragma. ----
    {
      code: `
        function ParentComponent() {
          return Preact.createElement(OutsideDefined, null);
        }
      `,
      settings: { react: { pragma: 'Preact' } },
    },
    // ---- settings.componentWrapperFunctions (string form) — a top-level
    // myMemo wrapper has no enclosing component, so it stays valid even
    // though the wrapper is registered. ----
    {
      code: `
        const TopLevel = myMemo(() => <div />);
      `,
      settings: { componentWrapperFunctions: ['myMemo'] },
    },
    // ---- settings.componentWrapperFunctions (object form) — same as
    // above but pragma-qualified. ----
    {
      code: `
        const TopLevel = MyLib.observer(() => <div />);
      `,
      settings: {
        componentWrapperFunctions: [{ property: 'observer', object: 'MyLib' }],
      },
    },
    {
      code: `
        function ParentComponent() {
          return <Table rowRenderer={(rowData) => <Row data={data} />} />
        }
      `,
      options: [{ propNamePattern: '*Renderer' }],
    },
    {
      code: `
        function ParentComponent() {
          return (
            <SomeComponent>
              {
                thing.match({
                  renderLoading: () => <div />,
                  renderSuccess: () => <div />,
                  renderFailure: () => <div />,
                })
              }
            </SomeComponent>
          )
        }
      `,
    },
    {
      code: `
        function createTestComponent(props) {
          return (
            <div />
          );
        }
      `,
    },
    {
      code: `
        function ParentComponent() {
          function getComponent() {
            return <div />;
          }
          return (
            <div>
              {getComponent()}
            </div>
          );
        }
      `,
    },
  ],
  invalid: [
    {
      code: `
        function ParentComponent() {
          function UnstableNestedFunctionComponent() {
            return <div />;
          }
          return (
            <div>
              <UnstableNestedFunctionComponent />
            </div>
          );
        }
      `,
      errors: [{ message: ERROR_MESSAGE }],
    },
    {
      code: `
        function ParentComponent() {
          const UnstableNestedVariableComponent = () => {
            return <div />;
          }
          return (
            <div>
              <UnstableNestedVariableComponent />
            </div>
          );
        }
      `,
      errors: [{ message: ERROR_MESSAGE }],
    },
    {
      code: `
        export default () => {
          function UnstableNestedFunctionComponent() {
            return <div />;
          }
          return (
            <div>
              <UnstableNestedFunctionComponent />
            </div>
          );
        }
      `,
      errors: [{ message: ERROR_MESSAGE_WITHOUT_NAME }],
    },
    {
      code: `
        function ParentComponent() {
          class UnstableNestedClassComponent extends React.Component {
            render() {
              return <div />;
            }
          };
          return (
            <div>
              <UnstableNestedClassComponent />
            </div>
          );
        }
      `,
      errors: [{ message: ERROR_MESSAGE }],
    },
    {
      code: `
        class ParentComponent extends React.Component {
          render() {
            class UnstableNestedClassComponent extends React.Component {
              render() {
                return <div />;
              }
            };
            return (
              <div>
                <UnstableNestedClassComponent />
              </div>
            );
          }
        }
      `,
      errors: [{ message: ERROR_MESSAGE }],
    },
    {
      code: `
        function ComponentWithProps(props) {
          return <div />;
        }

        function ParentComponent() {
          return (
            <ComponentWithProps footer={() => <div />} />
          );
        }
      `,
      errors: [{ message: ERROR_MESSAGE_COMPONENT_AS_PROPS }],
    },
    {
      code: `
        function ComponentForProps(props) {
          return <div />;
        }

        function ParentComponent() {
          return (
            <ComponentForProps notPrefixedWithRender={() => <div />} />
          );
        }
      `,
      errors: [{ message: ERROR_MESSAGE_COMPONENT_AS_PROPS }],
    },
    {
      code: `
        function ParentComponent() {
          return (
            <ComponentForProps someMap={{ Header: () => <div /> }} />
          );
        }
      `,
      errors: [{ message: ERROR_MESSAGE_COMPONENT_AS_PROPS }],
    },
    // ---- settings.react.pragma override — Preact.createElement
    // recognized as JSX-returning, so the inner function is detected as
    // a stateless component and its enclosing ParentComponent is named
    // in the diagnostic. ----
    {
      code: `
        function ParentComponent() {
          function UnstableNestedComponent() {
            return Preact.createElement("div", null);
          }
          return <UnstableNestedComponent />;
        }
      `,
      settings: { react: { pragma: 'Preact' } },
      errors: [{ message: ERROR_MESSAGE }],
    },
    // ---- settings.componentWrapperFunctions extends defaults — a
    // user-declared `myMemo(fn)` wrapper inside a component now flags. ----
    {
      code: `
        function ParentComponent() {
          const Unstable = myMemo(() => <div />);
          return <Unstable />;
        }
      `,
      settings: { componentWrapperFunctions: ['myMemo'] },
      errors: [{ message: ERROR_MESSAGE }],
    },
    // ---- settings.componentWrapperFunctions object form (pragma-qualified). ----
    {
      code: `
        function ParentComponent() {
          const Unstable = MyLib.observer(() => <div />);
          return <Unstable />;
        }
      `,
      settings: {
        componentWrapperFunctions: [{ property: 'observer', object: 'MyLib' }],
      },
      errors: [{ message: ERROR_MESSAGE }],
    },
    {
      code: `
        function ParentComponent() {
          const UnstableNestedComponent = React.memo(() => {
            return <div />;
          });
          return (
            <div>
              <UnstableNestedComponent />
            </div>
          );
        }
      `,
      errors: [{ message: ERROR_MESSAGE }],
    },
    {
      code: `
        function ParentComponent() {
          return (
            <SomeComponent>
              {
                thing.match({
                  loading: () => <div />,
                  success: () => <div />,
                  failure: () => <div />,
                })
              }
            </SomeComponent>
          )
        }
      `,
      errors: [
        { message: ERROR_MESSAGE_COMPONENT_AS_PROPS },
        { message: ERROR_MESSAGE_COMPONENT_AS_PROPS },
        { message: ERROR_MESSAGE_COMPONENT_AS_PROPS },
      ],
    },
  ],
});
