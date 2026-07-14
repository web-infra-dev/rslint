import { describe, test, expect } from '@rstest/core';
import fs from 'node:fs';
import path from 'node:path';
import { runRslint, createTempDir, cleanupTempDir } from './helpers.js';

interface Diagnostic {
  ruleName: string;
  filePath: string;
  severity: string;
}

async function lintJsonline(
  files: Record<string, string>,
  args: string[] = [],
): Promise<{
  diagnostics: Diagnostic[];
  stdout: string;
  stderr: string;
  cleanup: () => Promise<void>;
}> {
  const tempDir = await createTempDir(files);
  const result = await runRslint(['--format', 'jsonline', ...args], tempDir);
  const lines = result.stdout
    .trim()
    .split('\n')
    .filter((l) => l.trim());
  const diagnostics = lines.map((l) => JSON.parse(l) as Diagnostic);
  return {
    diagnostics,
    stdout: result.stdout,
    stderr: result.stderr,
    cleanup: () => cleanupTempDir(tempDir),
  };
}

function diagsAt(diagnostics: Diagnostic[], pathPart: string): Diagnostic[] {
  return diagnostics.filter(
    (d) => d.filePath === pathPart || d.filePath.startsWith(pathPart + '/'),
  );
}

const CONFIG_NO_CONSOLE = `export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];`;

describe('Gitignore: basic patterns', () => {
  test('simple directory pattern blocks dir contents', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'dist/\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'dist/bundle.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'dist').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('file extension pattern blocks matching files', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': '*.generated.ts\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'src/types.generated.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'src/types.generated.ts').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('bare directory name blocks matching dir', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'build\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'build/output.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'build').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('no .gitignore — no effect, no crash', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'dist/bundle.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'dist/bundle.ts').length).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('empty .gitignore — no effect', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': '',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'dist/bundle.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'dist/bundle.ts').length).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('comments and blank lines are ignored', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': '# Build output\ndist/\n\n# Logs\n*.log\n\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'dist/bundle.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'dist').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('multiple patterns all apply', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'dist/\ncoverage/\n*.tmp.ts\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'dist/bundle.ts': 'console.log("test");\n',
      'coverage/report.ts': 'console.log("test");\n',
      'src/temp.tmp.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'dist').length).toBe(0);
      expect(diagsAt(diagnostics, 'coverage').length).toBe(0);
      expect(diagsAt(diagnostics, 'src/temp.tmp.ts').length).toBe(0);
    } finally {
      await cleanup();
    }
  });
});

describe('Gitignore: inheritance (root patterns affect nested dirs)', () => {
  test('root dist/ blocks packages/app/dist/ too', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'dist/\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'dist/bundle.ts': 'console.log("test");\n',
      'packages/app/dist/output.ts': 'console.log("test");\n',
      'packages/app/src/index.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(
        diagsAt(diagnostics, 'packages/app/src/index.ts').length,
      ).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'dist').length).toBe(0);
      expect(diagsAt(diagnostics, 'packages/app/dist').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('root-anchored /dist only blocks root-level dist', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': '/dist\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'dist/bundle.ts': 'console.log("test");\n',
      'packages/app/dist/output.ts': 'console.log("test");\n',
      'src/index.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'dist').length).toBe(0);
      // /dist is root-anchored → packages/app/dist/ NOT ignored
      expect(
        diagsAt(diagnostics, 'packages/app/dist/output.ts').length,
      ).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });
});

