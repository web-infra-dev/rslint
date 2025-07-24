import { spawn, ChildProcess } from 'child_process';
import { readFile, writeFile, mkdir } from 'fs/promises';
import { join, dirname } from 'path';
import { JsonStreamParser } from './parser.js';
import { PortingResult, ClaudeResponse } from './types.js';

export class ClaudePorter {
  private parser = new JsonStreamParser();
  private promptTemplate: string = '';
  private verifyTemplate: string = '';
  private fixTemplate: string = '';
  private showProgress: boolean = false;

  async loadPromptTemplates(): Promise<void> {
    this.promptTemplate = await readFile(
      join(process.cwd(), 'prompts', 'convert-rule.md'),
      'utf-8'
    );
    this.verifyTemplate = await readFile(
      join(process.cwd(), 'prompts', 'verify-rule.md'),
      'utf-8'
    );
    this.fixTemplate = await readFile(
      join(process.cwd(), 'prompts', 'fix-test-failures.md'),
      'utf-8'
    );
  }

  setProgressMode(showProgress: boolean): void {
    this.showProgress = showProgress;
  }

  private preparePrompt(ruleName: string, ruleSource: string): string {
    return this.promptTemplate
      .replace(/{{RULE_SOURCE}}/g, ruleSource)
      .replace(/{{RULE_NAME_KEBAB}}/g, ruleName)
      .replace(/{{RULE_NAME_UNDERSCORED}}/g, ruleName.replace(/-/g, '_'))
      .replace(/{{RULE_NAME_PASCAL}}/g, this.toPascalCase(ruleName));
  }

  private toPascalCase(str: string): string {
    return str
      .split('-')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1))
      .join('');
  }

  private runClaude(prompt: string, systemPrompt: string): Promise<{responses: ClaudeResponse[], exitCode: number}> {
    return new Promise((resolve) => {
      const args = [
        '-p',
        '--verbose',
        '--output-format', 'stream-json',
        '--max-turns', '1',
        '--model', 'claude-sonnet-4-20250514',
        '--cwd', '/Users/bytedance/dev/rslint',
        '--dangerously-skip-permissions',
        '--system-prompt', systemPrompt
      ];

      const claudeProcess: ChildProcess = spawn('claude', args, {
        stdio: ['pipe', 'pipe', 'pipe'],
        env: {
          ...process.env,
          NODE_OPTIONS: undefined
        }
      });

      let allResponses: ClaudeResponse[] = [];
      let errorOccurred = false;
      let exitCode = 0;

      claudeProcess.stdout?.on('data', (chunk: Buffer) => {
        const responses = this.parser.processChunk(chunk.toString());
        allResponses.push(...responses);
        
        if (this.showProgress) {
          for (const response of responses) {
            console.log(JSON.stringify(response));
          }
        }
        
        for (const response of responses) {
          if (this.parser.isErrorResponse(response)) {
            errorOccurred = true;
            claudeProcess.kill();
            return;
          }
        }
      });

      claudeProcess.stderr?.on('data', (chunk: Buffer) => {
        if (!this.showProgress) {
          console.error(`Claude stderr: ${chunk.toString()}`);
        }
      });

      claudeProcess.on('close', (code) => {
        exitCode = code || 0;
        resolve({ responses: allResponses, exitCode });
      });

      // Send the prompt
      claudeProcess.stdin?.write(prompt);
      claudeProcess.stdin?.end();
    });
  }

  async portRule(ruleName: string, ruleSource: string): Promise<PortingResult> {
    if (!this.promptTemplate) {
      await this.loadPromptTemplates();
    }

    const prompt = this.preparePrompt(ruleName, ruleSource);
    const { responses, exitCode } = await this.runClaude(
      prompt,
      'You are an expert at converting TypeScript ESLint rules to Go. Follow the instructions exactly and provide only the requested Go code.'
    );

    if (exitCode !== 0) {
      return {
        ruleName,
        success: false,
        error: `Claude process exited with code ${exitCode}`
      };
    }

    try {
      const fullText = this.parser.extractTextFromResponses(responses);
      const goCode = this.parser.extractGoCode(fullText);
      
      // Save the Go code
      const outputPath = join(
        '/Users/bytedance/dev/rslint/internal/rules',
        ruleName.replace(/-/g, '_'),
        `${ruleName.replace(/-/g, '_')}.go`
      );
      
      await mkdir(dirname(outputPath), { recursive: true });
      await writeFile(outputPath, goCode);

      return {
        ruleName,
        success: true,
        goCode
      };
    } catch (error) {
      return {
        ruleName,
        success: false,
        error: `Failed to process response: ${error}`
      };
    }
  }

  async verifyRule(ruleName: string, ruleSource: string, goRulePath: string): Promise<boolean> {
    const prompt = this.verifyTemplate
      .replace(/{{RULE_SOURCE}}/g, ruleSource)
      .replace(/{{GO_RULE_PATH}}/g, goRulePath);

    const { responses, exitCode } = await this.runClaude(
      prompt,
      'You are reviewing a Go rule conversion. Read the file, verify it is correct, and either respond with "VERIFIED" or provide the corrected code.'
    );

    if (exitCode !== 0) {
      console.error(`Verification failed with exit code ${exitCode}`);
      return false;
    }

    try {
      const fullText = this.parser.extractTextFromResponses(responses);
      
      if (fullText.includes('VERIFIED')) {
        return true;
      }
      
      // Try to extract corrected code
      const goCode = this.parser.extractGoCode(fullText);
      await writeFile(goRulePath, goCode);
      console.log(`Rule ${ruleName} was corrected during verification`);
      return true;
    } catch (error) {
      console.error(`Verification error: ${error}`);
      return false;
    }
  }

  async fixTestFailures(ruleName: string, ruleSource: string, goRulePath: string, testOutput: string): Promise<boolean> {
    const prompt = this.fixTemplate
      .replace(/{{TEST_OUTPUT}}/g, testOutput)
      .replace(/{{GO_RULE_PATH}}/g, goRulePath)
      .replace(/{{RULE_SOURCE}}/g, ruleSource);

    const { responses, exitCode } = await this.runClaude(
      prompt,
      'You are fixing a Go rule that is failing tests. Analyze the test output and provide the corrected Go code.'
    );

    if (exitCode !== 0) {
      console.error(`Fix attempt failed with exit code ${exitCode}`);
      return false;
    }

    try {
      const fullText = this.parser.extractTextFromResponses(responses);
      const goCode = this.parser.extractGoCode(fullText);
      await writeFile(goRulePath, goCode);
      console.log(`Rule ${ruleName} was fixed based on test failures`);
      return true;
    } catch (error) {
      console.error(`Fix error: ${error}`);
      return false;
    }
  }
}