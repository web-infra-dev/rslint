#!/usr/bin/env node

const { spawn } = require('child_process');
const { readdir, readFile, writeFile, mkdir, unlink, access, rm, chmod } = require('fs/promises');
const { join, basename, dirname } = require('path');
const https = require('https');
const { randomBytes } = require('crypto');
const os = require('os');

// __dirname is available in CommonJS

// Configuration
const TEST_TIMEOUT = 120000; // 120 seconds (2 minutes)
const TEST_DIR = 'packages/rslint-test-tools/tests/typescript-eslint/rules';
const TSLINT_BASE_URL = 'https://raw.githubusercontent.com/typescript-eslint/typescript-eslint/main';

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
      await writeFile(workFile, JSON.stringify({
        id: i,
        test: items[i],
        status: 'pending',
        createdAt: Date.now()
      }));
    }
  }

  async claimWork(workerId) {
    const files = await readdir(this.workDir);
    const workFiles = files.filter(f => f.startsWith('work_') && f.endsWith('.json'));

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
    const workFiles = files.filter(f => f.startsWith('work_') && f.endsWith('.json'));

    let pending = 0, claimed = 0, completed = 0, failed = 0;

    for (const file of workFiles) {
      const work = JSON.parse(await readFile(join(this.workDir, file), 'utf8'));
      switch (work.status) {
        case 'pending': pending++; break;
        case 'claimed': claimed++; break;
        case 'completed': completed++; break;
        case 'failed': failed++; break;
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
  white: '\x1b[37m'
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

  console.log(`${colors.dim}${timestamp}${colors.reset} ${color}${prefix} ${message}${colors.reset}`);
}

function logProgress(message, data = {}) {
  // Special handling for Claude output
  if (data.phase && data.phase.startsWith('claude-')) {
    if (data.phase === 'claude-text' || data.phase === 'claude-text-final') {
      log(`Claude: ${data.content}`, 'claude');
      return;
    } else if (data.phase === 'claude-code' || data.phase === 'claude-code-final') {
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
  if (data.phase === 'test-start') {
    console.log('');
    log(`Testing ${data.testName} (attempt ${data.attempt}/${data.maxAttempts})`, 'progress');
  } else if (data.phase === 'test-pass') {
    log(`✓ ${data.testName} passed in ${data.durationMs}ms`, 'success');
  } else if (data.phase === 'test-fail') {
    log(`✗ ${data.testName} failed with exit code ${data.exitCode}`, 'error');
  } else if (data.phase === 'script-complete') {
    console.log('\n' + '='.repeat(60));
    log('Automation Complete', 'info');
    log(`Total Duration: ${data.totalDurationMinutes} minutes`, 'info');
    log(`Tests: ${data.testResults.passed}/${data.testResults.total} passed (${data.testResults.successRate}%)`,
        data.testResults.failed > 0 ? 'warning' : 'success');
    console.log('='.repeat(60));
  } else {
    log(message, 'info');
  }
}

async function fetchFromGitHub(url) {
  return new Promise((resolve, reject) => {
    https.get(url, (res) => {
      let data = '';
      res.on('data', (chunk) => { data += chunk; });
      res.on('end', () => {
        if (res.statusCode === 200) {
          resolve(data);
        } else {
          resolve(null); // Return null if not found
        }
      });
    }).on('error', (err) => {
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
    testContent
  };
}

async function runCommand(command, args, options = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      stdio: 'pipe',
      cwd: __dirname,
      ...options
    });

    let stdout = '';
    let stderr = '';

    child.stdout?.on('data', (data) => {
      stdout += data.toString();
    });

    child.stderr?.on('data', (data) => {
      stderr += data.toString();
    });

    const timeout = options.timeout ? setTimeout(() => {
      child.kill('SIGKILL');
      reject(new Error(`Command timed out after ${options.timeout}ms`));
    }, options.timeout) : null;

    child.on('close', (code) => {
      if (timeout) clearTimeout(timeout);
      resolve({ code, stdout, stderr });
    });

    child.on('error', (error) => {
      if (timeout) clearTimeout(timeout);
      reject(error);
    });
  });
}

async function runClaudeWithStreaming(prompt) {
  return new Promise((resolve) => {
    // Use same flags as porter: -p, --verbose, --output-format stream-json
    const settingsFile = join(__dirname, '.claude', 'settings.local.json');
    const args = [
      '-p',
      '--verbose',
      '--output-format', 'stream-json',
      '--max-turns', '500',
      '--settings', settingsFile,
      '--dangerously-skip-permissions'
    ];

    const child = spawn('claude', args, {
      stdio: ['pipe', 'pipe', 'pipe'],
      env: { ...process.env }
    });

    let fullOutput = '';
    let fullError = '';
    let jsonBuffer = '';

    // Process stdout stream for JSON
    child.stdout.on('data', (data) => {
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
                    process.stdout.write(`${colors.magenta}🤖 Claude: ${line}${colors.reset}\n`);
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
                const resultStr = typeof resultContent === 'string' ? resultContent : JSON.stringify(resultContent);
                const lines = resultStr.split('\n');
                const maxLines = 10;
                const displayLines = lines.slice(0, maxLines);

                log(`Tool result${lines.length > maxLines ? ` (showing first ${maxLines} lines)` : ''}:`, 'success');
                for (const line of displayLines) {
                  console.log(colors.dim + '   ' + line + colors.reset);
                }
                if (lines.length > maxLines) {
                  console.log(colors.dim + `   ... (${lines.length - maxLines} more lines)` + colors.reset);
                }
              }
            }
          } else if (json.type === 'result') {
            // Display final result
            if (json.subtype === 'success') {
              log('Claude completed successfully', 'success');
            } else if (json.subtype === 'error_max_turns') {
              log(`Claude reached max turns (${json.num_turns} turns)`, 'warning');
            } else if (json.subtype?.includes('error')) {
              log(`Claude error (${json.subtype})`, 'error');
              if (json.result?.message) {
                log(`Details: ${json.result.message}`, 'error');
              }
            }

            if (json.usage) {
              log(`Tokens used: ${json.usage.input_tokens} in, ${json.usage.output_tokens} out`, 'info');
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

    child.stderr.on('data', (data) => {
      const error = data.toString();
      fullError += error;
      if (error.trim()) {
        log(`Claude CLI error: ${error.trim()}`, 'error');
      }
    });

    child.on('close', (code) => {
      // Process any remaining JSON buffer
      if (jsonBuffer.trim()) {
        try {
          const json = JSON.parse(jsonBuffer);
          if (json.delta?.text) {
            process.stdout.write(`${colors.magenta}${json.delta.text}${colors.reset}\n`);
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
        stderr: fullError
      });
    });

    // Set timeout for 5 minutes (increased for complex fixes)
    const timeout = setTimeout(() => {
      child.kill('SIGKILL');
      log('Claude CLI timeout after 5 minutes', 'error');
      resolve({
        code: -1,
        stdout: fullOutput,
        stderr: 'Process timed out after 5 minutes'
      });
    }, 900000); // 5 minutes

    // Clear timeout on close
    child.on('close', () => clearTimeout(timeout));

    // Write prompt to stdin instead of passing as argument
    child.stdin.write(prompt);
    child.stdin.end();
  });
}

async function validatePortWithClaudeCLI(testFile, originalSources) {
  const testName = basename(testFile);
  const ruleName = testName.replace('.test.ts', '');
  
  // Find the corresponding Go implementation
  const goRulePath = join(__dirname, 'internal', 'rules', ruleName.replace(/-/g, '_'), ruleName.replace(/-/g, '_') + '.go');
  let goRuleContent = null;
  
  try {
    goRuleContent = await readFile(goRulePath, 'utf8');
  } catch (err) {
    log(`Could not find Go implementation at ${goRulePath}`, 'warning');
    return false;
  }

  let prompt = `Your task is to validate that the Go port of a TypeScript-ESLint rule is functionally correct by comparing implementations.

FOCUS: Study the TypeScript implementation and ensure the Go version captures the same logic, edge cases, and behavior.

## Rule: ${ruleName}

### Original TypeScript Implementation:
`;

  if (originalSources && originalSources.ruleContent) {
    prompt += `\`\`\`typescript\n${originalSources.ruleContent}\n\`\`\`\n`;
  } else {
    prompt += `(TypeScript implementation not available)\n`;
  }

  prompt += `\n### Go Port Implementation:
\`\`\`go\n${goRuleContent}\n\`\`\`\n`;

  if (originalSources && originalSources.testContent) {
    prompt += `\n### Original Test Cases (for reference):
\`\`\`typescript\n${originalSources.testContent}\n\`\`\`\n`;
  }

  // Also include the current RSLint test to see what we're testing
  try {
    const currentTestContent = await readFile(testFile, 'utf8');
    prompt += `\n### Current RSLint Test:
\`\`\`typescript\n${currentTestContent}\n\`\`\`\n`;
  } catch (err) {
    log(`Could not read current test file: ${err.message}`, 'warning');
  }

  prompt += `
## Validation Tasks:

1. **Core Logic Comparison**: Compare the core rule logic between TypeScript and Go implementations
2. **Edge Case Coverage**: Ensure the Go version handles the same edge cases as TypeScript
3. **AST Pattern Matching**: Verify the Go version correctly identifies the same AST patterns
4. **Error Messages**: Check that error messages are consistent
5. **Configuration Options**: Ensure rule options are handled equivalently
6. **Type Checking**: Verify TypeScript type-aware features are properly ported

## Analysis Framework:

For each aspect, provide:
- ✅ **CORRECT**: When Go implementation matches TypeScript behavior
- ⚠️ **POTENTIAL ISSUE**: When there might be a discrepancy 
- ❌ **INCORRECT**: When Go implementation differs from TypeScript

## Output Format:

Provide a structured analysis covering:

### Functional Equivalence Analysis
- Core rule logic comparison
- Edge case handling
- AST pattern matching

### Implementation Details
- Error message consistency  
- Configuration option handling
- Type checking behavior

### Recommendations
- Any fixes needed for the Go implementation
- Missing functionality that should be added
- Test cases that should be enhanced

Focus on ensuring the Go port captures all the nuances of the original TypeScript implementation.
`;

  log(`Validating port correctness for ${ruleName}...`, 'info');

  try {
    const result = await runClaudeWithStreaming(prompt);

    if (result.code === 0) {
      log(`Port validation completed for ${ruleName}`, 'success');
    } else {
      log(`Port validation failed for ${ruleName}`, 'error');
    }

    return result.code === 0;
  } catch (error) {
    log(`Port validation error: ${error.message}`, 'error');
    return false;
  }
}


async function getTestFiles() {
  try {
    const testPath = join(__dirname, TEST_DIR);
    const files = await readdir(testPath);
    return files
      .filter(file => file.endsWith('.test.ts'))
      .map(file => join(testPath, file));
  } catch (error) {
    log(`Error reading test directory: ${error.message}`, 'error');
    return [];
  }
}

async function runSingleValidation(testFile) {
  const testName = basename(testFile);
  const startTime = Date.now();

  logProgress('Validation started', {
    phase: 'validation-start',
    testName,
    testFile
  });

  // Fetch original TypeScript ESLint sources
  log('Fetching original sources from GitHub...', 'info');
  const originalSources = await fetchOriginalRule(testName);
  
  if (!originalSources.ruleContent) {
    log(`⚠️ Original TypeScript rule not found for ${testName}`, 'warning');
    // Still proceed with validation using available sources
  }

  if (originalSources.ruleContent || originalSources.testContent) {
    log(`✓ Original sources fetched (rule: ${originalSources.ruleContent ? 'yes' : 'no'}, test: ${originalSources.testContent ? 'yes' : 'no'})`, 'success');
  }

  // Validate the port
  try {
    const validationSuccess = await validatePortWithClaudeCLI(testFile, originalSources);
    
    const duration = Date.now() - startTime;

    if (validationSuccess) {
      logProgress('Validation completed', {
        phase: 'validation-complete',
        testName,
        durationMs: duration
      });
      completedTests++;
      return true;
    } else {
      logProgress('Validation failed', {
        phase: 'validation-fail',
        testName
      });
      failedTests++;
      return false;
    }
  } catch (error) {
    log(`Validation error: ${error.message}`, 'error');
    failedTests++;
    return false;
  }
}

async function runAllTests(concurrentMode = false, workerCount = DEFAULT_WORKERS) {
  const testFiles = await getTestFiles();
  totalTests = testFiles.length;

  if (!IS_WORKER) {
    console.log('\n' + '='.repeat(60));
    log(`Starting port validation for ${totalTests} rules`, 'info');
    console.log('='.repeat(60));
  }

  if (concurrentMode && !IS_WORKER) {
    // Main process in concurrent mode
    await runConcurrentTests(testFiles, workerCount);
  } else if (IS_WORKER) {
    // Worker process
    await runWorker();
  } else {
    // Sequential mode
    for (let i = 0; i < testFiles.length; i++) {
      const testFile = testFiles[i];
      const testName = basename(testFile);

      console.log(`\n${colors.bright}[${i + 1}/${totalTests}] ${testName}${colors.reset}`);
      console.log('-'.repeat(40));

      await runSingleValidation(testFile);

      // Show running totals
      console.log(`\n${colors.dim}Progress: ${completedTests} validated, ${failedTests} failed, ${totalTests - completedTests - failedTests} remaining${colors.reset}`);
    }

    logProgress('Port validation completed', {
      phase: 'script-complete',
      totalTests,
      completedTests,
      failedTests,
      successRate: totalTests > 0 ? Math.round((completedTests / totalTests) * 100) : 0,
      testResults: {
        passed: completedTests,
        failed: failedTests,
        total: totalTests,
        successRate: totalTests > 0 ? Math.round((completedTests / totalTests) * 100) : 0
      }
    });
  }
}

async function runConcurrentTests(testFiles, workerCount) {
  const workQueue = new WorkQueue(WORK_QUEUE_DIR);
  await workQueue.initialize();

  // Add all tests to work queue
  await workQueue.addWork(testFiles);
  log(`Added ${testFiles.length} tests to work queue`, 'info');

  // Create hook configuration
  const hookConfig = {
    hooks: {
      PreToolUse: [{
        matcher: "Write|Edit|MultiEdit",
        hooks: [{
          type: "command",
          command: join(__dirname, 'hooks', 'pre-tool-use.js')
        }]
      }],
      PostToolUse: [{
        matcher: "Write|Edit|MultiEdit",
        hooks: [{
          type: "command",
          command: join(__dirname, 'hooks', 'post-tool-use.js')
        }]
      }]
    }
  };

  // Create hooks directory and files
  await createHooks();

  // Write hook configuration
  const configPath = join(os.homedir(), '.config', 'claude-code', 'settings.json');
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
        RSLINT_WORK_QUEUE_DIR: WORK_QUEUE_DIR
      },
      stdio: 'inherit'
    });

    workers.push({ id: workerId, process: worker });
  }

  // Monitor progress
  const progressInterval = setInterval(async () => {
    const progress = await workQueue.getProgress();
    const successRate = progress.total > 0 ? Math.round((progress.completed / progress.total) * 100) : 0;
    log(`Progress: ${progress.completed + progress.failed}/${progress.total} (${successRate}% success) - ${progress.completed} passed, ${progress.failed} failed, ${progress.claimed} in progress`, 'progress');
  }, 10000); // Every 10 seconds

  // Wait for all workers to complete
  await Promise.all(workers.map(w => new Promise((resolve) => {
    w.process.on('exit', (code) => {
      log(`Worker ${w.id} exited with code ${code}`, code === 0 ? 'success' : 'error');
      resolve();
    });
  })));

  clearInterval(progressInterval);

  // Get final results
  const finalProgress = await workQueue.getProgress();
  completedTests = finalProgress.completed;
  failedTests = finalProgress.failed;

  logProgress('Port validation completed', {
    phase: 'script-complete',
    totalTests,
    completedTests,
    failedTests,
    successRate: totalTests > 0 ? Math.round((completedTests / totalTests) * 100) : 0,
    testResults: {
      passed: completedTests,
      failed: failedTests,
      total: totalTests,
      successRate: totalTests > 0 ? Math.round((completedTests / totalTests) * 100) : 0
    }
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

    log(`Worker ${WORKER_ID}: Processing ${basename(work.test)}`, 'info');

    try {
      const success = await runSingleValidation(work.test);
      await workQueue.completeWork(work.id, success);

      if (success) {
        log(`Worker ${WORKER_ID}: Validated ${basename(work.test)} successfully`, 'success');
      } else {
        log(`Worker ${WORKER_ID}: Failed validation of ${basename(work.test)}`, 'error');
      }
    } catch (err) {
      log(`Worker ${WORKER_ID}: Error processing ${basename(work.test)}: ${err.message}`, 'error');
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
  const workerCount = workerCountArg ? parseInt(workerCountArg.split('=')[1]) : DEFAULT_WORKERS;

  if (showHelp && !IS_WORKER) {
    console.log(`\nRSLint Port Validation Tool\n\nValidates that Go implementations correctly port TypeScript-ESLint rule logic.\n\nUsage: node automate-validate-check.js [options]\n\nOptions:\n  --concurrent      Run validations in parallel using multiple Claude instances\n  --workers=N       Number of parallel workers (default: ${DEFAULT_WORKERS})\n  --help, -h        Show this help message\n\nExamples:\n  node automate-validate-check.js                    # Sequential validation\n  node automate-validate-check.js --concurrent       # Parallel with ${DEFAULT_WORKERS} workers\n  node automate-validate-check.js --concurrent --workers=8  # Parallel with 8 workers\n`);
    process.exit(0);
  }

  if (!IS_WORKER) {
    console.clear();
    console.log(`${colors.bright}${colors.cyan}╔═══════════════════════════════════════════════════════════╗${colors.reset}`);
    console.log(`${colors.bright}${colors.cyan}║          RSLint Port Validation Tool                      ║${colors.reset}`);
    console.log(`${colors.bright}${colors.cyan}╚═══════════════════════════════════════════════════════════╝${colors.reset}\n`);

    log(`Script started (PID: ${process.pid}, Node: ${process.version})`, 'info');
    log(`Validating TypeScript->Go port correctness`, 'info');

    if (concurrentMode) {
      log(`Running in concurrent mode with ${workerCount} workers`, 'info');
    }
  } else {
    log(`Worker ${WORKER_ID} started (PID: ${process.pid})`, 'info');
  }

  // Run port validation
  if (!IS_WORKER) {
    console.log(`\n${colors.bright}=== PORT VALIDATION PHASE ===${colors.reset}`);
  }
  await runAllTests(concurrentMode, workerCount);

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
        successRate: totalTests > 0 ? Math.round((completedTests / totalTests) * 100) : 0
      }
    });

    return {
      success: failedTests === 0,
      totalTests,
      completedTests,
      failedTests,
      totalDuration
    };
  }
}

