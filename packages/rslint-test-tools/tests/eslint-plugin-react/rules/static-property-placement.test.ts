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
    {
      code: `const Box = { C: class extends React.Component {} };
        Box.C = {};
        Box.C.propTypes = {};`,
    },
    {
      code: `const Box = { C: class extends React.Component {} } as const;
        Box.C.propTypes = {};`,
    },
    {
      code: `/** @extends {React.Component} */
        class C { propTypes = {}; }`,
      filename: 'files/static-property-placement.jsx',
    },
    {
      code: `/** @extends React.Component */
        const C = class {
          // eslint-disable-next-line react/static-property-placement
          propTypes = {};
        };`,
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
      errors: [
        {
          messageId: 'notStaticClassProp',
          message: "'propTypes' should be declared as a static class property.",
        },
      ],
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
    {
      code: `/** @extends React.Component */
        const C = class { propTypes = {}; };`,
      filename: 'files/static-property-placement.jsx',
      errors: [
        {
          messageId: 'notStaticClassProp',
          message: "'propTypes' should be declared as a static class property.",
        },
      ],
    },
    {
      code: `const Box = { C: class extends React.Component {} };
        (Box).C.propTypes = {};`,
      options: ['static getter'],
      errors: [
        {
          messageId: 'notGetterClassFunc',
          message:
            "'propTypes' should be declared as a static getter class function.",
        },
      ],
    },
    {
      code: `/* @jsx Preact.h */
        class C extends Preact.Component { propTypes = {}; }`,
      errors: [
        {
          messageId: 'notStaticClassProp',
          message: "'propTypes' should be declared as a static class property.",
        },
      ],
    },
  ],
});
