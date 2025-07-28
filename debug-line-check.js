const code = `interface SomeType {
  prop: string;
}
function foo() {
  this.prop;
}`;

const lines = code.split('\n');
lines.forEach((line, index) => {
  console.log(`Line ${index + 1}: "${line}"`);
});

// Find where "this.prop" appears
lines.forEach((line, index) => {
  if (line.includes('this.prop')) {
    console.log(`\n"this.prop" found on line ${index + 1}`);
    console.log(`Line content: "${line}"`);
    console.log(`"this" starts at column ${line.indexOf('this') + 1}`);
    console.log(`"this" ends at column ${line.indexOf('this') + 4 + 1}`);
    console.log(`"this.prop;" starts at column ${line.indexOf('this.prop;') + 1}`);
    console.log(`"this.prop;" ends at column ${line.indexOf('this.prop;') + 'this.prop;'.length + 1}`);
  }
});