async function main() {
  const TOTAL_RUNS = 20;

  if (!IS_WORKER) {
    console.clear();
    console.log(`${colors.bright}${colors.magenta}╔═══════════════════════════════════════════════════════════╗${colors.reset}`);
    console.log(`${colors.bright}${colors.magenta}║          RSLint Port Validation Runner                    ║${colors.reset}`);
    console.log(`${colors.bright}${colors.magenta}╚═══════════════════════════════════════════════════════════╝${colors.reset}\n`);

    log(`Starting ${TOTAL_RUNS} consecutive validation runs`, 'info');
    console.log('='.repeat(80));
  }

  let allRunsResults = [];

  for (let runNumber = 1; runNumber <= TOTAL_RUNS; runNumber++) {
    if (!IS_WORKER) {
      console.log(`\n${colors.bright}${colors.yellow}╔═════════════════════════════════════════════════════════════════════════════╗${colors.reset}`);
      console.log(`${colors.bright}${colors.yellow}║                              RUN ${runNumber.toString().padStart(2)} OF ${TOTAL_RUNS}                                ║${colors.reset}`);
      console.log(`${colors.bright}${colors.yellow}╚═════════════════════════════════════════════════════════════════════════════╝${colors.reset}\n`);

      // Reset counters for each run
      totalTests = 0;
      completedTests = 0;
      failedTests = 0;
    }

    const runResult = await runCompleteProcess();

    if (!IS_WORKER) {
      allRunsResults.push({
        runNumber,
        ...runResult
      });

      console.log(`\n${colors.bright}${colors.cyan}RUN ${runNumber} COMPLETE:${colors.reset}`);
      console.log(`  • Tests: ${runResult.completedTests}/${runResult.totalTests} passed`);
      console.log(`  • Duration: ${Math.round(runResult.totalDuration / 60000)} minutes`);
      console.log(`  • Status: ${runResult.success ? colors.green + 'SUCCESS' + colors.reset : colors.red + 'FAILED' + colors.reset}`);

      // Brief pause between runs
      if (runNumber < TOTAL_RUNS) {
        log('Pausing 5 seconds before next run...', 'info');
        await new Promise(resolve => setTimeout(resolve, 5000));
      }
    }
  }

  if (!IS_WORKER) {
    // Final summary
    console.log(`\n\n${colors.bright}${colors.magenta}╔═════════════════════════════════════════════════════════════════════════════╗${colors.reset}`);
    console.log(`${colors.bright}${colors.magenta}║                            FINAL SUMMARY (${TOTAL_RUNS} RUNS)                             ║${colors.reset}`);
    console.log(`${colors.bright}${colors.magenta}╚═════════════════════════════════════════════════════════════════════════════╝${colors.reset}\n`);

    const totalSuccessfulRuns = allRunsResults.filter(r => r.success).length;
    const totalFailedRuns = TOTAL_RUNS - totalSuccessfulRuns;
    const averageDuration = allRunsResults.reduce((sum, r) => sum + r.totalDuration, 0) / TOTAL_RUNS;

    console.log(`${colors.bright}OVERALL RESULTS:${colors.reset}`);
    console.log(`  • Successful runs: ${colors.green}${totalSuccessfulRuns}/${TOTAL_RUNS}${colors.reset}`);
    console.log(`  • Failed runs: ${colors.red}${totalFailedRuns}/${TOTAL_RUNS}${colors.reset}`);
    console.log(`  • Success rate: ${totalSuccessfulRuns === TOTAL_RUNS ? colors.green : colors.yellow}${Math.round((totalSuccessfulRuns / TOTAL_RUNS) * 100)}%${colors.reset}`);
    console.log(`  • Average duration: ${Math.round(averageDuration / 60000)} minutes`);

    console.log(`\n${colors.bright}DETAILED RESULTS:${colors.reset}`);
    allRunsResults.forEach(result => {
      const status = result.success ? colors.green + '✓' + colors.reset : colors.red + '✗' + colors.reset;
      const duration = Math.round(result.totalDuration / 60000);
      const testSummary = `${result.completedTests}/${result.totalTests}`;
      console.log(`  Run ${result.runNumber.toString().padStart(2)}: ${status} ${testSummary.padEnd(8)} (${duration}m)`);
    });

    console.log(`\n${colors.bright}${colors.magenta}10x automation run completed!${colors.reset}`);
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
