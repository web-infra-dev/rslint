import { external, pending } from './types';

export function run() {
  debugger;
  pending();
  const unsafe: number = external;
  var count = 1;
  return unsafe + count;
}
