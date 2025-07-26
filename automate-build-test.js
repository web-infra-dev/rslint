#!/usr/bin/env node

const { spawn } = require('child_process');
const { readdir, readFile } = require('fs/promises');
const { join, basename } = require('path');
const https = require('https');

// __dirname is available in CommonJS

// Configuration
const BUILD_COMMAND = 'pnpm';
const BUILD_ARGS = ['-r', 'build'];
const TEST_TIMEOUT = 120000; // 120 seconds (2 minutes)
const TEST_DIR = 'packages/rslint-test-tools/tests/typescript-eslint/rules';
const TSLINT_BASE_URL = 'https://raw.githubusercontent.com/typescript-eslint/typescript-eslint/main';

// Progress tracking
let totalTests = 0;
let completedTests = 0;
let failedTests = 0;

function logProgress(message, data = {}) {
  const progressData = {
    timestamp: new Date().toISOString(),
    message,
    progress: {
      total: totalTests,
      completed: completedTests,
      failed: failedTests,
      percentage: totalTests > 0 ? Math.round((completedTests / totalTests) * 100) : 0
    },
    ...data
  };
  console.log(JSON.stringify(progressData));
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
      logProgress('GitHub fetch error', { url, error: err.message });
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
    const child = spawn('claude', [prompt], {
      stdio: ['pipe', 'pipe', 'pipe']
    });

    let fullOutput = '';
    let fullError = '';
    let lastChunk = '';

    // Process stdout stream
    child.stdout.on('data', (data) => {
      const chunk = data.toString();
      fullOutput += chunk;
      lastChunk += chunk;
      
      // Try to parse and display JSON chunks
      const lines = lastChunk.split('\n');
      lastChunk = lines.pop() || ''; // Keep incomplete line for next chunk
      
      for (const line of lines) {
        if (line.trim()) {
          try {
            const json = JSON.parse(line);
            if (json.type === 'message' && json.content) {
              logProgress('Claude response chunk', { 
                type: json.type,
                content: json.content.slice(0, 200) + (json.content.length > 200 ? '...' : '')
              });
            }
          } catch (e) {
            // Not JSON, just log as text
            if (line.length > 0) {
              logProgress('Claude output', { text: line.slice(0, 200) });
            }
          }
        }
      }
    });

    child.stderr.on('data', (data) => {
      fullError += data.toString();
    });

    child.on('close', (code) => {
      // Process any remaining chunk
      if (lastChunk.trim()) {
        try {
          const json = JSON.parse(lastChunk);
          if (json.type === 'message' && json.content) {
            logProgress('Claude final response chunk', { content: json.content.slice(0, 200) });
          }
        } catch (e) {
          // Not JSON
        }
      }
      
      resolve({
        code,
        stdout: fullOutput,
        stderr: fullError
      });
    });

    // Set timeout
    setTimeout(() => {
      child.kill('SIGKILL');
      resolve({
        code: -1,
        stdout: fullOutput,
        stderr: 'Process timed out after 60 seconds'
      });
    }, 60000);
  });
}

async function fixErrorWithClaudeCLI(errorOutput, command, originalSources = null, currentTestContent = null) {
  let prompt = `Error occurred:\n\n${errorOutput}\n\n`;
  
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
  
  prompt += `Fix the error above, then run: ${command}`;
  
  logProgress('Sending error to Claude CLI for fixing', {
    phase: 'start',
    command,
    hasOriginalSources: !!originalSources,
    hasCurrentTest: !!currentTestContent,
    errorPreview: errorOutput.slice(0, 500) + (errorOutput.length > 500 ? '...' : '')
  });

  try {
    const result = await runClaudeWithStreaming(prompt);
    
    logProgress('Claude CLI completed', {
      phase: 'complete',
      success: result.code === 0,
      exitCode: result.code
    });
    
    return result.code === 0;
  } catch (error) {
    logProgress('Claude CLI error', { 
      phase: 'error',
      error: error.message 
    });
    return false;
  }
}

async function runBuild() {
  logProgress('Starting build process');
  
  try {
    const result = await runCommand(BUILD_COMMAND, BUILD_ARGS, { timeout: 120000 });
    
    if (result.code === 0) {
      logProgress('Build successful');
      return true;
    } else {
      logProgress('Build failed', {
        exitCode: result.code,
        stderr: result.stderr,
        stdout: result.stdout
      });

      const buildCommand = `${BUILD_COMMAND} ${BUILD_ARGS.join(' ')}`;
      const errorOutput = result.stderr || result.stdout;
      
      // Try to fix with Claude CLI
      const fixed = await fixErrorWithClaudeCLI(errorOutput, buildCommand);
      
      if (fixed) {
        // Retry build after fix
        logProgress('Retrying build after Claude CLI fix');
        return await runBuild();
      } else {
        logProgress('Failed to fix build error with Claude CLI');
        return false;
      }
    }
  } catch (error) {
    logProgress('Build process error', { error: error.message });
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
    logProgress('Error reading test directory', { error: error.message });
    return [];
  }
}

