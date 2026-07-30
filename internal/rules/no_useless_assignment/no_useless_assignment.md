# no-useless-assignment

## Rule Details

Disallows assignments whose value is never read. A variable that is written and
then overwritten — or written on a path that ends without reading it again —
carries a value nothing observes, which usually means a mistake: a typo in the
variable name, a missing `return`, or leftover code from a refactor.

The rule only looks at variables it can follow end to end. It stays silent when
the variable is never read at all (that is `no-unused-vars`' job), when it is
read from another function, when it is exported, and when the assignment sits
inside a `try` block, where the block may be abandoned before the value is used.

Examples of **incorrect** code for this rule:

```javascript
function fn() {
	let v = 'used';
	console.log(v);
	v = 'unused';
}

function fn() {
	let v = 'unused';
	if (condition) {
		v = 'used';
		console.log(v);
		return;
	}
}

function fn() {
	let v = 'used';
	console.log(v);
	v = 'unused';
	v = 'used';
	console.log(v);
}
```

Examples of **correct** code for this rule:

```javascript
function fn() {
	let v = 'used';
	console.log(v);
	v = 'used-2';
	console.log(v);
}

function fn() {
	let v = 'used';
	if (condition) {
		v = 'used-2';
		console.log(v);
		return;
	}
	console.log(v);
}

function fn() {
	let v = 'used';
	console.log(v);
	setTimeout(() => console.log(v), 1);
	v = 'used in another scope';
}
```

## Differences from ESLint

- A variable read from inside a parameter decorator's arguments is treated as
  used, so assignments to it are never reported:
  `const pipe = {}; class C { handler(@Body(pipe) body) {} }`. ESLint reports
  `const pipe = {}` here.

## Original Documentation

- [ESLint no-useless-assignment](https://eslint.org/docs/latest/rules/no-useless-assignment)
