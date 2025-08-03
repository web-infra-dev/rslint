#!/usr/bin/env node

// RSLint Automated Rule Porter - Improved Version
//
// This script automatically ports missing TypeScript-ESLint rules to Go for RSLint.
// Improvements based on learnings from the autoporter branch experience:
// 1. Better prompts with specific patterns and common pitfalls to avoid
// 2. Enhanced error detection and recovery
// 3. Improved test verification with rstest framework support
// 4. Better handling of rule registration and configuration
// 5. More comprehensive context for Claude about RSLint patterns
// 6. Automatic download of missing test files from typescript-eslint repo
//
// It uses Claude CLI to:
// 1. Discover missing rules by comparing TypeScript-ESLint's rules with existing RSLint rules
// 2. Download rule source and test files from the TypeScript-ESLint repository
// 3. Port each rule from TypeScript to Go following RSLint patterns
// 4. Transform and adapt test files to work with RSLint's test framework
// 5. Register new rules in the appropriate configuration files
//
// Usage: node automated-port.js [--concurrent] [--workers=N] [--list] [--status]

const { spawn } = require('child_process');
const {
  readdir,
  readFile,
  writeFile,
  mkdir,
  unlink,
  rm,
  chmod,
} = require('fs/promises');
const { join } = require('path');
const https = require('https');
const { randomBytes } = require('crypto');
const os = require('os');

// Configuration
const MAX_PORT_ATTEMPTS = 3;
const GITHUB_RAW_BASE =
  'https://raw.githubusercontent.com/typescript-eslint/typescript-eslint/main';
const RULES_INDEX_URL = `${GITHUB_RAW_BASE}/packages/eslint-plugin/src/rules/index.ts`;

// Concurrent execution configuration
const WORK_QUEUE_DIR = join(os.tmpdir(), 'rslint-port-automation');
const WORKER_ID = process.env.RSLINT_WORKER_ID || null;
const IS_WORKER = !!WORKER_ID;
const DEFAULT_WORKERS = 3;

// Progress tracking
let totalRules = 0;
let completedRules = 0;
let failedRules = 0;

// Known issues and patterns from our experience
const KNOWN_ISSUES = {
  registration: [
    'dot_notation',
    'explicit_function_return_type',
    'method_signature_style',
    'no_type_alias',
    'no_var_requires',
  ],
  missingTests: [
    'method_signature_style',
    'explicit_function_return_type',
    'no_var_requires',
  ],
  commonPitfalls: {
    astNodeTypes:
      'Must import AST_NODE_TYPES from @typescript-eslint/utils in tests',
    testFramework: 'Tests must use rstest framework with describe/test blocks',
    registration:
      'Rules must be registered with both namespaced and non-namespaced names',
    todoPattern: 'Use TODO(port) for incomplete features that need attention',
    messageIds:
      'Ensure messageId is included in diagnostics for test compatibility',
  },
};

// Work queue management (reused from automate-build-test.js)
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
          rule: items[i],
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

// Colors for terminal output (reused from automate-build-test.js)
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
    case 'porter':
      prefix = '🔄';
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

