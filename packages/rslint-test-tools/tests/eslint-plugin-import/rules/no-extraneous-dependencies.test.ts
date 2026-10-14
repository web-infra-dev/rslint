// Upstream: eslint-plugin-import v2.32.0, all semantic groups.
// Flow typeof and executable webpack resolver cases are retained in the data
// with skipReason; the native runtime cannot execute those configurations.
import fs from 'node:fs';
import path from 'node:path';
import { RuleTester } from '../rule-tester.js';

const dataDir = path.resolve(
  import.meta.dirname,
  '../../../../../internal/plugins/import/rules/no_extraneous_dependencies/testdata',
);
const root = fs.mkdtempSync(
  path.join(process.cwd(), '.no-extraneous-dependencies-'),
);
const chunks = fs
  .readFileSync(path.join(dataDir, 'modules.txtar'), 'utf8')
  .split(/^-- (.+) --\r?$/m);
if (chunks.length < 3) throw new Error('Missing upstream fixtures');
for (let i = 1; i < chunks.length; i += 2) {
  const file = path.join(root, chunks[i]);
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, chunks[i + 1].replace(/^\r?\n/, ''));
}
afterAll(() => fs.rmSync(root, { recursive: true, force: true }));
const suites = JSON.parse(
  fs
    .readFileSync(path.join(dataDir, 'upstream-cases.json'), 'utf8')
    .replaceAll('/fixture', root),
);
function adapt(item: any) {
  const { fileName, skip, skipReason, ...rest } = item;
  // Bare globs in upstream resolve relative to the upstream checkout's cwd.
  // Anchor that spelling to this isolated fixture root without changing options
  // such as absolute globs, negation, or the imported file's identity.
  if (rest.options)
    for (const options of rest.options) {
      for (const key of [
        'devDependencies',
        'optionalDependencies',
        'peerDependencies',
        'bundledDependencies',
      ]) {
        if (Array.isArray(options[key]))
          options[key] = options[key].map((glob: string) =>
            glob.startsWith('*.') ? path.join(root, glob) : glob,
          );
      }
    }
  return { ...rest, filename: path.join(root, fileName) };
}
for (const suite of suites)
  new RuleTester().run('no-extraneous-dependencies', null as never, {
    valid: suite.valid.filter((c: any) => !c.skip).map(adapt),
    invalid: suite.invalid.filter((c: any) => !c.skip).map(adapt),
  });