describe('Gitignore: nested .gitignore files', () => {
  test('child .gitignore adds pattern scoped to subtree only', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'dist/\n',
      'packages/app/.gitignore': 'src/\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'packages/app/src/index.ts': 'console.log("test");\n',
      'packages/app/lib/utils.ts': 'console.log("test");\n',
    });
    try {
      // Root src/ NOT affected by child .gitignore
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      // Child src/ IS blocked by child .gitignore
      expect(diagsAt(diagnostics, 'packages/app/src').length).toBe(0);
      // Child lib/ not blocked
      expect(
        diagsAt(diagnostics, 'packages/app/lib/utils.ts').length,
      ).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('child .gitignore negates parent pattern with !', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'dist/\n',
      'packages/app/.gitignore': '!dist/\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'dist/root-bundle.ts': 'console.log("test");\n',
      'packages/app/dist/app-bundle.ts': 'console.log("test");\n',
      'packages/app/src/index.ts': 'console.log("test");\n',
    });
    try {
      // Root dist/ still blocked by root .gitignore
      expect(diagsAt(diagnostics, 'dist/root-bundle.ts').length).toBe(0);
      // Child dist/ re-included by child !dist/
      expect(
        diagsAt(diagnostics, 'packages/app/dist/app-bundle.ts').length,
      ).toBeGreaterThan(0);
      expect(
        diagsAt(diagnostics, 'packages/app/src/index.ts').length,
      ).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('child bare negation re-includes a parent-ignored directory', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'debug/\n',
      'packages/app/.gitignore': '!debug\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'debug/root.ts': 'console.log("test");\n',
      'packages/app/debug/app.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'debug/root.ts').length).toBe(0);
      expect(
        diagsAt(diagnostics, 'packages/app/debug/app.ts').length,
      ).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('three-level nested .gitignore cascade', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'dist/\n',
      'packages/app/.gitignore': 'tmp/\n',
      'packages/app/sub/.gitignore': 'cache/\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'packages/app/sub/src/index.ts': 'console.log("test");\n',
      'packages/app/sub/cache/data.ts': 'console.log("test");\n',
      'packages/app/tmp/temp.ts': 'console.log("test");\n',
      'dist/bundle.ts': 'console.log("test");\n',
    });
    try {
      expect(
        diagsAt(diagnostics, 'packages/app/sub/src/index.ts').length,
      ).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'packages/app/sub/cache').length).toBe(0);
      expect(diagsAt(diagnostics, 'packages/app/tmp').length).toBe(0);
      expect(diagsAt(diagnostics, 'dist').length).toBe(0);
    } finally {
      await cleanup();
    }
  });
});

describe('Gitignore: interaction with config ignores', () => {
  test('both .gitignore and config ignores apply independently', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'dist/\n',
      'rslint.config.mjs': `export default [
        { ignores: ["coverage/**"] },
        { files: ["**/*.ts"], rules: { "no-console": "error" } }
      ];`,
      'src/index.ts': 'console.log("test");\n',
      'dist/bundle.ts': 'console.log("test");\n',
      'coverage/report.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'dist').length).toBe(0);
      expect(diagsAt(diagnostics, 'coverage').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('config ! negation can override .gitignore', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'dist/\n',
      'rslint.config.mjs': `export default [
        { ignores: ["dist/**/*", "!dist/keep.ts"] },
        { files: ["**/*.ts"], rules: { "no-console": "error" } }
      ];`,
      'dist/keep.ts': 'console.log("test");\n',
      'dist/other.ts': 'console.log("test");\n',
      'src/index.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'dist/keep.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'dist/other.ts').length).toBe(0);
    } finally {
      await cleanup();
    }
  });
});

describe('Gitignore: monorepo scenarios', () => {
  test('package configs do not inherit root .gitignore', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'dist/\ncoverage/\n',
      'rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/app/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];`,
      'packages/lib/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];`,
      'packages/app/src/index.ts': 'console.log("test");\n',
      'packages/app/dist/bundle.ts': 'console.log("test");\n',
      'packages/lib/src/utils.ts': 'console.log("test");\n',
      'packages/lib/coverage/lcov.ts': 'console.log("test");\n',
    });
    try {
      expect(
        diagsAt(diagnostics, 'packages/app/src/index.ts').length,
      ).toBeGreaterThan(0);
      expect(
        diagsAt(diagnostics, 'packages/lib/src/utils.ts').length,
      ).toBeGreaterThan(0);
      expect(
        diagsAt(diagnostics, 'packages/app/dist/bundle.ts').length,
      ).toBeGreaterThan(0);
      expect(
        diagsAt(diagnostics, 'packages/lib/coverage/lcov.ts').length,
      ).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('per-package .gitignore — each scoped to own subtree', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'packages/app/.gitignore': 'tmp/\n',
      'packages/lib/.gitignore': 'cache/\n',
      'packages/app/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];`,
      'packages/lib/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];`,
      'packages/app/src/index.ts': 'console.log("test");\n',
      'packages/app/tmp/temp.ts': 'console.log("test");\n',
      'packages/app/cache/data.ts': 'console.log("test");\n',
      'packages/lib/src/utils.ts': 'console.log("test");\n',
      'packages/lib/cache/data.ts': 'console.log("test");\n',
      'packages/lib/tmp/temp.ts': 'console.log("test");\n',
    });
    try {
      // app: src linted, tmp blocked (app's gitignore), cache NOT blocked
      expect(
        diagsAt(diagnostics, 'packages/app/src/index.ts').length,
      ).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'packages/app/tmp').length).toBe(0);
      expect(
        diagsAt(diagnostics, 'packages/app/cache/data.ts').length,
      ).toBeGreaterThan(0);
      // lib: src linted, cache blocked (lib's gitignore), tmp NOT blocked
      expect(
        diagsAt(diagnostics, 'packages/lib/src/utils.ts').length,
      ).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'packages/lib/cache').length).toBe(0);
      expect(
        diagsAt(diagnostics, 'packages/lib/tmp/temp.ts').length,
      ).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });
});

