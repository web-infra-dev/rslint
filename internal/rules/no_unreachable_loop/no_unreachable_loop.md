# no-unreachable-loop

## Rule Details

Disallows a loop whose body can never run a second time. Every path through the
body leaves the loop — by `break`, `return`, or `throw` — so the loop is a
conditional block written as a loop, and the extra iterations it looks like it
performs never happen. That is usually a mistake: a `break` that belongs inside
an `if`, a `return` that should collect results instead of leaving on the first
element, or a condition that was never finished.

A loop iterates again as soon as one path flows back into it, so a `break` or
`return` guarded by an `if` is fine. A loop the surrounding code can never reach
is left alone.

Examples of **incorrect** code for this rule:

```javascript
for (let i = 0; i < arr.length; i++) {
	console.log(arr[i]);
	break;
}

while (foo) {
	doSomething(foo);
	foo = foo.parent;
	return;
}

function find(arr, target) {
	for (const item of arr) {
		return item === target;
	}
}
```

Examples of **correct** code for this rule:

```javascript
for (let i = 0; i < arr.length; i++) {
	console.log(arr[i]);
}

while (foo) {
	if (bar) {
		break;
	}
	foo = foo.parent;
}

function find(arr, target) {
	for (const item of arr) {
		if (item === target) {
			return item;
		}
	}
}
```

## Options

This rule accepts an options object with one property:

### `ignore`

An array of loop types to leave unchecked. Each entry is one of
`"WhileStatement"`, `"DoWhileStatement"`, `"ForStatement"`,
`"ForInStatement"`, or `"ForOfStatement"`. It defaults to an empty array.

`{ "ignore": ["ForInStatement", "ForOfStatement"] }` allows the idiom that
reads only the first entry of a collection:

```javascript
/* eslint no-unreachable-loop: ["error", { "ignore": ["ForInStatement", "ForOfStatement"] }] */

function firstKey(obj) {
	for (const key in obj) {
		return key;
	}
	return null;
}
```
