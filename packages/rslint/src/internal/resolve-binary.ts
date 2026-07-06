import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);

export function resolveRslintBinary(): string {
  const arch = process.arch;
  const tuples =
    process.platform === 'linux'
      ? [`linux-${arch}-gnu`, `linux-${arch}-musl`]
      : process.platform === 'win32'
        ? [`win32-${arch}-msvc`]
        : [`${process.platform}-${arch}`];

  for (const tuple of tuples) {
    try {
      return require.resolve(`@rslint/native-${tuple}/bin`);
    } catch {
      // Try the next supported tuple.
    }
  }

  throw new Error(
    `rslint: no native binary for ${process.platform}-${arch} ` +
      `(looked for @rslint/native-{${tuples.join(',')}})`,
  );
}
