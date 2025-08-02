#!/usr/bin/env node

const { spawn } = require('child_process');
const {
  readdir,
  readFile,
  writeFile,
  mkdir,
  unlink,
  access,
  rm,
  chmod,
} = require('fs/promises');
const { join, basename, dirname } = require('path');
const https = require('https');
const { randomBytes } = require('crypto');
const os = require('os');

// __dirname is available in CommonJS

// Configuration
const BUILD_COMMAND = 'pnpm';
const BUILD_ARGS = ['-r', 'build'];
const TEST_TIMEOUT = 120000; // 120 seconds (2 minutes)
const TEST_DIR = 'packages/rslint-test-tools/tests/typescript-eslint/rules';
const TSLINT_BASE_URL =
  'https://raw.githubusercontent.com/typescript-eslint/typescript-eslint/main';
const MAX_FIX_ATTEMPTS = 500; // Maximum attempts to fix a single test

// Concurrent execution configuration
const WORK_QUEUE_DIR = join(os.tmpdir(), 'rslint-automation');
const WORKER_ID = process.env.RSLINT_WORKER_ID || null;
const IS_WORKER = !!WORKER_ID;
const DEFAULT_WORKERS = 4;

// Progress tracking
let totalTests = 0;
let completedTests = 0;
let failedTests = 0;

// Work queue management
class WorkQueue {
  constructor(workDir) {
    this.workDir = workDir;
    this.lockDir = join(workDir, '.locks');
  }

  async initialize() {
    await mkdir(this.workDir, { recursive: true });
    await mkdir(this.lockDir, { recursive: true });
  }

  async addWork(items) {
    for (let i = 0; i < items.length; i++) {
      const workFile = join(this.workDir, `work_${i}.json`);
      await writeFile(
        workFile,
        JSON.stringify({
          id: i,
          test: items[i],
          status: 'pending',
          createdAt: Date.now(),
        }),
      );
    }
  }

  async claimWork(workerId) {
    const files = await readdir(this.workDir);
    const workFiles = files.filter(
      f => f.startsWith('work_') && f.endsWith('.json'),
    );

    for (const file of workFiles) {
      const lockFile = join(this.lockDir, `${file}.lock`);

      try {
        // Try to create lock file atomically
        await writeFile(lockFile, workerId, { flag: 'wx' });

        // Successfully got lock, read work item
        const workPath = join(this.workDir, file);
        const work = JSON.parse(await readFile(workPath, 'utf8'));

        if (work.status === 'pending') {
          // Update status
          work.status = 'claimed';
          work.workerId = workerId;
          work.claimedAt = Date.now();
          await writeFile(workPath, JSON.stringify(work, null, 2));

          return work;
        } else {
          // Already claimed by someone else, remove our lock
          await unlink(lockFile);
        }
      } catch (err) {
        if (err.code !== 'EEXIST') {
          log(`Error claiming work: ${err.message}`, 'error');
        }
        // Lock already exists or other error, try next file
      }
    }

    return null; // No work available
  }

  async completeWork(workId, success) {
    const workFile = join(this.workDir, `work_${workId}.json`);
    const lockFile = join(this.lockDir, `work_${workId}.json.lock`);

    const work = JSON.parse(await readFile(workFile, 'utf8'));
    work.status = success ? 'completed' : 'failed';
    work.completedAt = Date.now();
    await writeFile(workFile, JSON.stringify(work, null, 2));

    // Remove lock
    try {
      await unlink(lockFile);
    } catch (err) {
      // Lock might already be gone
    }
  }

  async getProgress() {
    const files = await readdir(this.workDir);
    const workFiles = files.filter(
      f => f.startsWith('work_') && f.endsWith('.json'),
    );

    let pending = 0,
      claimed = 0,
      completed = 0,
      failed = 0;

    for (const file of workFiles) {
      const work = JSON.parse(await readFile(join(this.workDir, file), 'utf8'));
      switch (work.status) {
        case 'pending':
          pending++;
          break;
        case 'claimed':
          claimed++;
          break;
        case 'completed':
          completed++;
          break;
        case 'failed':
          failed++;
          break;
      }
    }

    return { pending, claimed, completed, failed, total: workFiles.length };
  }

  async cleanup() {
    try {
      await rm(this.workDir, { recursive: true, force: true });
    } catch (err) {
      // Ignore cleanup errors
    }
  }
}

// Colors for terminal output
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  dim: '\x1b[2m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[37m',
};

function formatTime() {
  return new Date().toLocaleTimeString();
}

function log(message, type = 'info') {
  const timestamp = `[${formatTime()}]`;
  let prefix = '';
  let color = colors.white;

  switch (type) {
    case 'success':
      prefix = '✓';
      color = colors.green;
      break;
    case 'error':
      prefix = '✗';
      color = colors.red;
      break;
    case 'warning':
      prefix = '⚠';
      color = colors.yellow;
      break;
    case 'info':
      prefix = '→';
      color = colors.cyan;
      break;
    case 'claude':
      prefix = '🤖';
      color = colors.magenta;
      break;
    case 'progress':
      prefix = '◆';
      color = colors.blue;
      break;
  }

  console.log(
    `${colors.dim}${timestamp}${colors.reset} ${color}${prefix} ${message}${colors.reset}`,
  );
}

