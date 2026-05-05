// Circular: a re-exports default from b, b re-exports default from a.
// Without a real default anywhere in the cycle, the alias chain bottoms out.
export { default } from './circular-b';
