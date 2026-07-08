import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const useMemoMessage = (reason: string, description: string) =>
  `Error: ${reason}\n\n${description}.`;

const noParamsMessage = useMemoMessage(
  'useMemo() callbacks may not accept parameters',
  'useMemo() callbacks are called by React to cache calculations across re-renders. They should not take parameters. Instead, directly reference the props, state, or local variables needed for the computation',
);

const asyncGeneratorMessage = useMemoMessage(
  'useMemo() callbacks may not be async or generator functions',
  'useMemo() callbacks are called once and must synchronously return a value',
);

const reassignmentMessage = useMemoMessage(
  'useMemo() callbacks may not reassign variables declared outside of the callback',
  'useMemo() callbacks must be pure functions and cannot reassign variables defined outside of the callback function',
);

const arrayLiteralMessage = useMemoMessage(
  'Expected the dependency list for useMemo to be an array literal',
  'Expected the dependency list for useMemo to be an array literal',
);

const inlineFunctionMessage = useMemoMessage(
  'Expected the first argument to be an inline function expression',
  'Expected the first argument to be an inline function expression',
);

ruleTester.run('use-memo', {} as never, {
  valid: [
    {
      code: `
        function component(a) {
          let x = useMemo(() => [a], [a]);
          return <Foo x={x}></Foo>;
        }
      `,
    },
    {
      code: `
        import {useMemo as myMemo} from 'react';

        function Component({x}) {
          const v = myMemo(() => x * 2, [x]);
          return <div>{v}</div>;
        }
      `,
    },
    {
      code: `
        import {useMemo} from 'react';
        function Component(props) {
          return (
            // eslint-disable-next-line react-hooks/exhaustive-deps
            useMemo(() => {
              return [props.value];
            }) || []
          );
        }
      `,
    },
    {
      code: `
        function Component({ data }) {
          const processed = useMemo(() => {
            data.forEach(item => console.log(item));
          }, [data]);
          return <div>{processed}</div>;
        }
      `,
    },
    {
      code: `
        function Component() {
          const value = useMemo(() => computeValue(), []);
          return <div>{value}</div>;
        }
      `,
    },
    {
      code: `
        function Component(props) {
          // eslint-disable-next-line react-hooks/exhaustive-deps
          const value = useMemo(() => props.value);
          return <div>{value}</div>;
        }
      `,
    },
    {
      code: `
        function Component() {
          const value = useMemo(() => {
            return;
          }, []);
          return <div>{value}</div>;
        }
      `,
    },
    {
      code: `
        function Component() {
          const value = useMemo(() => {
            return null;
          }, []);
          return <div>{value}</div>;
        }
      `,
    },
    {
      code: `
        function Component({cond, a, b}) {
          const value = useMemo(() => {
            if (cond) {
              return a;
            }
            return b;
          }, [cond, a, b]);
          return <div>{value}</div>;
        }
      `,
    },
    {
      code: `
        function Component({kind, a, b}) {
          const value = useMemo(() => {
            switch (kind) {
              case 'a':
                return a;
              default:
                return b;
            }
          }, [kind, a, b]);
          return <div>{value}</div>;
        }
      `,
    },
    {
      code: `
        function Component({a}) {
          const value = useMemo(function named() {
            return a;
          }, [a]);
          return <div>{value}</div>;
        }
      `,
    },
  ],
  invalid: [
    {
      code: `
        function component(a, b) {
          let x = useMemo(c => c, []);
          return x;
        }
      `,
      errors: [{ message: noParamsMessage }],
    },
    {
      code: `
        function component(a, b) {
          let x = useMemo(async () => {
            return 1;
          }, []);
          return x;
        }
      `,
      errors: [{ message: asyncGeneratorMessage }],
    },
    {
      code: `
        function component(a, b) {
          let x = React.useMemo(async () => {
            return 1;
          }, []);
          return x;
        }
      `,
      errors: [{ message: asyncGeneratorMessage }],
    },
    {
      code: `
        /* eslint-disable react-hooks/exhaustive-deps */
        function component(a, b) {
          let x = useMemo(function* () {
            yield a;
          }, []);
          return x;
        }
      `,
      errors: [{ message: asyncGeneratorMessage }],
    },
    {
      code: `
        /* eslint-disable react-hooks/exhaustive-deps */
        function Component() {
          let x;
          const y = useMemo(() => {
            let z;
            x = [];
            z = true;
            return z;
          }, []);
          return [x, y];
        }
      `,
      errors: [{ message: reassignmentMessage }],
    },
    {
      code: `
        /* eslint-disable react-hooks/exhaustive-deps */
        import {useMemo} from 'react';

        function App({text, hasDeps}) {
          const resolvedText = useMemo(
            () => {
              return text.toUpperCase();
            },
            hasDeps ? null : [text],
          );
          return resolvedText;
        }
      `,
      errors: [{ message: arrayLiteralMessage }],
    },
    {
      code: `
        function Component(props) {
          // eslint-disable-next-line react-hooks/exhaustive-deps
          const x = useMemo(someHelper, []);
          return x;
        }
      `,
      errors: [{ message: inlineFunctionMessage }],
    },
  ],
});