// Fetch available rules from TypeScript-ESLint
async function fetchAvailableRules() {
  return new Promise((resolve, reject) => {
    https
      .get(RULES_INDEX_URL, res => {
        let data = '';
        res.on('data', chunk => {
          data += chunk;
        });
        res.on('end', () => {
          if (res.statusCode === 200) {
            try {
              // Extract rule names from export statements
              const rulePattern = /['"]([a-z-]+)['"]\s*:\s*\w+,?/g;
              const rules = [];
              let match;

              while ((match = rulePattern.exec(data)) !== null) {
                rules.push(match[1]);
              }

              resolve(rules.sort());
            } catch (err) {
              reject(new Error(`Failed to parse rules index: ${err.message}`));
            }
          } else {
            reject(
              new Error(`HTTP ${res.statusCode}: Failed to fetch rules index`),
            );
          }
        });
      })
      .on('error', err => {
        reject(new Error(`Network error: ${err.message}`));
      });
  });
}

// Get existing ported rules from internal/rules directory
async function getExistingRules() {
  try {
    const rulesPath = join(__dirname, 'internal', 'rules');
    const entries = await readdir(rulesPath, { withFileTypes: true });

    const rules = [];
    for (const entry of entries) {
      if (entry.isDirectory() && entry.name !== 'fixtures') {
        // Convert underscore back to hyphen for consistency
        const ruleName = entry.name.replace(/_/g, '-');
        rules.push(ruleName);
      }
    }

    return rules.sort();
  } catch (error) {
    log(`Failed to read existing rules: ${error.message}`, 'error');
    return [];
  }
}

// Find missing rules that need to be ported
async function findMissingRules() {
  log('Fetching available TypeScript-ESLint rules...', 'info');
  const availableRules = await fetchAvailableRules();

  log('Checking existing RSLint rules...', 'info');
  const existingRules = await getExistingRules();
  const existingSet = new Set(existingRules);

  const missingRules = availableRules.filter(rule => !existingSet.has(rule));

  log(`Found ${availableRules.length} total TypeScript-ESLint rules`, 'info');
  log(`Found ${existingRules.length} existing RSLint rules`, 'info');
  log(`Identified ${missingRules.length} missing rules to port`, 'info');

  return missingRules;
}

// Run command helper
async function runCommand(command, args, options = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      stdio: 'pipe',
      cwd: options.cwd || __dirname,
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

// Claude CLI integration for rule porting (similar to automate-build-test.js)
async function runClaudePortingWithStreaming(prompt) {
  return new Promise(resolve => {
    // Use same flags as automate-build-test.js
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
      '--settings',
      settingsFile,
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

          // Handle different types of streaming events
          if (json.type === 'system' && json.subtype === 'init') {
            log(`Claude initialized with model: ${json.model}`, 'porter');
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
                      `${colors.magenta}🔄 Claude: ${line}${colors.reset}\n`,
                    );
                  }
                }
              } else if (content.type === 'tool_use') {
                // Display tool usage
                log(`Using tool: ${content.name}`, 'porter');
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

    // Set timeout for 10 minutes (longer for porting)
    const timeout = setTimeout(() => {
      child.kill('SIGKILL');
      log('Claude CLI timeout after 10 minutes', 'error');
      resolve({
        code: -1,
        stdout: fullOutput,
        stderr: 'Process timed out after 10 minutes',
      });
    }, 600000); // 10 minutes

    // Clear timeout on close
    child.on('close', () => clearTimeout(timeout));

    // Write prompt to stdin
    child.stdin.write(prompt);
    child.stdin.end();
  });
}

// Fetch content from GitHub URLs
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
          } else if (res.statusCode === 404) {
            resolve(null); // File doesn't exist
          } else {
            reject(new Error(`HTTP ${res.statusCode}: Failed to fetch ${url}`));
          }
        });
      })
      .on('error', err => {
        reject(new Error(`Network error fetching ${url}: ${err.message}`));
      });
  });
}

// Download missing test files
async function downloadMissingTests() {
  log('Checking for missing test files...', 'info');

  const testDir = join(
    __dirname,
    'packages/rslint-test-tools/tests/typescript-eslint/rules',
  );

  // Get all existing rules from internal/rules
  const existingRules = await getExistingRules();

  let downloadedCount = 0;
  let skippedCount = 0;
  let failedCount = 0;

  for (const rule of existingRules) {
    const testFile = join(testDir, `${rule}.test.ts`);

    try {
      // Check if test file already exists
      await readFile(testFile, 'utf8');
      skippedCount++;
    } catch (err) {
      if (err.code === 'ENOENT') {
        // Test file doesn't exist, download it
        log(`Downloading missing test for ${rule}...`, 'porter');

        const testUrl = `${GITHUB_RAW_BASE}/packages/eslint-plugin/tests/rules/${rule}.test.ts`;
        const testContent = await fetchFromGitHub(testUrl);

        if (testContent) {
          // Transform the test content to use rstest format
          let transformedContent = testContent;

          // Add rstest imports if not present
          if (!transformedContent.includes('@rstest/core')) {
            transformedContent = `import { describe, test, expect } from '@rstest/core';
${transformedContent}`;
          }

          // Wrap RuleTester.run in describe/test blocks if needed
          if (
            !transformedContent.includes('describe(') &&
            transformedContent.includes('ruleTester.run(')
          ) {
            transformedContent = transformedContent.replace(
              /ruleTester\.run\(['"]([^'"]+)['"],/g,
              `describe('$1', () => {
  test('rule tests', () => {
    ruleTester.run('$1',`,
            );

            // Find the end of the ruleTester.run call and close the blocks
            transformedContent = transformedContent.replace(
              /\);\s*$/m,
              `);
  });
});`,
            );
          }

          await writeFile(testFile, transformedContent);
          log(`✓ Downloaded test for ${rule}`, 'success');
          downloadedCount++;
        } else {
          log(
            `✗ No test found for ${rule} in TypeScript-ESLint repo`,
            'warning',
          );
          failedCount++;
        }
      } else {
        log(`Error checking test for ${rule}: ${err.message}`, 'error');
        failedCount++;
      }
    }
  }

  log(
    `Test file check complete: ${downloadedCount} downloaded, ${skippedCount} already existed, ${failedCount} failed`,
    'info',
  );

  return { downloadedCount, skippedCount, failedCount };
}

