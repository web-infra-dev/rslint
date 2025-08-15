import { RspressPlugin } from '@rspress/core';
const RULE_MANIFEST_URL =
  'https://raw.githubusercontent.com/web-infra-dev/rslint/main/packages/rslint-test-tools/rule-manifest.json';
export function pluginPreloadRule(): RspressPlugin {
  return {
    name: 'preload-rules-data',
    async extendPageData(pageData) {
      try {
        const result = await fetch(RULE_MANIFEST_URL);
        (pageData as any).ruleManifest = await result.json();
      } catch (error) {
        // don't stop compile if fetch failed
        console.error('Failed to fetch rule manifest:', error);
      }
    },
  };
}
