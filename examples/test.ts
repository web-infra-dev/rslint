// Test file for inline comment support
async function test() {
    // This should trigger await-thenable rule
    await 42;
    
    // Disable rslint for the rest of the file
    /*rslint-disable */
    await 43;
    
    // This should also be disabled
    await 44;
}

// Disable the `await-thenable` rule for following line
/* rslint-disable await-thenable*/
await 45;

// Normal code after inline disable
await 46;