// Fetch original TypeScript-ESLint rule sources
async function fetchOriginalRule(ruleName) {
  const ruleUrl = `${GITHUB_RAW_BASE}/packages/eslint-plugin/src/rules/${ruleName}.ts`;
  const ruleContent = await fetchFromGitHub(ruleUrl);

  const testUrl = `${GITHUB_RAW_BASE}/packages/eslint-plugin/tests/rules/${ruleName}.test.ts`;
  const testContent = await fetchFromGitHub(testUrl);

  return {
    ruleName,
    ruleContent,
    testContent,
  };
}

// Verify rule registration
async function verifyRuleRegistration(ruleName) {
  try {
    const configPath = join(__dirname, 'internal', 'config', 'config.go');
    const configContent = await readFile(configPath, 'utf8');

    const snakeCaseRule = ruleName.replace(/-/g, '_');
    const namespacedName = `@typescript-eslint/${ruleName}`;

    // Check for both registrations
    const hasNamespaced = configContent.includes(`"${namespacedName}"`);
    const hasNonNamespaced = configContent.includes(`"${ruleName}"`);
    const hasImport = configContent.includes(`${snakeCaseRule}.Rule`);

    return {
      isRegistered: hasNamespaced && hasNonNamespaced && hasImport,
      hasNamespaced,
      hasNonNamespaced,
      hasImport,
    };
  } catch (error) {
    log(
      `Failed to verify registration for ${ruleName}: ${error.message}`,
      'error',
    );
    return { isRegistered: false };
  }
}

// Run Go build to verify compilation
async function runGoBuild() {
  try {
    log('Running go build...', 'info');
    const buildResult = await runCommand('go', ['build', './cmd/rslint'], {
      timeout: 60000,
      cwd: __dirname,
    });

    if (buildResult.code === 0) {
      log('✓ Go build successful', 'success');
      return true;
    } else {
      log(`✗ Go build failed (exit code ${buildResult.code})`, 'error');
      if (buildResult.stderr) {
        console.log(`${colors.red}Build errors:${colors.reset}`);
        console.log(buildResult.stderr);
      }
      return false;
    }
  } catch (error) {
    log(`Go build error: ${error.message}`, 'error');
    return false;
  }
}

// Run Go test for a specific rule
async function runGoTest(ruleName) {
  try {
    const ruleDir = ruleName.replace(/-/g, '_');
    log(`Running Go test for ${ruleName}...`, 'info');

    const testResult = await runCommand(
      'go',
      ['test', '-v', `./internal/rules/${ruleDir}`],
      {
        timeout: 120000,
        cwd: __dirname,
      },
    );

    if (testResult.code === 0) {
      log(`✓ Go test passed for ${ruleName}`, 'success');
      return { success: true, output: testResult.stdout };
    } else {
      log(
        `✗ Go test failed for ${ruleName} (exit code ${testResult.code})`,
        'error',
      );
      return {
        success: false,
        output: testResult.stdout,
        error: testResult.stderr,
      };
    }
  } catch (error) {
    log(`Go test error for ${ruleName}: ${error.message}`, 'error');
    return { success: false, error: error.message };
  }
}

// Run TypeScript test with rstest
async function runTypeScriptTest(ruleName) {
  try {
    log(`Running TypeScript test for ${ruleName}...`, 'info');

    // Use rstest instead of node --test
    const testResult = await runCommand(
      'npx',
      [
        'rstest',
        'run',
        `tests/typescript-eslint/rules/${ruleName}.test.ts`,
        '--testTimeout=0',
      ],
      {
        timeout: 120000,
        cwd: join(__dirname, 'packages/rslint-test-tools'),
      },
    );

    if (testResult.code === 0) {
      log(`✓ TypeScript test passed for ${ruleName}`, 'success');
      return { success: true, output: testResult.stdout };
    } else {
      log(
        `✗ TypeScript test failed for ${ruleName} (exit code ${testResult.code})`,
        'error',
      );
      return {
        success: false,
        output: testResult.stdout,
        error: testResult.stderr,
      };
    }
  } catch (error) {
    log(`TypeScript test error for ${ruleName}: ${error.message}`, 'error');
    return { success: false, error: error.message };
  }
}

