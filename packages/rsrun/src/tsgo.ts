import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import { createRequire } from 'node:module';
const esmRequire = createRequire(import.meta.url);

export function resolveTsgoExecutable(): string {
  const pkgRoot = path.dirname(esmRequire.resolve('@rslint/tsgo/package.json'));
  const localName = `tsgo${process.platform === 'win32' ? '.exe' : ''}`;
  const localPath = path.join(pkgRoot, 'bin', localName);
  if (fs.existsSync(localPath)) {
    return localPath;
  }
  const platformKey = `${process.platform}-${os.arch()}`;
  return esmRequire.resolve(
    `@rslint/tsgo-${platformKey}/lib/tsgo${process.platform === 'win32' ? '.exe' : ''}`,
  );
}
