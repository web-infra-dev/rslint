// Test file for auto fix code actions
const someValue: string = 'hello';

// This should trigger no-unnecessary-type-assertion rule (has auto fix)
const result = (someValue as string).toUpperCase();

// Another example that should trigger a fix
function example(): number {
  return 42 as number; // unnecessary type assertion
}