// Create improved porting prompt with all learnings
function createImprovedPortingPrompt(ruleName, originalSources) {
  let prompt = `Go into plan mode first to analyze this TypeScript-ESLint rule and plan the porting to Go.\n\n`;

  prompt += `Task: Port the TypeScript-ESLint rule "${ruleName}" to Go for the RSLint project.\n\n`;

  // Add known issues context
  if (KNOWN_ISSUES.registration.includes(ruleName)) {
    prompt += `⚠️ IMPORTANT: This rule was previously missing registration. Make sure to register it properly.\n\n`;
  }
  if (KNOWN_ISSUES.missingTests.includes(ruleName)) {
    prompt += `⚠️ IMPORTANT: This rule was previously missing tests. Make sure to create comprehensive tests.\n\n`;
  }

  // Add critical learnings section
  prompt += `--- CRITICAL LEARNINGS FROM PREVIOUS PORTS ---

1. **Rule Registration (MUST DO)**:
   - Register in /Users/bytedance/dev/rslint/internal/config/config.go
   - Add BOTH namespaced and non-namespaced versions:
     GlobalRuleRegistry.Register("@typescript-eslint/${ruleName}", ${ruleName.replace(/-/g, '_')}.Rule)
     GlobalRuleRegistry.Register("${ruleName}", ${ruleName.replace(/-/g, '_')}.Rule)
   - Import the rule package at the top of config.go

2. **Test Framework Requirements**:
   - Tests MUST use rstest framework with describe/test blocks
   - Import pattern MUST be:
     import { describe, test, expect } from '@rstest/core';
     import { noFormat, RuleTester } from '@typescript-eslint/rule-tester';
     import { getFixturesRootDir } from '../RuleTester.ts';
   - If tests use AST_NODE_TYPES, import from '@typescript-eslint/utils'
   - Structure:
     describe('${ruleName}', () => {
       test('rule tests', () => {
         ruleTester.run('${ruleName}', { ... });
       });
     });

3. **Common Implementation Patterns**:
   - Use utils.GetNameFromMember() for property name extraction
   - Include messageId in diagnostics: ctx.ReportNode(node, RuleMessage{MessageId: "camelCaseId", Description: "..."})
   - For TODO items, use TODO(port) to mark incomplete features
   - Access class members via node.Members() which returns []*ast.Node directly

4. **Position/Range Handling**:
   - RSLint uses 1-based line and column numbers
   - IPC API converts between 0-based and 1-based automatically
   - Be careful about what part of node to highlight in errors

5. **Testing Best Practices**:
   - Ensure all test cases from original are preserved
   - Don't skip test cases - mark with TODO(port) if complex
   - Test file must compile with TypeScript
   - Download missing test files from typescript-eslint repo if needed

--- END CRITICAL LEARNINGS ---\n\n`;

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

  prompt += `After analyzing in plan mode, port this rule to Go following these steps:

1. **Create the Go rule**:
   - Directory: /Users/bytedance/dev/rslint/internal/rules/${ruleName.replace(/-/g, '_')}/
   - Rule file: ${ruleName.replace(/-/g, '_')}.go
   - Test file: ${ruleName.replace(/-/g, '_')}_test.go
   - Follow patterns from existing rules like array_type, ban_ts_comment, consistent_indexed_object_style

2. **Transform the test file**:
   - Save to: /Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/rules/${ruleName}.test.ts
   - MUST use rstest format with describe/test blocks (see learnings above)
   - Preserve ALL test cases from original

3. **Register the rule**:
   - Add to /Users/bytedance/dev/rslint/internal/config/config.go
   - Register with BOTH namespaced and non-namespaced names
   - Add import for the rule package

4. **Implementation Guidelines**:
   - Maintain exact same rule logic and behavior as TypeScript version
   - Use RSLint utility functions (utils package) for common operations
   - Include proper error messages with messageId
   - Mark incomplete features with TODO(port)
   - Don't simplify - implement the full rule functionality

IMPORTANT: 
- ONLY create and edit files - do NOT run any commands
- Do NOT skip test cases or simplify implementations
- Focus on complete, production-ready implementation
- This script will handle all testing and verification`;

  return prompt;
}

