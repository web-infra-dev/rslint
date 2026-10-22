#!/usr/bin/env node
// Compare a rebuilt public AST API against a frozen baseline distribution.
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { createHash } from 'node:crypto';
import { createRequire } from 'node:module';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { parseArgs } from 'node:util';

const { values } = parseArgs({
  options: {
    reference: { type: 'string' },
    candidate: { type: 'string' },
    'repo-root': { type: 'string' },
  },
});
assert(
  values.reference,
  '--reference must identify a frozen @rslint/api dist directory',
);
const root = values['repo-root']
  ? path.resolve(values['repo-root'])
  : path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const reference = path.resolve(values.reference);
const candidate = values.candidate
  ? path.resolve(values.candidate)
  : path.join(root, 'packages/rslint-api/dist');
const ts = createRequire(path.join(root, 'package.json'))('typescript');
const hash = (value) => createHash('sha256').update(value).digest('hex');

function declarationTokens(text) {
  const parsed = ts.createSourceFile(
    'api.d.ts',
    text,
    ts.ScriptTarget.Latest,
    true,
  );
  assert.equal(parsed.parseDiagnostics.length, 0, 'Invalid API declarations');
  const scanner = ts.createScanner(
    ts.ScriptTarget.Latest,
    true,
    ts.LanguageVariant.Standard,
    text,
  );
  const result = [];
  for (
    let kind = scanner.scan();
    kind !== ts.SyntaxKind.EndOfFileToken;
    kind = scanner.scan()
  ) {
    result.push([
      kind,
      kind === ts.SyntaxKind.StringLiteral
        ? scanner.getTokenValue()
        : scanner.getTokenText(),
    ]);
  }
  return result;
}

function prototypeSurface(constructor) {
  const levels = [];
  for (
    let proto = constructor.prototype;
    proto !== Object.prototype;
    proto = Object.getPrototypeOf(proto)
  ) {
    levels.push(
      Reflect.ownKeys(proto)
        .filter((key) => key !== 'constructor')
        .map((key) => {
          const descriptor = Object.getOwnPropertyDescriptor(proto, key);
          return [
            String(key),
            Boolean(descriptor.get),
            Boolean(descriptor.set),
            typeof descriptor.value,
          ];
        })
        .sort((a, b) => a[0].localeCompare(b[0])),
    );
  }
  return levels;
}

const [before, after] = await Promise.all(
  [reference, candidate].map(async (directory) => ({
    declarations: await readFile(path.join(directory, 'index.d.ts'), 'utf8'),
    runtime: await import(
      pathToFileURL(path.join(directory, 'index.mjs')).href
    ),
  })),
);
const tokens = declarationTokens(before.declarations);
assert.deepEqual(
  declarationTokens(after.declarations),
  tokens,
  'Public declaration tokens changed',
);
assert.deepEqual(
  Object.keys(after.runtime).sort(),
  Object.keys(before.runtime).sort(),
  'Public runtime exports changed',
);
assert.deepEqual(
  after.runtime.SyntaxKind,
  before.runtime.SyntaxKind,
  'SyntaxKind numeric values changed',
);
assert.deepEqual(
  prototypeSurface(after.runtime.RemoteSourceFile),
  prototypeSurface(before.runtime.RemoteSourceFile),
  'AST decoder prototype surface changed',
);
console.log(
  JSON.stringify(
    {
      status: 'pass',
      declarationTokensSha256: hash(JSON.stringify(tokens)),
      tokenCount: tokens.length,
      exports: Object.keys(after.runtime).sort(),
      syntaxKindEntries: Object.keys(after.runtime.SyntaxKind).length,
      decoderPrototypeLevels: prototypeSurface(after.runtime.RemoteSourceFile)
        .length,
    },
    null,
    2,
  ),
);
