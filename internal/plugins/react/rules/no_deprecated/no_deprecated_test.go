package no_deprecated

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Common message builders (mirror upstream's errorMessage helper).
func msg(oldMethod, version, newMethod, refs string) string {
	out := oldMethod + " is deprecated since React " + version
	if newMethod != "" {
		out += ", use " + newMethod + " instead"
	}
	if refs != "" {
		out += ", see " + refs
	}
	return out
}

const lifecycleRefsTail = ". Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components."
const lifecycleRefsHead = "https://reactjs.org/docs/react-component.html#"

func TestNoDeprecatedRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDeprecatedRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid: not deprecated APIs ----
		{Code: `var element = React.createElement('p', {}, null);`, Tsx: true},
		{Code: `var clone = React.cloneElement(element);`, Tsx: true},
		{Code: `ReactDOM.cloneElement(child, container);`, Tsx: true},
		{Code: `ReactDOM.findDOMNode(instance);`, Tsx: true},
		{Code: `ReactDOM.createPortal(child, container);`, Tsx: true},
		{Code: `ReactDOMServer.renderToString(element);`, Tsx: true},
		{Code: `ReactDOMServer.renderToStaticMarkup(element);`, Tsx: true},

		// ---- Upstream valid: createReactClass with only render ----
		{Code: `
        var Foo = createReactClass({
          render: function() {}
        })
      `, Tsx: true},

		// ---- Upstream valid: Non-React patterns (not createReactClass) ----
		{Code: `
        var Foo = createReactClassNonReact({
          componentWillMount: function() {},
          componentWillReceiveProps: function() {},
          componentWillUpdate: function() {}
        });
      `, Tsx: true},
		{Code: `
        var Foo = {
          componentWillMount: function() {},
          componentWillReceiveProps: function() {},
          componentWillUpdate: function() {}
        };
      `, Tsx: true},
		{Code: `
        class Foo {
          constructor() {}
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `, Tsx: true},

		// ---- Upstream valid: deprecated in a later React version than settings ----
		{Code: `React.renderComponent()`, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "0.11.0"}}},
		{Code: `React.createClass()`, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "15.4.0"}}},
		{Code: `PropTypes`, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "15.4.0"}}},
		{Code: `
        class Foo extends React.Component {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.8.0"}}},

		// ---- Upstream valid: destructuring rest / default — no flagged name ----
		{Code: `
        import React from "react";

        let { default: defaultReactExport, ...allReactExports } = React;
      `, Tsx: true},

		// ---- Upstream valid: React < 18 (react-dom / react-dom/server pre-18 API is fine) ----
		{Code: `
        import { render, hydrate } from 'react-dom';
        import { renderToNodeStream } from 'react-dom/server';
        ReactDOM.render(element, container);
        ReactDOM.unmountComponentAtNode(container);
        ReactDOMServer.renderToNodeStream(element);
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "17.999.999"}}},

		// ---- Upstream valid: React 18 replacements are fine ----
		{Code: `
        import ReactDOM, { createRoot } from 'react-dom/client';
        ReactDOM.createRoot(container);
        const root = createRoot(container);
        root.unmount();
      `, Tsx: true},
		{Code: `
        import ReactDOM, { hydrateRoot } from 'react-dom/client';
        ReactDOM.hydrateRoot(container, <App/>);
        hydrateRoot(container, <App/>);
      `, Tsx: true},
		{Code: `
        import ReactDOMServer, { renderToPipeableStream } from 'react-dom/server';
        ReactDOMServer.renderToPipeableStream(<App />, {});
        renderToPipeableStream(<App />, {});
      `, Tsx: true},

		// ---- Upstream valid: renderToString stays on ReactDOMServer, not deprecated ----
		{Code: `
        import { renderToString } from 'react-dom/server';
      `, Tsx: true},
		{Code: `
        const { renderToString } = require('react-dom/server');
      `, Tsx: true},

		// ---- Edge: element access doesn't match (dotted-path only) ----
		{Code: `React['createClass']({})`, Tsx: true},

		// ---- Edge: non-React destructuring isn't flagged ----
		{Code: `var { createClass } = something();`, Tsx: true},
		{Code: `var { createClass } = require('not-react');`, Tsx: true},

		// ---- Edge: lifecycle inside nested non-component class is ignored ----
		// Only the nearest class decides ES6-component status, so an Inner
		// non-React class with componentWillMount does NOT inherit Outer's
		// component-ness.
		{Code: `
        class Outer extends React.Component {
          render() {
            class Inner {
              componentWillMount() {}
            }
            return null;
          }
        }
      `, Tsx: true},

		// ---- Edge: non-null assertion `React!.createClass` — chain breaks, not flagged ----
		// ESLint runs on JS and never sees this; rslint conservatively doesn't
		// traverse through TS-only NonNullExpression either, matching "upstream
		// has no opinion" semantics.
		{Code: `React!.createClass({})`, Tsx: true},

		// ---- Edge: `as any` cast breaks the dotted chain similarly ----
		{Code: `(React as any).createClass({})`, Tsx: true},

		// ---- Edge: renderToString from react-dom/server — canonical is
		// ReactDOMServer.renderToString, which is NOT in the deprecation map. ----
		{Code: `import { renderToString } from 'react-dom/server';`, Tsx: true},
		{Code: `ReactDOMServer.renderToString(element);`, Tsx: true},

		// ---- Edge: namespace import isn't a named specifier, so nothing flagged ----
		{Code: `import * as React from 'react';`, Tsx: true},

		// ---- Edge: default import alone — no named specifiers, nothing flagged ----
		{Code: `import React from 'react';`, Tsx: true},

		// ---- Edge: static lifecycle method on React class — same name matches map ----
		// Note: upstream also flags this (getComponentProperties returns all members,
		// the deprecation map key is just the bare name). Locked in for parity.
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream #1: React.renderComponent() ----
		{
			Code: `React.renderComponent()`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.renderComponent", "0.12.0", "React.render", ""),
				Line:      1, Column: 1,
			}},
		},

		// ---- Upstream #2: custom pragma via settings ----
		{
			Code:     `Foo.renderComponent()`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Foo"}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("Foo.renderComponent", "0.12.0", "Foo.render", ""),
				Line:      1, Column: 1,
			}},
		},

		// ---- Upstream #3: `@jsx Foo` comment overrides pragma ----
		{
			Code: `/** @jsx Foo */ Foo.renderComponent()`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("Foo.renderComponent", "0.12.0", "Foo.render", ""),
				Line:      1, Column: 17,
			}},
		},

		// ---- Upstream #4: `this.transferPropsTo()` ----
		{
			Code: `this.transferPropsTo()`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("this.transferPropsTo", "0.12.0", "spread operator ({...})", ""),
				Line:      1, Column: 1,
			}},
		},

		// ---- Upstream #5: nested member access — inner path matches ----
		{
			Code: `React.addons.TestUtils`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.addons.TestUtils", "15.5.0", "ReactDOM.TestUtils", ""),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `React.addons.classSet()`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.addons.classSet", "0.13.0", "the npm module classnames", ""),
				Line:      1, Column: 1,
			}},
		},

		// ---- Upstream #6: 0.14.0 migrations ----
		{
			Code: `React.render(element, container);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.render", "0.14.0", "ReactDOM.render", ""),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `React.unmountComponentAtNode(container);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.unmountComponentAtNode", "0.14.0", "ReactDOM.unmountComponentAtNode", ""),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `React.findDOMNode(instance);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.findDOMNode", "0.14.0", "ReactDOM.findDOMNode", ""),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `React.renderToString(element);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.renderToString", "0.14.0", "ReactDOMServer.renderToString", ""),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `React.renderToStaticMarkup(element);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.renderToStaticMarkup", "0.14.0", "ReactDOMServer.renderToStaticMarkup", ""),
				Line:      1, Column: 1,
			}},
		},

		// ---- Upstream #7: React.createClass / React.PropTypes / React.DOM ----
		{
			Code: `React.createClass({});`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.createClass", "15.5.0", "the npm module create-react-class", ""),
				Line:      1, Column: 1,
			}},
		},
		{
			Code:     `Foo.createClass({});`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Foo"}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("Foo.createClass", "15.5.0", "the npm module create-react-class", ""),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `React.PropTypes`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.PropTypes", "15.5.0", "the npm module prop-types", ""),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `React.DOM.div`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.DOM", "15.6.0", "the npm module react-dom-factories", ""),
				Line:      1, Column: 1,
			}},
		},

		// ---- Upstream #8: require() destructuring ----
		{
			Code: `var {createClass} = require('react');`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.createClass", "15.5.0", "the npm module create-react-class", ""),
				Line:      1, Column: 6,
			}},
		},
		{
			Code: `var {createClass, PropTypes} = require('react');`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Message:   msg("React.createClass", "15.5.0", "the npm module create-react-class", ""),
					Line:      1, Column: 6,
				},
				{
					MessageId: "deprecated",
					Message:   msg("React.PropTypes", "15.5.0", "the npm module prop-types", ""),
					Line:      1, Column: 19,
				},
			},
		},

		// ---- Upstream #9: named imports ----
		{
			Code: `import {createClass} from 'react';`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.createClass", "15.5.0", "the npm module create-react-class", ""),
				Line:      1, Column: 9,
			}},
		},
		{
			Code: `import {createClass, PropTypes} from 'react';`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Message:   msg("React.createClass", "15.5.0", "the npm module create-react-class", ""),
					Line:      1, Column: 9,
				},
				{
					MessageId: "deprecated",
					Message:   msg("React.PropTypes", "15.5.0", "the npm module prop-types", ""),
					Line:      1, Column: 22,
				},
			},
		},

		// ---- Upstream #10: default import + destructure of the namespace ----
		{
			Code: `
      import React from 'react';
      const {createClass, PropTypes} = React;
    `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Message:   msg("React.createClass", "15.5.0", "the npm module create-react-class", ""),
					Line:      3, Column: 14,
				},
				{
					MessageId: "deprecated",
					Message:   msg("React.PropTypes", "15.5.0", "the npm module prop-types", ""),
					Line:      3, Column: 27,
				},
			},
		},

		// ---- Upstream #11: react-addons-perf ----
		{
			Code: `import {printDOM} from 'react-addons-perf';`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("ReactPerf.printDOM", "15.0.0", "ReactPerf.printOperations", ""),
				Line:      1, Column: 9,
			}},
		},
		{
			Code: `
        import ReactPerf from 'react-addons-perf';
        const {printDOM} = ReactPerf;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("ReactPerf.printDOM", "15.0.0", "ReactPerf.printOperations", ""),
				Line:      3, Column: 16,
			}},
		},

		// ---- Upstream #12: lifecycle methods (ES6 class extending React.PureComponent) ----
		{
			Code: `
        class Bar extends React.PureComponent {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Message:   msg("componentWillMount", "16.9.0", "UNSAFE_componentWillMount", lifecycleRefsHead+"unsafe_componentwillmount"+lifecycleRefsTail),
					Line:      3, Column: 11,
				},
				{
					MessageId: "deprecated",
					Message:   msg("componentWillReceiveProps", "16.9.0", "UNSAFE_componentWillReceiveProps", lifecycleRefsHead+"unsafe_componentwillreceiveprops"+lifecycleRefsTail),
					Line:      4, Column: 11,
				},
				{
					MessageId: "deprecated",
					Message:   msg("componentWillUpdate", "16.9.0", "UNSAFE_componentWillUpdate", lifecycleRefsHead+"unsafe_componentwillupdate"+lifecycleRefsTail),
					Line:      5, Column: 11,
				},
			},
		},

		// ---- Upstream #13: class expression (extends React.PureComponent) ----
		{
			Code: `
        function Foo() {
          return class Bar extends React.PureComponent {
            componentWillMount() {}
            componentWillReceiveProps() {}
            componentWillUpdate() {}
          };
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 4, Column: 13},
				{MessageId: "deprecated", Line: 5, Column: 13},
				{MessageId: "deprecated", Line: 6, Column: 13},
			},
		},

		// ---- Upstream #14: bare PureComponent (unqualified) ----
		{
			Code: `
        class Bar extends PureComponent {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 3, Column: 11},
				{MessageId: "deprecated", Line: 4, Column: 11},
				{MessageId: "deprecated", Line: 5, Column: 11},
			},
		},

		// ---- Upstream #15: React.Component ----
		{
			Code: `
        class Foo extends React.Component {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 3, Column: 11},
				{MessageId: "deprecated", Line: 4, Column: 11},
				{MessageId: "deprecated", Line: 5, Column: 11},
			},
		},

		// ---- Upstream #16: bare Component ----
		{
			Code: `
        class Foo extends Component {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 3, Column: 11},
				{MessageId: "deprecated", Line: 4, Column: 11},
				{MessageId: "deprecated", Line: 5, Column: 11},
			},
		},

		// ---- Upstream #17: createReactClass lifecycle methods ----
		{
			Code: `
        var Foo = createReactClass({
          componentWillMount: function() {},
          componentWillReceiveProps: function() {},
          componentWillUpdate: function() {}
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 3, Column: 11},
				{MessageId: "deprecated", Line: 4, Column: 11},
				{MessageId: "deprecated", Line: 5, Column: 11},
			},
		},

		// ---- Upstream #18: with constructor (not reported) ----
		{
			Code: `
        class Foo extends React.Component {
          constructor() {}
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 4, Column: 11},
				{MessageId: "deprecated", Line: 5, Column: 11},
				{MessageId: "deprecated", Line: 6, Column: 11},
			},
		},

		// ---- Upstream #19–22: React 18 react-dom / react-dom-server deprecations ----
		{
			Code: `
        import { render } from 'react-dom';
        ReactDOM.render(<div></div>, container);
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Message:   msg("ReactDOM.render", "18.0.0", "createRoot", "https://reactjs.org/link/switch-to-createroot"),
					Line:      2, Column: 18,
				},
				{
					MessageId: "deprecated",
					Message:   msg("ReactDOM.render", "18.0.0", "createRoot", "https://reactjs.org/link/switch-to-createroot"),
					Line:      3, Column: 9,
				},
			},
		},
		{
			Code: `
        import { hydrate } from 'react-dom';
        ReactDOM.hydrate(<div></div>, container);
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Message:   msg("ReactDOM.hydrate", "18.0.0", "hydrateRoot", "https://reactjs.org/link/switch-to-createroot"),
					Line:      2, Column: 18,
				},
				{
					MessageId: "deprecated",
					Message:   msg("ReactDOM.hydrate", "18.0.0", "hydrateRoot", "https://reactjs.org/link/switch-to-createroot"),
					Line:      3, Column: 9,
				},
			},
		},
		{
			Code: `
        import { unmountComponentAtNode } from 'react-dom';
        ReactDOM.unmountComponentAtNode(container);
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Message:   msg("ReactDOM.unmountComponentAtNode", "18.0.0", "root.unmount", "https://reactjs.org/link/switch-to-createroot"),
					Line:      2, Column: 18,
				},
				{
					MessageId: "deprecated",
					Message:   msg("ReactDOM.unmountComponentAtNode", "18.0.0", "root.unmount", "https://reactjs.org/link/switch-to-createroot"),
					Line:      3, Column: 9,
				},
			},
		},
		{
			Code: `
        import { renderToNodeStream } from 'react-dom/server';
        ReactDOMServer.renderToNodeStream(element);
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Message:   msg("ReactDOMServer.renderToNodeStream", "18.0.0", "renderToPipeableStream", "https://reactjs.org/docs/react-dom-server.html#rendertonodestream"),
					Line:      2, Column: 18,
				},
				{
					MessageId: "deprecated",
					Message:   msg("ReactDOMServer.renderToNodeStream", "18.0.0", "renderToPipeableStream", "https://reactjs.org/docs/react-dom-server.html#rendertonodestream"),
					Line:      3, Column: 9,
				},
			},
		},

		// ---- Edge: parenthesized `(React).createClass` — rslint flags, ESLint would not ----
		// Locks in the Phase 1 Step 5.B divergence documented in the `.md`.
		{
			Code: `(React).createClass({})`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.createClass", "15.5.0", "the npm module create-react-class", ""),
				Line:      1, Column: 1,
			}},
		},

		// ---- Edge: version exactly at deprecation boundary is reported ----
		{
			Code:     `React.createClass({})`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "15.5.0"}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.createClass", "15.5.0", "the npm module create-react-class", ""),
				Line:      1, Column: 1,
			}},
		},

		// ---- Edge: async / generator / async-generator lifecycle variants ----
		{
			Code: `
        class Foo extends React.Component {
          async componentWillMount() {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 3, Column: 17},
			},
		},

		// ---- Edge: class field arrow lifecycle ----
		{
			Code: `
        class Foo extends React.Component {
          componentWillMount = () => {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 3, Column: 11},
			},
		},

		// ---- Edge: aliased destructuring `{createClass: fn}` — key name is from-module ----
		{
			Code: `const {createClass: fn} = require('react');`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.createClass", "15.5.0", "the npm module create-react-class", ""),
				Line:      1, Column: 8,
			}},
		},

		// ---- Edge: optional chain `React?.createClass()` — tsgo PropertyAccess
		// with OptionalChain flag is still KindPropertyAccessExpression, so the
		// dotted-path builder handles it transparently. ----
		{
			Code: `React?.createClass({})`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.createClass", "15.5.0", "the npm module create-react-class", ""),
				Line:      1, Column: 1,
			}},
		},
		{
			Code: `React?.addons?.TestUtils`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("React.addons.TestUtils", "15.5.0", "ReactDOM.TestUtils", ""),
				Line:      1, Column: 1,
			}},
		},

		// ---- Edge: createReactClass with multi-layer parens around the object argument ----
		{
			Code: `
        createReactClass(((({
          componentWillMount: function() {}
        }))));
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 3, Column: 11},
			},
		},

		// ---- Edge: `<pragma>.createClass` ≠ configured createClass (default
		// `createReactClass`), so the inner object is NOT recognized as an ES5
		// component and its lifecycle keys are NOT examined. Only the outer
		// `React.createClass` itself is flagged (as a deprecated member access).
		// Matches upstream: upstream's `componentUtil.isES5Component` also
		// compares against `settings.react.createClass` literally. ----
		{
			Code: `
        React.createClass({
          componentWillMount: function() {}
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 2, Column: 9},
			},
		},

		// ---- Edge: same shape but with `settings.react.createClass` set to
		// `"createClass"` — now the inner object IS an ES5 component and its
		// lifecycle prop is additionally flagged. ----
		{
			Code: `
        React.createClass({
          componentWillMount: function() {}
        });
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{"createClass": "createClass"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 2, Column: 9},
				{MessageId: "deprecated", Line: 3, Column: 11},
			},
		},

		// ---- Edge: @jsx wins over settings.react.pragma when both are present ----
		{
			Code:     `/** @jsx Bar */ Bar.createClass({})`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Foo"}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("Bar.createClass", "15.5.0", "the npm module create-react-class", ""),
				Line:      1, Column: 17,
			}},
		},

		// ---- Edge: class expression assigned to const — lifecycle still detected ----
		{
			Code: `
        const Hello = class extends React.Component {
          componentWillMount() {}
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 3, Column: 11},
			},
		},

		// ---- Edge: SpreadAssignment in createReactClass doesn't crash and doesn't mask siblings ----
		{
			Code: `
        const mixin = {};
        var Hello = createReactClass({
          ...mixin,
          componentWillUpdate: function() {}
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 5, Column: 11},
			},
		},

		// ---- Edge: ReactDOMServer.renderToNodeStream without any import (direct member access). ----
		{
			Code: `ReactDOMServer.renderToNodeStream(element);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "deprecated",
				Message:   msg("ReactDOMServer.renderToNodeStream", "18.0.0", "renderToPipeableStream", "https://reactjs.org/docs/react-dom-server.html#rendertonodestream"),
				Line:      1, Column: 1,
			}},
		},
	})
}
