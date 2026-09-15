import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('sort-comp', {} as never, {
  valid: [
    {
      code: `class Hello extends React.Component {
        static displayName = 'Hello';
        componentDidMount() {}
        render() {}
      }`,
    },
    {
      code: `var Hello = createReactClass({
        displayName: 'Hello',
        onClick() {},
        render() {},
      });`,
    },
    {
      code: `class Hello extends React.Component {
        onClick() {}
        render() {}
      }`,
      options: [{ order: ['lifecycle', 'everything-else', 'render'] }],
    },
  ],
  invalid: [
    {
      code: `class Hello extends React.Component {
        render() {}
        componentDidMount() {}
      }`,
      errors: [
        {
          message: 'render should be placed after componentDidMount',
        },
      ],
    },
    {
      code: `var Hello = createReactClass({
        render() {},
        displayName: 'Hello',
      });`,
      errors: [
        {
          message: 'render should be placed after displayName',
        },
      ],
    },
    {
      code: `class Hello extends React.Component {
        onClick() {}
        render() {}
      }`,
      options: [{ order: ['lifecycle', 'render'] }],
      errors: [
        {
          message: 'onClick should be placed after render',
        },
      ],
    },
  ],
});
