import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-deprecated', {} as never, {
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
  ],
  invalid: [
    // ---- Member access ----
    {
      code: `React.renderComponent()`,
      errors: [{ messageId: 'deprecated' }],
    },
    {
      code: `this.transferPropsTo()`,
      errors: [{ messageId: 'deprecated' }],
    },
    {
      code: `React.addons.TestUtils`,
      errors: [{ messageId: 'deprecated' }],
    },
    {
      code: `React.render(element, container);`,
      errors: [{ messageId: 'deprecated' }],
    },
    {
      code: `React.createClass({});`,
      errors: [{ messageId: 'deprecated' }],
    },
    {
      code: `React.PropTypes`,
      errors: [{ messageId: 'deprecated' }],
    },
    {
      code: `React.DOM.div`,
      errors: [{ messageId: 'deprecated' }],
    },

    // ---- Destructuring ----
    {
      code: `var {createClass} = require('react');`,
      errors: [{ messageId: 'deprecated' }],
    },
    {
      code: `var {createClass, PropTypes} = require('react');`,
      errors: [{ messageId: 'deprecated' }, { messageId: 'deprecated' }],
    },

    // ---- Imports ----
    {
      code: `import {createClass} from 'react';`,
      errors: [{ messageId: 'deprecated' }],
    },
    {
      code: `import {printDOM} from 'react-addons-perf';`,
      errors: [{ messageId: 'deprecated' }],
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
        { messageId: 'deprecated' },
        { messageId: 'deprecated' },
        { messageId: 'deprecated' },
      ],
    },
    {
      code: `
        var Foo = createReactClass({
          componentWillMount: function() {}
        })
      `,
      errors: [{ messageId: 'deprecated' }],
    },

    // ---- React 18 deprecations ----
    {
      code: `
        import { render } from 'react-dom';
        ReactDOM.render(<div></div>, container);
      `,
      errors: [{ messageId: 'deprecated' }, { messageId: 'deprecated' }],
    },
    {
      code: `
        import { renderToNodeStream } from 'react-dom/server';
        ReactDOMServer.renderToNodeStream(element);
      `,
      errors: [{ messageId: 'deprecated' }, { messageId: 'deprecated' }],
    },
  ],
});
