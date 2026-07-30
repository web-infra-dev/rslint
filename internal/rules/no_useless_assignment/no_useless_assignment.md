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
  `const pipe = {}` here — a false positive tracked at
  [eslint/eslint#20947](https://github.com/eslint/eslint/issues/20947). Its
  root cause turned out to be that `@typescript-eslint`'s scope manager does
  not register reads inside legacy parameter decorators as a reference at
  all, tracked separately at
  [typescript-eslint/typescript-eslint#12407](https://github.com/typescript-eslint/typescript-eslint/issues/12407);
  this rule's own reference tracking is not affected by that gap.
- This rule models exceptions and control flow with its own approximation
  rather than a literal port of ESLint's code-path analysis, so its findings
  can diverge from ESLint's for deeply nested `try`/`catch`/`finally`
  combinations — for instance a `switch` nested directly inside a `catch`
  clause:
  `function f() { let v = 0; try {} catch (e) { v = 1; switch (k) { case 1: use(v); v = 2; } } }`.
  ESLint reports only `v = 2` there; this rule also reports the initial
  `let v = 0`, which is unused on every path just as `v = 2` is. Other
  diverging shapes involve a `try` without its own `catch` nested inside
  another `try`. These shapes are uncommon in real-world code, and in the
  large majority of cases this rule reports the same findings as ESLint.
  ESLint's own code-path analysis has open, accepted-but-unfixed bugs in this
  exact area — see
  [eslint/eslint#17579](https://github.com/eslint/eslint/issues/17579), left
  open specifically because fixing it would be a breaking change — so
  matching it exactly for every nested shape isn't a well-defined target.

## Original Documentation

- [ESLint no-useless-assignment](https://eslint.org/docs/latest/rules/no-useless-assignment)
