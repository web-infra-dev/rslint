# no-unnecessary-type-conversion

## Rule Details

Disallow conversion idioms when they do not change the type or value of the expression. TypeScript already tracks the type of every expression, so calling `String(x)`, `Number(x)`, `Boolean(x)`, `BigInt(x)`, `.toString()`, `+x`, `!!x`, `~~x`, or concatenating with `''` when the source expression already has the target type is a no-op — it adds noise without changing the value or the type.

The rule never reports on `new String(...)`, `new Number(...)`, `new Boolean(...)`, or `new BigInt(...)`; those construct wrapper objects whose runtime type is `object`, not the primitive, so the call is not a no-op. It also opts out of the `.toString()` check when the receiver is an enum or enum member, since `.toString()` there is the documented way to read the enum's underlying string or number. A locally declared `String` / `Number` / `Boolean` / `BigInt` shadows the global and suppresses the call-form report.

Examples of **incorrect** code for this rule:

```typescript
String('asdf');
'asdf'.toString();
'asdf' + '';
'' + 'asdf';
let str = 'asdf';
str += '';

Number(123);
+123;
~~123;

Boolean(true);
!!true;

BigInt(3n);
```

Examples of **correct** code for this rule:

```typescript
String(1);
(1).toString();
`${1}`;
'' + 1;
1 + '';
let str = 1;
str += '';

Number('2');
+'2';
~~'2';
~~1.1;
~~(1 / 3);

Boolean(0);
!!0;

BigInt(3);

new String('asdf');
new Number(2);
new Boolean(true);
```

## Original Documentation

- [typescript-eslint no-unnecessary-type-conversion](https://typescript-eslint.io/rules/no-unnecessary-type-conversion)
