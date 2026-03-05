# no-misused-spread

## Rule Details

Disallows spread syntax (`...`) in places where the type being spread would produce unexpected behavior. This includes spreading strings in arrays (which can mishandle special characters and emojis), spreading arrays in object literals (producing a list of indices rather than values), spreading Promises in objects (which yields an empty object), spreading Maps in objects (also producing an empty object), spreading functions without properties, spreading class instances (losing the prototype), and spreading class declarations (only copying static properties).

Examples of **incorrect** code for this rule:

```typescript
// Spreading a string in an array
const chars = [...'hello'];

// Spreading an array in an object
const obj = { ...[1, 2, 3] };

// Spreading a Promise in an object
const data = { ...fetchData() };

// Spreading a Map in an object
const map = new Map([['a', 1]]);
const obj = { ...map };

// Spreading a class instance in an object
const instance = new MyClass();
const copy = { ...instance };
```

Examples of **correct** code for this rule:

```typescript
const arr = [...otherArray];

const obj = { ...otherObject };

const data = { ...(await fetchData()) };

const obj = Object.fromEntries(map);

const chars = Array.from(new Intl.Segmenter().segment('hello'));
```

## Original Documentation

- [typescript-eslint no-misused-spread](https://typescript-eslint.io/rules/no-misused-spread)
