import type { RsbuildPlugin, Rspack } from 'rstack/lib';
import { fileURLToPath } from 'node:url';
import ts from 'typescript';

const project = fileURLToPath(
  new URL('../compiler-syntax/tsconfig.json', import.meta.url),
);

// Upstream syntax sources have their own composite project and compiler
// settings. Prepare only its declarations before core's strict dts project.
export function compilerSyntaxDeclarationsPlugin(): RsbuildPlugin {
  return {
    name: 'rslint:compiler-syntax-declarations',
    setup(api) {
      let inputs: readonly string[] = [];
      api.modifyRspackConfig((config) => {
        config.plugins ??= [];
        config.plugins.push({
          apply(compiler: Rspack.Compiler) {
            compiler.hooks.afterCompile.tap(
              'rslint:compiler-syntax-declarations',
              (compilation) => {
                for (const file of inputs)
                  compilation.fileDependencies.add(file);
              },
            );
          },
        });
      });
      api.onBeforeBuild({
        order: 'pre',
        handler() {
          const diagnostics: ts.Diagnostic[] = [];
          const host = ts.createSolutionBuilderHost(
            ts.sys,
            undefined,
            (diagnostic) => {
              diagnostics.push(diagnostic);
            },
          );
          const configurations: string[] = [];
          const parsed = ts.getParsedCommandLineOfConfigFile(
            project,
            {},
            {
              ...ts.sys,
              readFile(file) {
                configurations.push(file);
                return ts.sys.readFile(file);
              },
              onUnRecoverableConfigFileDiagnostic(diagnostic) {
                diagnostics.push(diagnostic);
              },
            },
          );
          inputs = [
            ...configurations,
            ...(parsed?.fileNames ?? []),
            fileURLToPath(
              new URL(
                '../../../typescript-go/packages/typescript/package.json',
                import.meta.url,
              ),
            ),
          ];
          if (
            ts.createSolutionBuilder(host, [project], {}).build() !==
            ts.ExitStatus.Success
          ) {
            throw new Error(
              ts.formatDiagnostics(diagnostics, {
                getCanonicalFileName: (file) => file,
                getCurrentDirectory: ts.sys.getCurrentDirectory,
                getNewLine: () => ts.sys.newLine,
              }),
            );
          }
        },
      });
    },
  };
}
