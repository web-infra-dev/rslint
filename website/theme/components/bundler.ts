export enum Linter {
  RSLint = 'rslint',
  Rspack = 'rspack', // Legacy compatibility
}

// Legacy enum name for compatibility
export const Bundler = Linter;

export function getLinter(): Linter {
  // Always return RSLint since we're only tracking one linter
  return Linter.RSLint;
}

// Legacy function name for compatibility
export function getBundler(): Linter {
  return getLinter();
}
