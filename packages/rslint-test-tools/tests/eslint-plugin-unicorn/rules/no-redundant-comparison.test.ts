// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-redundant-comparison.js
import {
  RuleTester,
  type InvalidTestCase,
  type ValidTestCase,
} from '../rule-tester';

function valid(code: string, typescript = false): ValidTestCase {
  return {
    code,
    filename: typescript ? 'src/virtual.ts' : 'src/virtual.js',
    languageOptions: { sourceType: 'module' },
  };
}

// The wrapper checks messages; Go tests also assert ranges, fixes and suggestions.
function invalid(code: string, typescript = false): InvalidTestCase {
  return {
    ...valid(code, typescript),
    errors: [
      {
        messageId: 'no-redundant-comparison',
        message:
          'This comparison is redundant because it is already implied by the other conditions.',
      },
    ],
  };
}

new RuleTester().run('no-redundant-comparison', {} as never, {
  valid: [
    // Explicit side effects invalidate equality, strict inequality and bound implications
    valid(
      'let first = 1; let second = 1; first === second && first === 1 && (second = 2) === 2 && second === 1;',
    ),
    valid(
      'let first = 1; let second = 1; first === second && first === 1 && second++ === 1 && second === 1;',
    ),
    valid(
      'let first = 1; let second = 1; function change() { second = 2; return true; } first === second && first === 1 && change() && second === 1;',
    ),
    valid(
      'let first = 1; let second = 1; function change() { second = 2; return true; } first === second && first === 1 && change`value` && second === 1;',
    ),
    valid(
      'let first = 1; let second = 2; first === 1 && (first = 2) === 2 && second === 1 && first === second;',
    ),
    valid(
      'let first = 1; let second = 2; first === 1 && first !== second && (second = 1) === 1 && second !== 1;',
    ),
    valid(
      'let first = 1; let second = 1; first === second && first > 0 && (second = 0) === 0 && second > 0;',
    ),
    valid(
      'let first = 1; let second = 1; let third = 1; first === second && second === third && first === 1 && (third = 2) === 2 && third === 1;',
    ),
    // Explicit side effects and wrappers
    valid(
      'first === second && first === 1 && /* keep */ change() && second === 1;',
    ),
    valid(
      '((first === second) && (first === 1) && ((second = 2) === 2) && (second === 1));',
    ),
    valid(
      'first === second && first === 1 && (condition ? change() : true) && second === 1;',
    ),
    valid(
      'first === second && first === 1 && (second = third, true) && second === 1;',
    ),
    valid(
      'first === second && first === 1 && delete object.value && second === 1;',
    ),
    valid(
      'first === second && first === 1 && new SomeClass() && second === 1;',
    ),
    valid(
      'first === second && first === 1 && (change`value` === true) && second === 1;',
    ),
    valid(
      'first === second && first === 1 && (condition ? change`value` : true) && second === 1;',
    ),
    valid(
      'first === second && first === 1 && (change`value`, true) && second === 1;',
    ),
    valid(
      'first === second && first === 1 && /* keep */ ((change`value`)) && second === 1;',
    ),
    valid(
      'let first = 1; let second = 1; (first as number) === second && first === 1 && ((second = 2) as number) === 2 && second === 1;',
      true,
    ),
    valid(
      'first! === second && first === 1 && (change() satisfies boolean) && second === 1;',
      true,
    ),
    valid(
      'first! === second && first === 1 && (change`value` satisfies boolean) && second === 1;',
      true,
    ),
    // Snapshot cases
    valid('a > 0 || b > 0 || a === b'),
    valid('a > 0 ?? b > 0 ?? a === b'),
    valid('a > 0'),
    valid('a === b'),
    valid('a > 0 && a === b'),
    valid('a > 0 && b > 0'),
    valid('a !== b'),
    valid('a > 0 && b > 0 && a !== b'),
    valid('a > 0 && b < 0 && a === b'),
    valid('a === "foo" && b !== "bar" && a !== b'),
    valid('a !== "foo" && b !== "bar" && a === b'),
    valid('foo() && bar() && a === b'),
    valid('a && b && a === b'),
    valid('a === b && b === c && a === c'),
    valid('a?.x > 0 && b?.x > 0 && a?.x === b?.x'),
    valid('a > 0 && a > 0 && a === a'),
    valid('a > 0 && c > 0 && a === b'),
    valid('a == 0 && b == 0 && a === b'),
    valid('a > foo() && b > foo() && a === b'),
    valid("if (a === b && a !== 'x' && isEnabled() && b !== 'x') {}"),
    valid('a > 0 && b < 10 && a === b'),
    valid('a[f()] > 0 && b[f()] > 0 && a[f()] === b[f()]'),
    valid(
      'const modes = new Set([1]); modes.clear(); value === other && value > (modes.size ? 3 : 2) && other > 3;',
    ),
    valid(
      'const modes = new Set([1]); modes.clear(); value === other && value > -(modes.size ? 2 : 3) && other > -2;',
    ),
    valid(
      'const modes = new Set([1]); modes.clear(); value === other && value > ((modes.size && 3) || lowerBound) && other > 3;',
    ),
    valid(
      'const object = {value: true}; Object.defineProperty(object, "value", {get() { return false; }}); value === other && value > (object.value ? 3 : lowerBound) && other > 3;',
    ),
    valid(
      'const modes = new Set([1]); modes.clear(); value === other && value > ((modes.size ? 3 : 2) as number) && other > 3;',
      true,
    ),
    // Documentation examples
    valid('if (a > 0 && a === b) {}'),
    valid('if (b > 2 && a === b) {}'),
    valid("if (a !== 'test' && a === b) {}"),
    valid("if (a === 'test' && a !== b) {}"),
  ],
  invalid: [
    // Semicolon insertion
    invalid('foo()\nb !== 1 && (a === 1) && a !== b'),
    invalid('const previous = value\nb !== 1 && (a === 1) && a !== b'),
    invalid('foo() // keep\nb !== 1 && (a === 1) && a !== b'),
    invalid('foo();\nb !== 1 && (a === 1) && a !== b'),
    invalid('if (ready)\nb !== 1 && (a === 1) && a !== b'),
    invalid('foo()\nb !== 1 && (a as number) === 1 && a !== b', true),
    invalid('foo()\na > 0 && (a > 1) && a === b'),
    // pure strict comparisons retain autofixes
    invalid(
      'const first = 1; const second = 1; first === second && first === 1 && second === 1;',
    ),
    invalid(
      'const first = 1; const second = 2; first === 1 && first !== second && second !== 1;',
    ),
    invalid(
      'const first = 1; const second = 1; const third = 1; first === second && second === third && first === 1 && third === 1;',
    ),
    // implicit ordering coercion retains the existing suggestion
    invalid(
      'const first = 1; const second = 1; first === second && first > 0 && second > 0;',
    ),
    // Explicit side effects and wrappers
    invalid('first === second && first === 1 && /* keep */ second === 1;'),
    invalid('(first === second) && (first === 1) && (second === 1);'),
    invalid(
      'first === second && first === "value" && `value` && second === "value";',
    ),
    invalid(
      '((first as number) === second) && (first === 1) && (second! === 1);',
      true,
    ),
    // Snapshot cases
    invalid('a > 0 && b > 0 && a === b'),
    invalid('a > 0 && b > 2 && a === b'),
    invalid('a < c && b < c && a === b'),
    invalid('a !== "test" && b !== "test" && a === b'),
    invalid('a === "test" && b === "test" && a === b'),
    invalid('a === "test" && b !== "test" && a !== b'),
    invalid('a !== "test" && b === "test" && a !== b'),
    invalid('b > 0 && a > 0 && a === b'),
    invalid('a >= 5 && b > 0 && a === b'),
    invalid('a > 0 && b >= 5 && a === b'),
    invalid('a > 0 && b >= 0 && a === b'),
    invalid('a > -2 && b > 0 && a === b'),
    invalid('a < 10 && b < 100 && a === b'),
    invalid('obj.x > 0 && obj.y > 0 && obj.x === obj.y'),
    invalid('a[key] > 0 && b[key] > 0 && a[key] === b[key]'),
    invalid('a > 0 && b > 0 && a === b && c > 0'),
    invalid("if (a === b && a !== 'x' && b !== 'x' && isEnabled()) {}"),
    invalid("if (isEnabled() && a === b && a !== 'x' && b !== 'x') {}"),
    invalid("if (a === b && a !== 'x' && b !== 'x' && list.includes(a)) {}"),
    invalid('a > 0 && c > 0 && a === b && b === c'),
    invalid('0 < a && 0 < b && a === b'),
    invalid('a > b && c > b && a === c'),
    invalid('a < "foo" && b < "foo" && a === b'),
    invalid('this.x > 0 && this.y > 0 && this.x === this.y'),
    invalid('(a > 0 && b > 0 && a === b) || c'),
    invalid('if (\n\ta > 0 &&\n\tb > 0 &&\n\ta === b\n) {}'),
    invalid('a > 0 && /* keep */ b > 0 && a === b'),
    invalid('(a as number) > 0 && (b as number) > 0 && a === b', true),
    invalid('a! > 0 && b! > 0 && a === b', true),
    invalid(
      '(a satisfies number) > 0 && (b satisfies number) > 0 && a === b',
      true,
    ),
    invalid('a !== b && 1 === a && b !== 1'),
    invalid('a === b && c === d && b === c && a > 5 && d > 3'),
    invalid('a === b && b === c && a === c && a > 5 && c > 3'),
    invalid('a === b && c === a && c > 5 && b > 3'),
    invalid('a < 5 && b < 3 && a === b'),
    invalid('a <= 3 && b < 3 && a === b'),
    // Documentation examples
    invalid('if (a > 0 && b > 0 && a === b) {}'),
    invalid('if (a > 0 && b > 2 && a === b) {}'),
    invalid("if (a !== 'test' && b !== 'test' && a === b) {}"),
    invalid("if (a === 'test' && b !== 'test' && a !== b) {}"),
  ],
});