// Enhanced test failure fix prompt
function createImprovedFixPrompt(
  ruleName,
  goTestResult,
  tsTestResult,
  registrationStatus,
) {
  let prompt = `Fix the issues for the RSLint rule "${ruleName}". `;

  if (!registrationStatus.isRegistered) {
    prompt += `\n⚠️ CRITICAL: Rule is not properly registered! `;
    if (!registrationStatus.hasImport) {
      prompt += `Missing import in config.go. `;
    }
    if (
      !registrationStatus.hasNamespaced ||
      !registrationStatus.hasNonNamespaced
    ) {
      prompt += `Missing registration (needs both namespaced and non-namespaced). `;
    }
    prompt += `\n`;
  }

  prompt += `Here are the test results:\n\n`;

  if (!goTestResult.success) {
    prompt += `--- GO TEST FAILURE ---\n`;
    prompt += `Exit code: Non-zero\n`;
    if (goTestResult.output) {
      prompt += `Stdout:\n${goTestResult.output}\n`;
    }
    if (goTestResult.error) {
      prompt += `Stderr:\n${goTestResult.error}\n`;
    }
    prompt += `\n`;
  }

  if (!tsTestResult.success) {
    prompt += `--- TYPESCRIPT TEST FAILURE ---\n`;
    prompt += `Exit code: Non-zero\n`;
    if (tsTestResult.output) {
      // Check for common rstest issues
      if (tsTestResult.output.includes('No test suites found')) {
        prompt += `\n⚠️ Test file not using rstest format! Must wrap in describe/test blocks.\n`;
      }
      if (tsTestResult.output.includes('AST_NODE_TYPES is not defined')) {
        prompt += `\n⚠️ Missing AST_NODE_TYPES import from '@typescript-eslint/utils'\n`;
      }
      prompt += `Stdout:\n${tsTestResult.output}\n`;
    }
    if (tsTestResult.error) {
      prompt += `Stderr:\n${tsTestResult.error}\n`;
    }
    prompt += `\n`;
  }

  prompt += `Please fix these issues. Common fixes needed:

1. **Registration Issues**:
   - Add import in config.go: import "${ruleName.replace(/-/g, '_')} github.com/web-infra-dev/rslint/internal/rules/${ruleName.replace(/-/g, '_')}"
   - Register with both names in config.go init()

2. **Test Framework Issues**:
   - Ensure test uses rstest format with describe/test blocks
   - Import AST_NODE_TYPES from '@typescript-eslint/utils' if needed
   - Check for duplicate variable declarations

3. **Implementation Issues**:
   - Ensure messageId is included in all diagnostics
   - Fix any compilation errors
   - Implement missing functionality marked with TODO(port)

Focus on the most critical errors first. Do not run any commands - just edit the files to fix the issues.`;

  return prompt;
}

