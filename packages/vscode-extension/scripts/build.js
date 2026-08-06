const esbuild = require('esbuild');
const fs = require('fs');
const path = require('path');

const isWatchMode = process.argv.includes('--watch');
const extensionRoot = path.resolve(__dirname, '..');
const distDirectory = path.join(extensionRoot, 'dist');
const canonicalExtensionRoot = fs.realpathSync.native(extensionRoot);
const workspacePackagesRoot = path.dirname(canonicalExtensionRoot);

function isPathWithin(root, candidate) {
  const relative = path.relative(root, candidate);
  return (
    relative === '' ||
    (relative !== '..' &&
      !relative.startsWith(`..${path.sep}`) &&
      !path.isAbsolute(relative))
  );
}

function canonicalInput(input) {
  const absolute = path.resolve(extensionRoot, input);
  try {
    return fs.realpathSync.native(absolute);
  } catch {
    return absolute;
  }
}

const config = {
  absWorkingDir: extensionRoot,
  entryPoints: [path.join(extensionRoot, 'src/main.ts')],
  outfile: path.join(distDirectory, 'main.js'),
  format: 'cjs',
  bundle: true,
  sourcemap: true,
  metafile: true,
  platform: 'node',
  external: ['vscode'],
  plugins: [
    {
      name: 'forbid-bundled-rslint-runtime',
      setup(build) {
        build.onResolve({ filter: /^@rslint\// }, (args) => ({
          errors: [
            {
              text: `The VS Code extension must resolve ${args.path} at runtime, not bundle it`,
            },
          ],
        }));
        build.onEnd((result) => {
          if (!result.metafile) return;
          const forbidden = Object.keys(result.metafile.inputs)
            .map(canonicalInput)
            .filter(
              (input) =>
                isPathWithin(workspacePackagesRoot, input) &&
                !isPathWithin(canonicalExtensionRoot, input),
            );
          if (forbidden.length === 0) return;
          return {
            errors: forbidden.map((input) => ({
              text: `The VS Code extension must not bundle workspace runtime source ${input}`,
            })),
          };
        });
      },
    },
  ],
};

async function main() {
  // A previous platform-specific build may have left native/worker payloads in
  // dist. The extension is now code-only, so always start from an empty output
  // directory and make accidental VSIX bundling impossible.
  fs.rmSync(distDirectory, { recursive: true, force: true });
  if (isWatchMode) {
    const context = await esbuild.context(config);
    await context.watch();
    return;
  }
  await esbuild.build(config);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
