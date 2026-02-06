import { RspressPlugin } from '@rspress/core';
import { execSync } from 'node:child_process';
import path from 'node:path';

export function pluginRuleManifest(): RspressPlugin {
  return {
    name: 'rule-manifest',
    beforeBuild() {
      const scriptPath = path.resolve(
        __dirname,
        '../scripts/gen-rule-manifest.js',
      );
      execSync(`node "${scriptPath}"`, { stdio: 'inherit' });
    },
  };
}
