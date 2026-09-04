import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

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

ruleTester.run('no-arrow-function-lifecycle', {} as never, {
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
  ],
  invalid: [
    ...instanceLifecycles.map((name) => ({
      code: `var Hello = createReactClass({ ${name}: () => { return null; } });`,
      errors: [{ message: message(name) }],
    })),
    ...instanceLifecycles.map((name) => ({
      code: `class Hello extends React.Component { ${name} = () => { return null; } }`,
      errors: [{ message: message(name) }],
    })),
    {
      code: 'class Hello extends React.Component { static getDerivedStateFromProps = () => { return null; } }',
      errors: [{ message: message('getDerivedStateFromProps') }],
    },
  ],
});
