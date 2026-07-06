import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('component-hook-factories', {} as never, {
  valid: [
    {
      code: `
        function Component({ defaultValue }) {
          return <span>{defaultValue}</span>;
        }
      `,
    },
    {
      code: `
        function useData(endpoint) {
          return useState(endpoint)[0];
        }
      `,
    },
    {
      code: `
        function Button({color, children}) {
          return (
            <button style={{backgroundColor: color}}>
              {children}
            </button>
          );
        }

        function App() {
          return (
            <>
              <Button color="red">Red</Button>
              <Button color="blue">Blue</Button>
            </>
          );
        }
      `,
    },
    {
      code: `
        function createComponent(defaultValue) {
          return function Component() {
            return null;
          };
        }
      `,
    },
  ],
  invalid: [
    {
      code: `
        function createComponent(defaultValue) {
          return function Component() {
            return <span>{defaultValue}</span>;
          };
        }
      `,
      errors: [
        {
          message:
            "Components and hooks cannot be created dynamically. The function `Component` appears to be a React component, but it's defined inside `createComponent`. Components and Hooks should always be declared at module scope.",
        },
      ],
    },
    {
      code: `
        function Parent() {
          function Child() {
            return <div />;
          }
          return <Child />;
        }
      `,
      errors: [
        {
          message:
            "Components and hooks cannot be created dynamically. The function `Child` appears to be a React component, but it's defined inside `Parent`. Components and Hooks should always be declared at module scope.",
        },
      ],
    },
    {
      code: `
        function createCustomHook(endpoint) {
          return function useData() {
            return useState(endpoint)[0];
          };
        }
      `,
      errors: [
        {
          message:
            "Components and hooks cannot be created dynamically. The function `useData` appears to be a React hook, but it's defined inside `createCustomHook`. Components and Hooks should always be declared at module scope.",
        },
      ],
    },
    {
      code: `
        function makeButton(color) {
          return function Button({children}) {
            return (
              <button style={{backgroundColor: color}}>
                {children}
              </button>
            );
          };
        }

        const RedButton = makeButton('red');
        const BlueButton = makeButton('blue');
      `,
      errors: 1,
    },
  ],
});
