import { access } from 'fs/promises';
import { join } from 'path';

export class RuleChecker {
  private rslintRulesPath = '/Users/bytedance/dev/rslint/internal/rules';

  async isRuleAlreadyPorted(ruleName: string): Promise<boolean> {
    const ruleDir = join(this.rslintRulesPath, ruleName.replace(/-/g, '_'));
    const ruleFile = join(ruleDir, `${ruleName.replace(/-/g, '_')}.go`);
    
    try {
      await access(ruleFile);
      return true;
    } catch {
      return false;
    }
  }

  async getExistingRules(): Promise<string[]> {
    try {
      const { readdir } = await import('fs/promises');
      const entries = await readdir(this.rslintRulesPath, { withFileTypes: true });
      
      const rules: string[] = [];
      for (const entry of entries) {
        if (entry.isDirectory()) {
          // Convert underscore back to hyphen for consistency
          const ruleName = entry.name.replace(/_/g, '-');
          rules.push(ruleName);
        }
      }
      
      return rules.sort();
    } catch (error) {
      console.error(`Failed to read existing rules: ${error}`);
      return [];
    }
  }

  async filterUnportedRules(ruleNames: string[]): Promise<string[]> {
    const unported: string[] = [];
    
    for (const ruleName of ruleNames) {
      const exists = await this.isRuleAlreadyPorted(ruleName);
      if (!exists) {
        unported.push(ruleName);
      }
    }
    
    return unported;
  }
}