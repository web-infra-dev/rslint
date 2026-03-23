// Test file for multi-pass cascade fix:
// Pass 1: ban-types fixes String → string, Number → number, Boolean → boolean
// Pass 2: no-inferrable-types removes now-inferrable type annotations
const cascadeA: String = 'hello';
const cascadeB: Number = 42;
const cascadeC: Boolean = true;
export { cascadeA, cascadeB, cascadeC };
