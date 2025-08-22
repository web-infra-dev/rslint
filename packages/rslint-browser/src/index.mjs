import { load, WASI } from '@tybys/wasm-util';
import { Volume, createFsFromVolume } from 'memfs-browser';

async function main() {
console.log('start');
  const fs = createFsFromVolume(
    Volume.fromJSON({
      '/input.ts': "let a:number = 20",
      '/tsconfig.json': "{}",
    }),
  );

  const wasi = new WASI({
    args: ['rslint.wasm', "--browser"],
    returnOnExit:true,
    preopens: {
      '/': '/',
    },
    fs,
  });

  const imports = {
    wasi_snapshot_preview1: wasi.wasiImport,
  };

  const { module, instance } = await load('dist/rslint.wasm', imports);
  const exitCode = wasi.start(instance);
  return exitCode;
  // wasi.initialize(instance)
}

main();