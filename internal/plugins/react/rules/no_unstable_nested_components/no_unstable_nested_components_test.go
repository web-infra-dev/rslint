package no_unstable_nested_components

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Message constants mirror eslint-plugin-react v7.37.x byte-for-byte,
// including the typographic apostrophe (U+2019) in "subtree's" and the
// typographic double quotation marks (U+201C / U+201D) around the parent
// name. Verified against upstream installed source.
const (
	errorMessage                 = "Do not define components during render. React will see a new component type on every render and destroy the entire subtree’s DOM nodes and state (https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, move this component definition out of the parent component “ParentComponent” and pass data as props."
	errorMessageWithoutName      = "Do not define components during render. React will see a new component type on every render and destroy the entire subtree’s DOM nodes and state (https://reactjs.org/docs/reconciliation.html#elements-of-different-types). Instead, move this component definition out of the parent component and pass data as props."
	errorMessageComponentAsProps = errorMessage + " If you want to allow component creation in props, set allowAsProps option to true."
)

func TestNoUnstableNestedComponentsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnstableNestedComponentsRule, []rule_tester.ValidTestCase{
		// ---- Upstream: outside-defined components used in JSX ----
		{Code: `
        function ParentComponent() {
          return (
            <div>
              <OutsideDefinedFunctionComponent />
            </div>
          );
        }
      `, Tsx: true},
		{Code: `
        function ParentComponent() {
          return React.createElement(
            "div",
            null,
            React.createElement(OutsideDefinedFunctionComponent, null)
          );
        }
      `, Tsx: true},
		{Code: `
        function ParentComponent() {
          return (
            <SomeComponent
              footer={<OutsideDefinedComponent />}
              header={<div />}
              />
          );
        }
      `, Tsx: true},
		{Code: `
        function ParentComponent() {
          return React.createElement(SomeComponent, {
            footer: React.createElement(OutsideDefinedComponent, null),
            header: React.createElement("div", null)
          });
        }
      `, Tsx: true},

		// ---- Upstream: React.useCallback false-negatives (intentionally not flagged) ----
		{Code: `
        function ParentComponent() {
          const MemoizedNestedComponent = React.useCallback(() => <div />, []);
          return (
            <div>
              <MemoizedNestedComponent />
            </div>
          );
        }
      `, Tsx: true},
		{Code: `
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
      `, Tsx: true},
		{Code: `
        function ParentComponent() {
          const MemoizedNestedFunctionComponent = React.useCallback(
            function () { return <div />; },
            []
          );
          return (
            <div>
              <MemoizedNestedFunctionComponent />
            </div>
          );
        }
      `, Tsx: true},
		{Code: `
        function ParentComponent() {
          const MemoizedNestedFunctionComponent = React.useCallback(
            function () { return React.createElement("div", null); },
            []
          );
          return React.createElement(
            "div",
            null,
            React.createElement(MemoizedNestedFunctionComponent, null)
          );
        }
      `, Tsx: true},

		// ---- Upstream: handler functions that don't return JSX are not components ----
		{Code: `
        function ParentComponent(props) {
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
      `, Tsx: true},

		// ---- Upstream: lowercase "factory" functions are not components ----
		{Code: `
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
      `, Tsx: true},
		{Code: `
        function ParentComponent() {
          function getComponent() {
            return React.createElement("div", null);
          }
          return React.createElement("div", null, getComponent());
        }
      `, Tsx: true},

		// ---- Upstream: render-prop-as-children patterns ----
		{Code: `
        function ParentComponent() {
            return (
              <RenderPropComponent>
                {() => <div />}
              </RenderPropComponent>
            );
        }
      `, Tsx: true},
		{Code: `
        function ParentComponent() {
            return (
              <RenderPropComponent children={() => <div />} />
            );
        }
      `, Tsx: true},
		{Code: `
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
      `, Tsx: true},
		{Code: `
        function ParentComponent() {
          return React.createElement(
              RenderPropComponent,
              null,
              () => React.createElement("div", null)
          );
        }
      `, Tsx: true},

		// ---- Upstream: Array#map callbacks are not components ----
		{Code: `
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
      `, Tsx: true},
		{Code: `
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
      `, Tsx: true},
		{Code: `
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
      `, Tsx: true},
		{Code: `
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
      `, Tsx: true},
		{Code: `
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
      `, Tsx: true},

		// ---- Upstream: lowercase factory top-level functions are not components ----
		{Code: `
        function createTestComponent(props) {
          return (
            <div />
          );
        }
      `, Tsx: true},
		{Code: `
        function createTestComponent(props) {
          return React.createElement("div", null);
        }
      `, Tsx: true},

		// ---- Upstream: allowAsProps option ----
		{
			Code: `
        function ParentComponent() {
          return (
            <ComponentWithProps footer={() => <div />} />
          );
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"allowAsProps": true},
		},
		{
			Code: `
        function ParentComponent() {
          return React.createElement(ComponentWithProps, {
            footer: () => React.createElement("div", null)
          });
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"allowAsProps": true},
		},
		{
			Code: `
        function ParentComponent() {
          return (
            <SomeComponent item={{ children: () => <div /> }} />
          )
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"allowAsProps": true},
		},

		// ---- Upstream: render-prop pattern keys inside nested objects ----
		{Code: `
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
      `, Tsx: true},
		{Code: `
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
      `, Tsx: true},
		{
			Code: `
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
			Tsx:     true,
			Options: map[string]interface{}{"allowAsProps": true},
		},
		{
			Code: `
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
			Tsx:     true,
			Options: map[string]interface{}{"allowAsProps": true},
		},

		// ---- Upstream: renderX attribute (default pattern) ----
		{Code: `
        function ParentComponent() {
          return (
            <ComponentForProps renderFooter={() => <div />} />
          );
        }
      `, Tsx: true},
		{Code: `
        function ParentComponent() {
          return React.createElement(ComponentForProps, {
            renderFooter: () => React.createElement("div", null)
          });
        }
      `, Tsx: true},

		// ---- Upstream: useEffect cleanup callback is not a component ----
		{Code: `
        function ParentComponent() {
          useEffect(() => {
            return () => null;
          });
          return <div />;
        }
      `, Tsx: true},

		// ---- Upstream: `renderers` nested object — prop name matches pattern ----
		{Code: `
        function ParentComponent() {
          return (
            <SomeComponent renderers={{ Header: () => <div /> }} />
          )
        }
      `, Tsx: true},

		// ---- Upstream: nested render-prop with inner map ----
		{Code: `
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
      `, Tsx: true},

		// ---- Upstream: arrow inside array literal of a JSX attribute ----
		{Code: `
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
     `, Tsx: true},

		// ---- Upstream: direct value of render: key in a plain object ----
		{Code: `
        function ParentComponent() {
          const rows = [
            {
              name: 'A',
              render: (props) => <Row {...props} />
            },
          ];
          return <Table rows={rows} />;
        }
      `, Tsx: true},

		// ---- Upstream: arrow returning null whose property key doesn't match pattern ----
		{Code: `
        function ParentComponent() {
          return <SomeComponent renderers={{ notComponent: () => null }} />;
        }
      `, Tsx: true},

		// ---- Upstream: createReactClass statics field ----
		{Code: `
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
      `, Tsx: true},

		// ---- Upstream: allowAsProps with non-render prefix ----
		{
			Code: `
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
			Tsx:     true,
			Options: map[string]interface{}{"allowAsProps": true},
		},

		// ---- Upstream: propNamePattern override ----
		{
			Code: `
        function ParentComponent() {
          return <Table
            rowRenderer={(rowData) => <Row data={data} />}
          />
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"propNamePattern": "*Renderer"},
		},

		// =======================================================
		// Additional valid edge-case coverage (tsgo-specific) —
		// failures here clearly point at over-reporting rather than
		// upstream-aligned behavior.
		// =======================================================

		// ---- Optional chain `.map`: `items?.map(...)` is still a map call ----
		{Code: `
        function ParentComponent(props) {
          return (
            <ul>
              {props.items?.map(item => (<li key={item.id}>{item.name}</li>))}
            </ul>
          );
        }
      `, Tsx: true},

		// ---- Computed property key — anonymous arrow under a computed key
		// is NOT a render-prop pattern and the function isn't a stateless
		// component (computed keys go through Branch 9 + null-only check) ----
		{Code: `
        function ParentComponent() {
          const key = "Foo";
          return <Comp config={{ [key]: () => null }} />;
        }
      `, Tsx: true},

		// ---- Numeric / string-literal property key — neither matches the
		// render-prop glob; the inner arrow is not classified as a component ----
		{Code: `
        function ParentComponent() {
          return <Comp config={{ 0: () => null, "render-foo": () => null }} />;
        }
      `, Tsx: true},

		// ---- Class method returning JSX must not double-report:
		// outer class is the only acceptable diagnostic site ----
		{Code: `
        class StableTopLevel extends React.Component {
          render() { return <div />; }
        }
      `, Tsx: true},

		// ---- `useMemo(() => <div/>, [])` — hook callback in allowed
		// position (VariableDeclaration) but NOT a memo/forwardRef
		// wrapper, so the inner arrow isn't classified as a component ----
		{Code: `
        function ParentComponent() {
          const ui = useMemo(() => <div />, []);
          return ui;
        }
      `, Tsx: true},

		// ---- `useCallback` cleanup pattern — the returned arrow is a hook
		// return statement (skip via isReturnStatementOfHook) ----
		{Code: `
        function ParentComponent() {
          useEffect(() => {
            return function CleanupComp() { return <div />; };
          }, []);
          return <div />;
        }
      `, Tsx: true},

		// ---- Lowercase parent — even deeply nested capitalized children
		// inside a lowercase factory must NOT report (parent name lowercase
		// short-circuit) ----
		{Code: `
        function makeView(props) {
          function NestedInsideFactory() { return <div />; }
          return <NestedInsideFactory />;
        }
      `, Tsx: true},

		// ---- TS-as wrapper around createElement callee — pragma still
		// recognized after unwrap ----
		{Code: `
        function ParentComponent() {
          return ((React as any).createElement)("div", null);
        }
      `, Tsx: true},

		// ---- Outer paren wrapper around the JSX element / fragment ----
		{Code: `
        const ParentComponent = () => ((<div />));
      `, Tsx: true},

		// ---- Generator function nested in component — generators do not
		// classify as React components (they yield, not return JSX) so the
		// inner generator must not be reported. ----
		{Code: `
        function ParentComponent() {
          function* helperGen() {
            yield 1;
            yield 2;
          }
          return <div>{helperGen()}</div>;
        }
      `, Tsx: true},

		// ---- Inner abstract class declaration — abstract methods are
		// body-absent; the class itself doesn't extend React.Component so
		// it's not a component. ----
		{Code: `
        function ParentComponent() {
          abstract class Helper {
            abstract serialize(): string;
          }
          return <div />;
        }
      `, Tsx: true},

		// ---- Top-level React.memo — no parent component, so even though
		// the arrow looks like a stateless component it has no enclosing
		// component to attach to. ----
		{Code: `
        const TopLevel = React.memo(() => <div />);
      `, Tsx: true},

		// ---- Arrow with block body returning early `null` only — strict
		// isReturningJSX (ignoreNull=true) treats this as not-a-component
		// when it's used as a property value, so should not be flagged
		// even when the property key matches the render-prop pattern. ----
		{Code: `
        function ParentComponent() {
          return (
            <Comp renderEmpty={() => { return null; }} />
          );
        }
      `, Tsx: true},

		// ---- Bare `memo(fn)` without importing `memo` from the pragma
		// module — upstream skips this because `isPragmaComponentWrapper`
		// requires the binding to come from `react`, and our
		// `IsDestructuredFromPragmaImport` enforces the same gate. The
		// matching invalid form (with `import { memo } from 'react'`)
		// lives in the invalid suite. ----
		{Code: `
        function ParentComponent() {
          const UnstableMemo = memo(() => <div />);
          return <UnstableMemo />;
        }
      `, Tsx: true},

		// ---- Bare `memo(fn)` with `memo` imported from a non-pragma
		// module (e.g. `preact`) — name collides with React's `memo` but
		// upstream's import resolution rejects it, so we must too. ----
		{Code: `
        import { memo } from 'preact';
        function ParentComponent() {
          const UnstableMemo = memo(() => <div />);
          return <UnstableMemo />;
        }
      `, Tsx: true},

		// ---- Aliased import (`memo as m`) — upstream recognizes
		// wrappers by the local name, so `m(...)` does NOT match the
		// hardcoded wrapper list and the call is not treated as a
		// component. Locks in name-based matching, not symbol-based. ----
		{Code: `
        import { memo as m } from 'react';
        function ParentComponent() {
          const Aliased = m(() => <div />);
          return <Aliased />;
        }
      `, Tsx: true},

		// ---- Bare `createElement(...)` without importing it from
		// the pragma module — upstream's `isCreateElement` requires the
		// binding to resolve to React; without the import this is treated
		// as a regular function call and the helper is not flagged. ----
		{Code: `
        function ParentComponent() {
          function NotAComponent() { return createElement('div'); }
          return <NotAComponent />;
        }
      `, Tsx: true},

		// ---- React.memo wrapping a sibling/outer arrow component
		// (`nodeWrapsComponent` gate) — upstream's
		// `isPragmaComponentWrapper` short-circuits to false for the
		// MemberExpression form when the wrapped function returns JSX
		// whose root tag matches an already-detected sibling/outer
		// component, so the call is NOT registered. Locks in three
		// declaration shapes (arrow / class / function) and the bare /
		// import-aware variants. ----
		{Code: `
        const Inner = () => <div />;
        function Parent() {
          const Wrap = React.memo(() => <Inner />);
          return <Wrap />;
        }
      `, Tsx: true},
		{Code: `
        const Inner = () => <div />;
        function Parent() {
          const Wrap = React.memo(() => { return <Inner />; });
          return <Wrap />;
        }
      `, Tsx: true},
		{Code: `
        const Inner = () => <div />;
        function Parent() {
          const Wrap = React.memo(Inner);
          return <Wrap />;
        }
      `, Tsx: true},
		{Code: `
        class InnerCls extends React.Component { render() { return <div />; } }
        function Parent() {
          const Wrap = React.memo(() => <InnerCls />);
          return <Wrap />;
        }
      `, Tsx: true},
		{Code: `
        const Inner = () => <div />;
        function Parent() {
          const Wrap = React.forwardRef(() => <Inner />);
          return <Wrap />;
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: function declaration nested in function component ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},

		// ---- Upstream: arrow assigned to capitalized binding nested in function component ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 51},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 51},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},

		// ---- Upstream: anonymous default-exported arrow ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageWithoutName, Line: 3, Column: 11},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageWithoutName, Line: 3, Column: 11},
			},
		},

		// ---- Upstream: nested arrow assigned to capitalized binding (arrow parent) ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 51},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 51},
			},
		},

		// ---- Upstream: class nested in function component ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},

		// ---- Upstream: class nested in class component (via render method) ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 4, Column: 13},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 4, Column: 13},
			},
		},

		// ---- Upstream: function nested in class component render ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 4, Column: 13},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 4, Column: 13},
			},
		},

		// ---- Upstream: arrow assigned to capitalized binding inside class render ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 4, Column: 53},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 4, Column: 50},
			},
		},

		// ---- Upstream: nested function inside a nested getter-function ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 4, Column: 13},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 4, Column: 13},
			},
		},

		// ---- Upstream: function declaration as prop value (componentAsProps) ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageComponentAsProps, Line: 10, Column: 17},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageComponentAsProps, Line: 8, Column: 21},
			},
		},
		{
			Code: `
        function ComponentWithProps(props) {
          return <div />;
        }

        function ParentComponent() {
            return (
              <ComponentWithProps footer={() => <div />} />
            );
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageComponentAsProps, Line: 8, Column: 43},
			},
		},
		{
			Code: `
        function ComponentWithProps(props) {
          return React.createElement("div", null);
        }

        function ParentComponent() {
          return React.createElement(ComponentWithProps, {
            footer: () => React.createElement("div", null)
          });
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageComponentAsProps, Line: 8, Column: 21},
			},
		},

		// ---- Upstream: render-prop children with nested component ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 6, Column: 19},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 11, Column: 15},
			},
		},

		// ---- Upstream: non-render-prefixed attribute (componentAsProps) ----
		{
			Code: `
        function ComponentForProps(props) {
          return <div />;
        }

        function ParentComponent() {
          return (
            <ComponentForProps notPrefixedWithRender={() => <div />} />
          );
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageComponentAsProps, Line: 8, Column: 55},
			},
		},
		{
			Code: `
        function ComponentForProps(props) {
          return React.createElement("div", null);
        }

        function ParentComponent() {
          return React.createElement(ComponentForProps, {
            notPrefixedWithRender: () => React.createElement("div", null)
          });
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageComponentAsProps, Line: 8, Column: 36},
			},
		},

		// ---- Upstream: nested object with capitalized key under non-render attribute ----
		{
			Code: `
        function ParentComponent() {
          return (
            <ComponentForProps someMap={{ Header: () => <div /> }} />
          );
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageComponentAsProps, Line: 4, Column: 51},
			},
		},

		// ---- Upstream: single-error sanity check — class with nested List arrow ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 4, Column: 26},
			},
		},

		// ---- Upstream: nested thing.match with lowercase keys — 3 errors ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// Three distinct property values, each at the start of its
				// arrow expression. Column points at `() => …` after the
				// `<key>: ` prefix, so loading=26, success=26, failure=26;
				// lines progress by one each.
				{Message: errorMessageComponentAsProps, Line: 7, Column: 26},
				{Message: errorMessageComponentAsProps, Line: 8, Column: 26},
				{Message: errorMessageComponentAsProps, Line: 9, Column: 26},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageComponentAsProps, Line: 4, Column: 20},
				{Message: errorMessageComponentAsProps, Line: 5, Column: 20},
				{Message: errorMessageComponentAsProps, Line: 6, Column: 20},
			},
		},

		// ---- Upstream: rows in array with notPrefixedWithRender key ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageComponentAsProps, Line: 6, Column: 36},
			},
		},

		// ---- Upstream: React.memo wrapper patterns (arrow / function) ----
		// Wrapper-call diagnostic Pos lives at the START of the wrapper
		// CallExpression (`React.memo`), matching upstream's
		// `components.add(call, 2)` registration in Components.detect.
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 43},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 43},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 43},
			},
		},
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 43},
			},
		},

		// ============================================================
		// tsgo-specific edge cases beyond the upstream test suite
		// ============================================================

		// ---- TS wrapper: `(X as any)` callee on createElement should still
		// classify the inner arrow as JSX-returning ----
		{
			Code: `
        function ParentComponent() {
          function UnstableNestedComponent() {
            return (React as any).createElement("div", null);
          }
          return <UnstableNestedComponent />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},

		// ---- TS wrapper: `<div/> satisfies JSX.Element` return ----
		{
			Code: `
        function ParentComponent() {
          function UnstableNestedComponent() {
            return <div /> satisfies React.ReactNode;
          }
          return <UnstableNestedComponent />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},

		// ---- TS wrapper: non-null `!` on createElement call ----
		{
			Code: `
        function ParentComponent() {
          function UnstableNestedComponent() {
            return React.createElement("div", null)!;
          }
          return <UnstableNestedComponent />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},

		// ---- ParenthesizedExpression around React.memo argument ----
		{
			Code: `
        function ParentComponent() {
          const UnstableComp = React.memo((() => <div />));
          return <UnstableComp />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// Position points at the start of the wrapper CallExpression
				// (`React.memo`), matching upstream's component registration
				// site for memo/forwardRef wrappers.
				{Message: errorMessage, Line: 3, Column: 32},
			},
		},

		// (ClassExpression test removed — verified via differential testing
		// that upstream eslint-plugin-react v7.37.x does NOT listen on
		// ClassExpression. `const X = class extends React.Component {}` is
		// silent in upstream so we match by not registering the listener.
		// See `differential_test.go` for the alignment harness.)

		// ---- async function as nested component ----
		{
			Code: `
        function ParentComponent() {
          async function UnstableAsyncComponent() {
            return <div />;
          }
          return <UnstableAsyncComponent />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},

		// ---- forwardRef wrapper (sibling of React.memo) ----
		{
			Code: `
        function ParentComponent() {
          const UnstableForwarded = React.forwardRef((props, ref) => <div ref={ref} />);
          return <UnstableForwarded />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// Wrapper-call Pos: `React.forwardRef(...)` starts at col 37.
				{Message: errorMessage, Line: 3, Column: 37},
			},
		},

		// ---- Member-expression wrapper that does NOT wrap a known
		// sibling component — the gate only fires when the JSX root
		// tag resolves to a detected outer/sibling component. With
		// the inner tag unknown, `React.memo` is registered as usual. ----
		{
			Code: `
        function ParentComponent() {
          const Wrap = React.memo(() => <NotKnown />);
          return <Wrap />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 24},
			},
		},

		// ---- Bare `memo(fn)` with `import { memo } from 'react'` ----
		// Upstream's `isPragmaComponentWrapper` only treats a bare
		// `memo(...)` call as a pragma wrapper when the binding was
		// imported from / destructured from / required from the pragma
		// module. The corresponding "no-import" form is in the valid
		// suite — both shapes are covered to lock in the gate.
		{
			Code: `
        import { memo } from 'react';
        function ParentComponent() {
          const UnstableMemo = memo(() => <div />);
          return <UnstableMemo />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// Wrapper-call Pos: `memo(...)` starts at col 32.
				{Message: errorMessage, Line: 4, Column: 32},
			},
		},
		{
			Code: `
        const { memo } = require('react');
        function ParentComponent() {
          const UnstableMemo = memo(() => <div />);
          return <UnstableMemo />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 4, Column: 32},
			},
		},
		{
			Code: `
        import React from 'react';
        const { memo } = React;
        function ParentComponent() {
          const UnstableMemo = memo(() => <div />);
          return <UnstableMemo />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 5, Column: 32},
			},
		},
		{
			Code: `
        import { createElement } from 'react';
        function ParentComponent() {
          function UnstableNested() { return createElement('div'); }
          return <UnstableNested />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// FunctionDeclaration Pos for the nested helper.
				{Message: errorMessage, Line: 4, Column: 11},
			},
		},
		// Bare `createElement(...)` recognized via destructure-from-React
		// (`const { createElement } = React`). Closes the
		// `IsCreateElementCallWithChecker` second branch — TypeChecker
		// resolves the local `createElement` to its BindingElement, and
		// `IsDestructuredFromPragmaImport` then walks to the enclosing
		// VariableDeclaration and confirms its initializer is the pragma
		// identifier (`React`).
		{
			Code: `
        import React from 'react';
        const { createElement } = React;
        function ParentComponent() {
          function UnstableNested() { return createElement('div'); }
          return <UnstableNested />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// FunctionDeclaration Pos for the nested helper.
				{Message: errorMessage, Line: 5, Column: 11},
			},
		},
		// Optional-call form on a bare wrapper (`memo?.(...)`) with
		// `import { memo } from 'react'` — exercises the call-level
		// optional path of `MatchesAnyComponentWrapperWithChecker`.
		// Upstream classifies the call as a wrapper (the callee binding
		// resolves to the pragma module), but emits the message
		// WITHOUT a parent name: Babel wraps optional calls in a
		// ChainExpression sharing range[0] with the inner call, so
		// upstream's parent walk stops at the ChainExpression which has
		// no `.id`. We mirror via `isOptionalPragmaWrapperCall`'s
		// parentName-blanking branch.
		{
			Code: `
        import { memo } from 'react';
        function ParentComponent() {
          const Inner = memo?.(() => <div />);
          return <Inner />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageWithoutName, Line: 4, Column: 25},
			},
		},
		// `nodeWrapsComponent` is order-dependent: a sibling/outer arrow
		// declared AFTER the wrapper call hasn't been added to upstream's
		// detected-components list yet, so the gate doesn't fire and the
		// wrapper IS reported. Locks in the position guard inside
		// `sourceHasComponentNamedBefore`.
		{
			Code: `
        function ParentComponent() {
          const Wrap = React.memo(() => <Inner />);
          return <Wrap />;
        }
        const Inner = () => <div />;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 24},
			},
		},
		// Optional member-access wrapper (`React?.memo(...)`) — Babel
		// wraps the inner CallExpression in a ChainExpression sharing
		// `range[0]` with it, so upstream's parent walk stops at the
		// ChainExpression which has no `.id`. The wrapper IS classified
		// (the property name still matches) but the message is emitted
		// without a parent name — mirrored via `isOptionalPragmaWrapperCall`.
		{
			Code: `
        function ParentComponent() {
          const X = React?.memo(() => <div />);
          return <X />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageWithoutName, Line: 3, Column: 21},
			},
		},
		// Optional member-access on a user-configured wrapper
		// (`MyLib?.observer(...)`) — same ChainExpression quirk as the
		// pragma form; locks in that the parentName-blanking applies
		// regardless of whether the wrapper came from the hardcoded
		// defaults or `componentWrapperFunctions`.
		{
			Code: `
        function ParentComponent() {
          const X = MyLib?.observer(() => <div />);
          return <X />;
        }
      `,
			Tsx: true,
			Settings: map[string]any{
				"componentWrapperFunctions": []any{
					map[string]any{"object": "MyLib", "property": "observer"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageWithoutName, Line: 3, Column: 21},
			},
		},
		// Class component extending `React.PureComponent` qualifies as
		// the parent component. Upstream's `componentUtil.isES6Component`
		// accepts both `Component` and `PureComponent` superclass names
		// (qualified or bare); mirrored by `ExtendsReactComponent` here.
		{
			Code: `
        class ParentComponent extends React.PureComponent {
          render() {
            function UnstableInPure() { return <div />; }
            return <UnstableInPure />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 4, Column: 13},
			},
		},
		// Bare `PureComponent` (without `React.` prefix) — locks in
		// that `ExtendsReactComponent`'s identifier branch accepts
		// both `Component` and `PureComponent` names, mirroring
		// upstream's `componentUtil.isPureComponent` which treats
		// the bare identifier form the same as the qualified one.
		{
			Code: `
        class ParentComponent extends PureComponent {
          render() {
            function UnstableInBarePure() { return <div />; }
            return <UnstableInBarePure />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 4, Column: 13},
			},
		},

		// ---- JsxFragment children render-prop (parallel to upstream's JsxElement case) ----
		{
			Code: `
        function ParentComponent() {
          return (
            <>
              {() => {
                function UnstableInsideFragment() {
                  return <div />;
                }
                return <UnstableInsideFragment />;
              }}
            </>
          );
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 6, Column: 17},
			},
		},

		// ---- Custom react.pragma setting — `Preact.createElement` recognized ----
		{
			Code: `
        function ParentComponent() {
          function UnstableNestedComponent() {
            return Preact.createElement("div", null);
          }
          return <UnstableNestedComponent />;
        }
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},

		// ---- Object-literal shorthand method that returns JSX, in a non-render prop ----
		{
			Code: `
        function ParentComponent() {
          return (
            <ComponentForProps map={{
              Foo() { return <div />; },
            }} />
          );
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// Position narrowed to the `(` of the parameter list,
				// matching upstream's report on the inner FunctionExpression.
				{Message: errorMessageComponentAsProps, Line: 5, Column: 18},
			},
		},

		// ---- Arrow returning createElement directly (expression body) ----
		{
			Code: `
        function ParentComponent() {
          const UnstableArrow = () => React.createElement("div", null);
          return <UnstableArrow />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 33},
			},
		},

		// ---- Conditional return: `cond ? <div/> : null` qualifies as JSX-or-null ----
		{
			Code: `
        function ParentComponent(props) {
          function UnstableConditional() {
            return props.show ? <div /> : null;
          }
          return <UnstableConditional />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},

		// ---- Nested arrow as property of a deeply-nested object inside JSX attribute ----
		{
			Code: `
        function ParentComponent() {
          return (
            <Comp config={{
              ui: {
                Header: () => <div />,
              }
            }} />
          );
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageComponentAsProps, Line: 6, Column: 25},
			},
		},

		// ---- React.memo wrapping a TS-asserted arrow ----
		{
			Code: `
        function ParentComponent() {
          const UnstableMemoTS = React.memo((() => <div />) as any);
          return <UnstableMemoTS />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// Wrapper-call Pos: `React.memo(...)` starts at col 34.
				{Message: errorMessage, Line: 3, Column: 34},
			},
		},

		// ---- Doubly-wrapped pragma component:
		// React.memo(React.forwardRef(...)). The INNER wrapper is the
		// reported site — this matches upstream's behavior verified via
		// differential testing: ESLint's `Components.detect` registers the
		// inner forwardRef call as a component but the outer memo call is
		// not (its first argument is a CallExpression, not a FunctionLike,
		// so it fails the `isReturningJSXOrNull(inner)` gate). ----
		{
			Code: `
        function ParentComponent() {
          const UnstableDouble = React.memo(React.forwardRef((props, ref) => <div ref={ref} />));
          return <UnstableDouble />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				// Wrapper-call Pos: inner `React.forwardRef(...)` starts at
				// col 45. Inner-arrow self-suppresses (its parent IS a
				// wrapper call) so we don't double-fire.
				{Message: errorMessage, Line: 3, Column: 45},
			},
		},

		// ---- Class with class-static-block must not affect detection of
		// nested function components inside its render method ----
		{
			Code: `
        class ParentComponent extends React.Component {
          static {
            console.log("init");
          }
          render() {
            const UnstableInner = () => <div />;
            return <UnstableInner />;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 7, Column: 35},
			},
		},

		// ============================================================
		// Round-3 expansion: full container coverage with EndLine /
		// EndColumn assertions, LogicalExpression / Identifier
		// resolution, settings.componentWrapperFunctions support, and
		// other previously-missed paths.
		// ============================================================

		// ---- EndLine / EndColumn on a multi-line FunctionDeclaration body ----
		{
			Code: `
        function ParentComponent() {
          function UnstableMultiline() {
            return (
              <div>
                <span>nested</span>
              </div>
            );
          }
          return <UnstableMultiline />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Message: errorMessage,
					Line: 3, Column: 11, EndLine: 9, EndColumn: 12,
				},
			},
		},

		// ---- EndLine / EndColumn on a multi-line ClassDeclaration ----
		{
			Code: `
        function ParentComponent() {
          class UnstableMulti extends React.Component {
            render() {
              return (
                <div>
                  text
                </div>
              );
            }
          };
          return <UnstableMulti />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Message: errorMessage,
					Line: 3, Column: 11, EndLine: 11, EndColumn: 12,
				},
			},
		},

		// ---- LogicalExpression `cond && <div/>` return — strict
		// isReturningJSX accepts EITHER side as JSX. ----
		{
			Code: `
        function ParentComponent(props) {
          function UnstableLogical() {
            return props.show && <div />;
          }
          return <UnstableLogical />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},

		// ---- Nullish coalescing `?? <div/>` return ----
		{
			Code: `
        function ParentComponent(props) {
          function UnstableNullish() {
            return props.cached ?? <div />;
          }
          return <UnstableNullish />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},

		// ---- Comma sequence return `(setup(), <div/>)` ----
		{
			Code: `
        function ParentComponent() {
          function UnstableSeq() {
            return (init(), <div />);
          }
          return <UnstableSeq />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},

		// ---- Identifier resolution: function returns a local variable
		// that holds JSX. Mirrors upstream's `case 'Identifier'` arm in
		// jsxUtil.isReturningJSX via a one-hop initializer lookup. ----
		{
			Code: `
        function ParentComponent() {
          function UnstableIndirect() {
            const view = <div />;
            return view;
          }
          return <UnstableIndirect />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 3, Column: 11},
			},
		},

		// (Identifier resolution chain `const a = <div/>; const b = a;
		// return b;` was previously asserted as invalid here, but
		// differential testing confirmed upstream does NOT chain-resolve:
		// its `isJSX(variable)` accepts only JSXElement / JSXFragment
		// initializers — Identifier-to-Identifier chains return false. The
		// matching valid case lives in the rule's valid suite below.)

		// ---- Cross-block Identifier resolution (TypeChecker path):
		// the JSX-bound binding lives in an outer block; only a real
		// scope walk (not the local-block fallback) can resolve it. ----
		{
			Code: `
        function ParentComponent() {
          const sharedView = <div />;
          function UnstableCrossBlock() {
            if (true) {
              return sharedView;
            }
            return null;
          }
          return <UnstableCrossBlock />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 4, Column: 11},
			},
		},

		// ---- settings.componentWrapperFunctions extends defaults — a
		// user-declared `myMemo(fn)` call counts as a component wrapper. ----
		{
			Code: `
        function ParentComponent() {
          const UnstableCustomWrap = myMemo(() => <div />);
          return <UnstableCustomWrap />;
        }
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"componentWrapperFunctions": []interface{}{"myMemo"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				// Wrapper-call Pos: `myMemo(...)` starts at col 38.
				{Message: errorMessage, Line: 3, Column: 38},
			},
		},

		// ---- settings.componentWrapperFunctions object form
		// `{property, object}` — pragma-qualified custom wrapper. ----
		{
			Code: `
        function ParentComponent() {
          const UnstableNS = MyLib.observer(() => <div />);
          return <UnstableNS />;
        }
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"componentWrapperFunctions": []interface{}{
					map[string]interface{}{"property": "observer", "object": "MyLib"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				// Wrapper-call Pos: `MyLib.observer(...)` starts at col 30.
				{Message: errorMessage, Line: 3, Column: 30},
			},
		},

		// ---- Member-expression parent component name —
		// `Foo.Bar = function ...` host. Parent name resolution returns
		// "" because `Foo.Bar` is not a VariableDeclaration; the diagnostic
		// uses the no-name template. ----
		{
			Code: `
        var Foo = {};
        Foo.Bar = function () {
          function UnstableInside() { return <div />; }
          return <UnstableInside />;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessageWithoutName, Line: 4, Column: 11},
			},
		},

		// ---- TypeScript decorator on a nested class component does NOT
		// alter detection — the heritage clause still drives classification. ----
		{
			Code: `
        function deco(target: any) { return target; }
        function ParentComponent() {
          @deco
          class UnstableDecorated extends React.Component {
            render() { return <div />; }
          };
          return <UnstableDecorated />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: errorMessage, Line: 4, Column: 11},
			},
		},
	})
}

// TestNilCheckerFallback locks down the contract that every Identifier-
// resolving entry point used by this rule degrades safely when no
// TypeChecker is available — the case for tooling that runs the rule on
// a SourceFile-only host or with TypeScript project services disabled.
//
// Each helper must return its safe-default ("not recognized") value
// rather than panicking on a nil receiver. The cases are split by
// argument shape (nil fn vs nil checker on a present fn) because a
// future regression could land in either guard.
func TestNilCheckerFallback(t *testing.T) {
	t.Run("nil fn + nil checker", func(t *testing.T) {
		if reactutil.FunctionReturnsJSXOrNullWithChecker(nil, "", nil) {
			t.Fatal("nil fn + nil tc must yield false")
		}
		if reactutil.FunctionReturnsJSXWithChecker(nil, "", nil) {
			t.Fatal("nil fn + nil tc must yield false (strict)")
		}
	})

	t.Run("nil-typed fn + nil checker (no panic)", func(t *testing.T) {
		// The rule_tester suite above exercises the non-nil-fn path
		// with real ASTs; here we only verify the WithChecker entry
		// points accept an explicitly nil-typed *ast.Node without
		// panicking when the checker is also nil.
		var nilFn *ast.Node
		_ = reactutil.FunctionReturnsJSXOrNullWithChecker(nilFn, "Preact", nil)
		_ = reactutil.FunctionReturnsJSXWithChecker(nilFn, "", nil)
	})

	t.Run("import-aware helpers degrade safely with nil checker", func(t *testing.T) {
		if reactutil.IsDestructuredFromPragmaImport(nil, "React", nil) {
			t.Fatal("IsDestructuredFromPragmaImport(nil ident, nil tc) must return false")
		}
		if reactutil.IsCreateElementCallWithChecker(nil, "React", nil) {
			t.Fatal("IsCreateElementCallWithChecker(nil callee, nil tc) must return false")
		}
		if reactutil.MatchesAnyComponentWrapperWithChecker(nil, nil, nil, "React", nil) {
			t.Fatal("MatchesAnyComponentWrapperWithChecker(nil call, nil fn, nil tc) must return false")
		}
		if reactutil.IsStatelessReactComponentWithChecker(nil, "React", nil) {
			t.Fatal("IsStatelessReactComponentWithChecker(nil fn, nil tc) must return false")
		}
		if reactutil.IsStatelessReactComponentWithWrappers(nil, "React", nil, nil) {
			t.Fatal("IsStatelessReactComponentWithWrappers(nil fn, nil tc, nil wrappers) must return false")
		}
	})
}
