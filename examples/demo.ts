// Example test file demonstrating inline comment support

// This await should be reported (no disable comment)
await 1;

/*rslint-disable */
// Everything after this should be disabled for all rules
await 2; // Should NOT be reported
const x = await 3; // Should NOT be reported

// Test specific rule disable
/* rslint-disable await-thenable*/
await 4; // Should NOT be reported due to specific rule disable

// Test multiple rules
/* rslint-disable await-thenable,no-floating-promises*/
await 5; // Should NOT be reported for await-thenable
Promise.resolve(); // Should NOT be reported for no-floating-promises

// Test single-line comment version
//rslint-disable await-thenable
await 6; // Should NOT be reported

// Back to normal code
await 7; // Should still be disabled due to file-wide disable above