describe('Gitignore: real-world scenarios', () => {
  test('rspack-like project with target/ and *.d.ts', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'target/\n*.d.ts\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'target/output.ts': 'console.log("test");\n',
      'tests/unit.ts': 'console.log("test");\n',
      'src/types.d.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'tests/unit.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'target').length).toBe(0);
      expect(diagsAt(diagnostics, 'src/types.d.ts').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('standard project: dist, coverage, node_modules', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'dist/\ncoverage/\nnode_modules/\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'src/utils.ts': 'console.log("test");\n',
      'dist/bundle.ts': 'console.log("test");\n',
      'coverage/lcov.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'src/utils.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'dist').length).toBe(0);
      expect(diagsAt(diagnostics, 'coverage').length).toBe(0);
    } finally {
      await cleanup();
    }
  });
});

describe('Gitignore: edge cases', () => {
  test('trailing-slash only matches directories, not similar names', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'logs/\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'logs/debug.ts': 'console.log("test");\n',
      'logs-archive/old.ts': 'console.log("test");\n',
      'src/index.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'logs').length).toBe(0);
      // logs-archive/ is a different dir, not blocked
      expect(
        diagsAt(diagnostics, 'logs-archive/old.ts').length,
      ).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('wildcard patterns: **/*.test.ts', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': '**/*.test.ts\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'src/index.test.ts': 'console.log("test");\n',
      'src/deep/nested/helper.test.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'src/index.test.ts').length).toBe(0);
      expect(
        diagsAt(diagnostics, 'src/deep/nested/helper.test.ts').length,
      ).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('.gitignore works without .git directory', async () => {
    // createTempDir does not create .git — this is the default
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'dist/\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'dist/bundle.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'dist').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('gitignore negation re-includes subdirectory', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      // Keep dist itself reachable: Git cannot re-include anything below an
      // ignored parent directory, but a later rule can reopen a child that a
      // wildcard ignored.
      '.gitignore': 'dist/*\n!dist/types/\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'dist/bundle.ts': 'console.log("test");\n',
      'dist/types/index.ts': 'console.log("test");\n',
      'src/index.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'dist/bundle.ts').length).toBe(0);
      // dist/types/ re-included by !dist/types/
      expect(
        diagsAt(diagnostics, 'dist/types/index.ts').length,
      ).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('automatic and explicit discovery both load a gitignored config', async () => {
    const tempDir = await createTempDir({
      '.gitignore': 'rslint.config.mjs\n',
      'rslint.config.mjs': `
        import { writeFileSync } from 'node:fs';
        writeFileSync(new URL('./config-loaded.marker', import.meta.url), 'loaded');
        ${CONFIG_NO_CONSOLE}
      `,
      'src/index.ts': 'console.log("test");\n',
    });
    const marker = path.join(tempDir, 'config-loaded.marker');
    try {
      const automatic = await runRslint(['--format', 'jsonline'], tempDir);
      expect(fs.existsSync(marker)).toBe(true);
      expect(automatic.stdout).toContain('no-console');

      fs.rmSync(marker);

      const explicit = await runRslint(
        [
          '--config',
          'rslint.config.mjs',
          '--format',
          'jsonline',
          'src/index.ts',
        ],
        tempDir,
      );
      const diagnostics = explicit.stdout
        .trim()
        .split('\n')
        .filter((line) => line.trim().startsWith('{'))
        .map((line) => JSON.parse(line) as Diagnostic);
      expect(fs.existsSync(marker)).toBe(true);
      expect(
        diagsAt(diagnostics, 'src/index.ts').some(
          (diagnostic) => diagnostic.ruleName === 'no-console',
        ),
      ).toBe(true);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('trailing whitespace in patterns is stripped', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      // Leading whitespace is significant in Git patterns. Exercise only
      // unescaped trailing spaces and tabs here.
      '.gitignore': 'dist/  \n\t  \n# comment\ncoverage/\t \n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'dist/bundle.ts': 'console.log("test");\n',
      'coverage/report.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'dist').length).toBe(0);
      expect(diagsAt(diagnostics, 'coverage').length).toBe(0);
    } finally {
      await cleanup();
    }
  });
});

describe('Gitignore: CLI invocation variants', () => {
  test('directory arg respects gitignore', async () => {
    const tempDir = await createTempDir({
      '.gitignore': 'dist/\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'src/dist/output.ts': 'console.log("test");\n',
    });
    try {
      const result = await runRslint(['--format', 'jsonline', 'src/'], tempDir);
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l) as Diagnostic);

      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      // src/dist/ blocked by gitignore
      expect(diagsAt(diagnostics, 'src/dist').length).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('explicit file arg in .gitignore — warns and does not lint', async () => {
    const tempDir = await createTempDir({
      '.gitignore': 'dist/\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'dist/bundle.ts': 'console.log("test");\n',
    });
    try {
      const result = await runRslint(
        ['--format', 'jsonline', 'dist/bundle.ts'],
        tempDir,
      );
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim() && !l.includes('warning'))
        .map((l) => {
          try {
            return JSON.parse(l) as Diagnostic;
          } catch {
            return null;
          }
        })
        .filter(Boolean) as Diagnostic[];

      // gitignored file: 0 diagnostics, warning emitted
      expect(diagnostics.length).toBe(0);
      expect(result.stdout + result.stderr).toContain('ignored');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--fix does not touch gitignored files', async () => {
    const tempDir = await createTempDir({
      '.gitignore': 'dist/\n',
      'rslint.config.mjs': `export default [{
        files: ["**/*.ts"], plugins: ["@typescript-eslint"],
        rules: { "@typescript-eslint/no-inferrable-types": "error" }
      }];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['**/*.ts'],
      }),
      'src/fixable.ts': 'const x: number = 42;\n',
      'dist/fixable.ts': 'const y: number = 42;\n',
    });
    try {
      await runRslint(['--fix'], tempDir);

      const fs = await import('node:fs/promises');
      const path = await import('node:path');

      // src/fixable.ts should be fixed
      const srcContent = await fs.readFile(
        path.join(tempDir, 'src/fixable.ts'),
        'utf8',
      );
      expect(srcContent).not.toContain(': number');

      // dist/fixable.ts should NOT be touched
      const distContent = await fs.readFile(
        path.join(tempDir, 'dist/fixable.ts'),
        'utf8',
      );
      expect(distContent).toContain(': number');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('Gitignore: tsconfig + gitignore interaction', () => {
  test('file in tsconfig program but in .gitignore — not linted', async () => {
    // tsconfig includes dist/, but .gitignore ignores dist/
    // files-driven: gitignore in global ignores → GetConfigForFile returns nil → 0 rules
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'dist/\n',
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['**/*.ts'],
      }),
      'rslint.config.mjs': `export default [{
        files: ["**/*.ts"],
        plugins: ["@typescript-eslint"],
        languageOptions: { parserOptions: { projectService: false, project: ["./tsconfig.json"] } },
        rules: { "@typescript-eslint/ban-ts-comment": "error" }
      }];`,
      'src/index.ts': '// @ts-ignore\nconst a = 1;\n',
      'dist/output.ts': '// @ts-ignore\nconst b = 1;\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      // In tsconfig but gitignored → not linted
      expect(diagsAt(diagnostics, 'dist/output.ts').length).toBe(0);
    } finally {
      await cleanup();
    }
  });
});

