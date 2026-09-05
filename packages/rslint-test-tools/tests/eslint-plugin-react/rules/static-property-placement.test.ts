import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('static-property-placement', {} as never, {
  valid: [
    {
      code: `class MyComponent extends React.Component {
        static propTypes = {};
        static defaultProps = {};
        static contextType = Context;
        static contextTypes = {};
        static childContextTypes = {};
        static displayName = 'MyComponent';
      }`,
    },
    {
      code: `class MyComponent extends React.Component {
        static get propTypes() { return {}; }
        static get defaultProps() { return {}; }
        static get contextType() { return Context; }
      }`,
      options: ['static getter'],
    },
    {
      code: `class MyComponent extends React.Component {}
        MyComponent.propTypes = {};
        MyComponent.defaultProps = {};`,
      options: ['property assignment'],
    },
  ],
  invalid: [
    {
      code: `class MyComponent extends React.Component {
        static get propTypes() { return {}; }
      }`,
      errors: [{ messageId: 'notStaticClassProp' }],
    },
    {
      code: `class MyComponent extends React.Component {
        static propTypes = {};
      }`,
      options: ['static getter'],
      errors: [{ messageId: 'notGetterClassFunc' }],
    },
    {
      code: `class MyComponent extends React.Component {}
        MyComponent.propTypes = {};`,
      errors: [{ messageId: 'declareOutsideClass' }],
    },
    {
      code: `class MyComponent extends React.Component {
        static displayName = 'x';
        static get propTypes() { return {}; }
      }
      MyComponent.defaultProps = {};`,
      options: ['property assignment'],
      errors: [
        { messageId: 'declareOutsideClass' },
        { messageId: 'declareOutsideClass' },
      ],
    },
  ],
});
