#!/usr/bin/env node
import { program } from 'commander';
import chalk from 'chalk';
import ora from 'ora';
import { readFile } from 'fs/promises';
import { RuleFetcher } from './fetcher.js';
import { ClaudePorter } from './claude-porter.js';
import { TestRunner } from './test-runner.js';
import { RuleChecker } from './rule-checker.js';
import { PortingResult } from './types.js';
import { join } from 'path';

const fetcher = new RuleFetcher();
const porter = new ClaudePorter();
const testRunner = new TestRunner();
const ruleChecker = new RuleChecker();

async function portSingleRule(ruleName: string, showProgress: boolean = false, skipExisting: boolean = true): Promise<PortingResult> {
  // Check if rule already exists
  if (skipExisting) {
    const exists = await ruleChecker.isRuleAlreadyPorted(ruleName);
    if (exists) {
      if (!showProgress) {
        console.log(chalk.yellow(`⚠ Skipping ${ruleName} - already ported`));
      }
      return {
        ruleName,
        success: true,
        error: 'Already ported'
      };
    }
  }
  
  const spinner = showProgress ? null : ora(`Porting ${ruleName}`).start();
  
  try {
    // Fetch rule info
    if (spinner) spinner.text = `Fetching ${ruleName} files...`;
    const ruleInfo = await fetcher.fetchRuleFiles(ruleName);
    
    // Download rule source
    const ruleSource = await fetcher.downloadRuleSource(ruleInfo.ruleUrl);
    
    // Download and save test
    if (spinner) spinner.text = `Saving test for ${ruleName}...`;
    const testPath = await fetcher.downloadAndSaveTest(ruleInfo.testUrl, ruleName);
    
    let testSource = '';
    // Adapt the TypeScript test for rslint cross-validation (if test was downloaded)
    if (testPath) {
      if (spinner) spinner.text = `Adapting test for ${ruleName}...`;
      testSource = await readFile(testPath, 'utf-8');
      const adapted = await porter.adaptTestFile(ruleName, testSource);
      if (!adapted) {
        console.warn(`Warning: ${ruleName} test adaptation failed, continuing...`);
      }
    }
    
    // Port the rule (passing test source as reference)
    if (spinner) spinner.text = `Converting ${ruleName} to Go...`;
    porter.setProgressMode(showProgress);
    const result = await porter.portRule(ruleName, ruleSource, testSource);
    
    if (!result.success) {
      if (spinner) spinner.fail(`Failed to port ${ruleName}: ${result.error}`);
      return result;
    }
    
    // Get the Go rule path
    const goRulePath = join(
      '/Users/bytedance/dev/rslint/internal/rules',
      ruleName.replace(/-/g, '_'),
      `${ruleName.replace(/-/g, '_')}.go`
    );
    
    // Verify the rule with a second Claude pass
    if (spinner) spinner.text = `Verifying ${ruleName} conversion...`;
    const verified = await porter.verifyRule(ruleName, ruleSource, goRulePath);
    
    if (!verified) {
      console.warn(`Warning: ${ruleName} verification failed, continuing with tests...`);
    }
    
    // Run tests
    if (spinner) spinner.text = `Running tests for ${ruleName}...`;
    const testResult = await testRunner.runRuleTest(ruleName);
    
    if (testResult.success) {
      // Run cross-validation tests to ensure Go rule matches TypeScript behavior
      if (spinner) spinner.text = `Running cross-validation for ${ruleName}...`;
      const crossValidated = await porter.crossValidateRule(ruleName);
      
      if (crossValidated) {
        if (spinner) spinner.succeed(`Successfully ported, tested, and cross-validated ${ruleName}`);
        return { ...result, testPath };
      } else {
        console.warn(`Warning: ${ruleName} cross-validation failed, but Go tests passed`);
        if (spinner) spinner.succeed(`Successfully ported and tested ${ruleName} (cross-validation failed)`);
        return { ...result, testPath };
      }
    }
    
    // If tests failed, try to fix with Claude
    if (spinner) spinner.text = `Tests failed for ${ruleName}, attempting to fix...`;
    const fixed = await porter.fixTestFailures(ruleName, ruleSource, goRulePath, testResult.output);
    
    if (fixed) {
      // Run tests again
      const retestResult = await testRunner.runRuleTest(ruleName);
      if (retestResult.success) {
        // Run cross-validation tests after successful fix
        if (spinner) spinner.text = `Running cross-validation for ${ruleName}...`;
        const crossValidated = await porter.crossValidateRule(ruleName);
        
        if (crossValidated) {
          if (spinner) spinner.succeed(`Successfully fixed, tested, and cross-validated ${ruleName}`);
          return { ...result, testPath };
        } else {
          console.warn(`Warning: ${ruleName} cross-validation failed after fix, but Go tests passed`);
          if (spinner) spinner.succeed(`Successfully fixed and tested ${ruleName} (cross-validation failed)`);
          return { ...result, testPath };
        }
      }
    }
    
    if (spinner) spinner.fail(`Failed to fix ${ruleName} after test failures`);
    return {
      ...result,
      success: false,
      error: `Tests failed: ${testResult.output.substring(0, 500)}...`
    };
  } catch (error) {
    if (spinner) spinner.fail(`Error porting ${ruleName}: ${error}`);
    return {
      ruleName,
      success: false,
      error: String(error)
    };
  }
}