function logProgress(message, data = {}) {
  // Special handling for Claude output
  if (data.phase && data.phase.startsWith('claude-')) {
    if (data.phase === 'claude-text' || data.phase === 'claude-text-final') {
      log(`Claude: ${data.content}`, 'claude');
      return;
    } else if (
      data.phase === 'claude-code' ||
      data.phase === 'claude-code-final'
    ) {
      console.log(`${colors.dim}--- Claude Code Block ---${colors.reset}`);
      console.log(data.content);
      console.log(`${colors.dim}--- End Code Block ---${colors.reset}`);
      return;
    } else if (data.phase === 'claude-command') {
      log(`Claude executing: ${data.command}`, 'claude');
      return;
    } else if (data.phase === 'claude-error') {
      log(`Claude error: ${data.error}`, 'error');
      return;
    }
  }

  // Regular progress messages
  if (data.phase === 'go-test-start') {
    console.log('');
    log(
      `Testing Go package ${data.packageName} (attempt ${data.attempt}/${data.maxAttempts})`,
      'progress',
    );
  } else if (data.phase === 'go-test-pass') {
    log(
      `✓ Go package ${data.packageName} passed in ${data.durationMs}ms`,
      'success',
    );
  } else if (data.phase === 'go-test-fail') {
    log(
      `✗ Go package ${data.packageName} failed with exit code ${data.exitCode}`,
      'error',
    );
  } else if (data.phase === 'test-start') {
    console.log('');
    log(
      `Testing ${data.testName} (attempt ${data.attempt}/${data.maxAttempts})`,
      'progress',
    );
  } else if (data.phase === 'test-pass') {
    log(`✓ ${data.testName} passed in ${data.durationMs}ms`, 'success');
  } else if (data.phase === 'test-fail') {
    if (data.exitCode) {
      log(`✗ ${data.testName} failed with exit code ${data.exitCode}`, 'error');
    } else {
      let failureDetails = [];
      if (data.goFailed) failureDetails.push('Go test');
      if (data.tsFailed) failureDetails.push('TypeScript test');
      log(
        `✗ ${data.testName} failed (${failureDetails.join(' and ')})`,
        'error',
      );
    }
  } else if (data.phase === 'script-complete') {
    console.log('\n' + '='.repeat(60));
    log('Automation Complete', 'info');
    log(`Total Duration: ${data.totalDurationMinutes} minutes`, 'info');
    log(
      `Tests: ${data.testResults.passed}/${data.testResults.total} passed (${data.testResults.successRate}%)`,
      data.testResults.failed > 0 ? 'warning' : 'success',
    );
    console.log('='.repeat(60));
  } else {
    log(message, 'info');
  }
}

async function fetchFromGitHub(url) {
  return new Promise((resolve, reject) => {
    https
      .get(url, res => {
        let data = '';
        res.on('data', chunk => {
          data += chunk;
        });
        res.on('end', () => {
          if (res.statusCode === 200) {
            resolve(data);
          } else {
            resolve(null); // Return null if not found
          }
        });
      })
      .on('error', err => {
        log(`GitHub fetch error: ${err.message}`, 'error');
        resolve(null);
      });
  });
}

async function fetchOriginalRule(ruleName) {
  // Convert test filename to rule name (e.g., no-array-delete.test.ts -> no-array-delete)
  const cleanRuleName = ruleName.replace('.test.ts', '');

  // Try to fetch the rule implementation
  const ruleUrl = `${TSLINT_BASE_URL}/packages/eslint-plugin/src/rules/${cleanRuleName}.ts`;
  const ruleContent = await fetchFromGitHub(ruleUrl);

  // Try to fetch the test file
  const testUrl = `${TSLINT_BASE_URL}/packages/eslint-plugin/tests/rules/${cleanRuleName}.test.ts`;
  const testContent = await fetchFromGitHub(testUrl);

  return {
    ruleName: cleanRuleName,
    ruleContent,
    testContent,
  };
}

async function runCommand(command, args, options = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      stdio: 'pipe',
      cwd: __dirname,
      ...options,
    });

    let stdout = '';
    let stderr = '';

    child.stdout?.on('data', data => {
      stdout += data.toString();
    });

    child.stderr?.on('data', data => {
      stderr += data.toString();
    });

    const timeout = options.timeout
      ? setTimeout(() => {
          child.kill('SIGKILL');
          reject(new Error(`Command timed out after ${options.timeout}ms`));
        }, options.timeout)
      : null;

    child.on('close', code => {
      if (timeout) clearTimeout(timeout);
      resolve({ code, stdout, stderr });
    });

    child.on('error', error => {
      if (timeout) clearTimeout(timeout);
      reject(error);
    });
  });
}

async function runClaudeWithStreaming(prompt) {
  return new Promise(resolve => {
    // Use same flags as porter: -p, --verbose, --output-format stream-json
    const settingsFile = join(__dirname, '.claude', 'settings.local.json');
    const args = [
      '-p',
      '--verbose',
      '--output-format',
      'stream-json',
      '--model',
      'claude-sonnet-4-20250514',
      '--max-turns',
      '500',
      // '--settings', settingsFile,
      '--dangerously-skip-permissions',
    ];

    const child = spawn('claude', args, {
      stdio: ['pipe', 'pipe', 'pipe'],
      env: { ...process.env },
    });

    let fullOutput = '';
    let fullError = '';
    let jsonBuffer = '';

    // Process stdout stream for JSON
    child.stdout.on('data', data => {
      const chunk = data.toString();
      fullOutput += chunk;
      jsonBuffer += chunk;

      // Process complete lines (JSON objects are line-delimited)
      const lines = jsonBuffer.split('\n');
      jsonBuffer = lines.pop() || ''; // Keep incomplete line for next chunk

      for (const line of lines) {
        if (!line.trim()) continue;

        try {
          const json = JSON.parse(line);

          // Handle different types of streaming events based on porter's displayProgress
          if (json.type === 'system' && json.subtype === 'init') {
            log(`Claude initialized with model: ${json.model}`, 'claude');
            log(`Working directory: ${json.cwd}`, 'info');
          } else if (json.type === 'assistant' && json.message?.content) {
            // Display assistant messages
            for (const content of json.message.content) {
              if (content.type === 'text' && content.text) {
                // Display text content line by line
                const lines = content.text.split('\n');
                for (const line of lines) {
                  if (line.trim()) {
                    process.stdout.write(
                      `${colors.magenta}🤖 Claude: ${line}${colors.reset}\n`,
                    );
                  }
                }
              } else if (content.type === 'tool_use') {
                // Display tool usage
                log(`Using tool: ${content.name}`, 'claude');
                if (content.input) {
                  const inputStr = JSON.stringify(content.input, null, 2);
                  const lines = inputStr.split('\n');
                  for (const line of lines) {
                    console.log(colors.dim + '   ' + line + colors.reset);
                  }
                }
              }
            }
          } else if (json.type === 'user' && json.message?.content) {
            // Display tool results
            for (const content of json.message.content) {
              if (content.type === 'tool_result') {
                const resultContent = content.content || '';
                const resultStr =
                  typeof resultContent === 'string'
                    ? resultContent
                    : JSON.stringify(resultContent);
                const lines = resultStr.split('\n');
                const maxLines = 10;
                const displayLines = lines.slice(0, maxLines);

                log(
                  `Tool result${lines.length > maxLines ? ` (showing first ${maxLines} lines)` : ''}:`,
                  'success',
                );
                for (const line of displayLines) {
                  console.log(colors.dim + '   ' + line + colors.reset);
                }
                if (lines.length > maxLines) {
                  console.log(
                    colors.dim +
                      `   ... (${lines.length - maxLines} more lines)` +
                      colors.reset,
                  );
                }
              }
            }
          } else if (json.type === 'result') {
            // Display final result
            if (json.subtype === 'success') {
              log('Claude completed successfully', 'success');
            } else if (json.subtype === 'error_max_turns') {
              log(
                `Claude reached max turns (${json.num_turns} turns)`,
                'warning',
              );
            } else if (json.subtype?.includes('error')) {
              log(`Claude error (${json.subtype})`, 'error');
              if (json.result?.message) {
                log(`Details: ${json.result.message}`, 'error');
              }
            }

            if (json.usage) {
              log(
                `Tokens used: ${json.usage.input_tokens} in, ${json.usage.output_tokens} out`,
                'info',
              );
            }
          }
        } catch (e) {
          // Not valid JSON, might be partial data
          if (line.length > 0 && !line.startsWith('{')) {
            // Sometimes non-JSON output comes through
            log(`Claude output: ${line}`, 'info');
          }
        }
      }
    });

    child.stderr.on('data', data => {
      const error = data.toString();
      fullError += error;
      if (error.trim()) {
        log(`Claude CLI error: ${error.trim()}`, 'error');
      }
    });

    child.on('close', code => {
      // Process any remaining JSON buffer
      if (jsonBuffer.trim()) {
        try {
          const json = JSON.parse(jsonBuffer);
          if (json.delta?.text) {
            process.stdout.write(
              `${colors.magenta}${json.delta.text}${colors.reset}\n`,
            );
          }
        } catch (e) {
          // Not JSON, just log it
          if (jsonBuffer.trim()) {
            log(`Remaining output: ${jsonBuffer}`, 'info');
          }
        }
      }

      resolve({
        code,
        stdout: fullOutput,
        stderr: fullError,
      });
    });

    // Set timeout for 5 minutes (increased for complex fixes)
    const timeout = setTimeout(() => {
      child.kill('SIGKILL');
      log('Claude CLI timeout after 5 minutes', 'error');
      resolve({
        code: -1,
        stdout: fullOutput,
        stderr: 'Process timed out after 5 minutes',
      });
    }, 900000); // 5 minutes

    // Clear timeout on close
    child.on('close', () => clearTimeout(timeout));

    // Write prompt to stdin instead of passing as argument
    child.stdin.write(prompt);
    child.stdin.end();
  });
}

