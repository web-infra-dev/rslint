export function directValue() {
  return 'direct';
}

export interface DirectType {
  value: string;
}

export { barrelValue as renamedBarrelValue, otherBarrelValue } from './values';
export type { BarrelType, RuntimeTypeOnly } from './values';
