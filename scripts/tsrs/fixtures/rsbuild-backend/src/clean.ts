import { pending } from './types';

export async function run() {
  await pending();
  const safe = 1;
  return safe;
}