async function fixRuleTestsWithClaude(
  ruleName,
  goTestResult,
  tsTestResult,
  originalSources = null,
  currentTestContent = null,
) {
  let prompt = `Go into plan mode first to analyze these test failures and plan the fix, performing deep research into the problem and paying attention to the TypeScript versions as a reference but also carefully considering the Go environment.\n\n`;

  prompt += `Rule: ${ruleName}\n\n`;

  // Add Go test results if available
  if (goTestResult && !goTestResult.success) {
    prompt += `--- GO TEST FAILURE ---\n`;
    prompt += `Command: ${goTestResult.command}\n`;
    if (goTestResult.output) {
      prompt += `Stdout:\n${goTestResult.output}\n`;
    }
    if (goTestResult.error) {
      prompt += `Stderr:\n${goTestResult.error}\n`;
    }
    prompt += `\n`;
  }

  // Add TypeScript test results if available
  if (tsTestResult && !tsTestResult.success) {
    prompt += `--- TYPESCRIPT TEST FAILURE ---\n`;
    prompt += `Command: ${tsTestResult.command}\n`;
    if (tsTestResult.output) {
      prompt += `Stdout:\n${tsTestResult.output}\n`;
    }
    if (tsTestResult.error) {
      prompt += `Stderr:\n${tsTestResult.error}\n`;
    }
    prompt += `\n`;
  }

  if (currentTestContent) {
    prompt += `\n--- CURRENT RSLINT TEST FILE ---\n`;
    prompt += `\`\`\`typescript\n${currentTestContent}\n\`\`\`\n`;
    prompt += `\n--- END CURRENT TEST ---\n\n`;
  }

  if (originalSources) {
    prompt += `\n--- ORIGINAL TYPESCRIPT-ESLINT IMPLEMENTATION ---\n`;

    if (originalSources.ruleContent) {
      prompt += `\nOriginal rule implementation (${originalSources.ruleName}.ts) from GitHub:\n`;
      prompt += `\`\`\`typescript\n${originalSources.ruleContent}\n\`\`\`\n`;
    }

    if (originalSources.testContent) {
      prompt += `\nOriginal test file (${originalSources.ruleName}.test.ts) from GitHub:\n`;
      prompt += `\`\`\`typescript\n${originalSources.testContent}\n\`\`\`\n`;
    }

    prompt += `\n--- END ORIGINAL SOURCES ---\n\n`;
  }

  prompt += `After analyzing in plan mode, fix the test failures above.

IMPORTANT: 
- ONLY edit files to fix the issues
- Focus on fixing BOTH Go and TypeScript test failures if both exist
- Ensure consistency between Go implementation and TypeScript test expectations
- Common issues include:
  - Go rule logic not matching TypeScript expectations
  - Missing test cases in Go tests
  - Incorrect assertions or expected values
  - Type mismatches between Go and TypeScript
  - Missing imports or dependencies
- This script will handle re-running the tests after your fixes
`;

  log(
    `Sending rule test failures to Claude CLI for fixing: ${ruleName}`,
    'info',
  );

  try {
    const result = await runClaudeWithStreaming(prompt);

    if (result.code === 0) {
      log('Claude CLI completed successfully', 'success');
    } else {
      log(`Claude CLI exited with code ${result.code}`, 'error');
    }

    return result.code === 0;
  } catch (error) {
    log(`Claude CLI error: ${error.message}`, 'error');
    return false;
  }
}

