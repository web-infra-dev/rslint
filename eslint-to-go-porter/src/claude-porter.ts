import { spawn, ChildProcess } from 'child_process';
import { readFile, writeFile, mkdir } from 'fs/promises';
import { join, dirname } from 'path';
import chalk from 'chalk';
import { JsonStreamParser } from './parser.js';
import { PortingResult, ClaudeResponse } from './types.js';

export class ClaudePorter {
  private parser = new JsonStreamParser();
  private promptTemplate: string = '';
  private verifyTemplate: string = '';
  private fixTemplate: string = '';
  private adaptTestTemplate: string = '';
  private crossValidateTemplate: string = '';
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
    this.adaptTestTemplate = await readFile(
      join(process.cwd(), 'prompts', 'adapt-test.md'),
      'utf-8'
    );
    this.crossValidateTemplate = await readFile(
      join(process.cwd(), 'prompts', 'cross-validate.md'),
      'utf-8'
    );
  }

  setProgressMode(showProgress: boolean): void {
    this.showProgress = showProgress;
  }

  private preparePrompt(ruleName: string, ruleSource: string, testSource: string = ''): string {
    return this.promptTemplate
      .replace(/{{RULE_SOURCE}}/g, ruleSource)
      .replace(/{{TEST_SOURCE}}/g, testSource)
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

  private runClaudeInDirectory(prompt: string, systemPrompt: string, workingDir: string): Promise<{responses: ClaudeResponse[], exitCode: number}> {
    return new Promise((resolve) => {
      const args = [
        '-p',
        '--verbose',
        '--output-format', 'stream-json',
        '--max-turns', '500',
        '--model', 'claude-sonnet-4-20250514',
        '--dangerously-skip-permissions',
        '--system-prompt', systemPrompt
      ];

      const claudeProcess: ChildProcess = spawn('claude', args, {
        stdio: ['pipe', 'pipe', 'pipe'],
        cwd: workingDir,
        env: {
          ...process.env,
          NODE_OPTIONS: undefined,
          // Ensure Claude doesn't access parent directories
          CLAUDE_WORKING_DIR: workingDir
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
            this.displayProgress(response);
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
        const msg = chunk.toString().trim();
        if (msg) {
          console.error(chalk.red(`Claude stderr: ${msg}`));
        }
      });

      claudeProcess.on('close', (code) => {
        exitCode = code || 0;
        resolve({ responses: allResponses, exitCode });
      });

      claudeProcess.stdin?.write(prompt);
      claudeProcess.stdin?.end();
    });
  }

  private runClaude(prompt: string, systemPrompt: string): Promise<{responses: ClaudeResponse[], exitCode: number}> {
    return this.runClaudeInDirectory(prompt, systemPrompt, '/Users/bytedance/dev/rslint/internal/rules');
  }

  async portRule(ruleName: string, ruleSource: string, testSource: string = ''): Promise<PortingResult> {
    if (!this.promptTemplate) {
      await this.loadPromptTemplates();
    }

    // Save TypeScript rule source to temp file so Claude can reference it
    const tsRulePath = join(
      '/Users/bytedance/dev/rslint/internal/rules',
      ruleName.replace(/-/g, '_'),
      `${ruleName.replace(/-/g, '_')}.ts`
    );
    
    try {
      // Ensure directory exists
      await mkdir(dirname(tsRulePath), { recursive: true });
      // Write TypeScript source file
      await writeFile(tsRulePath, ruleSource);
    } catch (error) {
      console.warn(`Warning: Could not save TypeScript source for ${ruleName}: ${error}`);
    }

    const prompt = this.preparePrompt(ruleName, ruleSource, testSource);
    const { responses, exitCode } = await this.runClaude(
      prompt,
      `You are an expert at converting TypeScript ESLint rules to Go. You are working in the /Users/bytedance/dev/rslint/internal/rules directory. IMPORTANT: Only access files within this rules directory and its subdirectories. Do NOT navigate to parent directories or other parts of the filesystem. Do NOT attempt to run, compile, or execute any Go code. The original TypeScript rule source is available at ${tsRulePath} for reference. Follow the instructions exactly and provide only the requested Go code files.`
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
      
      // Check if Claude created files using Write tool
      const outputPath = join(
        '/Users/bytedance/dev/rslint/internal/rules',
        ruleName.replace(/-/g, '_'),
        `${ruleName.replace(/-/g, '_')}.go`
      );
      
      let goCode: string;
      try {
        // Try to extract code from response text first
        goCode = this.parser.extractGoCode(fullText);
      } catch (extractError) {
        // If extraction fails, check if Claude created the file directly
        try {
          goCode = await readFile(outputPath, 'utf-8');
          console.log(`✓ Rule file created by Claude: ${outputPath}`);
        } catch (fileError) {
          throw new Error(`No Go code found in response and file not created: ${extractError}`);
        }
      }

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
    } finally {
      // Clean up temporary TypeScript file
      try {
        const { unlink } = await import('fs/promises');
        await unlink(tsRulePath);
      } catch (cleanupError) {
        // Ignore cleanup errors
      }
    }
  }

  async verifyRule(ruleName: string, ruleSource: string, goRulePath: string): Promise<boolean> {
    const prompt = this.verifyTemplate
      .replace(/{{RULE_SOURCE}}/g, ruleSource)
      .replace(/{{GO_RULE_PATH}}/g, goRulePath);

    const { responses, exitCode } = await this.runClaude(
      prompt,
      'You are reviewing a Go rule conversion. Read the file, verify it is correct, and either respond with "VERIFIED" or provide the corrected code. IMPORTANT: Do NOT attempt to run, compile, or execute any Go code during verification.'
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

  private displayProgress(response: ClaudeResponse): void {
    // Display initialization
    if (response.type === 'system' && response.subtype === 'init') {
      console.log(chalk.gray(`\n🚀 Claude initialized with model: ${response.model}`));
      console.log(chalk.gray(`📁 Working directory: ${response.cwd}\n`));
      return;
    }

    // Display assistant messages
    if (response.type === 'assistant' && response.message?.content) {
      for (const content of response.message.content) {
        if (content.type === 'text' && content.text) {
          // Display text content with proper formatting
          const lines = content.text.split('\n');
          for (const line of lines) {
            if (line.trim()) {
              console.log(chalk.blue('Claude: ') + line);
            }
          }
        } else if (content.type === 'tool_use') {
          // Display tool usage
          console.log(chalk.yellow(`\n🔧 Using tool: ${content.name}`) + chalk.gray(` (${content.id})`));
          if (content.input && this.showProgress) {
            const inputStr = JSON.stringify(content.input, null, 2);
            const lines = inputStr.split('\n');
            for (const line of lines) {
              console.log(chalk.gray('   ' + line));
            }
          }
        }
      }
    }

    // Display tool results
    if (response.type === 'user' && response.message?.content) {
      for (const content of response.message.content) {
        if (content.type === 'tool_result') {
          const resultContent = content.content || '';
          const resultStr = typeof resultContent === 'string' ? resultContent : JSON.stringify(resultContent);
          const lines = resultStr.split('\n');
          const maxLines = 10;
          const displayLines = lines.slice(0, maxLines);
          
          console.log(chalk.green(`✓ Tool result${lines.length > maxLines ? ` (showing first ${maxLines} lines)` : ''}:`));
          for (const line of displayLines) {
            console.log(chalk.gray('   ' + line));
          }
          if (lines.length > maxLines) {
            console.log(chalk.gray(`   ... (${lines.length - maxLines} more lines)`));
          }
        }
      }
    }

    // Display final result
    if (response.type === 'result') {
      if (response.subtype === 'success') {
        console.log(chalk.green(`\n✅ Claude completed successfully`));
      } else if (response.subtype === 'error_max_turns') {
        console.log(chalk.yellow(`\n⚠️  Claude reached max turns (${response.num_turns} turns)`));
      } else if (response.subtype === 'error_during_execution') {
        console.log(chalk.red(`\n❌ Claude execution error: Tool execution failed or encountered an error`));
        if (response.result?.message) {
          console.log(chalk.red(`   Details: ${response.result.message}`));
        }
      } else if (response.subtype === 'error_permission_denied') {
        console.log(chalk.red(`\n❌ Claude permission denied: Check file permissions and access rights`));
      } else if (response.subtype?.includes('error')) {
        console.log(chalk.red(`\n❌ Claude error (${response.subtype}): An unexpected error occurred`));
        if (response.result?.message) {
          console.log(chalk.red(`   Details: ${response.result.message}`));
        }
      }
      
      if (response.usage) {
        console.log(chalk.gray(`📊 Tokens used: ${response.usage.input_tokens} in, ${response.usage.output_tokens} out`));
      }
    }
  }

  async adaptTestFile(ruleName: string, testSource: string): Promise<boolean> {
    const prompt = this.adaptTestTemplate
      .replace(/{{TEST_SOURCE}}/g, testSource)
      .replace(/{{RULE_NAME}}/g, ruleName);

    const { responses, exitCode } = await this.runClaudeInDirectory(
      prompt,
      'You are adapting a TypeScript ESLint test file to work with rslint cross-validation framework. You are working in the /Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/rules directory. Only access files within this test directory. Create the adapted test file and do NOT attempt to run or execute any code.',
      '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/rules'
    );

    if (exitCode !== 0) {
      console.error(`Test adaptation failed with exit code ${exitCode}`);
      return false;
    }

    try {
      // Check if Claude created the adapted test file
      const adaptedTestPath = `/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/rules/${ruleName}.test.ts`;
      try {
        await readFile(adaptedTestPath, 'utf-8');
        console.log(`✓ Adapted test file created: ${adaptedTestPath}`);
        return true;
      } catch (fileError) {
        console.error(`Adapted test file not created: ${fileError}`);
        return false;
      }
    } catch (error) {
      console.error(`Test adaptation error: ${error}`);
      return false;
    }
  }
}