import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const voidUseMemoMessage = (reason: string, description: string) =>
  `Error: ${reason}\n\n${description}.`;

const missingReturnMessage = voidUseMemoMessage(
  'useMemo() callbacks must return a value',
  "This useMemo() callback doesn't return a value. useMemo() is for computing and caching values, not for arbitrary side effects",
);

const unusedResultMessage = voidUseMemoMessage(
  'useMemo() result is unused',
  'This useMemo() value is unused. useMemo() is for computing and caching values, not for arbitrary side effects',
);

ruleTester.run('void-use-memo', {} as never, {
  valid: [
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
        function Component({items}) {
          const value = useMemo(() => {
            for (let item of items) {
              if (item.match) return item;
            }
            return null;
          }, [items]);
          return <div>{value}</div>;
        }
      `,
    },
    {
      code: `
        function Component(props) {
          // eslint-disable-next-line react-hooks/exhaustive-deps
          const x = useMemo(() => {
            let y;
            switch (props.switch) {
              case 'foo': {
                return 'foo';
              }
              case 'bar': {
                y = 'bar';
                break;
              }
              default: {
                y = props.y;
              }
            }
            return y;
          });
          return x;
        }
      `,
    },
  ],
  invalid: [
    {
      code: `
        function Component() {
          const value = useMemo(() => {
            console.log('computing');
          }, []);
          const value2 = React.useMemo(() => {
            console.log('computing');
          }, []);
          return (
            <div>
              {value}
              {value2}
            </div>
          );
        }
      `,
      errors: [
        { message: missingReturnMessage },
        { message: missingReturnMessage },
      ],
    },
    {
      code: `
        function Component() {
          useMemo(() => {
            return [];
          }, []);
          return <div />;
        }
      `,
      errors: [{ message: unusedResultMessage }],
    },
    {
      code: `
        /* eslint-disable react-hooks/exhaustive-deps */
        function component(a) {
          let x = useMemo(() => {
            mutate(a);
          }, []);
          return x;
        }
      `,
      errors: [{ message: missingReturnMessage }],
    },
  ],
});
