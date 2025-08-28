// Test case 1: should NOT report (default ignoreTernaryTests is true)
declare let x: string | null;
const result1 = x ? x : 'default';

// Test case 2: this is not a ternary, should report
declare let y: string | null;
const result2 = y || 'default';
