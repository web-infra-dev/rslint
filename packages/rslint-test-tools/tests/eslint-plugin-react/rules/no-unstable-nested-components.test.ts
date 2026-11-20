import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: '',
    message:
      'Do not define components during render. React will see a new component type on every render and destroy the entire subtree’s DOM nodes and state (https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, move this component definition out of the parent component “ParentComponent” and pass data as props.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: '',
    message:
      'Do not define components during render. React will see a new component type on every render and destroy the entire subtree’s DOM nodes and state (https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, move this component definition out of the parent component and pass data as props.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: '',
    message:
      'Do not define components during render. React will see a new component type on every render and destroy the entire subtree’s DOM nodes and state (https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, move this component definition out of the parent component “ParentComponent” and pass data as props. If you want to allow component creation in props, set allowAsProps option to true.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

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

ruleTester.run('no-unstable-nested-components', {
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

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        function ParentComponent() {
          return React.createElement(
            "div",
            null,
            React.createElement(OutsideDefinedFunctionComponent, null)
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        function ParentComponent() {
          return React.createElement(SomeComponent, {
            footer: React.createElement(OutsideDefinedComponent, null),
            header: React.createElement("div", null)
          });
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        function ParentComponent() {
          const MemoizedNestedComponent = React.useCallback(() => <div />, []);

          return (
            <div>
              <MemoizedNestedComponent />
            </div>
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        function ParentComponent() {
          const MemoizedNestedComponent = React.useCallback(
            () => React.createElement("div", null),
            []
          );

          return React.createElement(
            "div",
            null,
            React.createElement(MemoizedNestedComponent, null)
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        function ParentComponent() {
          const MemoizedNestedFunctionComponent = React.useCallback(
            function () {
              return <div />;
            },
            []
          );

          return (
            <div>
              <MemoizedNestedFunctionComponent />
            </div>
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        function ParentComponent() {
          const MemoizedNestedFunctionComponent = React.useCallback(
            function () {
              return React.createElement("div", null);
            },
            []
          );

          return React.createElement(
            "div",
            null,
            React.createElement(MemoizedNestedFunctionComponent, null)
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        function ParentComponent(props) {
          // Should not interfere handler declarations
          function onClick(event) {
            props.onClick(event.target.value);
          }

          const onKeyPress = () => null;

          function getOnHover() {
            return function onHover(event) {
              props.onHover(event.target);
            }
          }

          return (
            <div>
              <button
                onClick={onClick}
                onKeyPress={onKeyPress}
                onHover={getOnHover()}

                // These should not be considered as components
                maybeComponentOrHandlerNull={() => null}
                maybeComponentOrHandlerUndefined={() => undefined}
                maybeComponentOrHandlerBlank={() => ''}
                maybeComponentOrHandlerString={() => 'hello-world'}
                maybeComponentOrHandlerNumber={() => 42}
                maybeComponentOrHandlerArray={() => []}
                maybeComponentOrHandlerObject={() => {}} />
            </div>
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        function ParentComponent() {
          function getComponent() {
            return React.createElement("div", null);
          }

          return React.createElement("div", null, getComponent());
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        function ParentComponent() {
            return (
              <RenderPropComponent>
                {() => <div />}
              </RenderPropComponent>
            );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        function ParentComponent() {
            return (
              <RenderPropComponent children={() => <div />} />
            );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        function ParentComponent() {
          return (
            <ComplexRenderPropComponent
              listRenderer={data.map((items, index) => (
                <ul>
                  {items[index].map((item) =>
                    <li>
                      {item}
                    </li>
                  )}
                </ul>
              ))
              }
            />
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        function ParentComponent() {
          return React.createElement(
              RenderPropComponent,
              null,
              () => React.createElement("div", null)
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        function ParentComponent(props) {
          return (
            <List items={props.items.map(item => {
              return (
                <li key={item.id}>
                  {item.name}
                </li>
              );
            })}
            />
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        function ParentComponent(props) {
          return React.createElement(
            "ul",
            null,
            props.items.map(() =>
              React.createElement(
                "li",
                { key: item.id },
                item.name
              )
            )
          )
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        function ParentComponent(props) {
          return (
            <ul>
              {props.items.map(function Item(item) {
                return (
                  <li key={item.id}>
                    {item.name}
                  </li>
                );
              })}
            </ul>
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        function ParentComponent(props) {
          return React.createElement(
            "ul",
            null,
            props.items.map(function Item() {
              return React.createElement(
                "li",
                { key: item.id },
                item.name
              );
            })
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        function createTestComponent(props) {
          return React.createElement("div", null);
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        function ParentComponent() {
          return React.createElement(ComponentWithProps, {
            footer: () => React.createElement("div", null)
          });
        }
      `,
      options: [{ allowAsProps: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        function ParentComponent() {
          return (
            <SomeComponent item={{ children: () => <div /> }} />
          )
        }
      `,
      options: [{ allowAsProps: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
      function ParentComponent() {
        const thingElement = thing.match({
          renderLoading: () => <div />,
          renderSuccess: () => <div />,
          renderFailure: () => <div />,
        });
        return (
          <SomeComponent>
            {thingElement}
          </SomeComponent>
        )
      }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
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
      options: [{ allowAsProps: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
      function ParentComponent() {
        const thingElement = thing.match({
          loading: () => <div />,
          success: () => <div />,
          failure: () => <div />,
        });
        return (
          <SomeComponent>
            {thingElement}
          </SomeComponent>
        )
      }
      `,
      options: [{ allowAsProps: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        function ParentComponent() {
          return React.createElement(ComponentForProps, {
            renderFooter: () => React.createElement("div", null)
          });
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        function ParentComponent() {
          useEffect(() => {
            return () => null;
          });

          return <div />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        function ParentComponent() {
          return (
            <SomeComponent renderMenu={() => (
              <RenderPropComponent>
                {items.map(item => (
                  <li key={item}>{item}</li>
                ))}
              </RenderPropComponent>
            )} />
          )
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
        const ParentComponent = () => (
          <SomeComponent
            components={[
              <ul>
                {list.map(item => (
                  <li key={item}>{item}</li>
                ))}
              </ul>,
            ]}
          />
        );
     `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
        function ParentComponent() {
          const rows = [
            {
              name: 'A',
              render: (props) => <Row {...props} />
            },
          ];

          return <Table rows={rows} />;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
      code: `
        const ParentComponent = createReactClass({
          displayName: "ParentComponent",
          statics: {
            getSnapshotBeforeUpdate: function () {
              return null;
            },
          },
          render() {
            return <div />;
          },
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
      code: `
        function ParentComponent() {
          const rows = [
            {
              name: 'A',
              notPrefixedWithRender: (props) => <Row {...props} />
            },
          ];

          return <Table rows={rows} />;
        }
      `,
      options: [{ allowAsProps: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 39 (7.37.5)',
      code: `
        function ParentComponent() {
          return <Table
            rowRenderer={(rowData) => <Row data={data} />}
          />
        }
      `,
      options: [{ propNamePattern: '*Renderer' }],
      languageOptions: upstreamModuleOptions,
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
      errors: [expectError1([3, 11, 5, 12])],
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
      errors: [expectError1([3, 51, 5, 12])],
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
      errors: [expectError2([3, 11, 5, 12])],
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
      errors: [expectError1([3, 11, 7, 12])],
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
      errors: [expectError1([4, 13, 8, 14])],
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
      errors: [expectError3([8, 41, 8, 54])],
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
      errors: [expectError3([8, 55, 8, 68])],
    },
    {
      code: `
        function ParentComponent() {
          return (
            <ComponentForProps someMap={{ Header: () => <div /> }} />
          );
        }
      `,
      errors: [expectError3([4, 51, 4, 64])],
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
      errors: [expectError1([3, 11, 5, 12])],
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
      errors: [expectError1([3, 28, 3, 49])],
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
      errors: [expectError1([3, 28, 3, 57])],
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
      errors: [expectError1([3, 43, 5, 13])],
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
        expectError3([7, 28, 7, 41]),
        expectError3([8, 28, 8, 41]),
        expectError3([9, 28, 9, 41]),
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
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
      errors: [expectError1([3, 11, 5, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        function ParentComponent() {
          function UnstableNestedFunctionComponent() {
            return React.createElement("div", null);
          }

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedFunctionComponent, null)
          );
        }
      `,
      errors: [expectError1([3, 11, 5, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
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
      errors: [expectError1([3, 51, 5, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        function ParentComponent() {
          const UnstableNestedVariableComponent = () => {
            return React.createElement("div", null);
          }

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedVariableComponent, null)
          );
        }
      `,
      errors: [expectError1([3, 51, 5, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        const ParentComponent = () => {
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
      errors: [expectError1([3, 11, 5, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        const ParentComponent = () => {
          function UnstableNestedFunctionComponent() {
            return React.createElement("div", null);
          }

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedFunctionComponent, null)
          );
        }
      `,
      errors: [expectError1([3, 11, 5, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
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
      errors: [expectError2([3, 11, 5, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        export default () => {
          function UnstableNestedFunctionComponent() {
            return React.createElement("div", null);
          }

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedFunctionComponent, null)
          );
        };
      `,
      errors: [expectError2([3, 11, 5, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        const ParentComponent = () => {
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
      errors: [expectError1([3, 51, 5, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        const ParentComponent = () => {
          const UnstableNestedVariableComponent = () => {
            return React.createElement("div", null);
          }

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedVariableComponent, null)
          );
        }
      `,
      errors: [expectError1([3, 51, 5, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
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
      errors: [expectError1([3, 11, 7, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        function ParentComponent() {
          class UnstableNestedClassComponent extends React.Component {
            render() {
              return React.createElement("div", null);
            }
          }

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedClassComponent, null)
          );
        }
      `,
      errors: [expectError1([3, 11, 7, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
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
      errors: [expectError1([4, 13, 8, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
        class ParentComponent extends React.Component {
          render() {
            class UnstableNestedClassComponent extends React.Component {
              render() {
                return React.createElement("div", null);
              }
            }

            return React.createElement(
              "div",
              null,
              React.createElement(UnstableNestedClassComponent, null)
            );
          }
        }
      `,
      errors: [expectError1([4, 13, 8, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        class ParentComponent extends React.Component {
          render() {
            function UnstableNestedFunctionComponent() {
              return <div />;
            }

            return (
              <div>
                <UnstableNestedFunctionComponent />
              </div>
            );
          }
        }
      `,
      errors: [expectError1([4, 13, 6, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        class ParentComponent extends React.Component {
          render() {
            function UnstableNestedClassComponent() {
              return React.createElement("div", null);
            }

            return React.createElement(
              "div",
              null,
              React.createElement(UnstableNestedClassComponent, null)
            );
          }
        }
      `,
      errors: [expectError1([4, 13, 6, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        class ParentComponent extends React.Component {
          render() {
            const UnstableNestedVariableComponent = () => {
              return <div />;
            }

            return (
              <div>
                <UnstableNestedVariableComponent />
              </div>
            );
          }
        }
      `,
      errors: [expectError1([4, 53, 6, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
        class ParentComponent extends React.Component {
          render() {
            const UnstableNestedClassComponent = () => {
              return React.createElement("div", null);
            }

            return React.createElement(
              "div",
              null,
              React.createElement(UnstableNestedClassComponent, null)
            );
          }
        }
      `,
      errors: [expectError1([4, 50, 6, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        function ParentComponent() {
          function getComponent() {
            function NestedUnstableFunctionComponent() {
              return <div />;
            };

            return <NestedUnstableFunctionComponent />;
          }

          return (
            <div>
              {getComponent()}
            </div>
          );
        }
      `,
      errors: [expectError1([4, 13, 6, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
        function ParentComponent() {
          function getComponent() {
            function NestedUnstableFunctionComponent() {
              return React.createElement("div", null);
            }

            return React.createElement(NestedUnstableFunctionComponent, null);
          }

          return React.createElement("div", null, getComponent());
        }
      `,
      errors: [expectError1([4, 13, 6, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        function ComponentWithProps(props) {
          return <div />;
        }

        function ParentComponent() {
          return (
            <ComponentWithProps
              footer={
                function SomeFooter() {
                  return <div />;
                }
              } />
          );
        }
      `,
      errors: [expectError3([10, 17, 12, 18])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
        function ComponentWithProps(props) {
          return React.createElement("div", null);
        }

        function ParentComponent() {
          return React.createElement(ComponentWithProps, {
            footer: function SomeFooter() {
              return React.createElement("div", null);
            }
          });
        }
      `,
      errors: [expectError3([8, 21, 10, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
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
      errors: [expectError3([8, 43, 8, 56])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        function ComponentWithProps(props) {
          return React.createElement("div", null);
        }

        function ParentComponent() {
          return React.createElement(ComponentWithProps, {
            footer: () => React.createElement("div", null)
          });
        }
      `,
      errors: [expectError3([8, 21, 8, 59])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
        function ParentComponent() {
            return (
              <RenderPropComponent>
                {() => {
                  function UnstableNestedComponent() {
                    return <div />;
                  }

                  return (
                    <div>
                      <UnstableNestedComponent />
                    </div>
                  );
                }}
              </RenderPropComponent>
            );
        }
      `,
      errors: [expectError1([6, 19, 8, 20])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
        function RenderPropComponent(props) {
          return props.render({});
        }

        function ParentComponent() {
          return React.createElement(
            RenderPropComponent,
            null,
            () => {
              function UnstableNestedComponent() {
                return React.createElement("div", null);
              }

              return React.createElement(
                "div",
                null,
                React.createElement(UnstableNestedComponent, null)
              );
            }
          );
        }
      `,
      errors: [expectError1([11, 15, 13, 16])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
        function ComponentForProps(props) {
          return React.createElement("div", null);
        }

        function ParentComponent() {
          return React.createElement(ComponentForProps, {
            notPrefixedWithRender: () => React.createElement("div", null)
          });
        }
      `,
      errors: [expectError3([8, 36, 8, 74])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `
        class ParentComponent extends React.Component {
          render() {
            const List = (props) => {
              const items = props.items
                .map((item) => (
                  <li key={item.key}>
                    <span>{item.name}</span>
                  </li>
                ));

              return <ul>{items}</ul>;
            };

            return <List {...this.props} />;
          }
        }
      `,
      errors: [expectError1([4, 26, 13, 14])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
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
        expectError3([7, 26, 7, 39]),
        expectError3([8, 26, 8, 39]),
        expectError3([9, 26, 9, 39]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `
      function ParentComponent() {
        const thingElement = thing.match({
          loading: () => <div />,
          success: () => <div />,
          failure: () => <div />,
        });
        return (
          <SomeComponent>
            {thingElement}
          </SomeComponent>
        )
      }
      `,
      errors: [
        expectError3([4, 20, 4, 33]),
        expectError3([5, 20, 5, 33]),
        expectError3([6, 20, 6, 33]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `
      function ParentComponent() {
        const rows = [
          {
            name: 'A',
            notPrefixedWithRender: (props) => <Row {...props} />
          },
        ];

        return <Table rows={rows} />;
      }
      `,
      errors: [expectError3([6, 36, 6, 65])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
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
      errors: [expectError1([3, 43, 5, 13])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
        function ParentComponent() {
          const UnstableNestedComponent = React.memo(
            () => React.createElement("div", null),
          );

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedComponent, null)
          );
        }
      `,
      errors: [expectError1([3, 43, 5, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
        function ParentComponent() {
          const UnstableNestedComponent = React.memo(
            function () {
              return <div />;
            }
          );

          return (
            <div>
              <UnstableNestedComponent />
            </div>
          );
        }
      `,
      errors: [expectError1([3, 43, 7, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `
        function ParentComponent() {
          const UnstableNestedComponent = React.memo(
            function () {
              return React.createElement("div", null);
            }
          );

          return React.createElement(
            "div",
            null,
            React.createElement(UnstableNestedComponent, null)
          );
        }
      `,
      errors: [expectError1([3, 43, 7, 12])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
