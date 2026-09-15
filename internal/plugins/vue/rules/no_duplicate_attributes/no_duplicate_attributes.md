# no-duplicate-attributes

## Rule Details

Disallow naming the same attribute twice on one element.

Vue keeps the first of two identical attributes and silently drops the second,
so a duplicate is always either dead code or a bug. The comparison spans both
spellings, because `v-bind:foo` binds the attribute `foo`: writing `foo` and
`:foo` on one element names the same thing twice.

Examples of **incorrect** code for this rule:

```vue
<template>
  <div foo="a" foo="b" />
  <div foo="a" :foo="b" />
  <div :foo="a" v-bind:foo="b" />
  <div class="a" class="b" />
</template>
```

Examples of **correct** code for this rule:

```vue
<template>
  <div foo="a" bar="b" />

  <!-- Vue merges a static class with a bound one, so this is the idiom. -->
  <div class="a" :class="b" />

  <!-- The keys of a spread are only known at run time. -->
  <div v-bind="attrs" foo="a" />

  <!-- Two handlers for one event both run. -->
  <div @click="a" @click="b" />
</template>
```

## Options

```json
{
  "vue/no-duplicate-attributes": [
    "error",
    { "allowCoexistClass": true, "allowCoexistStyle": true }
  ]
}
```

- `allowCoexistClass` (default `true`): allow a static `class` beside a bound
  `:class`.
- `allowCoexistStyle` (default `true`): allow a static `style` beside a bound
  `:style`.

Neither allowance permits two attributes of the *same* kind: `class="a"
class="b"` and `:class="a" :class="b"` are duplicates however these are set.

## Differences from ESLint

None known.
