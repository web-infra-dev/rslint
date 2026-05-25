import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('destructuring-assignment', {} as never, {
  valid: [
    {
      code: `
        const MyComponent = ({ id, className }) => (
          <div id={id} className={className} />
        );
      `,
    },
    {
      code: `
        const MyComponent = (props) => {
          const { id, className } = props;
          return <div id={id} className={className} />
        };
      `,
      options: ['always'],
    },
    {
      code: `
        const Foo = class extends React.PureComponent {
          render() {
            return <div>{this.props.foo}</div>;
          }
        };
      `,
      options: ['never'],
    },
    {
      code: `
        class Foo extends React.Component {
          bar = this.props.bar
        }
      `,
      options: ['always', { ignoreClassFields: true }],
    },
    {
      code: `
        function Foo(props) {
          const {a} = props;
          return <Goo {...props}>{a}</Goo>;
        }
      `,
      options: ['always', { destructureInSignature: 'always' }],
    },
  ],
  invalid: [
    {
      code: `
        const MyComponent = (props) => {
          return (<div id={props.id} />)
        };
      `,
      errors: [{ messageId: 'useDestructAssignment' }],
    },
    {
      code: `
        const MyComponent = ({ id, className }) => (
          <div id={id} className={className} />
        );
      `,
      options: ['never'],
      errors: [{ messageId: 'noDestructPropsInSFCArg' }],
    },
    {
      code: `
        const Foo = class extends React.PureComponent {
          render() {
            return <div>{this.props.foo}</div>;
          }
        };
      `,
      errors: [{ messageId: 'useDestructAssignment' }],
    },
    {
      code: `
        const Foo = class extends React.PureComponent {
          render() {
            return <div>{this.state.foo}</div>;
          }
        };
      `,
      errors: [{ messageId: 'useDestructAssignment' }],
    },
    {
      code: `
        const Foo = class extends React.PureComponent {
          render() {
            const { foo } = this.props;
            return <div>{foo}</div>;
          }
        };
      `,
      options: ['never'],
      errors: [{ messageId: 'noDestructAssignment' }],
    },
    {
      code: `
        function Foo(props) {
          const {a} = props;
          return <p>{a}</p>;
        }
      `,
      options: ['always', { destructureInSignature: 'always' }],
      errors: [{ messageId: 'destructureInSignature' }],
    },
  ],
});