async function fixErrorWithClaudeCLI(
  errorOutput,
  command,
  originalSources = null,
  currentTestContent = null,
) {
  let prompt = `Go into plan mode first to analyze this error and plan the fix, performing deep research into the problem and paying attention to the ts versions as a refernce but also carefully considering the go environment.\n\n`;

  prompt += `Error occurred:\n\n${errorOutput}\n\n`;

  if (currentTestContent) {
    prompt += `\n--- CURRENT RSLINT TEST FILE ---\n`;
    prompt += `\`\`\`typescript\n${currentTestContent}\n\`\`\`\n`;
    prompt += `\n--- END CURRENT TEST ---\n\n`;
  }

  if (originalSources) {
    prompt += `\n--- ORIGINAL TYPESCRIPT-ESLINT IMPLEMENTATION ---\n`;

    if (originalSources.ruleContent) {
      prompt += `\nOriginal rule implementation (${originalSources.ruleName}.ts) from GitHub:\n`;
      prompt += `\`\`\`typescript\n${originalSources.ruleContent}\n\`\`\`\n`;
    }

    if (originalSources.testContent) {
      prompt += `\nOriginal test file (${originalSources.ruleName}.test.ts) from GitHub:\n`;
      prompt += `\`\`\`typescript\n${originalSources.testContent}\n\`\`\`\n`;
    }

    prompt += `\n--- END ORIGINAL SOURCES ---\n\n`;
  }

  prompt += `After analyzing in plan mode, fix the error above.

IMPORTANT: 
- ONLY edit files to fix the issue
- Focus solely on file editing and code fixes
- This script will handle running: ${command}
`;

  log('Sending error to Claude CLI for fixing...', 'info');

  try {
    const result = await runClaudeWithStreaming(prompt);

    if (result.code === 0) {
      log('Claude CLI completed successfully', 'success');
    } else {
      log(`Claude CLI exited with code ${result.code}`, 'error');
    }

    return result.code === 0;
  } catch (error) {
    log(`Claude CLI error: ${error.message}`, 'error');
    return false;
  }
}

// Helper function to chunk error output into equal pieces for workers
function chunkErrorOutput(errorOutput, numChunks) {
  const lines = errorOutput.split('\n');
  const chunks = [];

  // Calculate lines per chunk
  const linesPerChunk = Math.ceil(lines.length / numChunks);

  for (let i = 0; i < numChunks; i++) {
    const start = i * linesPerChunk;
    const end = Math.min(start + linesPerChunk, lines.length);

    if (start < lines.length) {
      const chunkLines = lines.slice(start, end);
      if (chunkLines.length > 0) {
        chunks.push(chunkLines.join('\n'));
      }
    }
  }

  return chunks;
}

async function runBuild(concurrentMode = false, workerCount = DEFAULT_WORKERS) {
  log('Starting build process...', 'info');

  try {
    const result = await runCommand(BUILD_COMMAND, BUILD_ARGS, {
      timeout: 120000,
    });

    if (result.code === 0) {
      log('Build successful', 'success');
      return true;
    } else {
      log(`Build failed with exit code ${result.code}`, 'error');
      if (result.stderr) {
        console.log(`${colors.red}Build stderr:${colors.reset}`);
        console.log(result.stderr);
      }
      if (result.stdout) {
        console.log(`${colors.yellow}Build stdout:${colors.reset}`);
        console.log(result.stdout);
      }

      const buildCommand = `${BUILD_COMMAND} ${BUILD_ARGS.join(' ')}`;
      const errorOutput = result.stderr || result.stdout;

      // Check if error output is very long and we're in concurrent mode
      const errorLines = errorOutput.split('\n').length;
      const shouldChunk = errorLines > 100 && concurrentMode && workerCount > 1;

      if (shouldChunk) {
        log(
          `Build error is large (${errorLines} lines), splitting into ${workerCount} chunks for parallel fixing...`,
          'info',
        );

        // Chunk the error output based on worker count
        const errorChunks = chunkErrorOutput(errorOutput, workerCount);
        log(
          `Split into ${errorChunks.length} chunks for parallel processing`,
          'info',
        );

        // Process chunks in parallel using Promise.all
        const fixPromises = errorChunks.map(async (chunk, index) => {
          log(
            `Processing error chunk ${index + 1}/${errorChunks.length}...`,
            'info',
          );

          // Create a more focused prompt for each chunk
          const chunkPrompt = `This is error chunk ${index + 1} of ${errorChunks.length} from a build failure.
Focus on fixing ONLY the errors in this specific chunk:

${chunk}

Build command: ${buildCommand}`;

          return await fixErrorWithClaudeCLI(chunkPrompt, buildCommand);
        });

        // Wait for all chunks to be processed
        const results = await Promise.all(fixPromises);

        // If any chunk was fixed, retry the build
        if (results.some(fixed => fixed)) {
          log('At least one error chunk was fixed, retrying build...', 'info');
          return await runBuild(concurrentMode, workerCount);
        } else {
          log('Failed to fix any build error chunks', 'error');
          return false;
        }
      } else {
        // Original single-threaded approach for smaller errors or non-concurrent mode
        const fixed = await fixErrorWithClaudeCLI(errorOutput, buildCommand);

        if (fixed) {
          // Retry build after fix
          log('Retrying build after Claude CLI fix...', 'info');
          return await runBuild(concurrentMode, workerCount);
        } else {
          log('Failed to fix build error with Claude CLI', 'error');
          return false;
        }
      }
    }
  } catch (error) {
    log(`Build process error: ${error.message}`, 'error');
    return false;
  }
}

async function getGoTestPackages() {
  try {
    // Get list of all Go packages with tests
    const result = await runCommand('go', ['list', './internal/...'], {
      cwd: __dirname,
    });

    if (result.code === 0) {
      const packages = result.stdout
        .trim()
        .split('\n')
        .filter(pkg => pkg.length > 0);
      return packages;
    }

    return [];
  } catch (error) {
    log(`Error listing Go packages: ${error.message}`, 'error');
    return [];
  }
}

