import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const immutableMessage = (description: string) =>
  `Error: This value cannot be modified\n\n${description}`;

const stateMessage = immutableMessage(
  "Modifying a value returned from 'useState()', which should not be modified directly. Use the setter function to update instead.",
);

const propsMessage = immutableMessage(
  'Modifying component props or hook arguments is not allowed. Consider using a local variable instead.',
);

ruleTester.run('immutability', {} as never, {
  valid: [
    {
      code: `
        function Button(props) {
          return null;
        }
      `,
    },
    {
      code: `
        function Button(props) {
          const scrollView = React.useRef<ScrollView>(null);
          return <Button thing={scrollView} />;
        }
      `,
    },
    {
      code: `
        function helper(obj) {
          obj.key = 'value';
          return obj;
        }
      `,
    },
    {
      code: `
        const processData = (input) => {
          input.modified = true;
          return input;
        };
      `,
    },
  ],
  invalid: [
    {
      code: `
        import { useState } from 'react';
        function Component(props) {
          const x: \`foo\${1}\` = 'foo1';
          const [state, setState] = useState({a: 0});
          state.a = 1;
          return <div>{props.foo}</div>;
        }
      `,
      errors: [{ message: stateMessage }],
    },
    {
      code: `
        function MyComponent({a}) {
          a.key = 'value';
          return <div />;
        }
      `,
      errors: [{ message: propsMessage }],
    },
    {
      code: `
        const MyComponent = ({a}) => {
          a.key = 'value';
          return <div />;
        };
      `,
      errors: [{ message: propsMessage }],
    },
    {
      code: `
        const MyComponent = function({a}) {
          a.key = 'value';
          return <div />;
        };
      `,
      errors: [{ message: propsMessage }],
    },
    {
      code: `
        export function MyComponent({a}) {
          a.key = 'value';
          return <div />;
        }
      `,
      errors: [{ message: propsMessage }],
    },
    {
      code: `
        export const MyComponent = ({a}) => {
          a.key = 'value';
          return <div />;
        };
      `,
      errors: [{ message: propsMessage }],
    },
    {
      code: `
        export default function MyComponent({a}) {
          a.key = 'value';
          return <div />;
        }
      `,
      errors: [{ message: propsMessage }],
    },
  ],
});
