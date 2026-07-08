import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const preserveManualMemoizationMessage = (
  dependency: string,
  sourceDependencies: string,
  detail: string,
) =>
  `Compilation Skipped: Existing memoization could not be preserved\n\n` +
  `React Compiler has skipped optimizing this component because the existing manual memoization could not be preserved. ` +
  `The inferred dependencies did not match the manually specified dependencies, which could cause the value to change more or less frequently than expected. ` +
  `The inferred dependency was \`${dependency}\`, but the source dependencies were [${sourceDependencies}]. ${detail}`;

const preserveManualMemoizationMutationMessage =
  `Compilation Skipped: Existing memoization could not be preserved\n\n` +
  `React Compiler has skipped optimizing this component because the existing manual memoization could not be preserved. ` +
  `This dependency may be mutated later, which could cause the value to change unexpectedly.`;

ruleTester.run('preserve-manual-memoization', {} as never, {
  valid: [
    {
      code: `
        function Component({ data, filter }) {
          const filtered = useMemo(
            () => data.filter(filter),
            [data, filter]
          );
          return <List items={filtered} />;
        }
      `,
    },
    {
      code: `
        function Component({ onUpdate, value }) {
          const handleClick = useCallback(() => {
            onUpdate(value);
          }, [onUpdate, value]);
          return <button onClick={handleClick} />;
        }
      `,
    },
    {
      code: `
        function Component() {
          const ref = useRef(null);
          const value = useMemo(() => ref.current, []);
          return <div>{value}</div>;
        }
      `,
    },
  ],
  invalid: [
    {
      code: `
        function Component({ propA }) {
          const value = useMemo(() => {
            return propA.x();
          }, [propA.x]);
          return <span>{value}</span>;
        }
      `,
      errors: [
        {
          message: preserveManualMemoizationMessage(
            'propA',
            'propA.x',
            'Inferred less specific property than source.',
          ),
        },
        {
          message:
            "React Hook useMemo has a missing dependency: 'propA'. Either include it or remove the dependency array.",
        },
      ],
    },
    {
      code: `
        function Component({ items }) {
          const value = useMemo(() => items.length, [items]);
          items.push(1);
          return <span>{value}</span>;
        }
      `,
      errors: [
        {
          message: preserveManualMemoizationMutationMessage,
        },
      ],
    },
  ],
});
