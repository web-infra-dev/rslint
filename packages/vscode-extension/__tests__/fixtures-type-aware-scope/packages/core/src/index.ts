import { dependency } from './dependency';

export async function coreFunction(): Promise<void> {
  console.log('hello');
  dependency.value;
}
