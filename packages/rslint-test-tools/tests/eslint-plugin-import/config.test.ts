import path from 'node:path';
import { parse } from '@typescript-eslint/parser';
import { expect, test } from 'rstack/test';
import {
  fixturePathLiteral,
  fixtureUpperRoot,
  resolveFixturePaths,
} from './config';

const roots = [
  { name: 'ordinary path', root: '/tmp/fixture' },
  { name: 'apostrophe', root: "/tmp/O'Neil" },
  { name: 'double quote', root: '/tmp/a"b' },
  { name: 'backslash before quotes', root: '/tmp/a\\\'b\\"c' },
  { name: 'trailing backslash', root: '/tmp/fixture\\' },
  { name: 'line breaks', root: '/tmp/a\nb\rc\u2028d\u2029e' },
  { name: 'control characters', root: '/tmp/a\u00001\t\b\f' },
  { name: 'non-BMP character', root: '/tmp/\u{1f600}' },
  { name: 'literal escape sequence', root: '/tmp/\\u0027' },
  { name: 'replacement tokens', root: "/tmp/$&-$`-$'-$$" },
  { name: 'placeholder in path', root: '/tmp/<fixture>-<fixture-upper>' },
  { name: 'Windows path', root: "C:\\Users\\O'Neil\\Temp" },
];

for (const quote of ["'", '"'] as const) {
  for (const prefix of ['<fixture>', '<fixture-upper>']) {
    test.each(roots)(`${quote} ${prefix}: $name`, ({ root }) => {
      const value = `${prefix}/foo.js`;
      const literal = fixturePathLiteral(value, root, quote);
      const code = `// fixture import\nimport f from ${literal};`;
      const ast = parse(code, { sourceType: 'module', loc: true, range: true });
      const statement = ast.body[0];
      expect(statement.type).toBe('ImportDeclaration');
      if (statement.type !== 'ImportDeclaration')
        throw new Error('Expected an import declaration');
      const expectedPath =
        (prefix === '<fixture-upper>'
          ? fixtureUpperRoot(root)
          : root.replaceAll(path.sep, '/')) + '/foo.js';
      expect(statement.source.value).toBe(expectedPath);
      expect(literal.startsWith(quote) && literal.endsWith(quote)).toBe(true);
      expect(statement.source.loc).toEqual({
        start: { line: 2, column: 14 },
        end: { line: 2, column: 14 + literal.length },
      });
      expect(code.slice(...statement.source.range)).toBe(literal);
      expect(resolveFixturePaths(`Casing of ${value}`, root)).toBe(
        `Casing of ${expectedPath}`,
      );
    });
  }
}
