import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectLifecycleError1(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'lifecycle',
    'render is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
    range,
  );
}

function expectLifecycleError2(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'lifecycle',
    'getInitialState is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

const instanceLifecycles = [
  'getDefaultProps',
  'getInitialState',
  'getChildContext',
  'componentWillMount',
  'UNSAFE_componentWillMount',
  'componentDidMount',
  'componentWillReceiveProps',
  'UNSAFE_componentWillReceiveProps',
  'shouldComponentUpdate',
  'componentWillUpdate',
  'UNSAFE_componentWillUpdate',
  'getSnapshotBeforeUpdate',
  'componentDidUpdate',
  'componentDidCatch',
  'componentWillUnmount',
  'render',
];

const message = (name: string) =>
  `${name} is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.`;

ruleTester.run('no-arrow-function-lifecycle', {
  valid: [
    {
      code: 'var Hello = createReactClass({ render: function() { return null; } });',
    },
    {
      code: 'var Hello = createReactClass({ getDerivedStateFromProps: () => null });',
    },
    {
      code: 'class Hello extends React.Component { render() { return null; } }',
    },
    {
      code: 'class Hello extends React.Component { getDerivedStateFromProps = () => null; }',
    },
    { code: 'class Hello extends React.Component { onChange: () => void; }' },
    ...instanceLifecycles.map((name) => ({
      code: `class Hello extends React.Component { ${name}() { return null; } }`,
    })),

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        var Hello = createReactClass({
          getDefaultProps: function() { return {}; },
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        var Hello = createReactClass({
          getInitialState: function() { return {}; },
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        var Hello = createReactClass({
          getChildContext: function() { return {}; },
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        var Hello = createReactClass({
          getDerivedStateFromProps: function() { return {}; },
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillMount: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        var Hello = createReactClass({
          UNSAFE_componentWillMount: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillReceiveProps: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        var Hello = createReactClass({
          UNSAFE_componentWillReceiveProps: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        var Hello = createReactClass({
          shouldComponentUpdate: function() { return true; },
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillUpdate: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        var Hello = createReactClass({
          UNSAFE_componentWillUpdate: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        var Hello = createReactClass({
          getSnapshotBeforeUpdate: function() { return {}; },
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentDidCatch: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillUnmount: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDefaultProps() { return {}; }
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getInitialState() { return {}; }
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getChildContext() { return {}; }
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDerivedStateFromProps() { return {}; }
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillMount() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillMount() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidMount() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillReceiveProps() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillReceiveProps() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          shouldComponentUpdate() { return true; }
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUpdate() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillUpdate() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getSnapshotBeforeUpdate() { return {}; }
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidUpdate() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidCatch() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUnmount() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDerivedStateFromProps = () => { return {}; } // not a lifecycle method
          static getDerivedStateFromProps() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
        var Hello = createReactClass({
          getDerivedStateFromProps: () => { return {}; },
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `
        class MyComponent extends React.Component {
          onChange: () => void;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    ...instanceLifecycles.map((name) => ({
      code: `var Hello = createReactClass({ ${name}: () => { return null; } });`,
      errors: [
        diagnostic('lifecycle', message(name), [1, 32, 1, name.length + 56]),
      ],
      output: `var Hello = createReactClass({ ${name}: function() { return null; } });`,
    })),
    ...instanceLifecycles.map((name) => ({
      code: `class Hello extends React.Component { ${name} = () => { return null; } }`,
      errors: [
        diagnostic('lifecycle', message(name), [1, 39, 1, name.length + 64]),
      ],
      output: `class Hello extends React.Component { ${name}() { return null; } }`,
    })),
    {
      code: 'class Hello extends React.Component { static getDerivedStateFromProps = () => { return null; } }',
      errors: [
        diagnostic(
          'lifecycle',
          'getDerivedStateFromProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [1, 39, 1, 95],
        ),
      ],
      output: `class Hello extends React.Component { static getDerivedStateFromProps() { return null; } }`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: () => { return <div />; }
        });
      `,
      errors: [expectLifecycleError1([3, 11, 3, 44])],
      output: `
        var Hello = createReactClass({
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        var Hello = createReactClass({
          getDefaultProps: () => { return {}; },
          render: function() { return <div />; }
        });
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'getDefaultProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [3, 11, 3, 48],
        ),
      ],
      output: `
        var Hello = createReactClass({
          getDefaultProps: function() { return {}; },
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        var Hello = createReactClass({
          getInitialState: () => { return {}; },
          render: function() { return <div />; }
        });
      `,
      errors: [expectLifecycleError2([3, 11, 3, 48])],
      output: `
        var Hello = createReactClass({
          getInitialState: function() { return {}; },
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        var Hello = createReactClass({
          getChildContext: () => { return {}; },
          render: function() { return <div />; }
        });
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'getChildContext is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [3, 11, 3, 48],
        ),
      ],
      output: `
        var Hello = createReactClass({
          getChildContext: function() { return {}; },
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillMount: () => {},
          render: function() { return <div />; }
        });
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'componentWillMount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [3, 11, 3, 39],
        ),
      ],
      output: `
        var Hello = createReactClass({
          componentWillMount: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        var Hello = createReactClass({
          UNSAFE_componentWillMount: () => {},
          render: function() { return <div />; }
        });
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'UNSAFE_componentWillMount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [3, 11, 3, 46],
        ),
      ],
      output: `
        var Hello = createReactClass({
          UNSAFE_componentWillMount: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentDidMount: () => {},
          render: function() { return <div />; }
        });
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'componentDidMount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [3, 11, 3, 38],
        ),
      ],
      output: `
        var Hello = createReactClass({
          componentDidMount: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillReceiveProps: () => {},
          render: function() { return <div />; }
        });
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'componentWillReceiveProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [3, 11, 3, 46],
        ),
      ],
      output: `
        var Hello = createReactClass({
          componentWillReceiveProps: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        var Hello = createReactClass({
          UNSAFE_componentWillReceiveProps: () => {},
          render: function() { return <div />; }
        });
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'UNSAFE_componentWillReceiveProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [3, 11, 3, 53],
        ),
      ],
      output: `
        var Hello = createReactClass({
          UNSAFE_componentWillReceiveProps: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        var Hello = createReactClass({
          shouldComponentUpdate: () => { return true; },
          render: function() { return <div />; }
        });
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'shouldComponentUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [3, 11, 3, 56],
        ),
      ],
      output: `
        var Hello = createReactClass({
          shouldComponentUpdate: function() { return true; },
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillUpdate: () => {},
          render: function() { return <div />; }
        });
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'componentWillUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [3, 11, 3, 40],
        ),
      ],
      output: `
        var Hello = createReactClass({
          componentWillUpdate: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        var Hello = createReactClass({
          UNSAFE_componentWillUpdate: () => {},
          render: function() { return <div />; }
        });
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'UNSAFE_componentWillUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [3, 11, 3, 47],
        ),
      ],
      output: `
        var Hello = createReactClass({
          UNSAFE_componentWillUpdate: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        var Hello = createReactClass({
          getSnapshotBeforeUpdate: () => { return {}; },
          render: function() { return <div />; }
        });
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'getSnapshotBeforeUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [3, 11, 3, 56],
        ),
      ],
      output: `
        var Hello = createReactClass({
          getSnapshotBeforeUpdate: function() { return {}; },
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentDidUpdate: () => {},
          render: function() { return <div />; }
        });
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'componentDidUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [3, 11, 3, 39],
        ),
      ],
      output: `
        var Hello = createReactClass({
          componentDidUpdate: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentDidCatch: () => {},
          render: function() { return <div />; }
        });
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'componentDidCatch is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [3, 11, 3, 38],
        ),
      ],
      output: `
        var Hello = createReactClass({
          componentDidCatch: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentWillUnmount: () => {},
          render: function() { return <div />; }
        });
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'componentWillUnmount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [3, 11, 3, 41],
        ),
      ],
      output: `
        var Hello = createReactClass({
          componentWillUnmount: function() {},
          render: function() { return <div />; }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          render = () => { return <div />; }
        }
      `,
      errors: [expectLifecycleError1([4, 11, 4, 45])],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDefaultProps = () => { return {}; }
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'getDefaultProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 49],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDefaultProps() { return {}; }
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getInitialState = () => { return {}; }
          render = () => { return <div />; }
        }
      `,
      errors: [
        expectLifecycleError2([4, 11, 4, 49]),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getInitialState() { return {}; }
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getChildContext = () => { return {}; }
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'getChildContext is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 49],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getChildContext() { return {}; }
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          static getDerivedStateFromProps = () => { return {}; }
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'getDerivedStateFromProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 65],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          static getDerivedStateFromProps() { return {}; }
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillMount = () => {}
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'componentWillMount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 40],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillMount() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillMount = () => {}
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'UNSAFE_componentWillMount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 47],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillMount() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidMount = () => {}
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'componentDidMount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 39],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidMount() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillReceiveProps = () => {}
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'componentWillReceiveProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 47],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillReceiveProps() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillReceiveProps = () => {}
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'UNSAFE_componentWillReceiveProps is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 54],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillReceiveProps() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          shouldComponentUpdate = () => { return true; }
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'shouldComponentUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 57],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          shouldComponentUpdate() { return true; }
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUpdate = () => {}
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'componentWillUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 41],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUpdate() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillUpdate = () => {}
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'UNSAFE_componentWillUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 48],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillUpdate() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getSnapshotBeforeUpdate = () => { return {}; }
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'getSnapshotBeforeUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 57],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getSnapshotBeforeUpdate() { return {}; }
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidUpdate = (prevProps) => {}
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'componentDidUpdate is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 49],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidUpdate(prevProps) {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidCatch = () => {}
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'componentDidCatch is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 39],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidCatch() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUnmount = () => {}
          render = () => { return <div />; }
        }
      `,
      errors: [
        diagnostic(
          'lifecycle',
          'componentWillUnmount is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead.',
          [4, 11, 4, 42],
        ),
        expectLifecycleError1([5, 11, 5, 45]),
      ],
      output: `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUnmount() {}
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render = () => <div />
        }
      `,
      errors: [expectLifecycleError1([3, 11, 3, 33])],
      output: `
        class Hello extends React.Component {
          render() { return <div />; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
        class Hello extends React.Component {
          render = () => /*first*/<div />/*second*/
        }
      `,
      errors: [expectLifecycleError1([3, 11, 3, 42])],
      output: `
        class Hello extends React.Component {
          render() { return /*first*/<div />/*second*/; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
        export default class Root extends Component {
          getInitialState = () => ({
            errorImporting: null,
            errorParsing: null,
            errorUploading: null,
            file: null,
            fromExtension: false,
            importSuccess: false,
            isImporting: false,
            isParsing: false,
            isUploading: false,
            parsedResults: null,
            showLongRunningMessage: false,
          });
        }
      `,
      errors: [expectLifecycleError2([3, 11, 15, 14])],
      output: `
        export default class Root extends Component {
          getInitialState() { return {
            errorImporting: null,
            errorParsing: null,
            errorUploading: null,
            file: null,
            fromExtension: false,
            importSuccess: false,
            isImporting: false,
            isParsing: false,
            isUploading: false,
            parsedResults: null,
            showLongRunningMessage: false,
          }; }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
});
