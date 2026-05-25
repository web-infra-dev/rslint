import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('rules-of-hooks', {} as never, {
  valid: [
    {
      code: `
        function ComponentWithHook() {
          useHook();
        }
      `,
    },
    {
      code: `
        function createComponentWithHook() {
          return function ComponentWithHook() {
            useHook();
          };
        }
      `,
    },
    {
      code: `
        function useHookWithHook() {
          useHook();
        }
      `,
    },
    {
      code: `
        function ComponentWithNormalFunction() {
          doSomething();
        }
      `,
    },
    {
      code: `
        function functionThatStartsWithUseButIsntAHook() {
          if (cond) {
            userFetch();
          }
        }
      `,
    },
    {
      code: `
        function useUnreachable() {
          return;
          useHook();
        }
      `,
    },
    {
      code: `
        function useHook() { useState(); }
        const whatever = function useHook() { useState(); };
        const useHook1 = () => { useState(); };
        let useHook2 = () => useState();
        useHook2 = () => { useState(); };
        ({useHook: () => { useState(); }});
        ({useHook() { useState(); }});
        const {useHook3 = () => { useState(); }} = {};
        ({useHook = () => { useState(); }} = {});
        Namespace.useHook = () => { useState(); };
      `,
    },
    {
      code: `
        const FancyButton = React.forwardRef((props, ref) => {
          useHook();
          return <button {...props} ref={ref} />
        });
      `,
    },
    {
      code: `
        const MemoizedFunction = React.memo(props => {
          useHook();
          return <button {...props} />
        });
      `,
    },
    {
      code: `
        class C {
          m() {
            this.useHook();
            super.useHook();
          }
        }
      `,
    },
    {
      code: `
        jest.useFakeTimers();
        beforeEach(() => {
          jest.useRealTimers();
        })
      `,
    },
    {
      code: `
        function App() {
          const text = use(Promise.resolve('A'));
          return <Text text={text} />
        }
      `,
    },
    {
      code: `
        function App() {
          let data = [];
          for (const query of queries) {
            const text = use(item);
            data.push(text);
          }
          return <Child data={data} />
        }
      `,
    },
    {
      code: `
        function MyComponent({ theme }) {
          const onClick = useEffectEvent(() => {
            showNotification(theme);
          });
          useEffect(() => {
            onClick();
          });
        }
      `,
    },
    {
      code: `
        function MyComponent({ theme }) {
          const onClick = useEffectEvent(() => {
            showNotification(theme);
          });
          useMyEffect(() => {
            onClick();
          });
        }
      `,
      settings: {
        'react-hooks': {
          additionalEffectHooks: '(useMyEffect|useServerEffect)',
        },
      },
    },
    {
      code: `
        function RegressionTest() {
          if (page == null) {
            throw new Error('oh no!');
          }
          useState();
        }
      `,
    },
  ],
  invalid: [
    {
      code: `
        function ComponentWithConditionalHook() {
          if (cond) {
            useConditionalHook();
          }
        }
      `,
      errors: [
        {
          message: `React Hook "useConditionalHook" is called conditionally. React Hooks must be called in the exact same order in every component render.`,
        },
      ],
    },
    {
      code: `
        Hook.useState();
        Hook.useHook();
      `,
      errors: 2,
    },
    {
      code: `
        class C {
          m() {
            This.useHook();
          }
        }
      `,
      errors: [
        {
          message: `React Hook "This.useHook" cannot be called in a class component. React Hooks must be called in a React function component or a custom React Hook function.`,
        },
      ],
    },
    {
      code: `
        function ComponentWithHookInsideLoop() {
          while (cond) {
            useHookInsideLoop();
          }
        }
      `,
      errors: [
        {
          message: `React Hook "useHookInsideLoop" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`,
        },
      ],
    },
    {
      code: `
        function useHook() {
          if (a) return;
          useState();
        }
      `,
      errors: [
        {
          message: `React Hook "useState" is called conditionally. React Hooks must be called in the exact same order in every component render. Did you accidentally call a React Hook after an early return?`,
        },
      ],
    },
    {
      code: `
        async function AsyncComponent() {
          useState();
        }
      `,
      errors: [
        {
          message: `React Hook "useState" cannot be called in an async function.`,
        },
      ],
    },
    {
      code: `
        function App({p1}) {
          try {
            use(p1);
          } catch (error) {
            console.error(error);
          }
          return <div>App</div>;
        }
      `,
      errors: [
        {
          message: `React Hook "use" cannot be called in a try/catch block.`,
        },
      ],
    },
    {
      code: `
        function MyComponent({ theme }) {
          const onClick = useEffectEvent(() => {
            showNotification(theme);
          });
          return <Child onClick={onClick}></Child>;
        }
      `,
      errors: [
        {
          message:
            '`onClick` is a function created with React Hook "useEffectEvent", and can only be called from Effects and Effect Events in the same component. It cannot be assigned to a variable or passed down.',
        },
      ],
    },
    {
      code: `
        function MyComponent({ theme }) {
          return <Child onClick={useEffectEvent(() => {
            showNotification(theme);
          })} />;
        }
      `,
      errors: [
        {
          message: `React Hook "useEffectEvent" can only be called at the top level of your component. It cannot be passed down.`,
        },
      ],
    },
  ],
});
