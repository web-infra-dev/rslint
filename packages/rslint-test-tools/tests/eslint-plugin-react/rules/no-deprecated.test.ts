import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectDeprecatedError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'deprecated',
    message:
      'React.createClass is deprecated since React 15.5.0, use the npm module create-react-class instead',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectDeprecatedError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'deprecated',
    message:
      'React.PropTypes is deprecated since React 15.5.0, use the npm module prop-types instead',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectDeprecatedError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'deprecated',
    message:
      'componentWillMount is deprecated since React 16.9.0, use UNSAFE_componentWillMount instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillmount. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectDeprecatedError4(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'deprecated',
    message:
      'componentWillReceiveProps is deprecated since React 16.9.0, use UNSAFE_componentWillReceiveProps instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillreceiveprops. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectDeprecatedError5(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'deprecated',
    message:
      'componentWillUpdate is deprecated since React 16.9.0, use UNSAFE_componentWillUpdate instead, see https://reactjs.org/docs/react-component.html#unsafe_componentwillupdate. Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles to automatically update your components.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-deprecated', {
  valid: [
    // ---- Not deprecated ----
    { code: `var element = React.createElement('p', {}, null);` },
    { code: `var clone = React.cloneElement(element);` },
    { code: `ReactDOM.cloneElement(child, container);` },
    { code: `ReactDOM.findDOMNode(instance);` },
    { code: `ReactDOM.createPortal(child, container);` },
    { code: `ReactDOMServer.renderToString(element);` },
    { code: `ReactDOMServer.renderToStaticMarkup(element);` },

    // ---- createReactClass with only render ----
    {
      code: `
        var Foo = createReactClass({
          render: function() {}
        })
      `,
    },

    // ---- Non-React patterns ----
    {
      code: `
        var Foo = createReactClassNonReact({
          componentWillMount: function() {}
        });
      `,
    },
    {
      code: `
        var Foo = { componentWillMount: function() {} };
      `,
    },
    {
      code: `
        class Foo {
          componentWillMount() {}
        }
      `,
    },

    // ---- React 18 replacements ----
    {
      code: `
        import ReactDOM, { createRoot } from 'react-dom/client';
        ReactDOM.createRoot(container);
        const root = createRoot(container);
        root.unmount();
      `,
    },
    {
      code: `
        import { renderToString } from 'react-dom/server';
      `,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        var Foo = createReactClassNonReact({
          componentWillMount: function() {},
          componentWillReceiveProps: function() {},
          componentWillUpdate: function() {}
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        var Foo = {
          componentWillMount: function() {},
          componentWillReceiveProps: function() {},
          componentWillUpdate: function() {}
        };
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        class Foo {
          constructor() {}
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `React.renderComponent()`,
      settings: { react: { version: '0.11.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `React.createClass()`,
      settings: { react: { version: '15.4.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `PropTypes`,
      settings: { react: { version: '15.4.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        class Foo extends React.Component {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `,
      settings: { react: { version: '16.8.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        import React from "react";

        let { default: defaultReactExport, ...allReactExports } = React;
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        import { render, hydrate } from 'react-dom';
        import { renderToNodeStream } from 'react-dom/server';
        ReactDOM.render(element, container);
        ReactDOM.unmountComponentAtNode(container);
        ReactDOMServer.renderToNodeStream(element);
      `,
      settings: { react: { version: '17.999.999' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        import ReactDOM, { hydrateRoot } from 'react-dom/client';
        ReactDOM.hydrateRoot(container, <App/>);
        hydrateRoot(container, <App/>);
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        import ReactDOMServer, { renderToPipeableStream } from 'react-dom/server';
        ReactDOMServer.renderToPipeableStream(<App />, {});
        renderToPipeableStream(<App />, {});
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        const { renderToString } = require('react-dom/server');
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // ---- Member access ----
    {
      code: `React.renderComponent()`,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'React.renderComponent is deprecated since React 0.12.0, use React.render instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: `this.transferPropsTo()`,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'this.transferPropsTo is deprecated since React 0.12.0, use spread operator ({...}) instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: `React.addons.TestUtils`,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'React.addons.TestUtils is deprecated since React 15.5.0, use ReactDOM.TestUtils instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: `React.render(element, container);`,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'React.render is deprecated since React 0.14.0, use ReactDOM.render instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: `React.createClass({});`,
      errors: [expectDeprecatedError1([1, 1, 1, 18])],
    },
    {
      code: `React.PropTypes`,
      errors: [expectDeprecatedError2([1, 1, 1, 16])],
    },
    {
      code: `React.DOM.div`,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'React.DOM is deprecated since React 15.6.0, use the npm module react-dom-factories instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },

    // ---- Destructuring ----
    {
      code: `var {createClass} = require('react');`,
      errors: [expectDeprecatedError1([1, 6, 1, 17])],
    },
    {
      code: `var {createClass, PropTypes} = require('react');`,
      errors: [
        expectDeprecatedError1([1, 6, 1, 17]),
        expectDeprecatedError2([1, 19, 1, 28]),
      ],
    },

    // ---- Imports ----
    {
      code: `import {createClass} from 'react';`,
      errors: [expectDeprecatedError1([1, 9, 1, 20])],
    },
    {
      code: `import {printDOM} from 'react-addons-perf';`,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'ReactPerf.printDOM is deprecated since React 15.0.0, use ReactPerf.printOperations instead',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },

    // ---- Lifecycle methods ----
    {
      code: `
        class Foo extends React.Component {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `,
      errors: [
        expectDeprecatedError3([3, 11, 3, 29]),
        expectDeprecatedError4([4, 11, 4, 36]),
        expectDeprecatedError5([5, 11, 5, 30]),
      ],
    },
    {
      code: `
        var Foo = createReactClass({
          componentWillMount: function() {}
        })
      `,
      errors: [expectDeprecatedError3([3, 11, 3, 29])],
    },

    // ---- React 18 deprecations ----
    {
      code: `
        import { render } from 'react-dom';
        ReactDOM.render(<div></div>, container);
      `,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'ReactDOM.render is deprecated since React 18.0.0, use createRoot instead, see https://reactjs.org/link/switch-to-createroot',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 24,
        },
        {
          messageId: 'deprecated',
          message:
            'ReactDOM.render is deprecated since React 18.0.0, use createRoot instead, see https://reactjs.org/link/switch-to-createroot',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 24,
        },
      ],
    },
    {
      code: `
        import { renderToNodeStream } from 'react-dom/server';
        ReactDOMServer.renderToNodeStream(element);
      `,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'ReactDOMServer.renderToNodeStream is deprecated since React 18.0.0, use renderToPipeableStream instead, see https://reactjs.org/docs/react-dom-server.html#rendertonodestream',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 36,
        },
        {
          messageId: 'deprecated',
          message:
            'ReactDOMServer.renderToNodeStream is deprecated since React 18.0.0, use renderToPipeableStream instead, see https://reactjs.org/docs/react-dom-server.html#rendertonodestream',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 42,
        },
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `Foo.renderComponent()`,
      settings: { react: { pragma: 'Foo' } },
      errors: [
        {
          messageId: 'deprecated',
          message:
            'Foo.renderComponent is deprecated since React 0.12.0, use Foo.render instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `/** @jsx Foo */ Foo.renderComponent()`,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'Foo.renderComponent is deprecated since React 0.12.0, use Foo.render instead',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 36,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `React.addons.classSet()`,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'React.addons.classSet is deprecated since React 0.13.0, use the npm module classnames instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `React.unmountComponentAtNode(container);`,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'React.unmountComponentAtNode is deprecated since React 0.14.0, use ReactDOM.unmountComponentAtNode instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `React.findDOMNode(instance);`,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'React.findDOMNode is deprecated since React 0.14.0, use ReactDOM.findDOMNode instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `React.renderToString(element);`,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'React.renderToString is deprecated since React 0.14.0, use ReactDOMServer.renderToString instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `React.renderToStaticMarkup(element);`,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'React.renderToStaticMarkup is deprecated since React 0.14.0, use ReactDOMServer.renderToStaticMarkup instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `Foo.createClass({});`,
      settings: { react: { pragma: 'Foo' } },
      errors: [
        {
          messageId: 'deprecated',
          message:
            'Foo.createClass is deprecated since React 15.5.0, use the npm module create-react-class instead',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `import {createClass, PropTypes} from 'react';`,
      errors: [
        expectDeprecatedError1([1, 9, 1, 20]),
        expectDeprecatedError2([1, 22, 1, 31]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
      import React from 'react';
      const {createClass, PropTypes} = React;
    `,
      errors: [
        expectDeprecatedError1([3, 14, 3, 25]),
        expectDeprecatedError2([3, 27, 3, 36]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        import ReactPerf from 'react-addons-perf';
        const {printDOM} = ReactPerf;
      `,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'ReactPerf.printDOM is deprecated since React 15.0.0, use ReactPerf.printOperations instead',
          line: 3,
          column: 16,
          endLine: 3,
          endColumn: 24,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `
        class Bar extends React.PureComponent {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        };
      `,
      errors: [
        expectDeprecatedError3([3, 11, 3, 29]),
        expectDeprecatedError4([4, 11, 4, 36]),
        expectDeprecatedError5([5, 11, 5, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        function Foo() {
          return class Bar extends React.PureComponent {
            componentWillMount() {}
            componentWillReceiveProps() {}
            componentWillUpdate() {}
          };
        }
      `,
      errors: [
        expectDeprecatedError3([4, 13, 4, 31]),
        expectDeprecatedError4([5, 13, 5, 38]),
        expectDeprecatedError5([6, 13, 6, 32]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
        class Bar extends PureComponent {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        };
      `,
      errors: [
        expectDeprecatedError3([3, 11, 3, 29]),
        expectDeprecatedError4([4, 11, 4, 36]),
        expectDeprecatedError5([5, 11, 5, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
        class Foo extends Component {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `,
      errors: [
        expectDeprecatedError3([3, 11, 3, 29]),
        expectDeprecatedError4([4, 11, 4, 36]),
        expectDeprecatedError5([5, 11, 5, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
        var Foo = createReactClass({
          componentWillMount: function() {},
          componentWillReceiveProps: function() {},
          componentWillUpdate: function() {}
        })
      `,
      errors: [
        expectDeprecatedError3([3, 11, 3, 29]),
        expectDeprecatedError4([4, 11, 4, 36]),
        expectDeprecatedError5([5, 11, 5, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `
        class Foo extends React.Component {
          constructor() {}
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `,
      errors: [
        expectDeprecatedError3([4, 11, 4, 29]),
        expectDeprecatedError4([5, 11, 5, 36]),
        expectDeprecatedError5([6, 11, 6, 30]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `
        import { hydrate } from 'react-dom';
        ReactDOM.hydrate(<div></div>, container);
      `,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'ReactDOM.hydrate is deprecated since React 18.0.0, use hydrateRoot instead, see https://reactjs.org/link/switch-to-createroot',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 25,
        },
        {
          messageId: 'deprecated',
          message:
            'ReactDOM.hydrate is deprecated since React 18.0.0, use hydrateRoot instead, see https://reactjs.org/link/switch-to-createroot',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 25,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `
        import { unmountComponentAtNode } from 'react-dom';
        ReactDOM.unmountComponentAtNode(container);
      `,
      errors: [
        {
          messageId: 'deprecated',
          message:
            'ReactDOM.unmountComponentAtNode is deprecated since React 18.0.0, use root.unmount instead, see https://reactjs.org/link/switch-to-createroot',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 40,
        },
        {
          messageId: 'deprecated',
          message:
            'ReactDOM.unmountComponentAtNode is deprecated since React 18.0.0, use root.unmount instead, see https://reactjs.org/link/switch-to-createroot',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 40,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
