/** Test-only source adapter for the plugin lint runtime. */
import {
  lintFile as runFrame,
  type LintFileRequest,
} from '../../../rslint/src/eslint-plugin/linter/ecma-language-plugin.js';
import { projectFrame } from '../languages/js/project-frame.js';
import type { LoadedPlugins } from '../../../rslint/src/eslint-plugin/plugin/plugin-loader.js';

type TestRequest = Omit<LintFileRequest, 'frame'> & { text: string };
export function lintFile(req: TestRequest, plugins: LoadedPlugins) {
  const { text, ...metadata } = req;
  return runFrame(
    {
      ...metadata,
      frame: projectFrame(
        req.filePath,
        text,
        req.languageOptions?.sourceType ?? 'module',
      ),
    },
    plugins,
  );
}