async function portMultipleRules(ruleNames: string[], showProgress: boolean = false, skipExisting: boolean = true): Promise<void> {
  let rulesToPort = ruleNames;
  
  if (skipExisting) {
    const unportedRules = await ruleChecker.filterUnportedRules(ruleNames);
    const skippedCount = ruleNames.length - unportedRules.length;
    
    if (!showProgress && skippedCount > 0) {
      console.log(chalk.yellow(`⚠ Skipping ${skippedCount} already ported rules`));
    }
    
    rulesToPort = unportedRules;
  }
  
  if (rulesToPort.length === 0) {
    if (!showProgress) {
      console.log(chalk.green('✓ All rules are already ported!'));
    }
    return;
  }
  
  if (!showProgress) {
    console.log(chalk.blue(`\nPorting ${rulesToPort.length} rules...\n`));
  }
  
  const results: PortingResult[] = [];
  
  for (const ruleName of rulesToPort) {
    const result = await portSingleRule(ruleName, showProgress, skipExisting);
    results.push(result);
    
    // Add a small delay between rules to avoid rate limiting
    if (result.error !== 'Already ported') {
      await new Promise(resolve => setTimeout(resolve, 2000));
    }
  }
  
  if (!showProgress) {
    // Summary
    console.log(chalk.blue('\n=== Porting Summary ===\n'));
    
    const successful = results.filter(r => r.success && r.error !== 'Already ported');
    const failed = results.filter(r => !r.success);
    const skipped = results.filter(r => r.error === 'Already ported');
    
    if (successful.length > 0) {
      console.log(chalk.green(`✓ Successfully ported ${successful.length} rules:`));
      successful.forEach(r => {
        console.log(chalk.gray(`  - ${r.ruleName}`));
        if (r.testPath) {
          console.log(chalk.gray(`    Test saved to: ${r.testPath}`));
        }
      });
    }
    
    if (skipped.length > 0) {
      console.log(chalk.yellow(`\n⚠ Skipped ${skipped.length} already ported rules`));
    }
    
    if (failed.length > 0) {
      console.log(chalk.red(`\n✗ Failed to port ${failed.length} rules:`));
      failed.forEach(r => {
        console.log(chalk.gray(`  - ${r.ruleName}: ${r.error}`));
      });
    }
  }
}

program
  .name('eslint-to-go-porter')
  .description('Port TypeScript ESLint rules to Go using Claude Code')
  .version('1.0.0');

program
  .command('port <rules...>')
  .description('Port one or more ESLint rules to Go')
  .option('-a, --all', 'Port all available rules')
  .option('-p, --progress', 'Show JSON streaming progress')
  .option('-f, --force', 'Force re-port existing rules')
  .action(async (rules: string[], options) => {
    try {
      await porter.loadPromptTemplates();
      const skipExisting = !options.force;
      
      if (options.all) {
        if (!options.progress) {
          console.log(chalk.blue('Fetching all available rules...'));
        }
        const allRules = await fetcher.fetchAvailableRules();
        if (!options.progress) {
          console.log(chalk.gray(`Found ${allRules.length} rules`));
        }
        await portMultipleRules(allRules, options.progress, skipExisting);
      } else {
        await portMultipleRules(rules, options.progress, skipExisting);
      }
    } catch (error) {
      if (!options.progress) {
        console.error(chalk.red(`Error: ${error}`));
      } else {
        console.error(JSON.stringify({ error: String(error) }));
      }
      process.exit(1);
    }
  });

program
  .command('list')
  .description('List all available ESLint rules')
  .action(async () => {
    try {
      const spinner = ora('Fetching available rules...').start();
      const rules = await fetcher.fetchAvailableRules();
      spinner.succeed(`Found ${rules.length} rules`);
      
      console.log(chalk.blue('\nAvailable rules:'));
      rules.forEach(rule => {
        console.log(chalk.gray(`  - ${rule}`));
      });
    } catch (error) {
      console.error(chalk.red(`Error: ${error}`));
      process.exit(1);
    }
  });

program
  .command('status')
  .description('Show porting status of all rules')
  .action(async () => {
    try {
      const spinner = ora('Checking porting status...').start();
      
      // Get all available rules
      const allRules = await fetcher.fetchAvailableRules();
      
      // Get existing ported rules
      const existingRules = await ruleChecker.getExistingRules();
      const existingSet = new Set(existingRules);
      
      // Calculate statistics
      const portedRules = allRules.filter(rule => existingSet.has(rule));
      const unportedRules = allRules.filter(rule => !existingSet.has(rule));
      
      spinner.succeed('Status check complete');
      
      console.log(chalk.blue('\n=== Porting Status ===\n'));
      console.log(chalk.green(`✓ Ported: ${portedRules.length}/${allRules.length} (${Math.round(portedRules.length / allRules.length * 100)}%)`));
      console.log(chalk.yellow(`⚠ Remaining: ${unportedRules.length}`));
      
      if (unportedRules.length > 0) {
        console.log(chalk.blue('\nUnported rules:'));
        unportedRules.forEach(rule => {
          console.log(chalk.gray(`  - ${rule}`));
        });
      } else {
        console.log(chalk.green('\n🎉 All rules have been ported!'));
      }
    } catch (error) {
      console.error(chalk.red(`Error: ${error}`));
      process.exit(1);
    }
  });

program.parse();