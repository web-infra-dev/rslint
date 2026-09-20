# rslim/require-dynamic-import-entry

Require `@entry` on declarations used at runtime through dynamic imports so Rslim can preserve them.

When the import argument is a string literal or a template literal without substitutions, the rule resolves the module and checks the exports used by the importing code.

If the argument cannot be resolved, including variable arguments, template literals with substitutions, and missing modules, the rule reports a diagnostic asking you to check the imported declarations manually. Its severity follows the rule's `warn` or `error` configuration.

After checking all possible target modules and adding `@entry` where needed, manually add an ignore comment:

```ts
// rslint-disable-next-line rslim/require-dynamic-import-entry -- Checked @entry on the imported declarations.
const mod = await import(modulePath);
```

The diagnostic points to the import argument. For a multiline import, put the ignore comment immediately before the argument's line:

```ts
const mod = await import(
  // rslint-disable-next-line rslim/require-dynamic-import-entry -- Checked @entry on the imported declarations.
  modulePath
);
```

The rule does not automatically add annotations or ignore comments.
