// Test file for fixAll code action (auto-fix on save)
const value1: string = 'hello';
const value2: number = 42;

// Multiple auto-fixable issues: no-unnecessary-type-assertion
const result1 = (value1 as string).toUpperCase();
const result2 = (value2 as number).toFixed(2);
