export function barrelValue() {
  return 'barrel';
}

export const otherBarrelValue = 'other';

export interface BarrelType {
  value: string;
}

export class RuntimeTypeOnly {
  value = 'runtime';
}
