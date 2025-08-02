// Test file for disable rule code actions
// These should trigger unsafe rules (no auto fix available)

const obj: any = {};
const value = obj.someProperty.nested; // no-unsafe-member-access

function takesString(s: string) {}
const anyValue: any = 'hello';
takesString(anyValue); // no-unsafe-argument
