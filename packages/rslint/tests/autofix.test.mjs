import { Rslint } from '@rslint/core';
import { describe, test, expect } from 'rstack/test';
import { mkdtemp, rm, symlink, writeFile } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';

describe('Rslint autofix', () => {
  test('remaining multi-edit metadata targets the final output', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-api-fix-limit-'));
    const pluginPath = path.join(tmp, 'blocking-plugin.mjs');
    await writeFile(
      pluginPath,
      `export default {
		rules: {
			block: {
				meta: {
					fixable: 'code',
					schema: [],
					messages: { block: 'block' },
				},
				create(context) {
					return { Program(node) {
						context.report({
								node,
								messageId: 'block',
								fix(fixer) {
									const text = context.sourceCode.text;
									const offset = text.indexOf('if (value)');
									return fixer.replaceTextRange(
										[0, text.length],
										text.slice(0, offset) + ' ' + text.slice(offset),
									);
								},
							});
						} };
					},
				},
			},
		};\n`,
    );
    await writeFile(
      path.join(tmp, 'rslint.config.mjs'),
      `import blocking from ${JSON.stringify(pathToFileURL(pluginPath).href)};
		export default [{
			plugins: { blocking },
			rules: {
				'blocking/block': 'error',
				'no-else-return': 'error',
				'no-prototype-builtins': 'error',
			},
		}];\n`,
    );

    const rslint = new Rslint({ cwd: tmp, fix: true });
    const nativeFixer = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [{ rules: { 'no-else-return': 'error' } }],
      fix: true,
    });
    const BOM = String.fromCharCode(0xfeff);
    const source =
      BOM +
      `const prefix = 0;
function f(value) {
  const ownsKey = value.hasOwnProperty('key');
  if (value) {
    return 1;
  } else {
    return 2;
  }
}\n`;
    try {
      const [result] = await rslint.lintText(source, { filePath: 'probe.js' });
      let finalSource = source;
      for (let round = 0; round < 10; round++) {
        finalSource = finalSource.replace('if (value)', ' if (value)');
      }
      expect(result.output).toBe(finalSource);
      const finalText = finalSource.slice(BOM.length);
      const message = result.messages.find(
        (candidate) => candidate.ruleId === 'no-else-return',
      );
      expect(message).toBeDefined();
      expect(message.fix).toBeDefined();
      const [expected] = await nativeFixer.lintText(finalSource, {
        filePath: 'expected.js',
      });
      const applied =
        finalText.slice(0, message.fix.range[0]) +
        message.fix.text +
        finalText.slice(message.fix.range[1]);
      expect(applied).toBe(expected.output.slice(BOM.length));

      const suggestionMessage = result.messages.find(
        (candidate) => candidate.ruleId === 'no-prototype-builtins',
      );
      const suggestion = suggestionMessage?.suggestions?.[0]?.fix;
      expect(suggestion).toBeDefined();
      const suggestionApplied =
        finalText.slice(0, suggestion.range[0]) +
        suggestion.text +
        finalText.slice(suggestion.range[1]);
      expect(suggestionApplied).toBe(
        finalText.replace(
          "value.hasOwnProperty('key')",
          "Object.prototype.hasOwnProperty.call(value, 'key')",
        ),
      );
    } finally {
      await Promise.all([rslint.close(), nativeFixer.close()]);
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('fix:true preserves an explicitly empty output', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-api-empty-fix-'));
    const pluginPath = path.join(tmp, 'empty-plugin.mjs');
    await writeFile(
      pluginPath,
      `export default {
			rules: {
				empty: {
					meta: { fixable: 'code', schema: [], messages: { empty: 'empty' } },
					create(context) {
						return { Program(node) {
							if (context.sourceCode.text.length === 0) return;
							context.report({
								node,
								messageId: 'empty',
								fix: (fixer) => fixer.removeRange([0, context.sourceCode.text.length]),
							});
						} };
					},
				},
			},
		};\n`,
    );
    await writeFile(
      path.join(tmp, 'rslint.config.mjs'),
      `import empty from ${JSON.stringify(pathToFileURL(pluginPath).href)};
		export default [{ plugins: { empty }, rules: { 'empty/empty': 'error' } }];\n`,
    );

    const rslint = new Rslint({ cwd: tmp, fix: true });
    try {
      const [result] = await rslint.lintText('removeMe();', {
        filePath: 'probe.js',
      });
      expect(result.output).toBe('');
      expect(result.messages).toEqual([]);
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('fix:false returns community-plugin edits without applying them', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-api-edit-info-'));
    const pluginPath = path.join(tmp, 'edit-info-plugin.mjs');
    await writeFile(
      pluginPath,
      `export default {
        rules: {
          rename: {
            meta: {
              type: 'suggestion',
              fixable: 'code',
              schema: [],
              messages: { rename: 'Rename bad' },
            },
            create(context) {
              return {
                Identifier(node) {
                  if (node.name !== 'bad') return;
                  context.report({
                    node,
                    messageId: 'rename',
                    fix: (fixer) => fixer.replaceText(node, 'good'),
                  });
                },
              };
            },
          },
          choices: {
            meta: {
              type: 'suggestion',
              hasSuggestions: true,
              schema: [],
              messages: {
                choose: 'Choose uppercase names',
                uppercase: 'Uppercase {{ first }} and {{ second }}',
                empty: 'Empty suggestion',
              },
            },
            create(context) {
              let left;
              let right;
              return {
                Identifier(node) {
                  if (node.name === 'left') left = node;
                  if (node.name === 'right') right = node;
                },
                'Program:exit'() {
                  if (!left || !right) return;
                  context.report({
                    node: left,
                    messageId: 'choose',
                    suggest: [
                      {
                        messageId: 'uppercase',
                        data: { first: 'left', second: 'right' },
                        fix: (fixer) => [
                          fixer.replaceText(left, 'LEFT'),
                          fixer.replaceText(right, 'RIGHT'),
                        ],
                      },
                      {
                        messageId: 'empty',
                        fix: () => null,
                      },
                    ],
                  });
                },
              };
            },
          },
        },
      };\n`,
    );
    await writeFile(
      path.join(tmp, 'rslint.config.mjs'),
      `import editInfo from ${JSON.stringify(pathToFileURL(pluginPath).href)};
      export default [{
        plugins: { probe: editInfo },
        rules: {
          'probe/rename': 'error',
          'probe/choices': 'warn',
        },
      }];\n`,
    );

    const source = 'const bad = 1;\nconst choice = left + right;\n';
    const rslint = new Rslint({ cwd: tmp, fix: false });
    try {
      const [result] = await rslint.lintText(source, {
        filePath: 'probe.js',
      });
      const rename = result.messages.find(
        (message) => message.ruleId === 'probe/rename',
      );
      expect(rename?.fix).toBeDefined();
      expect(
        source.slice(0, rename.fix.range[0]) +
          rename.fix.text +
          source.slice(rename.fix.range[1]),
      ).toBe('const good = 1;\nconst choice = left + right;\n');

      const choices = result.messages.find(
        (message) => message.ruleId === 'probe/choices',
      );
      expect(choices?.suggestions).toHaveLength(1);
      expect(choices.suggestions[0].messageId).toBe('uppercase');
      expect(choices.suggestions[0].desc).toBe('Uppercase left and right');
      const suggestion = choices.suggestions[0].fix;
      expect(
        source.slice(0, suggestion.range[0]) +
          suggestion.text +
          source.slice(suggestion.range[1]),
      ).toBe('const bad = 1;\nconst choice = LEFT + RIGHT;\n');

      expect(result.errorCount).toBe(1);
      expect(result.warningCount).toBe(1);
      expect(result.fixableErrorCount).toBe(1);
      expect(result.fixableWarningCount).toBe(0);
      expect(result.output).toBeUndefined();
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles omits atomic multi-edits when their source disappears', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-edit-source-'),
    );
    const target = path.join(tmp, 'probe.js');
    const pluginPath = path.join(tmp, 'delete-source-plugin.mjs');
    const source = [
      'const f = (function () { return 1; }).bind(this);',
      "value.hasOwnProperty('key');",
      '',
    ].join('\n');
    await writeFile(target, source);
    await writeFile(
      pluginPath,
      `import { unlinkSync } from 'node:fs';
      export default {
        rules: {
          source: {
            meta: { schema: [] },
            create() {
              return { Program() { unlinkSync(${JSON.stringify(target)}); } };
            },
          },
        },
      };\n`,
    );
    await writeFile(
      path.join(tmp, 'rslint.config.mjs'),
      `import deleteSource from ${JSON.stringify(pathToFileURL(pluginPath).href)};
      export default [{
        plugins: { destroy: deleteSource },
        rules: {
          'destroy/source': 'error',
          'no-extra-bind': 'error',
          'no-prototype-builtins': 'error',
        },
      }];\n`,
    );

    const rslint = new Rslint({ cwd: tmp, fix: false });
    try {
      const [result] = await rslint.lintFiles(target);
      const autofix = result.messages.find(
        (message) => message.ruleId === 'no-extra-bind',
      );
      const suggestion = result.messages.find(
        (message) => message.ruleId === 'no-prototype-builtins',
      );
      expect(autofix).toBeDefined();
      expect(autofix.fix).toBeUndefined();
      expect(suggestion).toBeDefined();
      expect(suggestion.suggestions).toBeUndefined();
      expect(result.fixableErrorCount).toBe(0);
      expect(result.output).toBeUndefined();
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles merges aliased multi-edits from the canonical virtual source', async () => {
    // Both native rules emit two wire edits spanning source text. Their gap
    // text must come from the canonical-path overlay because the ranges
    // describe that source generation, even when lintFiles selected an alias
    // whose physical file contains different text.
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-overlay-fix-'));
    const realTarget = path.join(tmp, 'real.js');
    const aliasTarget = path.join(tmp, 'alias.js');
    const source = (marker, callGap, eol) =>
      [
        'function choose(value) {',
        `  const marker = ${JSON.stringify(marker)};`,
        `  const ownsKey = value.hasOwnProperty${callGap}('key');`,
        '  if (value) {',
        '    return 1;',
        '  } else {',
        '    return 2;',
        '  }',
        '}',
        '',
      ].join(eol);
    const diskSource = source('diskMarker', ' /* disk gap */ ', '\n');
    const overlaySource = source(
      'overlayMarker🧪',
      ' /* overlay gap 🧪 */ ',
      '\r\n',
    );
    await writeFile(realTarget, diskSource);
    try {
      await symlink(realTarget, aliasTarget);
    } catch {
      await rm(tmp, { recursive: true, force: true });
      return;
    }

    const fixInspector = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [{ rules: { 'no-else-return': 'error' } }],
      virtualFiles: { [realTarget]: overlaySource },
    });
    // This request deliberately has no other multi-edit diagnostic: source
    // loading therefore depends on the suggestion branch itself.
    const suggestionInspector = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [{ rules: { 'no-prototype-builtins': 'error' } }],
      virtualFiles: { [realTarget]: overlaySource },
    });
    const fixer = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [{ rules: { 'no-else-return': 'error' } }],
      fix: true,
    });
    try {
      const [fixResult] = await fixInspector.lintFiles('alias.js');
      expect(fixResult.filePath).toBe(path.normalize(aliasTarget));
      const message = fixResult.messages.find(
        (candidate) => candidate.ruleId === 'no-else-return',
      );
      expect(message).toBeDefined();
      expect(message.fix).toBeDefined();

      const [fixed] = await fixer.lintText(overlaySource, {
        filePath: 'probe.js',
      });
      const applied =
        overlaySource.slice(0, message.fix.range[0]) +
        message.fix.text +
        overlaySource.slice(message.fix.range[1]);
      expect(applied).toBe(fixed.output);
      expect(applied).toContain('overlayMarker🧪');
      expect(applied).not.toContain('diskMarker');

      const [suggestionResult] =
        await suggestionInspector.lintFiles('alias.js');
      expect(suggestionResult.output).toBeUndefined();
      const suggestionMessage = suggestionResult.messages.find(
        (candidate) => candidate.ruleId === 'no-prototype-builtins',
      );
      const suggestion = suggestionMessage?.suggestions?.[0]?.fix;
      expect(suggestion).toBeDefined();
      const suggestionApplied =
        overlaySource.slice(0, suggestion.range[0]) +
        suggestion.text +
        overlaySource.slice(suggestion.range[1]);
      expect(suggestionApplied).toBe(
        overlaySource.replace(
          "value.hasOwnProperty /* overlay gap 🧪 */ ('key')",
          "Object.prototype.hasOwnProperty.call /* overlay gap 🧪 */ (value, 'key')",
        ),
      );
      expect(suggestionApplied).toContain('overlayMarker🧪');
      expect(suggestionApplied).toContain('overlay gap 🧪');
      expect(suggestionApplied).not.toContain('disk gap');
      expect(suggestionApplied).not.toContain('diskMarker');
    } finally {
      await Promise.all([
        fixInspector.close(),
        suggestionInspector.close(),
        fixer.close(),
      ]);
      await rm(tmp, { recursive: true, force: true });
    }
  });
});
