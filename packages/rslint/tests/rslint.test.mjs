import { Rslint } from '@rslint/core';
import { lint } from '@rslint/core/internal';
import { describe, test, expect } from '@rstest/core';
import path from 'node:path';
import os from 'node:os';
import { writeFile, rm, mkdtemp, mkdir, readFile, cp } from 'node:fs/promises';

const fixturesDir = path.resolve(import.meta.dirname, '../fixtures');

// A self-contained config (overrideConfigFile:true → no discovery, only
// overrideConfig). array-type is non-type-aware so it runs even on a gap file.
const arrayTypeConfig = [
  {
    files: ['**/*.ts'],
    languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
    rules: { '@typescript-eslint/array-type': 'error' },
    plugins: ['@typescript-eslint'],
  },
];

describe('Rslint class', () => {
  test('published declarations expose async disposal under ES2022 consumer libs', async () => {
    const distDir = path.resolve(import.meta.dirname, '../dist');
    const packageJson = JSON.parse(
      await readFile(
        path.resolve(import.meta.dirname, '../package.json'),
        'utf8',
      ),
    );
    const distDts = await readFile(path.join(distDir, 'index.d.ts'), 'utf8');
    expect(distDts).toContain('reference lib="esnext.disposable"');
    expect(distDts).toContain('[Symbol.asyncDispose](): Promise<void>');

    const ts = await import('typescript');
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-dts-consumer-'));
    try {
      const packageRoot = path.join(tmp, 'node_modules', '@rslint', 'core');
      await mkdir(packageRoot, { recursive: true });
      await writeFile(
        path.join(packageRoot, 'package.json'),
        JSON.stringify({
          name: packageJson.name,
          type: packageJson.type,
          exports: {
            '.': packageJson.exports['.'],
          },
        }),
      );
      await cp(distDir, path.join(packageRoot, 'dist'), { recursive: true });
      const entry = path.join(tmp, 'index.ts');
      await writeFile(
        entry,
        [
          "import { Rslint } from '@rslint/core';",
          'const rslint: Rslint = new Rslint();',
          'rslint.close();',
          '',
        ].join('\n'),
      );

      const options = {
        target: ts.ScriptTarget.ES2022,
        module: ts.ModuleKind.NodeNext,
        moduleResolution: ts.ModuleResolutionKind.NodeNext,
        lib: ['lib.es2022.d.ts'],
        noEmit: true,
        strict: true,
        skipLibCheck: false,
      };
      const program = ts.createProgram([entry], options);
      const diagnostics = ts.getPreEmitDiagnostics(program);
      const rendered = diagnostics
        .map((d) => ts.flattenDiagnosticMessageText(d.messageText, '\n'))
        .join('\n');
      expect(rendered).not.toContain('asyncDispose');
      expect(diagnostics).toEqual([]);
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintText returns ESLint-shaped LintResult[]', async () => {
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: arrayTypeConfig,
    });
    try {
      const results = await rslint.lintText('let a: Array<string> = [];', {
        filePath: 'gap-rslint.ts',
      });
      expect(results).toHaveLength(1);
      const r = results[0];
      // filePath is absolute (ESLint contract; outputFixes guards on it).
      expect(path.isAbsolute(r.filePath)).toBe(true);
      expect(r.filePath.endsWith('gap-rslint.ts')).toBe(true);
      expect(r.errorCount).toBe(1);
      expect(r.warningCount).toBe(0);
      expect(r.messages).toHaveLength(1);
      const m = r.messages[0];
      expect(m.ruleId).toBe('@typescript-eslint/array-type');
      expect(m.severity).toBe(2); // error → 2
      expect(m.line).toBe(1);
      expect(m.column).toBe(8);
      expect(m.endLine).toBe(1);
      expect(m.endColumn).toBe(21);
      expect(m.messageId).toBe('errorStringArray');
      // fix is a flat UTF-16 range + replacement text (Array<string> → string[]).
      expect(m.fix.range).toEqual([7, 20]);
      expect(m.fix.text).toBe('string[]');
    } finally {
      await rslint.close();
    }
  });

  test('warn level maps to severity 1', async () => {
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          rules: { '@typescript-eslint/array-type': 'warn' },
          plugins: ['@typescript-eslint'],
        },
      ],
    });
    try {
      const results = await rslint.lintText('let a: Array<string> = [];', {
        filePath: 'gap-rslint-warn.ts',
      });
      const r = results[0];
      expect(r.errorCount).toBe(0);
      expect(r.warningCount).toBe(1);
      expect(r.messages[0].severity).toBe(1); // warn → 1
    } finally {
      await rslint.close();
    }
  });

  test('fix:true returns per-result output + fixable count', async () => {
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: arrayTypeConfig,
      fix: true,
    });
    try {
      const results = await rslint.lintText('let a: Array<string> = [];', {
        filePath: 'gap-rslint-fix.ts',
      });
      const r = results[0];
      expect(r.output).toBe('let a: string[] = [];');
      expect(r.fixableErrorCount).toBe(1);
      // Single-pass (design §8): messages still report the pre-fix diagnostics
      // — the fixable error is in `output` AND still listed, unlike ESLint's
      // post-fix re-lint. Pinned so this reverse-of-ESLint behavior is explicit.
      expect(r.messages).toHaveLength(1);
      expect(r.errorCount).toBe(1);
    } finally {
      await rslint.close();
    }
  });

  test('a clean file yields a result with zero messages', async () => {
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: arrayTypeConfig,
    });
    try {
      const results = await rslint.lintText('let a: string[] = [];', {
        filePath: 'gap-rslint-clean.ts',
      });
      expect(results).toHaveLength(1);
      expect(results[0].messages).toHaveLength(0);
      expect(results[0].errorCount).toBe(0);
    } finally {
      await rslint.close();
    }
  });

  test('lintFiles globs files and returns one result per file', async () => {
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: arrayTypeConfig,
    });
    try {
      const results = await rslint.lintFiles('src/index.ts');
      expect(results).toHaveLength(1);
      expect(path.isAbsolute(results[0].filePath)).toBe(true);
      expect(results[0].filePath.endsWith(path.join('src', 'index.ts'))).toBe(
        true,
      );
    } finally {
      await rslint.close();
    }
  });

  test('outputFixes writes output to absolute paths, skips no-output/relative', async () => {
    const { readFile, rm, mkdtemp } = await import('node:fs/promises');
    const os = await import('node:os');
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-outputfixes-'));
    const target = path.join(tmp, 'a.ts');
    const noOutput = path.join(tmp, 'b.ts');
    const base = {
      messages: [],
      errorCount: 0,
      warningCount: 0,
      fixableErrorCount: 0,
      fixableWarningCount: 0,
    };
    try {
      await Rslint.outputFixes([
        { ...base, filePath: target, output: 'let a: string[] = [];' },
        { ...base, filePath: noOutput }, // no output → skipped
        { ...base, filePath: '__rslint_relskip_probe__.ts', output: 'x' }, // relative → skipped
      ]);
      expect(await readFile(target, 'utf8')).toBe('let a: string[] = [];');
      let bWritten = true;
      try {
        await readFile(noOutput);
      } catch {
        bWritten = false;
      }
      expect(bWritten).toBe(false);
      // Relative path is skipped (isAbsolute guard): nothing written to cwd.
      let relWritten = true;
      try {
        await readFile(
          path.resolve(process.cwd(), '__rslint_relskip_probe__.ts'),
        );
      } catch {
        relWritten = false;
      }
      expect(relWritten).toBe(false);
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintText converts a rule suggestion to ESLint shape', async () => {
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          rules: { eqeqeq: 'error' },
        },
      ],
    });
    try {
      const results = await rslint.lintText(
        'const a = 1;\nconst b = 2;\nconst c = a == b;',
        { filePath: 'gap-eqeqeq.ts' },
      );
      const m = results[0].messages.find((x) => x.ruleId === 'eqeqeq');
      expect(m).toBeDefined();
      expect(m.suggestions).toHaveLength(1);
      const s = m.suggestions[0];
      expect(s.messageId).toBe('replaceOperator');
      expect(s.desc).toBe("Use '===' instead of '=='.");
      expect(s.fix.range).toHaveLength(2);
      expect(s.fix.text).toBe('===');
    } finally {
      await rslint.close();
    }
  });

  test('lintText without filePath uses the <text> sentinel, outputFixes skips it', async () => {
    const { readdir } = await import('node:fs/promises');
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: arrayTypeConfig,
      fix: true,
    });
    try {
      const results = await rslint.lintText('let a: Array<string> = [];'); // no filePath
      expect(results[0].filePath).toBe('<text>');
      expect(path.isAbsolute(results[0].filePath)).toBe(false);
      const before = (await readdir(fixturesDir)).length;
      await Rslint.outputFixes(results); // must skip the non-absolute <text>
      const after = await readdir(fixturesDir);
      expect(after.length).toBe(before); // no phantom __text__.ts written
      expect(after.includes('__text__.ts')).toBe(false);
    } finally {
      await rslint.close();
    }
  });

  test('auto-discovers config and appends overrideConfig', async () => {
    const { mkdtemp, writeFile, rm } = await import('node:fs/promises');
    const os = await import('node:os');
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-discover-'));
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        "export default [{ files: ['**/*.ts'], rules: { eqeqeq: 'error' } }];\n",
      );
      // No overrideConfigFile → auto-discover; overrideConfig appends no-var.
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfig: [{ files: ['**/*.ts'], rules: { 'no-var': 'error' } }],
      });
      try {
        const results = await rslint.lintText('var c = 1 == 2;', {
          filePath: 'a.ts',
        });
        const ruleIds = new Set(results[0].messages.map((m) => m.ruleId));
        expect(ruleIds.has('eqeqeq')).toBe(true); // from discovered config
        expect(ruleIds.has('no-var')).toBe(true); // from appended overrideConfig
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles returns one result per matched file, routed correctly', async () => {
    const { mkdtemp, writeFile, mkdir, rm } = await import('node:fs/promises');
    const os = await import('node:os');
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-multifile-'));
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        "export default [{ files: ['**/*.ts'], rules: { 'no-var': 'error' } }];\n",
      );
      await mkdir(path.join(tmp, 'sub'), { recursive: true });
      await writeFile(path.join(tmp, 'dirty.ts'), 'var x = 1;\n'); // 1 no-var
      await writeFile(path.join(tmp, 'sub', 'clean.ts'), 'const y = 1;\n'); // 0
      const rslint = new Rslint({ cwd: tmp });
      try {
        const results = await rslint.lintFiles('**/*.ts');
        const byBase = new Map(
          results.map((r) => [path.basename(r.filePath), r]),
        );
        expect(byBase.size).toBe(2);
        expect(byBase.get('dirty.ts').messages).toHaveLength(1);
        expect(byBase.get('dirty.ts').messages[0].ruleId).toBe('no-var');
        // clean file still produces a (zero-message) result.
        expect(byBase.get('clean.ts').messages).toHaveLength(0);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('multi-edit fix: merged message.fix applied equals Go output', async () => {
    // no-extra-bind emits TWO fix edits (drop `.bind` + the `(this)` arg);
    // mergeFixes collapses them into one span. Applying that JS-merged fix to
    // the source must equal Go's in-band output — exercises both the multi-edit
    // merge branch and JS↔Go fix agreement.
    const code = 'const f = (function () { return 1; }).bind(this);';
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          rules: { 'no-extra-bind': 'error' },
        },
      ],
      fix: true,
    });
    try {
      const results = await rslint.lintText(code, { filePath: 'gap-bind.ts' });
      const r = results[0];
      const m = r.messages.find((x) => x.ruleId === 'no-extra-bind');
      expect(m).toBeDefined();
      expect(m.fix).toBeDefined();
      expect(m.fix.range).toHaveLength(2);
      const applied =
        code.slice(0, m.fix.range[0]) + m.fix.text + code.slice(m.fix.range[1]);
      expect(applied).toBe(r.output);
    } finally {
      await rslint.close();
    }
  });

  test('cross-layer oracle: low-level lint() and Rslint.lintText agree field-by-field', async () => {
    const code = 'let a: Array<string> = [];';
    const cfg = [
      {
        files: ['**/*.ts'],
        languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
        rules: { '@typescript-eslint/array-type': 'error' },
        plugins: ['@typescript-eslint'],
      },
    ];
    const filePath = path.join(fixturesDir, 'gap-oracle.ts');
    // Low-level lint() → wire (Go) shape.
    const wire = await lint({
      config: cfg,
      configDirectory: fixturesDir,
      workingDirectory: fixturesDir,
      files: [filePath],
      fileContents: { [filePath]: code },
    });
    const d = wire.diagnostics.find(
      (x) => x.ruleName === '@typescript-eslint/array-type',
    );
    // High-level Rslint.lintText → ESLint (JS) shape.
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: cfg,
    });
    try {
      const results = await rslint.lintText(code, {
        filePath: 'gap-oracle.ts',
      });
      const m = results[0].messages.find(
        (x) => x.ruleId === '@typescript-eslint/array-type',
      );
      expect(d).toBeDefined();
      expect(m).toBeDefined();
      // Assert the reshape MAPPING (not two hardcoded equal constants):
      expect(m.severity).toBe(d.severity === 'error' ? 2 : 1);
      expect(m.ruleId).toBe(d.ruleName);
      expect(m.column).toBe(d.range.start.column);
      expect(m.endColumn).toBe(d.range.end.column);
      expect(m.fix.range[0]).toBe(d.fixes[0].startPos);
      expect(m.fix.range[1]).toBe(d.fixes[0].endPos);
      expect(m.fix.text).toBe(d.fixes[0].text);
    } finally {
      await rslint.close();
    }
  });

  test('lintFiles keeps fix.range BOM-stripped and re-prepends the BOM only to output', async () => {
    // A disk file whose bytes start with a UTF-8 BOM: Go reads it BOM-stripped,
    // so its fix offsets and Output carry no BOM. fix.range stays BOM-stripped
    // (matching ESLint v10, whose fix.range is relative to BOM-stripped source,
    // and the message column); only `output` gets the BOM re-prepended so
    // outputFixes writes back the real on-disk file. (lintText is unaffected —
    // its overlay keeps the BOM and Go's offsets already include it.)
    const BOM = String.fromCharCode(0xfeff);
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-bom-'));
    try {
      await writeFile(path.join(tmp, 'tsconfig.json'), '{}');
      await writeFile(
        path.join(tmp, 'bom.ts'),
        BOM + 'let a: Array<string> = [];\n',
      );
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: arrayTypeConfig,
        fix: true,
      });
      try {
        const results = await rslint.lintFiles('bom.ts');
        expect(results).toHaveLength(1);
        const m = results[0].messages[0];
        // Identical to the no-BOM lintText case (line 47): BOM-stripped [7, 20].
        expect(m.fix.range).toEqual([7, 20]);
        expect(m.fix.text).toBe('string[]');
        // Only output carries the BOM.
        expect(results[0].output).toBe(BOM + 'let a: string[] = [];\n');
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles multi-edit fix on a BOM-prefixed file: merged range BOM-stripped, output keeps BOM', async () => {
    // no-extra-bind emits a multi-edit fix; on a BOM-prefixed disk file the
    // merged range must stay BOM-stripped (Go's coordinate space) so applying it
    // to the BOM-stripped body reproduces output minus its BOM. Exercises the
    // multi-edit merge branch together with the strip-then-re-prepend path.
    const BOM = String.fromCharCode(0xfeff);
    const body = 'const f = (function () { return 1; }).bind(this);\n';
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-bom-multi-'));
    try {
      await writeFile(path.join(tmp, 'tsconfig.json'), '{}');
      await writeFile(path.join(tmp, 'bind.ts'), BOM + body);
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          {
            files: ['**/*.ts'],
            languageOptions: {
              parserOptions: { project: ['./tsconfig.json'] },
            },
            rules: { 'no-extra-bind': 'error' },
          },
        ],
        fix: true,
      });
      try {
        const results = await rslint.lintFiles('bind.ts');
        const m = results[0].messages.find((x) => x.ruleId === 'no-extra-bind');
        expect(m).toBeDefined();
        expect(m.fix).toBeDefined();
        // BOM-stripped range: applying it to the BOM-stripped body equals output
        // minus its re-prepended BOM.
        const applied =
          body.slice(0, m.fix.range[0]) +
          m.fix.text +
          body.slice(m.fix.range[1]);
        expect(results[0].output).toBe(BOM + applied);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles returns no result when every match is config-ignored', async () => {
    // Regression guard for the empty-lintedFiles wire case: when all glob
    // matches are excluded by config `ignores`, Go returns an empty (non-nil)
    // lintedFiles array — NOT an omitted field — so the class yields zero
    // results instead of falling back to the glob matches and seeding phantom
    // empty results.
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-ignored-'));
    try {
      await writeFile(path.join(tmp, 'tsconfig.json'), '{}');
      await writeFile(
        path.join(tmp, 'ignored.ts'),
        'let a: Array<string> = [];\n',
      );
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          { ignores: ['ignored.ts'] },
          {
            files: ['**/*.ts'],
            languageOptions: {
              parserOptions: { project: ['./tsconfig.json'] },
            },
            rules: { '@typescript-eslint/array-type': 'error' },
            plugins: ['@typescript-eslint'],
          },
        ],
      });
      try {
        const results = await rslint.lintFiles('ignored.ts');
        expect(results).toEqual([]);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  // Regression for the unref() fix: a script that lints and never calls close()
  // must still let the Node process exit. Without unref the resident `--api`
  // child + its stdio pipes keep the event loop alive and the process hangs.
  test('lintText without close() lets the process exit on its own (no hang)', async () => {
    const { spawn } = await import('node:child_process');
    const script = path.resolve(
      import.meta.dirname,
      'fixtures',
      'no-close-exit.mjs',
    );
    const child = spawn(process.execPath, [script], {
      cwd: path.resolve(import.meta.dirname, '..'),
      stdio: 'inherit',
    });
    const code = await new Promise((resolve) => {
      const timer = setTimeout(() => {
        child.kill('SIGKILL');
        resolve('TIMEOUT-HANG');
      }, 20000);
      child.on('exit', (c) => {
        clearTimeout(timer);
        resolve(c);
      });
    });
    // 0 = the fixture linted (it asserts exactly 1 message) AND the process
    // exited without close(); 'TIMEOUT-HANG' = unref regressed.
    expect(code).toBe(0);
  });

  // Fully in-memory (issue #1106): config object + in-memory tsconfig via
  // `virtualFiles`, type-aware rule, ZERO disk. Empty temp dir as cwd + path.join
  // keys so the tsconfig and the config's `project` resolve to one path on every OS.
  test('lintText runs type-aware rules with an in-memory tsconfig (zero disk)', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-virtual-'));
    const rslint = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        // Explicit tsconfig `files` (not an `include` glob): a glob is expanded
        // against the overlay-over-real-FS, which would scan the cwd on disk.
        [path.join(tmp, 'tsconfig.json')]: JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./src/probe.ts'],
        }),
      },
    });
    try {
      const results = await rslint.lintText(
        'const arr = [1, 2, 3];\nfor (const i in arr) {\n}\n',
        { filePath: path.join(tmp, 'src', 'probe.ts') },
      );
      expect(results).toHaveLength(1);
      const messages = results[0].messages;
      // no-for-in-array is type-aware (it asks the TypeChecker whether `arr` is
      // an array), so it can only fire if the in-memory tsconfig built a real
      // program over the overlay — proving fully-in-memory type-aware linting.
      expect(messages).toHaveLength(1);
      expect(messages[0].ruleId).toBe('@typescript-eslint/no-for-in-array');
      expect(messages[0].severity).toBe(2);
      expect(messages[0].messageId).toBe('forInViolation');
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  // Cross-file in-memory type resolution: the linted buffer imports a symbol
  // from ANOTHER in-memory overlay file (dep.ts). no-for-in-array fires only if
  // the checker resolved `nums`'s array type across the in-memory import — i.e.
  // the overlay is one connected program, not isolated files.
  test('lintText resolves type info across in-memory dependency files', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-virtual-'));
    const rslint = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        [path.join(tmp, 'tsconfig.json')]: JSON.stringify({
          compilerOptions: { strict: true, moduleResolution: 'bundler' },
          files: ['./probe.ts', './dep.ts'],
        }),
        [path.join(tmp, 'dep.ts')]:
          'export const nums: number[] = [1, 2, 3];\n',
      },
    });
    try {
      const results = await rslint.lintText(
        "import { nums } from './dep';\nfor (const i in nums) {\n}\n",
        { filePath: path.join(tmp, 'probe.ts') },
      );
      expect(results).toHaveLength(1);
      const messages = results[0].messages;
      expect(messages).toHaveLength(1);
      expect(messages[0].ruleId).toBe('@typescript-eslint/no-for-in-array');
      expect(messages[0].severity).toBe(2);
      expect(messages[0].messageId).toBe('forInViolation');
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  // lintFiles + virtualFiles: glob a REAL on-disk source file but supply the
  // tsconfig in-memory via the instance overlay — the overlay is threaded into
  // lintFiles too (not just lintText), so type-aware rules run over disk files
  // with no tsconfig on disk.
  test('lintFiles runs type-aware rules with an in-memory tsconfig overlay', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-virtual-'));
    await writeFile(
      path.join(tmp, 'probe.ts'),
      'const arr = [1, 2, 3];\nfor (const i in arr) {\n}\n',
    );
    const rslint = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        [path.join(tmp, 'tsconfig.json')]: JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./probe.ts'],
        }),
      },
    });
    try {
      const results = await rslint.lintFiles('probe.ts');
      expect(results).toHaveLength(1);
      const messages = results[0].messages;
      expect(messages).toHaveLength(1);
      expect(messages[0].ruleId).toBe('@typescript-eslint/no-for-in-array');
      expect(messages[0].severity).toBe(2);
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  // Path-style coverage: relative `virtualFiles` keys + relative filePath resolve
  // against cwd, same as absolute keys, on every OS.
  test('lintText accepts relative virtualFiles keys', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-virtual-'));
    const rslint = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        'tsconfig.json': JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./probe.ts'],
        }),
      },
    });
    try {
      const results = await rslint.lintText(
        'const arr = [1, 2, 3];\nfor (const i in arr) {\n}\n',
        { filePath: 'probe.ts' },
      );
      expect(results[0].messages).toHaveLength(1);
      expect(results[0].messages[0].ruleId).toBe(
        '@typescript-eslint/no-for-in-array',
      );
      expect(results[0].messages[0].severity).toBe(2);
      expect(results[0].messages[0].messageId).toBe('forInViolation');
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  // A bare POSIX-absolute key with cwd '/' is re-anchored against cwd (→ `C:\…` on
  // Windows), matching where the config resolves — the cross-platform key fix.
  test('lintText re-anchors a POSIX-absolute virtualFiles key (cwd "/")', async () => {
    const rslint = new Rslint({
      cwd: '/',
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        '/tsconfig.json': JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./probe.ts'],
        }),
      },
    });
    try {
      const results = await rslint.lintText(
        'const arr = [1, 2, 3];\nfor (const i in arr) {\n}\n',
        { filePath: '/probe.ts' },
      );
      expect(results[0].messages).toHaveLength(1);
      expect(results[0].messages[0].ruleId).toBe(
        '@typescript-eslint/no-for-in-array',
      );
      expect(results[0].messages[0].severity).toBe(2);
      expect(results[0].messages[0].messageId).toBe('forInViolation');
    } finally {
      await rslint.close();
    }
  });

  // The real fully-in-memory shape: cwd '/' (no process.cwd()) with all paths
  // relative — virtualFiles key, filePath, project, and the tsconfig `files`.
  // Relative paths anchor to cwd identically on every OS; Windows CI is the
  // real cross-platform check (a macOS host can't simulate the drive letter).
  test('lintText runs fully in-memory with cwd "/" and all relative paths', async () => {
    const rslint = new Rslint({
      cwd: '/',
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        'tsconfig.json': JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./probe.ts'],
        }),
      },
    });
    try {
      const results = await rslint.lintText(
        'const arr = [1, 2, 3];\nfor (const i in arr) {\n}\n',
        { filePath: 'probe.ts' },
      );
      expect(results[0].messages).toHaveLength(1);
      expect(results[0].messages[0].ruleId).toBe(
        '@typescript-eslint/no-for-in-array',
      );
      expect(results[0].messages[0].severity).toBe(2);
      expect(results[0].messages[0].messageId).toBe('forInViolation');
    } finally {
      await rslint.close();
    }
  });

  // ESLint's lintText returns exactly one result — for the linted buffer. An
  // in-memory dependency file that matches the config and carries its own
  // violation must NOT leak a second result (which outputFixes would then write).
  test('lintText returns a single result even if an overlay dependency file has violations', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-virtual-'));
    const rslint = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        [path.join(tmp, 'tsconfig.json')]: JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./probe.ts', './dep.ts'],
        }),
        // dep.ts itself violates no-for-in-array.
        [path.join(tmp, 'dep.ts')]:
          'export const xs = [1, 2, 3];\nfor (const k in xs) {\n}\n',
      },
    });
    try {
      const results = await rslint.lintText(
        "import { xs } from './dep';\nvoid xs;\n",
        { filePath: path.join(tmp, 'probe.ts') },
      );
      expect(results).toHaveLength(1);
      expect(results[0].filePath).toBe(path.join(tmp, 'probe.ts'));
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  // lintText reports BOM-INCLUSIVE offsets for BOM-prefixed code (known
  // limitation, one ahead of ESLint v10 which strips the BOM). Pinned without
  // hardcoding offsets: the same code with a leading BOM shifts every offset +1.
  test('lintText reports BOM-inclusive offsets for BOM-prefixed code (known limitation)', async () => {
    const rslint = new Rslint({
      cwd: '/',
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/array-type': 'error' },
        },
      ],
      virtualFiles: {
        'tsconfig.json': JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./probe.ts'],
        }),
      },
    });
    try {
      const code = 'let a: Array<string> = [];\n';
      const [plain] = await rslint.lintText(code, { filePath: 'probe.ts' });
      const [bom] = await rslint.lintText('\uFEFF' + code, {
        filePath: 'probe.ts',
      });
      expect(plain.messages).toHaveLength(1);
      expect(bom.messages).toHaveLength(1);
      const pm = plain.messages[0];
      const bm = bom.messages[0];
      expect(pm.ruleId).toBe('@typescript-eslint/array-type');
      // The BOM shifts every offset by exactly one UTF-16 unit.
      expect(bm.column).toBe(pm.column + 1);
      expect(bm.fix.range[0]).toBe(pm.fix.range[0] + 1);
      expect(bm.fix.range[1]).toBe(pm.fix.range[1] + 1);
    } finally {
      await rslint.close();
    }
  });

  // errorCount/warningCount/fixable* are bucketed by severity; a file mixing an
  // error and a warning must split them, not collapse into one bucket.
  test('lintText splits error/warning counts for a file mixing severities', async () => {
    const rslint = new Rslint({
      cwd: '/',
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: {
            '@typescript-eslint/array-type': 'error', // fixable error
            '@typescript-eslint/no-for-in-array': 'warn', // non-fixable warning
          },
        },
      ],
      virtualFiles: {
        'tsconfig.json': JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./probe.ts'],
        }),
      },
    });
    try {
      const [r] = await rslint.lintText(
        'const a: Array<number> = [1];\nfor (const k in a) {\n}\n',
        { filePath: 'probe.ts' },
      );
      expect(r.errorCount).toBe(1);
      expect(r.warningCount).toBe(1);
      expect(r.fixableErrorCount).toBe(1); // array-type is fixable
      expect(r.fixableWarningCount).toBe(0); // no-for-in-array has no fix
    } finally {
      await rslint.close();
    }
  });
});
