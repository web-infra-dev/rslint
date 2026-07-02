// Triggers all three configured rules:
//   null literal           -> local/no-null        (plugin)
//   .filter member access  -> local/prefer-array-some (plugin)
//   console.*              -> no-console           (native)
const value = null;
const numbers = [1, 2, 3];
const hasPositive = numbers.filter((n) => n > 0).length > 0;
console.log(value, hasPositive);
