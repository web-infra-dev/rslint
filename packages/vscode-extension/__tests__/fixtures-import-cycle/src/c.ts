import { fromA } from './a';

export var witnessC = 1;

export function fromC(): number {
  return witnessC;
}

export function backToA(): number {
  return fromA();
}
