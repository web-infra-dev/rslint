// Regenerate with: node generate.mjs /path/to/eslint-plugin-n-v18.3.0
import { readFile, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { pathToFileURL } from 'node:url';
import { execFileSync } from 'node:child_process';

const upstream = resolve(process.argv[2]);
const pkg = JSON.parse(
  await readFile(resolve(upstream, 'package.json'), 'utf8'),
);
if (pkg.name !== 'eslint-plugin-n' || pkg.version !== '18.3.0') {
  throw new Error(
    'Expected an eslint-plugin-n v18.3.0 checkout with dependencies installed',
  );
}
const load = (file) => import(pathToFileURL(resolve(upstream, file)).href);
const { NodeBuiltinGlobals, NodeBuiltinModules, NodeBuiltinImportMeta } =
  await load('lib/unsupported-features/node-builtins.js');
const { default: rule } = await load(
  'lib/rules/no-unsupported-features/node-builtins.js',
);
const { READ, CALL, CONSTRUCT } = await load(
  'node_modules/@eslint-community/eslint-utils/index.mjs',
);
const quote = JSON.stringify;
const { default: semver } = await load('node_modules/semver/index.js');
const versions = (values) => semver.rsort([...(values ?? [])]).join(','); // cspell:ignore rsort
let output = `// Code generated from eslint-plugin-n v18.3.0 by generate.mjs; DO NOT EDIT.
// https://github.com/eslint-community/eslint-plugin-n/tree/v18.3.0/lib/unsupported-features
// cspell:ignore CLEVEL Cipheriv Decipheriv Diffie Fips Keypress Naptr Spkac Tlsa Uncloneable alpn btlazy btopt btultra cpus debuglog dfast diffie dlopen execve fdatasync fips freemem initgroups loadavg onresourcetimingbufferfull ppid readv timerify totalmem webcrypto
package node_builtins

`;
const aliases = [];
for (const [name, tree] of Object.entries({
  globalAPIs: NodeBuiltinGlobals,
  moduleAPIs: NodeBuiltinModules,
  importMetaAPIs: NodeBuiltinImportMeta,
})) {
  output += `var ${name} = newBuiltinAPIs([]builtinFeature{\n`;
  function visit(node, path, stack = new Map()) {
    if (stack.has(node)) {
      aliases.push([name, path, stack.get(node)]);
      return;
    }
    stack = new Map(stack).set(node, path);
    for (const [kind, name] of [
      [READ, path.join('.')],
      [CALL, path.join('.') + '()'],
      [CONSTRUCT, 'new ' + path.join('.') + '()'],
    ]) {
      if (node[kind]) {
        output += `\t{${quote(name)}, ${quote(versions(node[kind].supported))}, ${quote(versions(node[kind].experimental))}},\n`;
      }
    }
    for (const [key, value] of Object.entries(node))
      visit(value, [...path, key], stack);
  }
  visit(tree, []);
  output += '})\n\n';
}
output += 'func init() {\n';
const access = (root, path) =>
  root + path.map((key) => `.properties[${quote(key)}]`).join('');
for (const [root, path, target] of aliases)
  output += `\t${access(root, path)} = ${access(root, target)}\n`;
output += '}\n';
const dataFile = new URL('apis_generated.go', import.meta.url);
await writeFile(dataFile, output);
execFileSync('gofmt', ['-w', dataFile.pathname]);
await writeFile(
  new URL('node_builtins.schema.json', import.meta.url),
  JSON.stringify(
    { type: 'array', items: rule.meta.schema, minItems: 0, maxItems: 1 },
    null,
    2,
  ) + '\n',
);
