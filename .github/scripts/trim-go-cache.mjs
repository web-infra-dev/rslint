import { readdirSync, statSync, unlinkSync } from 'node:fs';
import { join } from 'node:path';

const [cachePath, limitMiB] = process.argv.slice(2);
const limit = Number(limitMiB) * 1024 * 1024;
if (!cachePath || !Number.isSafeInteger(limit) || limit <= 0) {
  throw new Error('Usage: trim-go-cache.mjs <cache-path> <positive-limit-MiB>');
}

const files = [];
function collect(directory) {
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) {
      collect(path);
    } else if (entry.isFile()) {
      const { size, mtimeMs } = statSync(path);
      files.push({ path, size, mtimeMs });
    }
  }
}
collect(cachePath);

let size = files.reduce((total, file) => total + file.size, 0);
const originalSize = size;
// Go updates cache modification times on use (at most once per hour). Missing entries are
// rebuilt normally, so discard the least recently used files first.
files.sort((a, b) => a.mtimeMs - b.mtimeMs);
let removed = 0;
for (const file of files) {
  if (size <= limit) break;
  unlinkSync(file.path);
  size -= file.size;
  removed++;
}

const toMiB = (bytes) => (bytes / 1024 / 1024).toFixed(1);
console.log(
  `Go build cache: ${toMiB(originalSize)} MiB -> ${toMiB(size)} MiB; removed ${removed} files (limit ${limitMiB} MiB)`,
);