// Port a single rule using Claude CLI with build and test verification
async function portSingleRule(ruleName, attemptNumber = 1) {
  const startTime = Date.now();

  log(
    `Porting ${ruleName} (attempt ${attemptNumber}/${MAX_PORT_ATTEMPTS})`,
    'porter',
  );

  // Fetch original TypeScript-ESLint sources for context
  let originalSources = null;

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
    // Re-fetch on subsequent attempts as context might be needed
    originalSources = await fetchOriginalRule(ruleName);
  }

  try {
    // Create improved porting prompt
    const prompt = createImprovedPortingPrompt(ruleName, originalSources);
    const result = await runClaudePortingWithStreaming(prompt);

    if (result.code !== 0) {
      log(
        `✗ Claude CLI failed for ${ruleName} (exit code ${result.code})`,
        'error',
      );

      if (result.stderr) {
        console.log(`${colors.red}Claude stderr:${colors.reset}`);
        console.log(result.stderr);
      }

      if (attemptNumber < MAX_PORT_ATTEMPTS) {
        log(
          `Retrying ${ruleName} (attempt ${attemptNumber + 1}/${MAX_PORT_ATTEMPTS})...`,
          'warning',
        );
        await new Promise(resolve => setTimeout(resolve, 10000));
        return await portSingleRule(ruleName, attemptNumber + 1);
      } else {
        log(
          `Failed to port ${ruleName} after ${attemptNumber} attempts`,
          'error',
        );
        failedRules++;
        return false;
      }
    }

    log(
      `✓ Rule porting completed for ${ruleName}, now verifying...`,
      'success',
    );

    // Verify registration
    const registrationStatus = await verifyRuleRegistration(ruleName);
    if (!registrationStatus.isRegistered) {
      log(`⚠️ Rule ${ruleName} is not properly registered!`, 'warning');
    }

    // Run build
    const buildResult = await runCommand('go', ['build', './cmd/rslint'], {
      timeout: 60000,
      cwd: __dirname,
    });

    const buildSuccess = buildResult.code === 0;
    if (!buildSuccess) {
      log(`Build failed after porting ${ruleName}`, 'error');
    }

    // Run Go test
    const ruleDir = ruleName.replace(/-/g, '_');
    const goTestResult = await runCommand(
      'go',
      ['test', '-v', `./internal/rules/${ruleDir}`],
      {
        timeout: 120000,
        cwd: __dirname,
      },
    );

    // Run TypeScript test with rstest
    const tsTestResult = await runTypeScriptTest(ruleName);

    // Check results
    if (
      buildSuccess &&
      goTestResult.code === 0 &&
      tsTestResult.success &&
      registrationStatus.isRegistered
    ) {
      const duration = Date.now() - startTime;
      log(
        `✓ Successfully ported ${ruleName} in ${Math.round(duration / 1000)}s`,
        'success',
      );
      completedRules++;
      return true;
    }

    // Attempt to fix issues
    log(`Issues detected for ${ruleName}, attempting fixes...`, 'warning');

    const fixPrompt = createImprovedFixPrompt(
      ruleName,
      {
        success: goTestResult.code === 0,
        output: goTestResult.stdout,
        error: goTestResult.stderr,
      },
      tsTestResult,
      registrationStatus,
    );

    const fixResult = await runClaudePortingWithStreaming(fixPrompt);

    if (fixResult.code === 0) {
      // Re-verify after fix
      const newRegistrationStatus = await verifyRuleRegistration(ruleName);
      const reBuildResult = await runCommand('go', ['build', './cmd/rslint'], {
        timeout: 60000,
        cwd: __dirname,
      });

      if (reBuildResult.code === 0) {
        const reGoTestResult = await runCommand(
          'go',
          ['test', '-v', `./internal/rules/${ruleDir}`],
          {
            timeout: 120000,
            cwd: __dirname,
          },
        );
        const reTsTestResult = await runTypeScriptTest(ruleName);

        if (
          reGoTestResult.code === 0 &&
          reTsTestResult.success &&
          newRegistrationStatus.isRegistered
        ) {
          const duration = Date.now() - startTime;
          log(
            `✓ Successfully ported and fixed ${ruleName} in ${Math.round(duration / 1000)}s`,
            'success',
          );
          completedRules++;
          return true;
        }
      }
    }

    // If we get here, tests failed and fix didn't work
    if (attemptNumber < MAX_PORT_ATTEMPTS) {
      log(
        `Retrying complete port for ${ruleName} (attempt ${attemptNumber + 1}/${MAX_PORT_ATTEMPTS})...`,
        'warning',
      );
      await new Promise(resolve => setTimeout(resolve, 10000));
      return await portSingleRule(ruleName, attemptNumber + 1);
    } else {
      log(
        `Failed to port ${ruleName} after ${attemptNumber} attempts with working tests`,
        'error',
      );
      failedRules++;
      return false;
    }
  } catch (error) {
    if (error.message.includes('timed out')) {
      log(`Rule ${ruleName} timed out after 10 minutes`, 'error');
    } else {
      log(`Error porting ${ruleName}: ${error.message}`, 'error');
    }

    if (attemptNumber < MAX_PORT_ATTEMPTS) {
      log(`Retrying ${ruleName} after error...`, 'warning');
      await new Promise(resolve => setTimeout(resolve, 10000));
      return await portSingleRule(ruleName, attemptNumber + 1);
    }

    failedRules++;
    return false;
  }
}

