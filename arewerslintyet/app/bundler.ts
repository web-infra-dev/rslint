export enum Linter {
  RSLint = 'rslint',
}

export function getLinter(): Linter {
  // Always return RSLint since we're only tracking one linter
  return Linter.RSLint;
}

// Legacy function name for compatibility
export function getBundler(): Linter {
  return getLinter();
}
