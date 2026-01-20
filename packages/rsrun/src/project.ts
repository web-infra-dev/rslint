import fs from 'node:fs';
import path from 'node:path';

export function findProject(startDir: string): string | null {
  let current = path.resolve(startDir);
  for (;;) {
    const tsconfig = path.join(current, 'tsconfig.json');
    if (fs.existsSync(tsconfig)) {
      return tsconfig;
    }
    const tsconfigJsonc = path.join(current, 'tsconfig.jsonc');
    if (fs.existsSync(tsconfigJsonc)) {
      return tsconfigJsonc;
    }
    const parent = path.dirname(current);
    if (parent === current) {
      return null;
    }
    current = parent;
  }
}

export function isInNodeModules(filePath: string): boolean {
  return (
    filePath.includes(`${path.sep}node_modules${path.sep}`) ||
    filePath.includes('/node_modules/')
  );
}
