import { fromC } from './c';

export var witnessB = 1;

export function fromB(): number {
  return fromC() + witnessB;
}
