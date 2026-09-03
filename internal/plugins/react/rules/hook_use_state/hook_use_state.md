# hook-use-state

Require `useState` calls to be destructured into a value and matching setter
pair.

## Rule Details

This rule ensures that a React `useState` call uses a symmetric
`[value, setValue]` destructure. Returning a `useState` result directly is
allowed.

Examples of **incorrect** code for this rule:

```javascript
import { useState } from "react";

const color = useState("blue");
const [color, updateColor] = useState("blue");
```

Examples of **correct** code for this rule:

```javascript
import { useState } from "react";

const [color, setColor] = useState("blue");

function useColor() {
  return useState("blue");
}
```

## Options

`allowDestructuredState` defaults to `false`. When enabled, the value part may
itself be destructured, provided the setter remains a simple binding.

Examples of **correct** code for this rule with
`{ "allowDestructuredState": true }`:

```json
{ "react/hook-use-state": ["error", { "allowDestructuredState": true }] }
```

```javascript
import { useState } from "react";

const [{ name }, setUser] = useState({ name: "Ada" });
```

## Original Documentation

- [eslint-plugin-react: hook-use-state](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/hook-use-state.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/hook-use-state.js)
