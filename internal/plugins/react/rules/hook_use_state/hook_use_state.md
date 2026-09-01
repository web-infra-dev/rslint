# Ensure destructuring and symmetric naming of useState hook value and setter variables (`react/hook-use-state`)

This rule checks that a React `useState` call is destructured into a value and
matching setter pair, such as `[color, setColor]`.

## Rule Details

Examples of **incorrect** code:

```js
import { useState } from "react";
const color = useState("blue");
const [color, updateColor] = useState("blue");
```

Examples of **correct** code:

```js
import { useState } from "react";
const [color, setColor] = useState("blue");
```

Returning a `useState` result directly is also allowed.

## Options

`allowDestructuredState` defaults to `false`. When enabled, the value part may
itself be destructured, provided the setter remains a simple binding.

```js
/* react/hook-use-state: ["error", { allowDestructuredState: true }] */
const [{ name }, setUser] = useState({ name: "Ada" });
```

## Suggestions

For a malformed pair, the rule can suggest a matching setter name. For a
single-value destructure with one initializer argument, it can also suggest a
`useMemo` replacement.

## References

- [eslint-plugin-react v7.37.5 documentation](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/hook-use-state.md)
