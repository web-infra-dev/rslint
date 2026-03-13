import fastGlob from 'fast-glob';
import path from 'node:path';
import Mocha from 'mocha';

export function run(
  testPath: string,
  callback: (error: unknown, failures?: number) => void,
) {
  const files = fastGlob.sync('**/*.test.js', {
    cwd: testPath,
  });
  const mocha = new Mocha({
    ui: 'tdd',
  });

  files.forEach(file => {
    mocha.addFile(path.join(testPath, file));
  });

  try {
    mocha.run(failures => {
      callback(null, failures);
    });
  } catch (error) {
    callback(error);
  }
}
