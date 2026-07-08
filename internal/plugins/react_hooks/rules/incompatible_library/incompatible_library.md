# incompatible-library

## Rule Details

Validates against usage of libraries with APIs that are incompatible with memoization, including React Compiler's automatic memoization.

This rule currently reports known incompatible APIs from React Hook Form,
TanStack Table, and TanStack Virtual.

Like the upstream rule, it reports usage inside React Compiler target
components and hooks. Top-level module code and non-component helpers are
ignored.

Examples of **incorrect** code for this rule:

```javascript
import { useForm } from "react-hook-form";

function Form() {
  const { watch } = useForm();
  const name = watch("name");
  return <div>{name}</div>;
}
```

```javascript
import { useReactTable } from "@tanstack/react-table";

function Table({ data, columns }) {
  const table = useReactTable({ data, columns });
  return <Grid table={table} />;
}
```

Examples of **correct** code for this rule:

```javascript
import { useForm } from "react-hook-form";

function Form() {
  const { register } = useForm();
  return <input {...register("name")} />;
}
```

```javascript
function Component() {
  return <div />;
}
```

## Original Documentation

- [react.dev - incompatible-library](https://react.dev/reference/eslint-plugin-react-hooks/lints/incompatible-library)
- [Source code](https://github.com/facebook/react/blob/1ddff43c41147b880c22eb363e07aade5a71c5d9/compiler/packages/babel-plugin-react-compiler/src/HIR/DefaultModuleTypeProvider.ts)
