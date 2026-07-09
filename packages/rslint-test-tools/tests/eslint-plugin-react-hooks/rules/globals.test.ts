import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const globalMessage = (name: string) =>
  `Error: Cannot reassign variables declared outside of the component/hook\n\nVariable \`${name}\` is declared outside of the component/hook. Reassigning this value during render is a form of side effect, which can cause unpredictable behavior depending on when the component happens to re-render. If this variable is used in rendering, use useState instead. Otherwise, consider updating it in an effect. (https://react.dev/reference/rules/components-and-hooks-must-be-pure#side-effects-must-run-outside-of-render).`;

ruleTester.run('globals', {} as never, {
  valid: [
    {
      code: `
        function Component() {
          const onClick = () => {
            someUnknownGlobal = true;
            moduleLocal = true;
          };
          return <div onClick={onClick} />;
        }
      `,
    },
    {
      code: `
        function Component(props) {
          const x = {};
          const y = Boolean(x);
          return [x, y];
        }
      `,
    },
    {
      code: `
        function Component(props) {
          const x = {};
          const y = Number(x);
          return [x, y];
        }
      `,
    },
    {
      code: `
        function Component(props) {
          const x = {};
          const y = String(x);
          return [x, y];
        }
      `,
    },
    {
      code: `
        import {useState as _useState, useCallback, useEffect} from 'react';

        function useState(value) {
          const [state, setState] = _useState(value);
          return [state, setState];
        }

        function Component() {
          const [state, setState] = useState('hello');

          return <div onClick={() => setState('goodbye')}>{state}</div>;
        }
      `,
    },
    {
      code: `
        function Component() {
          someUnknownGlobal = true;
          return null;
        }
      `,
    },
    {
      code: `
        function Component() {
          let moduleLocal;
          moduleLocal = true;
          return <div />;
        }
      `,
    },
    {
      code: `
        function Component(moduleLocal) {
          moduleLocal = true;
          return <div />;
        }
      `,
    },
    {
      code: `
        function Component() {
          window.location.href = "/";
          return <div />;
        }
      `,
    },
    {
      code: `
        function Component() {
          const onClick = () => {
            someGlobal = true;
          };
          return <button onClick={onClick} />;
        }
      `,
    },
    {
      code: `
        function Component() {
          useEffect(() => {
            someGlobal = true;
          });
          return <div />;
        }
      `,
    },
    {
      code: `
        function Component() {
          useCallback(() => {
            someGlobal = true;
          }, []);
          return <div />;
        }
      `,
    },
    {
      code: `
        function Component() {
          someGlobal++;
          ++otherGlobal;
          return <div />;
        }
      `,
    },
  ],
  invalid: [
    {
      code: `
        function Component() {
          someUnknownGlobal = true;
          moduleLocal = true;
          return <div />;
        }
      `,
      errors: [
        { message: globalMessage('someUnknownGlobal') },
        { message: globalMessage('moduleLocal') },
      ],
    },
    {
      code: `
        // @enableNewMutationAliasingModel
        function Component() {
          someUnknownGlobal = true;
          moduleLocal = true;
          return <div />;
        }
      `,
      errors: [
        { message: globalMessage('someUnknownGlobal') },
        { message: globalMessage('moduleLocal') },
      ],
    },
    {
      code: `
        let moduleLocal;
        function Component() {
          moduleLocal = true;
          return <div />;
        }
      `,
      errors: [{ message: globalMessage('moduleLocal') }],
    },
    {
      code: `
        function Component() {
          const foo = () => {
            someUnknownGlobal = true;
            moduleLocal = true;
          };
          foo();
          return <div />;
        }
      `,
      errors: [
        { message: globalMessage('someUnknownGlobal') },
        { message: globalMessage('moduleLocal') },
      ],
    },
    {
      code: `
        // @enableNewMutationAliasingModel
        function Component() {
          const foo = () => {
            someUnknownGlobal = true;
            moduleLocal = true;
          };
          foo();
          return <div />;
        }
      `,
      errors: [
        { message: globalMessage('someUnknownGlobal') },
        { message: globalMessage('moduleLocal') },
      ],
    },
    {
      code: `
        function Component() {
          const foo = () => {
            someGlobal = true;
          };
          return <Foo>{foo}</Foo>;
        }
      `,
      errors: [{ message: globalMessage('someGlobal') }],
    },
    {
      code: `
        function Component() {
          useMemo(() => {
            someGlobal = true;
            return 1;
          }, []);
          return <div />;
        }
      `,
      errors: [{ message: globalMessage('someGlobal') }],
    },
  ],
});