describe('Gitignore: monorepo gap file interaction', () => {
  test('root .gitignore does not prune child config discovery', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'build/\n',
      'rslint.config.mjs': `export default [{ files: ["*.ts"], rules: { "no-console": "error" } }];`,
      'packages/app/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/app/src/index.ts': 'debugger;\n',
      'packages/app/build/output.ts': 'debugger;\n',
    });
    try {
      expect(
        diagsAt(diagnostics, 'packages/app/src/index.ts').length,
      ).toBeGreaterThan(0);
      expect(
        diagsAt(diagnostics, 'packages/app/build/output.ts').length,
      ).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });
});

describe('Gitignore: overlap and edge cases', () => {
  test('gitignore and config ignores same pattern — no conflict', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': 'dist/\n',
      'rslint.config.mjs': `export default [
        { ignores: ["dist/**"] },
        { files: ["**/*.ts"], rules: { "no-console": "error" } }
      ];`,
      'src/index.ts': 'console.log("test");\n',
      'dist/bundle.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'dist').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('*.d.ts pattern with multiple dots', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      '.gitignore': '*.d.ts\n',
      'rslint.config.mjs': CONFIG_NO_CONSOLE,
      'src/index.ts': 'console.log("test");\n',
      'src/types.d.ts': 'console.log("test");\n',
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      expect(diagsAt(diagnostics, 'src/types.d.ts').length).toBe(0);
    } finally {
      await cleanup();
    }
  });
});
