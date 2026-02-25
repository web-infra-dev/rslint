import { readFileSync, writeFileSync } from 'fs';
import { gzipSync } from 'zlib';

const input = readFileSync('rslint.wasm');
const compressed = gzipSync(input);
writeFileSync('rslint.wasm.gz', compressed);

console.log(
  `Compressed rslint.wasm: ${(input.length / 1024 / 1024).toFixed(1)}MB -> ${(compressed.length / 1024 / 1024).toFixed(1)}MB`,
);