async function getRuleTestPairs() {
  try {
    // Get TypeScript test files
    const testPath = join(__dirname, TEST_DIR);
    const tsTestFiles = await readdir(testPath);
    const tsTests = tsTestFiles
      .filter(file => file.endsWith('.test.ts'))
      .map(file => ({
        tsTestFile: join(testPath, file),
        ruleName: file.replace('.test.ts', ''),
        goRuleName: file.replace('.test.ts', '').replace(/-/g, '_'),
      }));

    // Get Go packages with tests
    const goPackages = await getGoTestPackages();

    // Create rule test pairs by matching rule names
    const ruleTestPairs = [];

    for (const tsTest of tsTests) {
      // Find corresponding Go package
      const goPackage = goPackages.find(pkg =>
        pkg.includes(`/internal/rules/${tsTest.goRuleName}`),
      );

      if (goPackage) {
        ruleTestPairs.push({
          ruleName: tsTest.ruleName,
          goRuleName: tsTest.goRuleName,
          goPackage: goPackage,
          tsTestFile: tsTest.tsTestFile,
        });
      } else {
        // TypeScript test without corresponding Go test
        log(
          `Warning: TypeScript test ${tsTest.ruleName} has no corresponding Go test`,
          'warning',
        );
        ruleTestPairs.push({
          ruleName: tsTest.ruleName,
          goRuleName: tsTest.goRuleName,
          goPackage: null,
          tsTestFile: tsTest.tsTestFile,
        });
      }
    }

    // Add Go-only packages (those without TypeScript tests)
    for (const goPackage of goPackages) {
      const goRuleName = goPackage.split('/').pop();
      const tsRuleName = goRuleName.replace(/_/g, '-');

      const alreadyIncluded = ruleTestPairs.find(
        pair => pair.goPackage === goPackage,
      );
      if (!alreadyIncluded) {
        log(
          `Warning: Go package ${goRuleName} has no corresponding TypeScript test`,
          'warning',
        );
        ruleTestPairs.push({
          ruleName: tsRuleName,
          goRuleName: goRuleName,
          goPackage: goPackage,
          tsTestFile: null,
        });
      }
    }

    return ruleTestPairs;
  } catch (error) {
    log(`Error getting rule test pairs: ${error.message}`, 'error');
    return [];
  }
}

async function runGoTestForRule(packagePath) {
  const packageName = packagePath.split('/').pop();

  try {
    const result = await runCommand('go', ['test', packagePath, '-v'], {
      timeout: 120000, // 2 minutes per package
      cwd: __dirname,
    });

    if (result.code === 0) {
      return { success: true, output: result.stdout };
    } else {
      return {
        success: false,
        output: result.stdout,
        error: result.stderr,
        command: `go test ${packagePath} -v`,
      };
    }
  } catch (error) {
    return {
      success: false,
      error: error.message,
      command: `go test ${packagePath} -v`,
    };
  }
}

async function runTsTestForRule(testFile) {
  const testName = basename(testFile);

  try {
    const result = await runCommand(
      'node',
      ['--import=tsx/esm', '--test', testFile],
      {
        timeout: TEST_TIMEOUT,
        cwd: join(__dirname, 'packages/rslint-test-tools'),
      },
    );

    if (result.code === 0) {
      return { success: true, output: result.stdout };
    } else {
      return {
        success: false,
        output: result.stdout,
        error: result.stderr,
        command: `node --import=tsx/esm --test ${testFile}`,
      };
    }
  } catch (error) {
    return {
      success: false,
      error: error.message,
      command: `node --import=tsx/esm --test ${testFile}`,
    };
  }
}

async function runSingleRuleTest(ruleTestPair, attemptNumber = 1) {
  const { ruleName, goRuleName, goPackage, tsTestFile } = ruleTestPair;
  const startTime = Date.now();

  logProgress('Rule test execution started', {
    phase: 'test-start',
    testName: ruleName,
    attempt: attemptNumber,
    maxAttempts: MAX_FIX_ATTEMPTS,
  });

  // Fetch original TypeScript ESLint sources (only on first attempt)
  let originalSources = null;
  let currentTestContent = null;

  if (attemptNumber === 1) {
    log('Fetching original sources from GitHub...', 'info');

    originalSources = await fetchOriginalRule(ruleName);
    if (originalSources.ruleContent || originalSources.testContent) {
      log(
        `✓ Original sources fetched (rule: ${originalSources.ruleContent ? 'yes' : 'no'}, test: ${originalSources.testContent ? 'yes' : 'no'})`,
        'success',
      );
    }
  } else {
    // Re-fetch on subsequent attempts as it might have been fixed
    originalSources = await fetchOriginalRule(ruleName);
  }

  // Always read the current RSLint test file if it exists
  if (tsTestFile) {
    try {
      currentTestContent = await readFile(tsTestFile, 'utf8');
      log(
        `✓ Current test file read (${currentTestContent.length} bytes)`,
        'success',
      );
    } catch (err) {
      log(`Failed to read current test file: ${err.message}`, 'error');
    }
  }

  // Run both Go and TypeScript tests
  let goTestResult = null;
  let tsTestResult = null;

  // Run Go test if package exists
  if (goPackage) {
    log(`Running Go test for ${goRuleName}...`, 'info');
    goTestResult = await runGoTestForRule(goPackage);

    if (goTestResult.success) {
      log(`✓ Go test passed for ${goRuleName}`, 'success');
    } else {
      log(`✗ Go test failed for ${goRuleName}`, 'error');
    }
  }

  // Run TypeScript test if file exists
  if (tsTestFile) {
    log(`Running TypeScript test for ${ruleName}...`, 'info');
    tsTestResult = await runTsTestForRule(tsTestFile);

    if (tsTestResult.success) {
      log(`✓ TypeScript test passed for ${ruleName}`, 'success');
    } else {
      log(`✗ TypeScript test failed for ${ruleName}`, 'error');
    }
  }

  // Check if both tests passed (or if only one test exists and it passed)
  const goSuccess = !goPackage || goTestResult.success;
  const tsSuccess = !tsTestFile || tsTestResult.success;

  if (goSuccess && tsSuccess) {
    const duration = Date.now() - startTime;
    logProgress('Test passed', {
      phase: 'test-pass',
      testName: ruleName,
      durationMs: duration,
    });
    completedTests++;
    return true;
  }

  // If we get here, at least one test failed
  logProgress('Test failed', {
    phase: 'test-fail',
    testName: ruleName,
    goFailed: !goSuccess,
    tsFailed: !tsSuccess,
  });

  if (attemptNumber < MAX_FIX_ATTEMPTS) {
    // Try to fix with Claude CLI, passing both Go and TypeScript test results
    log(
      `Attempting to fix rule test failures (attempt ${attemptNumber}/${MAX_FIX_ATTEMPTS})...`,
      'warning',
    );

    const fixed = await fixRuleTestsWithClaude(
      ruleName,
      goTestResult,
      tsTestResult,
      originalSources,
      currentTestContent,
    );

    if (fixed) {
      // Claude thinks it fixed the issues, let's rebuild and retry
      log('Claude completed fix attempt, rebuilding...', 'info');

      // Run build again to ensure any changes are compiled
      const buildSuccess = await runBuild(false, 1); // Use non-concurrent mode for individual rule fixes

      if (!buildSuccess) {
        log(`Build failed after fix attempt ${attemptNumber}`, 'error');
        failedTests++;
        return false;
      }

      // Retry test with incremented attempt number
      log(`Retrying rule tests after fix and rebuild...`, 'info');
      return await runSingleRuleTest(ruleTestPair, attemptNumber + 1);
    } else {
      log(`Claude CLI failed to fix the rule test issues`, 'error');
    }
  }

  // Max attempts reached or fix failed
  log(`Rule test failed after ${attemptNumber} attempts`, 'error');
  failedTests++;
  return false;
}

