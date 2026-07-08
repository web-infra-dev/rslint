import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const incompatibleMessage = (detail: string) =>
  `Compilation Skipped: Use of incompatible library\n\n` +
  `This API returns functions which cannot be memoized without leading to stale UI. ` +
  `To prevent this, by default React Compiler will skip memoizing this component/hook. ` +
  `However, you may see issues if values from this API are passed to other components/hooks that are memoized.\n\n` +
  detail;

const reactHookFormWatchMessage = incompatibleMessage(
  "React Hook Form's `useForm()` API returns a `watch()` function which cannot be memoized safely.",
);

const tanStackTableMessage = incompatibleMessage(
  "TanStack Table's `useReactTable()` API returns functions that cannot be memoized safely",
);

const tanStackVirtualMessage = incompatibleMessage(
  "TanStack Virtual's `useVirtualizer()` API returns functions that cannot be memoized safely",
);

ruleTester.run('incompatible-library', {} as never, {
  valid: [
    {
      code: `
        import { useForm } from 'react-hook-form';
        function Form() {
          const form = useForm();
          return <form />;
        }
      `,
    },
    {
      code: `
        import { useForm } from 'react-hook-form';
        function Form() {
          const { register } = useForm();
          return <input {...register('name')} />;
        }
      `,
    },
    {
      code: `
        import { useReactTable } from '@example/react-table';
        function Table() {
          const table = useReactTable();
          return <Grid table={table} />;
        }
      `,
    },
    {
      code: `
        import { useWatch } from 'react-hook-form';
        function Form() {
          const name = useWatch({ name: 'name' });
          return <div>{name}</div>;
        }
      `,
    },
    {
      code: `
        import { observer } from 'mobx-react-lite';
        const TodoView = observer(function TodoView({ todo }) {
          return <div>{todo.title}</div>;
        });
      `,
    },
  ],
  invalid: [
    {
      code: `
        import { useForm } from 'react-hook-form';
        function Form() {
          const form = useForm();
          const name = form.watch('name');
          return <div>{name}</div>;
        }
      `,
      options: [{}],
      errors: [{ message: reactHookFormWatchMessage }],
    },
    {
      code: `
        import { useForm } from 'react-hook-form';
        function Form() {
          const { watch } = useForm();
          const name = watch('name');
          return <div>{name}</div>;
        }
      `,
      errors: [{ message: reactHookFormWatchMessage }],
    },
    {
      code: `
        import { useMemo } from 'react';
        import { useForm } from 'react-hook-form';
        function Form() {
          const { watch } = useForm();
          const name = useMemo(() => watch('name'), [watch]);
          return <div>{name}</div>;
        }
      `,
      errors: [{ message: reactHookFormWatchMessage }],
    },
    {
      code: `
        import { useReactTable, getCoreRowModel } from '@tanstack/react-table';
        function Table({ data, columns }) {
          const table = useReactTable({
            data,
            columns,
            getCoreRowModel: getCoreRowModel(),
          });
          return <Grid table={table} />;
        }
      `,
      errors: [{ message: tanStackTableMessage }],
    },
    {
      code: `
        import { useVirtualizer } from '@tanstack/react-virtual';
        function List({ count }) {
          const virtualizer = useVirtualizer({ count });
          return <Grid virtualizer={virtualizer} />;
        }
      `,
      errors: [{ message: tanStackVirtualMessage }],
    },
  ],
});
