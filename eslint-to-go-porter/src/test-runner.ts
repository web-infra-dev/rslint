import { spawn } from 'child_process';
import { join } from 'path';

export interface TestResult {
  success: boolean;
  output: string;
  error?: string;
}

export class TestRunner {
  async runRuleTest(ruleName: string): Promise<TestResult> {
    return new Promise((resolve) => {
      const testPath = join(
        '/Users/bytedance/dev/rslint/internal/rules',
        ruleName.replace(/-/g, '_')
      );

      const testProcess = spawn('go', ['test', '-v', './...'], {
        cwd: testPath,
        stdio: ['ignore', 'pipe', 'pipe']
      });

      let stdout = '';
      let stderr = '';

      testProcess.stdout.on('data', (chunk: Buffer) => {
        stdout += chunk.toString();
      });

      testProcess.stderr.on('data', (chunk: Buffer) => {
        stderr += chunk.toString();
      });

      testProcess.on('close', (code) => {
        const output = stdout + '\n' + stderr;
        
        if (code === 0) {
          resolve({
            success: true,
            output
          });
        } else {
          resolve({
            success: false,
            output,
            error: `Test exited with code ${code}`
          });
        }
      });

      testProcess.on('error', (err) => {
        resolve({
          success: false,
          output: '',
          error: `Failed to run test: ${err.message}`
        });
      });
    });
  }
}