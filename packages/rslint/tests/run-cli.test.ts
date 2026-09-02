import { describe, expect, test } from 'rstack/test';
import { runCLI } from '../src/index.js';

describe('runCLI', () => {
  test('is exported from the package root', () => {
    expect(typeof runCLI).toBe('function');
  });
});
