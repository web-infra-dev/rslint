import fetch from 'node-fetch';
import { writeFile, mkdir } from 'fs/promises';
import { dirname, join } from 'path';
import { RuleInfo } from './types.js';

const GITHUB_RAW_BASE = 'https://raw.githubusercontent.com/typescript-eslint/typescript-eslint/main';
const RULES_INDEX_URL = `${GITHUB_RAW_BASE}/packages/eslint-plugin/src/rules/index.ts`;

export class RuleFetcher {
  async fetchAvailableRules(): Promise<string[]> {
    try {
      const response = await fetch(RULES_INDEX_URL);
      const content = await response.text();
      
      // Extract rule names from export statements
      const rulePattern = /['"]([a-z-]+)['"]\s*:\s*\w+,?/g;
      const rules: string[] = [];
      let match;
      
      while ((match = rulePattern.exec(content)) !== null) {
        rules.push(match[1]);
      }
      
      return rules.sort();
    } catch (error) {
      throw new Error(`Failed to fetch available rules: ${error}`);
    }
  }

  async fetchRuleFiles(ruleName: string): Promise<RuleInfo> {
    const ruleUrl = `${GITHUB_RAW_BASE}/packages/eslint-plugin/src/rules/${ruleName}.ts`;
    const testUrl = `${GITHUB_RAW_BASE}/packages/eslint-plugin/tests/rules/${ruleName}.test.ts`;
    
    return {
      name: ruleName,
      ruleUrl,
      testUrl
    };
  }

  async downloadRuleSource(ruleUrl: string): Promise<string> {
    try {
      const response = await fetch(ruleUrl);
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      return await response.text();
    } catch (error) {
      throw new Error(`Failed to download rule source: ${error}`);
    }
  }

  async downloadAndSaveTest(testUrl: string, ruleName: string): Promise<string> {
    try {
      const response = await fetch(testUrl);
      if (!response.ok) {
        console.warn(`Test file not found for ${ruleName}: ${response.status}`);
        return '';
      }
      
      const content = await response.text();
      const testPath = join(
        '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/rules',
        `${ruleName}.test.ts`
      );
      
      await mkdir(dirname(testPath), { recursive: true });
      await writeFile(testPath, content);
      
      return testPath;
    } catch (error) {
      console.warn(`Failed to download test for ${ruleName}: ${error}`);
      return '';
    }
  }
}