async function runAllRuleTests(
  concurrentMode = false,
  workerCount = DEFAULT_WORKERS,
) {
  const ruleTestPairs = await getRuleTestPairs();
  totalTests = ruleTestPairs.length;

  if (!IS_WORKER) {
    console.log('\n' + '='.repeat(60));
    log(`Starting combined rule test suite with ${totalTests} rules`, 'info');
    console.log('='.repeat(60));
  }

  if (concurrentMode && !IS_WORKER) {
    // Main process in concurrent mode
    await runConcurrentRuleTests(ruleTestPairs, workerCount);
  } else if (IS_WORKER) {
    // Worker process
    await runWorker();
  } else {
    // Sequential mode
    for (let i = 0; i < ruleTestPairs.length; i++) {
      const ruleTestPair = ruleTestPairs[i];
      const ruleName = ruleTestPair.ruleName;

      console.log(
        `\n${colors.bright}[${i + 1}/${totalTests}] ${ruleName}${colors.reset}`,
      );
      console.log('-'.repeat(40));

      await runSingleRuleTest(ruleTestPair);

      // Show running totals
      console.log(
        `\n${colors.dim}Progress: ${completedTests} passed, ${failedTests} failed, ${totalTests - completedTests - failedTests} remaining${colors.reset}`,
      );
    }

    logProgress('Rule test suite completed', {
      phase: 'script-complete',
      totalTests,
      completedTests,
      failedTests,
      successRate:
        totalTests > 0 ? Math.round((completedTests / totalTests) * 100) : 0,
      testResults: {
        passed: completedTests,
        failed: failedTests,
        total: totalTests,
        successRate:
          totalTests > 0 ? Math.round((completedTests / totalTests) * 100) : 0,
      },
    });
  }
}

async function runConcurrentRuleTests(ruleTestPairs, workerCount) {
  const workQueue = new WorkQueue(WORK_QUEUE_DIR);
  await workQueue.initialize();

  // Add all rule test pairs to work queue
  await workQueue.addWork(ruleTestPairs);
  log(`Added ${ruleTestPairs.length} rule test pairs to work queue`, 'info');

  // Create hook configuration
  const hookConfig = {
    hooks: {
      PreToolUse: [
        {
          matcher: 'Write|Edit|MultiEdit',
          hooks: [
            {
              type: 'command',
              command: join(__dirname, 'hooks', 'pre-tool-use.js'),
            },
          ],
        },
      ],
      PostToolUse: [
        {
          matcher: 'Write|Edit|MultiEdit',
          hooks: [
            {
              type: 'command',
              command: join(__dirname, 'hooks', 'post-tool-use.js'),
            },
          ],
        },
      ],
    },
  };

  // Create hooks directory and files
  await createHooks();

  // Write hook configuration
  const configPath = join(
    os.homedir(),
    '.config',
    'claude-code',
    'settings.json',
  );
  await mkdir(dirname(configPath), { recursive: true });
  await writeFile(configPath, JSON.stringify(hookConfig, null, 2));

  // Start workers
  const workers = [];
  for (let i = 0; i < workerCount; i++) {
    const workerId = `worker_${i}_${randomBytes(4).toString('hex')}`;
    log(`Starting worker ${i + 1}: ${workerId}`, 'info');

    const worker = spawn(process.argv[0], [__filename], {
      env: {
        ...process.env,
        RSLINT_WORKER_ID: workerId,
        RSLINT_WORK_QUEUE_DIR: WORK_QUEUE_DIR,
      },
      stdio: 'inherit',
    });

    workers.push({ id: workerId, process: worker });
  }

  // Monitor progress
  const progressInterval = setInterval(async () => {
    const progress = await workQueue.getProgress();
    const successRate =
      progress.total > 0
        ? Math.round((progress.completed / progress.total) * 100)
        : 0;
    log(
      `Progress: ${progress.completed + progress.failed}/${progress.total} (${successRate}% success) - ${progress.completed} passed, ${progress.failed} failed, ${progress.claimed} in progress`,
      'progress',
    );
  }, 10000); // Every 10 seconds

  // Wait for all workers to complete
  await Promise.all(
    workers.map(
      w =>
        new Promise(resolve => {
          w.process.on('exit', code => {
            log(
              `Worker ${w.id} exited with code ${code}`,
              code === 0 ? 'success' : 'error',
            );
            resolve();
          });
        }),
    ),
  );

  clearInterval(progressInterval);

  // Get final results
  const finalProgress = await workQueue.getProgress();
  completedTests = finalProgress.completed;
  failedTests = finalProgress.failed;

  logProgress('Test suite completed', {
    phase: 'script-complete',
    totalTests,
    completedTests,
    failedTests,
    successRate:
      totalTests > 0 ? Math.round((completedTests / totalTests) * 100) : 0,
    testResults: {
      passed: completedTests,
      failed: failedTests,
      total: totalTests,
      successRate:
        totalTests > 0 ? Math.round((completedTests / totalTests) * 100) : 0,
    },
  });

  // Cleanup
  await workQueue.cleanup();
}

