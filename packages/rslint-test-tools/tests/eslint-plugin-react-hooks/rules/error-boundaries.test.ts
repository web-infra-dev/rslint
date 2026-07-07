import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const errorBoundariesMessage =
  'Error: Avoid constructing JSX within try/catch\n\n' +
  'React does not immediately render components when JSX is rendered, so any errors from this component will not be caught by the try/catch. ' +
  'To catch errors in rendering a given component, wrap that component in an error boundary. ' +
  '(https://react.dev/reference/react/Component#catching-rendering-errors-with-an-error-boundary).';

ruleTester.run('error-boundaries', {} as never, {
  valid: [
    {
      code: `
        function Parent() {
          return <ErrorBoundary><ChildComponent /></ErrorBoundary>;
        }
      `,
    },
    {
      code: `
        function Parent() {
          try {
            doSomething();
          } catch {
            return <div>Error occurred</div>;
          }
        }
      `,
    },
    {
      code: `
        function helper() {
          try {
            return <Child />;
          } catch {
            return null;
          }
        }
      `,
    },
    {
      code: `
        try {
          const el = <Child />;
        } catch {}
      `,
    },
  ],
  invalid: [
    {
      code: `
        function Parent() {
          try {
            return <ChildComponent />;
          } catch (error) {
            return <div>Error occurred</div>;
          }
        }
      `,
      errors: [{ message: errorBoundariesMessage }],
    },
    {
      code: `
        function Component() {
          let el;
          try {
            el = <div />;
          } catch {
            return null;
          }
          return el;
        }
      `,
      errors: [{ message: errorBoundariesMessage }],
    },
    {
      code: `
        function Component(props) {
          let el;
          try {
            let value;
            try {
              value = identity(props.foo);
            } catch {
              el = <div value={value} />;
            }
          } catch {
            return null;
          }
          return el;
        }
      `,
      errors: [{ message: errorBoundariesMessage }],
    },
    {
      code: `
        function Component() {
          try {
            return <>x</>;
          } catch {
            return null;
          }
        }
      `,
      errors: [{ message: errorBoundariesMessage }],
    },
  ],
});
