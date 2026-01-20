# rsrun

Runtime TypeScript register based on `@rslint/tsgo`.

## Usage

```bash
node -r @rslint/rsrun/register path/to/script.ts
```

```js
const { register } = require("@rslint/rsrun");

register({
  project: "tsconfig.json",
  module: "commonjs",
  target: "es2020",
  jsx: "react-jsx",
  typecheck: true,
});
```

## Environment

- `RSRUN_PROJECT`: path to `tsconfig.json` or `tsconfig.jsonc`
- `RSRUN_MODULE`: module kind override (e.g. `commonjs`, `esnext`, `nodenext`)
- `RSRUN_TARGET`: script target override (e.g. `es2020`, `esnext`)
- `RSRUN_JSX`: jsx emit override (e.g. `react`, `react-jsx`, `preserve`)
- `RSRUN_TYPECHECK`: `true`/`false` to enable typechecking (fail on errors)
- `RSRUN_INLINE_SOURCE_MAP`: `true`/`false` to toggle inline source maps
- `RSRUN_SOURCE_MAP`: `true`/`false` to emit source maps
