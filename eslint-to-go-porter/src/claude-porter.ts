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
  private gitCommitTemplate: string = '';
  private registerRuleTemplate: string = '';
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
    this.gitCommitTemplate = await readFile(
      join(process.cwd(), 'prompts', 'git-commit.md'),
      'utf-8'
    );
    this.registerRuleTemplate = await readFile(
      join(process.cwd(), 'prompts', 'register-rule.md'),
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
    return this.runClaudeInDirectory(prompt, systemPrompt, '/Users/bytedance/dev/rslint');
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
      `You are an expert at converting TypeScript ESLint rules to Go. You are working in the /Users/bytedance/dev/rslint project root directory. Create the rule implementation and test files in the internal/rules subdirectory. Do NOT attempt to run, compile, or execute any Go code. The original TypeScript rule source is available at ${tsRulePath} for reference. Focus only on creating the rule and test files - rule registration will be handled separately.`
    );

    if (exitCode !== 0) {
      return {
        ruleName,
        success: false,
        error: `Claude process exited with code ${exitCode}`
      };
    }

    // Check if Claude created the required files
    const outputPath = join(
      '/Users/bytedance/dev/rslint/internal/rules',
      ruleName.replace(/-/g, '_'),
      `${ruleName.replace(/-/g, '_')}.go`
    );
    
    const testPath = join(
      '/Users/bytedance/dev/rslint/internal/rules',
      ruleName.replace(/-/g, '_'),
      `${ruleName.replace(/-/g, '_')}_test.go`
    );
    
    try {
      // Verify files were created
      await readFile(outputPath, 'utf-8');
      await readFile(testPath, 'utf-8');
      
      if (!this.showProgress) {
        console.log(`✓ Rule files created successfully`);
      }
      
      return {
        ruleName,
        success: true
      };
    } catch (error) {
      return {
        ruleName,
        success: false,
        error: `Rule files not created. Claude may have encountered an error. Check if files exist at: ${outputPath}`
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
      'You are reviewing a Go rule conversion. You are working in the /Users/bytedance/dev/rslint project root directory. Read the file, verify it is correct, and either respond with "VERIFIED" or provide the corrected code. IMPORTANT: Do NOT attempt to run, compile, or execute any Go code during verification.'
    );

    if (exitCode !== 0) {
      console.error(`Verification failed with exit code ${exitCode}`);
      return false;
    }

    // Claude will either verify the rule or fix it autonomously
    // We just need to check if the process completed successfully
    return true;
  }

  async fixTestFailures(ruleName: string, ruleSource: string, goRulePath: string, testOutput: string): Promise<boolean> {
    const prompt = this.fixTemplate
      .replace(/{{TEST_OUTPUT}}/g, testOutput)
      .replace(/{{GO_RULE_PATH}}/g, goRulePath)
      .replace(/{{RULE_SOURCE}}/g, ruleSource)
      .replace(/{{RULE_NAME}}/g, ruleName)
      .replace(/{{RULE_NAME_UNDERSCORED}}/g, ruleName.replace(/-/g, '_'))
      .replace(/{{RULE_NAME_PASCAL}}/g, this.toPascalCase(ruleName));

    const { responses, exitCode } = await this.runClaude(
      prompt,
      'You are fixing a Go rule that is failing tests. You are working in the /Users/bytedance/dev/rslint project root directory. Analyze the test output and provide the corrected Go code.'
    );

    if (exitCode !== 0) {
      console.error(`Fix attempt failed with exit code ${exitCode}`);
      return false;
    }

    // Claude will fix the rule autonomously
    // We just need to check if the process completed successfully
    if (!this.showProgress) {
      console.log(`Claude attempted to fix ${ruleName} based on test failures`);
    }
    return true;
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
      'You are adapting a TypeScript ESLint test file to work with rslint cross-validation framework. You are working from the /Users/bytedance/dev/rslint project root directory. Create the adapted test file in packages/rslint-test-tools/tests/typescript-eslint/rules/ and do NOT attempt to run or execute any code.',
      '/Users/bytedance/dev/rslint'
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

  async crossValidateRule(ruleName: string): Promise<boolean> {
    const prompt = this.crossValidateTemplate
      .replace(/{{RULE_NAME}}/g, ruleName);

    const { responses, exitCode } = await this.runClaudeInDirectory(
      prompt,
      'You are running cross-validation tests to ensure the Go rule behaves identically to the TypeScript ESLint rule. You are working from the /Users/bytedance/dev/rslint project root directory. Run the tests and report the results. Do NOT attempt to modify rule implementations.',
      '/Users/bytedance/dev/rslint'
    );

    if (exitCode !== 0) {
      console.error(`Cross-validation failed with exit code ${exitCode}`);
      return false;
    }

    try {
      const fullText = this.parser.extractTextFromResponses(responses);
      
      // Check if tests passed based on the response
      if (fullText.includes('PASS') || fullText.includes('✓') || fullText.toLowerCase().includes('all tests passed')) {
        console.log(`✓ Cross-validation tests passed for ${ruleName}`);
        return true;
      } else if (fullText.includes('FAIL') || fullText.includes('✗') || fullText.toLowerCase().includes('test failed')) {
        console.error(`✗ Cross-validation tests failed for ${ruleName}`);
        console.log('Cross-validation output:');
        console.log(fullText);
        return false;
      } else {
        console.warn(`⚠️  Cross-validation results unclear for ${ruleName}`);
        console.log('Cross-validation output:');
        console.log(fullText);
        return false;
      }
    } catch (error) {
      console.error(`Cross-validation error: ${error}`);
      return false;
    }
  }

  async gitCommitRule(ruleName: string): Promise<boolean> {
    const prompt = this.gitCommitTemplate
      .replace(/{{RULE_NAME}}/g, ruleName)
      .replace(/{{RULE_NAME_UNDERSCORED}}/g, ruleName.replace(/-/g, '_'))
      .replace(/{{RULE_NAME_PASCAL}}/g, this.toPascalCase(ruleName));

    const { responses, exitCode } = await this.runClaudeInDirectory(
      prompt,
      'You are creating a git commit for the newly ported rule. You are working in the /Users/bytedance/dev/rslint project root directory. IMPORTANT: First verify the rule is registered in internal/config/config.go with BOTH namespaced and non-namespaced versions. Then add ALL files with git add -A and create a clean commit. Do NOT push the commit.',
      '/Users/bytedance/dev/rslint'
    );

    if (exitCode !== 0) {
      console.error(`Git commit failed with exit code ${exitCode}`);
      return false;
    }

    try {
      const fullText = this.parser.extractTextFromResponses(responses);
      
      // Check if commit was successful based on the response
      if (fullText.toLowerCase().includes('commit') && 
          (fullText.includes('files changed') || fullText.includes('create mode') || fullText.includes('committed'))) {
        console.log(`✓ Successfully committed ${ruleName} rule implementation`);
        return true;
      } else {
        console.warn(`⚠️  Git commit may have failed for ${ruleName}`);
        console.log('Git output:');
        console.log(fullText);
        return false;
      }
    } catch (error) {
      console.error(`Git commit error: ${error}`);
      return false;
    }
  }
}