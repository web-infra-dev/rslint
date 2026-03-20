import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-constant-binary-expression', {
  valid: [
    // --- Variable references (not constant) ---
    'bar && foo',
    'bar || foo',
    'bar ?? foo',
    'foo == true',
    'foo === true',
    'true ? foo : bar',

    // --- Function return values (not constant) ---
    'foo() && bar',
    'foo() || bar',
    'foo() ?? bar',

    // --- Property access (not constant) ---
    'foo[0] && bar',
    'foo.bar && baz',

    // --- Template literals with expressions (not constant) ---
    'var a = `${bar}` && foo',

    // --- Compound assignment (not constant) ---
    '(x += 1) && foo',
    '(x -= 1) || bar',

    // --- Delete operations (not constant for isConstant) ---
    'delete bar.baz && foo',

    // --- Nullish coalescing edge cases ---
    'foo ?? null ?? bar',

    // --- Shadowed built-in functions ---
    'function Boolean(n: any) { return n; } Boolean(x) ?? foo',
    'function Boolean(n: any) { return n; } Boolean(x) && foo',
    'var Boolean = (n: any) => n; Boolean(x) ?? foo',

    // --- Valid comparisons ---
    'x === null',
    'null === x',
    'x == null',
    'x == undefined',
    'x !== null',
    'x != undefined',

    // --- Logical NOT of non-constant is not constant (#552) ---
    '!foo && bar',
    '!foo || bar',
    '!module || !module[pluginName]',
    '!!foo && bar',

    // --- For ==, alwaysNew only applies when BOTH sides are always new ---
    'x == /[a-z]/',
    'x == []',

    // --- new with user-defined constructors (not guaranteed always new) ---
    'new Foo() == true',

    // --- PostfixUnary / PrefixUnary ++ / -- (not constant) ---
    'x++ && bar',
    'x-- || bar',
    '++x && bar',

    // --- Boolean(variable) is not constant ---
    'Boolean(foo) && bar',

    // --- Unary +/- of variable is not constant ---
    '+foo && bar',
    '-foo || bar',

    // --- Comma expression with variable last ---
    '(1, x) && foo',

    // --- Logical assignment with non-identity rhs (not constant) ---
    '(x ||= foo) && bar',
    '(x &&= foo) || bar',

    // --- Single-element array has variable loose boolean comparison ---
    '[x] == true',

    // --- new user-defined constructor is not always new for === ---
    'new Foo() === x',

    // --- delete returns boolean, comparison to boolean varies ---
    'delete x.y === true',
  ],
  invalid: [
    // ============================
    // Constant short-circuit: &&
    // ============================
    // 2-char string is truthy (regression: quote stripping bug)
    {
      code: '"ab" && foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: '[] && greeting',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: 'true && hello',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: "'' && foo",
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: '100 && foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: '/[a-z]/ && foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: 'Boolean([]) && foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: '({}) && foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: '(() => {}) && foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: 'new Foo() && foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: 'undefined && foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    // Negation of constant
    {
      code: '!true && bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: '!undefined && bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: '![] && bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    // void is always constant
    {
      code: 'void 0 && bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    // typeof in boolean position is always constant
    {
      code: 'typeof foo && bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    // Constant binary expressions
    {
      code: '(1 + 2) && bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    // Assignment with constant right side
    {
      code: '(x = true) && bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    // null is constant
    {
      code: 'null && bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    // Class expression is constant
    {
      code: '(class {}) && bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },

    // ============================
    // Constant short-circuit: ||
    // ============================
    {
      code: '[] || greeting',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: 'true || hello',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: '0 || foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: "'' || foo",
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: '!true || bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },

    // ============================
    // Constant short-circuit: ??
    // ============================
    {
      code: '({}) ?? foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: '1 ?? foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: 'null ?? foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: 'undefined ?? foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    // Comparison operators always produce non-nullish boolean
    {
      code: '(x > 0) ?? fallback',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: '(x === y) ?? fallback',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    // Unary expressions always have constant nullishness
    {
      code: '!foo ?? bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: 'typeof foo ?? bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    // String(), Number() always return non-nullish
    {
      code: 'String(x) ?? foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: 'Number(x) ?? foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    // Boolean() always non-nullish
    {
      code: 'Boolean(x) ?? foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    // void produces undefined (always nullish)
    {
      code: 'void 0 ?? bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },

    // ============================
    // Constant binary operand: ==
    // ============================
    {
      code: '[] == true',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: 'true == []',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: '({}) == true',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: '({}) == null',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: '({}) == undefined',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: 'undefined == true',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: 'true == true',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: '"" == true',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: '"hello" == false',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: '0 == true',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: '1 == false',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    // Multi-element array in loose comparison
    {
      code: '[1, 2] == true',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    // void 0 == undefined is constant
    {
      code: 'void 0 == undefined',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },

    // ============================
    // Constant binary operand: !=
    // ============================
    {
      code: '({}) != true',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: '[] != null',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },

    // ============================
    // Constant binary operand: ===
    // ============================
    {
      code: 'true === true',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: '[] === null',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: 'null === null',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: '({}) === undefined',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: '({}) === null',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: 'true === false',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: '"" === true',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: '42 === true',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    // typeof is string, strict comparison with boolean is constant
    {
      code: 'typeof x === true',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    // !true is constant static boolean
    {
      code: '!true == 42',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },

    // ============================
    // Constant binary operand: !==
    // ============================
    {
      code: '[] !== null',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },
    {
      code: '({}) !== undefined',
      errors: [{ messageId: 'constantBinaryOperand' }],
    },

    // ============================
    // Both always new (== only)
    // ============================
    {
      code: '[a] == [a]',
      errors: [{ messageId: 'bothAlwaysNew' }],
    },
    {
      code: '({}) == []',
      errors: [{ messageId: 'bothAlwaysNew' }],
    },

    // ============================
    // Always new (=== / !==)
    // ============================
    {
      code: '[] === []',
      errors: [{ messageId: 'alwaysNew' }],
    },
    {
      code: '({}) === ({})',
      errors: [{ messageId: 'alwaysNew' }],
    },
    {
      code: 'x === {}',
      errors: [{ messageId: 'alwaysNew' }],
    },
    {
      code: 'x === []',
      errors: [{ messageId: 'alwaysNew' }],
    },
    {
      code: 'x === (() => {})',
      errors: [{ messageId: 'alwaysNew' }],
    },
    {
      code: 'x === /[a-z]/',
      errors: [{ messageId: 'alwaysNew' }],
    },
    {
      code: '({}) === x',
      errors: [{ messageId: 'alwaysNew' }],
    },
    {
      code: '/[a-z]/ === x',
      errors: [{ messageId: 'alwaysNew' }],
    },
    // Class expression is always new
    {
      code: '(class {}) === x',
      errors: [{ messageId: 'alwaysNew' }],
    },
    // Conditional expression with both branches always new
    {
      code: 'x === (true ? [] : {})',
      errors: [{ messageId: 'alwaysNew' }],
    },

    // ============================
    // Boolean constructor calls
    // ============================
    {
      code: 'Boolean(true) && foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: 'Boolean(false) || foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },

    // ============================
    // Parenthesized expressions
    // ============================
    {
      code: '(!true) && bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: '(null) ?? bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },

    // ============================
    // Logical assignment with identity rhs
    // ============================
    {
      code: '(x ||= 1) && foo',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
    {
      code: '(x &&= 0) || bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },

    // ============================
    // Comma (sequence) expression
    // ============================
    {
      code: '(1, 2) && bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },

    // ============================
    // Assignment with constant rhs
    // ============================
    {
      code: '(x = []) && bar',
      errors: [{ messageId: 'constantShortCircuit' }],
    },
  ],
});
