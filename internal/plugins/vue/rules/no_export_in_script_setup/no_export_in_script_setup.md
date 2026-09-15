# no-export-in-script-setup

## Rule Details

Disallow ES module exports inside `<script setup>`.

A `<script setup>` block is compiled into the component's `setup()` function
rather than kept as a module, so an export written in it has nothing to attach
to and the Vue compiler rejects the component. A plain `<script>` block beside
it is an ordinary module, and its exports are perfectly fine.

Examples of **incorrect** code for this rule:

```vue
<script setup>
export const value = 1;
</script>
```

```vue
<script setup>
export default { name: 'App' };
</script>
```

Examples of **correct** code for this rule:

```vue
<script setup>
const value = 1;
</script>
```

```vue
<!-- Exports belong in the plain block, which is a module. -->
<script>
export const shared = 1;
</script>

<script setup>
const local = 2;
</script>
```

A type-only export emits nothing, so it never reaches the Vue compiler and is
allowed:

```vue
<script setup lang="ts">
export type Value = number;
export interface Shape {
  a: number;
}
</script>
```

## When Not To Use It

There is no reason to turn this rule off: the code it reports does not compile.