async function runSingleTest(testFile) {
  const testName = basename(testFile);
  const startTime = Date.now();
  
  logProgress('Test execution started', { 
    phase: 'test-start',
    testName,
    testFile 
  });

  // Fetch original TypeScript ESLint sources
  logProgress('Fetching original sources', { 
    phase: 'fetch-start',
    testName 
  });
  
  const originalSources = await fetchOriginalRule(testName);
  if (originalSources.ruleContent || originalSources.testContent) {
    logProgress('Original sources fetched', {
      phase: 'fetch-complete',
      testName,
      hasRule: !!originalSources.ruleContent,
      hasTest: !!originalSources.testContent,
      ruleSizeBytes: originalSources.ruleContent?.length || 0,
      testSizeBytes: originalSources.testContent?.length || 0
    });
  }

  // Also read the current RSLint test file
  let currentTestContent = null;
  try {
    currentTestContent = await readFile(testFile, 'utf8');
    logProgress('Current test file read', {
      phase: 'read-test',
      testName,
      sizeBytes: currentTestContent.length
    });
  } catch (err) {
    logProgress('Failed to read current test file', { 
      phase: 'read-test-error',
      testName,
      error: err.message 
    });
  }

  try {
    const result = await runCommand('node', [
      '--import=tsx/esm',
      '--test',
      testFile
    ], { 
      timeout: TEST_TIMEOUT,
      cwd: join(__dirname, 'packages/rslint-test-tools')
    });

    if (result.code === 0) {
      const duration = Date.now() - startTime;
      logProgress('Test passed', {
        phase: 'test-pass',
        testName,
        durationMs: duration
      });
      completedTests++;
      return true;
    } else {
      logProgress('Test failed', {
        phase: 'test-fail',
        testName,
        exitCode: result.code,
        stderr: result.stderr?.slice(0, 1000),
        stdout: result.stdout?.slice(0, 1000)
      });

      const testCommand = `node --import=tsx/esm --test ${testFile}`;
      const errorOutput = result.stderr || result.stdout;
      
      // Try to fix with Claude CLI, including original sources and current test
      const fixed = await fixErrorWithClaudeCLI(errorOutput, testCommand, originalSources, currentTestContent);
      
      if (fixed) {
        // Retry test after fix
        logProgress('Retrying test after fix', {
          phase: 'test-retry',
          testName
        });
        return await runSingleTest(testFile);
      } else {
        logProgress('Failed to fix test error', {
          phase: 'test-fix-failed',
          testName
        });
        failedTests++;
        return false;
      }
    }
  } catch (error) {
    if (error.message.includes('timed out')) {
      logProgress('Test timed out', { 
        phase: 'test-timeout',
        testName,
        timeoutMs: TEST_TIMEOUT,
        error: error.message 
      });
      
      const testCommand = `node --import=tsx/esm --test ${testFile}`;
      const timeoutError = `Test timed out after ${TEST_TIMEOUT}ms`;
      
      const fixed = await fixErrorWithClaudeCLI(timeoutError, testCommand, originalSources, currentTestContent);
      
      if (fixed) {
        logProgress('Retrying test after timeout fix', {
          phase: 'test-retry-timeout',
          testName
        });
        return await runSingleTest(testFile);
      }
    } else {
      logProgress('Test error', { 
        phase: 'test-error',
        testName,
        error: error.message 
      });
    }
    
    failedTests++;
    return false;
  }
}

async function runAllTests() {
  const testFiles = await getTestFiles();
  totalTests = testFiles.length;
  
  logProgress('Starting test suite', { 
    phase: 'start',
    totalTests,
    testFiles: testFiles.map(f => basename(f))
  });

  for (let i = 0; i < testFiles.length; i++) {
    const testFile = testFiles[i];
    logProgress('Test progress', {
      phase: 'progress',
      current: i + 1,
      total: totalTests,
      currentTest: basename(testFile),
      completed: completedTests,
      failed: failedTests
    });
    
    await runSingleTest(testFile);
  }

  logProgress('Test suite completed', {
    phase: 'complete',
    totalTests,
    completedTests,
    failedTests,
    successRate: totalTests > 0 ? Math.round((completedTests / totalTests) * 100) : 0,
    summary: {
      passed: completedTests,
      failed: failedTests,
      total: totalTests
    }
  });
}

async function main() {
  const scriptStartTime = Date.now();
  
  logProgress('Automation script started', {
    phase: 'script-start',
    pid: process.pid,
    nodeVersion: process.version
  });

  // Step 1: Build
  logProgress('Build phase starting', { phase: 'build-phase-start' });
  const buildSuccess = await runBuild();
  
  if (!buildSuccess) {
    logProgress('Build failed, stopping automation', {
      phase: 'build-phase-failed',
      duration: Date.now() - scriptStartTime
    });
    process.exit(1);
  }

  logProgress('Build phase completed', { 
    phase: 'build-phase-complete',
    duration: Date.now() - scriptStartTime
  });

  // Step 2: Run tests
  logProgress('Test phase starting', { phase: 'test-phase-start' });
  await runAllTests();

  const totalDuration = Date.now() - scriptStartTime;
  logProgress('Automation completed', {
    phase: 'script-complete',
    totalDurationMs: totalDuration,
    totalDurationMinutes: Math.round(totalDuration / 60000),
    buildSuccess,
    testResults: {
      total: totalTests,
      passed: completedTests,
      failed: failedTests,
      successRate: totalTests > 0 ? Math.round((completedTests / totalTests) * 100) : 0
    }
  });

  process.exit(failedTests > 0 ? 1 : 0);
}

// Handle graceful shutdown
process.on('SIGINT', () => {
  logProgress('Script interrupted by user');
  process.exit(130);
});

process.on('SIGTERM', () => {
  logProgress('Script terminated');
  process.exit(143);
});

main().catch(error => {
  logProgress('Script failed with unhandled error', { error: error.message });
  process.exit(1);
});