// Run all missing rules through the porter
async function portAllMissingRules(
  concurrentMode = false,
  workerCount = DEFAULT_WORKERS,
) {
  const missingRules = await findMissingRules();
  totalRules = missingRules.length;

  if (totalRules === 0) {
    log(
      'No missing rules found - all TypeScript-ESLint rules are already ported!',
      'success',
    );
    return;
  }

  if (!IS_WORKER) {
    console.log('\n' + '='.repeat(60));
    log(`Starting rule porting with ${totalRules} missing rules`, 'info');
    log(
      `Known issues to address: ${KNOWN_ISSUES.registration.length} unregistered, ${KNOWN_ISSUES.missingTests.length} missing tests`,
      'warning',
    );
    console.log('='.repeat(60));
  }

  if (concurrentMode && !IS_WORKER) {
    // Main process in concurrent mode
    await runConcurrentPorting(missingRules, workerCount);
  } else if (IS_WORKER) {
    // Worker process
    await runWorker();
  } else {
    // Sequential mode
    for (let i = 0; i < missingRules.length; i++) {
      const ruleName = missingRules[i];

      console.log(
        `\n${colors.bright}[${i + 1}/${totalRules}] ${ruleName}${colors.reset}`,
      );
      console.log('-'.repeat(40));

      await portSingleRule(ruleName);

      // Show running totals
      console.log(
        `\n${colors.dim}Progress: ${completedRules} ported, ${failedRules} failed, ${totalRules - completedRules - failedRules} remaining${colors.reset}`,
      );

      // Add delay between rules to avoid rate limiting
      if (i < missingRules.length - 1) {
        await new Promise(resolve => setTimeout(resolve, 3000));
      }
    }

    // Final summary
    console.log('\n' + '='.repeat(60));
    log('Rule Porting Complete', 'info');
    log(
      `Total: ${totalRules}, Ported: ${completedRules}, Failed: ${failedRules}`,
      'info',
    );
    log(
      `Success Rate: ${totalRules > 0 ? Math.round((completedRules / totalRules) * 100) : 0}%`,
      failedRules > 0 ? 'warning' : 'success',
    );
    console.log('='.repeat(60));
  }
}

// Concurrent processing implementation
async function runConcurrentPorting(missingRules, workerCount) {
  const workQueue = new WorkQueue(WORK_QUEUE_DIR);
  await workQueue.initialize();

  // Add all rules to work queue
  await workQueue.addWork(missingRules);
  log(`Added ${missingRules.length} rules to work queue`, 'info');

  // Create hooks directory and files for file locking
  await createHooks();

  // Start workers
  const workers = [];
  for (let i = 0; i < workerCount; i++) {
    const workerId = `porter_${i}_${randomBytes(4).toString('hex')}`;
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
      `Progress: ${progress.completed + progress.failed}/${progress.total} (${successRate}% success) - ${progress.completed} ported, ${progress.failed} failed, ${progress.claimed} in progress`,
      'progress',
    );
  }, 15000); // Every 15 seconds

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
  completedRules = finalProgress.completed;
  failedRules = finalProgress.failed;

  // Final summary
  console.log('\n' + '='.repeat(60));
  log('Concurrent Rule Porting Complete', 'info');
  log(
    `Total: ${totalRules}, Ported: ${completedRules}, Failed: ${failedRules}`,
    'info',
  );
  log(
    `Success Rate: ${totalRules > 0 ? Math.round((completedRules / totalRules) * 100) : 0}%`,
    failedRules > 0 ? 'warning' : 'success',
  );
  console.log('='.repeat(60));

  // Cleanup
  await workQueue.cleanup();
}

// Worker process implementation
async function runWorker() {
  const workQueueDir = process.env.RSLINT_WORK_QUEUE_DIR || WORK_QUEUE_DIR;
  const workQueue = new WorkQueue(workQueueDir);

  while (true) {
    const work = await workQueue.claimWork(WORKER_ID);

    if (!work) {
      log(`Worker ${WORKER_ID}: No more work available, exiting`, 'info');
      break;
    }

    log(`Worker ${WORKER_ID}: Processing ${work.rule}`, 'info');

    try {
      const success = await portSingleRule(work.rule);
      await workQueue.completeWork(work.id, success);

      if (success) {
        log(
          `Worker ${WORKER_ID}: Completed ${work.rule} successfully`,
          'success',
        );
      } else {
        log(`Worker ${WORKER_ID}: Failed ${work.rule}`, 'error');
      }

      // Add delay between rules to avoid rate limiting
      await new Promise(resolve => setTimeout(resolve, 3000));
    } catch (err) {
      log(
        `Worker ${WORKER_ID}: Error processing ${work.rule}: ${err.message}`,
        'error',
      );
      await workQueue.completeWork(work.id, false);
    }
  }

  process.exit(0);
}

// Create hooks for file locking (reused from automate-build-test.js)
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

