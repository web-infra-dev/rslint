import { fromB } from './b';

export var witnessA = 1;

export function fromA(): number {
  return fromB() + witnessA;
}