async function runWorker() {
  const workQueueDir = process.env.RSLINT_WORK_QUEUE_DIR || WORK_QUEUE_DIR;
  const workQueue = new WorkQueue(workQueueDir);

  while (true) {
    const work = await workQueue.claimWork(WORKER_ID);

    if (!work) {
      log(`Worker ${WORKER_ID}: No more work available, exiting`, 'info');
      break;
    }

    log(`Worker ${WORKER_ID}: Processing ${work.test.ruleName}`, 'info');

    try {
      const success = await runSingleRuleTest(work.test);
      await workQueue.completeWork(work.id, success);

      if (success) {
        log(
          `Worker ${WORKER_ID}: Completed ${work.test.ruleName} successfully`,
          'success',
        );
      } else {
        log(`Worker ${WORKER_ID}: Failed ${work.test.ruleName}`, 'error');
      }
    } catch (err) {
      log(
        `Worker ${WORKER_ID}: Error processing ${work.test.ruleName}: ${err.message}`,
        'error',
      );
      await workQueue.completeWork(work.id, false);
    }
  }

  process.exit(0);
}

async function createHooks() {
  const hooksDir = join(__dirname, 'hooks');
  await mkdir(hooksDir, { recursive: true });

  // Pre-tool-use hook for file locking
  const preToolUseHook = `#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const crypto = require('crypto');

// Parse input from stdin
let input = '';
process.stdin.on('data', chunk => input += chunk);
process.stdin.on('end', async () => {
  try {
    const data = JSON.parse(input);
    const { tool, params } = data;
    
    // Lock files when they're being edited
    if (tool === 'Edit' || tool === 'MultiEdit' || tool === 'Write') {
      const filePath = params.file_path || params.path;
      if (filePath && filePath.includes('rslint')) {
        const lockFile = filePath + '.lock.' + process.env.RSLINT_WORKER_ID;
        const lockDir = path.dirname(lockFile);
        
        // Try to acquire lock
        let locked = false;
        for (let i = 0; i < 10; i++) {
          try {
            // Check for other locks
            const files = fs.readdirSync(lockDir).filter(f => 
              f.startsWith(path.basename(filePath) + '.lock.') && 
              f !== path.basename(lockFile)
            );
            
            if (files.length === 0) {
              // No other locks, create ours
              fs.writeFileSync(lockFile, process.env.RSLINT_WORKER_ID || 'main', { flag: 'wx' });
              locked = true;
              break;
            }
          } catch (err) {
            // Directory might not exist or lock already exists
          }
          
          if (!locked && i < 9) {
            // Wait before retry
            await new Promise(resolve => setTimeout(resolve, 500 + Math.random() * 500));
          }
        }
        
        if (!locked) {
          console.error(JSON.stringify({
            error: 'Could not acquire file lock',
            file: filePath,
            worker: process.env.RSLINT_WORKER_ID
          }));
          process.exit(1);
        }
      }
    }
    
    // Allow tool to proceed
    console.log(JSON.stringify({ allow: true }));
  } catch (err) {
    console.error(JSON.stringify({ error: err.message }));
    process.exit(1);
  }
});
`;

  // Post-tool-use hook for releasing locks
  const postToolUseHook = `#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Parse input from stdin
let input = '';
process.stdin.on('data', chunk => input += chunk);
process.stdin.on('end', () => {
  try {
    const data = JSON.parse(input);
    const { tool, params } = data;
    
    // Release file locks
    if (tool === 'Edit' || tool === 'MultiEdit' || tool === 'Write') {
      const filePath = params.file_path || params.path;
      if (filePath && filePath.includes('rslint')) {
        const lockFile = filePath + '.lock.' + process.env.RSLINT_WORKER_ID;
        
        try {
          fs.unlinkSync(lockFile);
        } catch (err) {
          // Lock might already be gone
        }
      }
    }
    
    // Always allow
    console.log(JSON.stringify({ allow: true }));
  } catch (err) {
    console.error(JSON.stringify({ error: err.message }));
    process.exit(1);
  }
});
`;

  await writeFile(join(hooksDir, 'pre-tool-use.js'), preToolUseHook);
  await writeFile(join(hooksDir, 'post-tool-use.js'), postToolUseHook);

  // Make hooks executable
  await chmod(join(hooksDir, 'pre-tool-use.js'), 0o755);
  await chmod(join(hooksDir, 'post-tool-use.js'), 0o755);
}

async function runCompleteProcess() {
  const scriptStartTime = Date.now();

  // Parse command line arguments
  const args = process.argv.slice(2);
  const showHelp = args.includes('--help') || args.includes('-h');
  const concurrentMode = args.includes('--concurrent');
  const workerCountArg = args.find(arg => arg.startsWith('--workers='));
  const workerCount = workerCountArg
    ? parseInt(workerCountArg.split('=')[1])
    : DEFAULT_WORKERS;

  if (showHelp && !IS_WORKER) {
    console.log(
      `\nRSLint Automated Build & Test Runner\n\nUsage: node automate-build-test.js [options]\n\nOptions:\n  --concurrent      Run tests in parallel using multiple Claude instances\n  --workers=N       Number of parallel workers (default: ${DEFAULT_WORKERS})\n  --help, -h        Show this help message\n\nExamples:\n  node automate-build-test.js                    # Sequential execution\n  node automate-build-test.js --concurrent       # Parallel with ${DEFAULT_WORKERS} workers\n  node automate-build-test.js --concurrent --workers=8  # Parallel with 8 workers\n`,
    );
    process.exit(0);
  }

  if (!IS_WORKER) {
    console.clear();
    console.log(
      `${colors.bright}${colors.cyan}╔═══════════════════════════════════════════════════════════╗${colors.reset}`,
    );
    console.log(
      `${colors.bright}${colors.cyan}║          RSLint Automated Build & Test Runner             ║${colors.reset}`,
    );
    console.log(
      `${colors.bright}${colors.cyan}╚═══════════════════════════════════════════════════════════╝${colors.reset}\n`,
    );

    log(
      `Script started (PID: ${process.pid}, Node: ${process.version})`,
      'info',
    );
    log(`Max fix attempts per test: ${MAX_FIX_ATTEMPTS}`, 'info');

    if (concurrentMode) {
      log(`Running in concurrent mode with ${workerCount} workers`, 'info');
    }
  } else {
    log(`Worker ${WORKER_ID} started (PID: ${process.pid})`, 'info');
  }

  // Step 1: Build (only for main process)
  if (!IS_WORKER) {
    console.log(`\n${colors.bright}=== BUILD PHASE ===${colors.reset}`);
    const buildSuccess = await runBuild(concurrentMode, workerCount);

    if (!buildSuccess) {
      log('Build failed, stopping automation', 'error');
      process.exit(1);
    }
  }

  // Step 2: Run combined rule tests (Go + TypeScript together)
  if (!IS_WORKER) {
    console.log(
      `\n${colors.bright}=== COMBINED RULE TEST PHASE ===${colors.reset}`,
    );
  }
  await runAllRuleTests(concurrentMode, workerCount);

  if (!IS_WORKER) {
    const totalDuration = Date.now() - scriptStartTime;
    logProgress('Automation completed', {
      phase: 'script-complete',
      totalDurationMs: totalDuration,
      totalDurationMinutes: Math.round(totalDuration / 60000),
      buildSuccess: true,
      testResults: {
        total: totalTests,
        passed: completedTests,
        failed: failedTests,
        successRate:
          totalTests > 0 ? Math.round((completedTests / totalTests) * 100) : 0,
      },
    });

    return {
      success: failedTests === 0,
      totalTests,
      completedTests,
      failedTests,
      totalDuration,
    };
  }
}

