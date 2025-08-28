// Test case 1: This should NOT report (OR in ternary condition, ignoreConditionalTests=true by default)
declare let x: string | null;
const result1 = x || 'foo' ? null : null;

// Test case 2: This SHOULD report (simple ternary with nullable type)
declare let y: string | null;
const result2 = y ? y : 'default';