async function main() {
  const scriptStartTime = Date.now();

  // Parse command line arguments
  const args = process.argv.slice(2);
  const showHelp = args.includes('--help') || args.includes('-h');
  const concurrentMode = args.includes('--concurrent');
  const workerCountArg = args.find(arg => arg.startsWith('--workers='));
  const workerCount = workerCountArg
    ? parseInt(workerCountArg.split('=')[1])
    : DEFAULT_WORKERS;
  const listOnly = args.includes('--list');
  const statusOnly = args.includes('--status');

  if (showHelp && !IS_WORKER) {
    console.log(
      `\nRSLint Automated Rule Porter\n\nUsage: node automated-port.js [options]\n\nOptions:\n  --concurrent      Port rules in parallel using multiple porter instances\n  --workers=N       Number of parallel workers (default: ${DEFAULT_WORKERS})\n  --list            List missing rules only (no porting)\n  --status          Show porting status\n  --help, -h        Show this help message\n\nExamples:\n  node automated-port.js                      # Sequential porting\n  node automated-port.js --concurrent         # Parallel with ${DEFAULT_WORKERS} workers\n  node automated-port.js --concurrent --workers=5  # Parallel with 5 workers\n  node automated-port.js --list               # Just list missing rules\n  node automated-port.js --status             # Show current status\n`,
    );
    process.exit(0);
  }

  if (!IS_WORKER) {
    console.clear();
    console.log(
      `${colors.bright}${colors.cyan}╔═══════════════════════════════════════════════════════════╗${colors.reset}`,
    );
    console.log(
      `${colors.bright}${colors.cyan}║             RSLint Automated Rule Porter                  ║${colors.reset}`,
    );
    console.log(
      `${colors.bright}${colors.cyan}╚═══════════════════════════════════════════════════════════╝${colors.reset}\n`,
    );

    log(
      `Script started (PID: ${process.pid}, Node: ${process.version})`,
      'info',
    );

    if (concurrentMode) {
      log(`Running in concurrent mode with ${workerCount} workers`, 'info');
    }
  } else {
    log(`Worker ${WORKER_ID} started (PID: ${process.pid})`, 'info');
  }

  try {
    if (listOnly && !IS_WORKER) {
      // Just list missing rules
      const missingRules = await findMissingRules();
      if (missingRules.length > 0) {
        console.log(`\n${colors.yellow}Missing rules to port:${colors.reset}`);
        missingRules.forEach((rule, i) => {
          const marker = KNOWN_ISSUES.registration.includes(rule) ? ' ⚠️' : '';
          console.log(
            `${colors.gray}  ${i + 1}. ${rule}${marker}${colors.reset}`,
          );
        });
      } else {
        console.log(
          `\n${colors.green}✓ All TypeScript-ESLint rules are already ported!${colors.reset}`,
        );
      }
      return;
    }

    if (statusOnly && !IS_WORKER) {
      // Show status
      const availableRules = await fetchAvailableRules();
      const existingRules = await getExistingRules();
      const missingRules = availableRules.filter(
        rule => !new Set(existingRules).has(rule),
      );

      console.log(`\n${colors.blue}=== Porting Status ===${colors.reset}`);
      console.log(
        `${colors.green}✓ Ported: ${existingRules.length}/${availableRules.length} (${Math.round((existingRules.length / availableRules.length) * 100)}%)${colors.reset}`,
      );
      console.log(
        `${colors.yellow}⚠ Remaining: ${missingRules.length}${colors.reset}`,
      );
      console.log(
        `${colors.red}⚠ Known unregistered: ${KNOWN_ISSUES.registration.length}${colors.reset}`,
      );
      console.log(
        `${colors.red}⚠ Known missing tests: ${KNOWN_ISSUES.missingTests.length}${colors.reset}`,
      );

      if (missingRules.length > 0) {
        console.log(`\n${colors.blue}Next rules to port:${colors.reset}`);
        missingRules.slice(0, 10).forEach((rule, i) => {
          console.log(`${colors.gray}  ${i + 1}. ${rule}${colors.reset}`);
        });
        if (missingRules.length > 10) {
          console.log(
            `${colors.gray}  ... and ${missingRules.length - 10} more${colors.reset}`,
          );
        }
      }
      return;
    }

    // Download any missing test files first
    if (!IS_WORKER) {
      await downloadMissingTests();
    }

    // Main porting process
    await portAllMissingRules(concurrentMode, workerCount);

    if (!IS_WORKER) {
      const totalDuration = Date.now() - scriptStartTime;
      log(
        `Automation completed in ${Math.round(totalDuration / 60000)} minutes`,
        'info',
      );

      process.exit(failedRules > 0 ? 1 : 0);
    }
  } catch (error) {
    log(`Script failed with unhandled error: ${error.message}`, 'error');
    console.error(error.stack);
    process.exit(1);
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