async function main() {
  const TOTAL_RUNS = 20;

  if (!IS_WORKER) {
    console.clear();
    console.log(
      `${colors.bright}${colors.magenta}╔═══════════════════════════════════════════════════════════╗${colors.reset}`,
    );
    console.log(
      `${colors.bright}${colors.magenta}║          RSLint 10x Automated Build & Test Runner         ║${colors.reset}`,
    );
    console.log(
      `${colors.bright}${colors.magenta}╚═══════════════════════════════════════════════════════════╝${colors.reset}\n`,
    );

    log(`Starting ${TOTAL_RUNS} consecutive automation runs`, 'info');
    console.log('='.repeat(80));
  }

  let allRunsResults = [];

  for (let runNumber = 1; runNumber <= TOTAL_RUNS; runNumber++) {
    if (!IS_WORKER) {
      console.log(
        `\n${colors.bright}${colors.yellow}╔═════════════════════════════════════════════════════════════════════════════╗${colors.reset}`,
      );
      console.log(
        `${colors.bright}${colors.yellow}║                              RUN ${runNumber.toString().padStart(2)} OF ${TOTAL_RUNS}                                ║${colors.reset}`,
      );
      console.log(
        `${colors.bright}${colors.yellow}╚═════════════════════════════════════════════════════════════════════════════╝${colors.reset}\n`,
      );

      // Reset counters for each run
      totalTests = 0;
      completedTests = 0;
      failedTests = 0;
    }

    const runResult = await runCompleteProcess();

    if (!IS_WORKER) {
      allRunsResults.push({
        runNumber,
        ...runResult,
      });

      console.log(
        `\n${colors.bright}${colors.cyan}RUN ${runNumber} COMPLETE:${colors.reset}`,
      );
      console.log(
        `  • Tests: ${runResult.completedTests}/${runResult.totalTests} passed`,
      );
      console.log(
        `  • Duration: ${Math.round(runResult.totalDuration / 60000)} minutes`,
      );
      console.log(
        `  • Status: ${runResult.success ? colors.green + 'SUCCESS' + colors.reset : colors.red + 'FAILED' + colors.reset}`,
      );

      // Brief pause between runs
      if (runNumber < TOTAL_RUNS) {
        log('Pausing 5 seconds before next run...', 'info');
        await new Promise(resolve => setTimeout(resolve, 5000));
      }
    }
  }

  if (!IS_WORKER) {
    // Final summary
    console.log(
      `\n\n${colors.bright}${colors.magenta}╔═════════════════════════════════════════════════════════════════════════════╗${colors.reset}`,
    );
    console.log(
      `${colors.bright}${colors.magenta}║                            FINAL SUMMARY (${TOTAL_RUNS} RUNS)                             ║${colors.reset}`,
    );
    console.log(
      `${colors.bright}${colors.magenta}╚═════════════════════════════════════════════════════════════════════════════╝${colors.reset}\n`,
    );

    const totalSuccessfulRuns = allRunsResults.filter(r => r.success).length;
    const totalFailedRuns = TOTAL_RUNS - totalSuccessfulRuns;
    const averageDuration =
      allRunsResults.reduce((sum, r) => sum + r.totalDuration, 0) / TOTAL_RUNS;

    console.log(`${colors.bright}OVERALL RESULTS:${colors.reset}`);
    console.log(
      `  • Successful runs: ${colors.green}${totalSuccessfulRuns}/${TOTAL_RUNS}${colors.reset}`,
    );
    console.log(
      `  • Failed runs: ${colors.red}${totalFailedRuns}/${TOTAL_RUNS}${colors.reset}`,
    );
    console.log(
      `  • Success rate: ${totalSuccessfulRuns === TOTAL_RUNS ? colors.green : colors.yellow}${Math.round((totalSuccessfulRuns / TOTAL_RUNS) * 100)}%${colors.reset}`,
    );
    console.log(
      `  • Average duration: ${Math.round(averageDuration / 60000)} minutes`,
    );

    console.log(`\n${colors.bright}DETAILED RESULTS:${colors.reset}`);
    allRunsResults.forEach(result => {
      const status = result.success
        ? colors.green + '✓' + colors.reset
        : colors.red + '✗' + colors.reset;
      const duration = Math.round(result.totalDuration / 60000);
      const testSummary = `${result.completedTests}/${result.totalTests}`;
      console.log(
        `  Run ${result.runNumber.toString().padStart(2)}: ${status} ${testSummary.padEnd(8)} (${duration}m)`,
      );
    });

    console.log(
      `\n${colors.bright}${colors.magenta}10x automation run completed!${colors.reset}`,
    );
    console.log('='.repeat(80));

    process.exit(totalFailedRuns > 0 ? 1 : 0);
  }
}

// Handle graceful shutdown
process.on('SIGINT', () => {
  console.log(`\n${colors.yellow}Script interrupted by user${colors.reset}`);
  process.exit(130);
});

process.on('SIGTERM', () => {
  console.log(`\n${colors.yellow}Script terminated${colors.reset}`);
  process.exit(143);
});

main().catch(error => {
  log(`Script failed with unhandled error: ${error.message}`, 'error');
  console.error(error.stack);
  process.exit(1);
});
