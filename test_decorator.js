const { RSLintService } = require('./packages/rslint/dist/index.js');
const path = require('path');
const fs = require('fs');

async function test() {
  const tempDir = fs.mkdtempSync(path.join(require('os').tmpdir(), 'rslint-test-'));
  const configPath = path.join(tempDir, 'rslint.json');
  const tsconfigPath = path.join(tempDir, 'tsconfig.json');
  const testFile = path.join(tempDir, 'test.ts');
  
  fs.writeFileSync(tsconfigPath, JSON.stringify({
    compilerOptions: {
      strictNullChecks: true,
      experimentalDecorators: true,
      emitDecoratorMetadata: false // This is key - when false, should report error
    },
    include: ["test.ts"]
  }));
  
  fs.writeFileSync(configPath, JSON.stringify([{
    language: "typescript",
    files: ["test.ts"],
    languageOptions: {
      parserOptions: {
        project: ["./tsconfig.json"],
        projectService: false
      }
    },
    rules: {
      "consistent-type-imports": "error"
    }
  }]));
  
  // Test case with decorator
  const code = `import Foo from 'foo';
@deco
class A {
  constructor(foo: Foo) {}
}`;
  
  fs.writeFileSync(testFile, code);
  
  console.log('Testing consistent-type-imports rule with decorator...');
  console.log('emitDecoratorMetadata: false, experimentalDecorators: true');
  const service = new RSLintService({ workingDirectory: tempDir });
  
  try {
    const result = await service.lint({
      files: [testFile],
      config: configPath,
      workingDirectory: tempDir
    });
    
    console.log('Result:', JSON.stringify(result, null, 2));
    console.log('Expected: Should report error since emitDecoratorMetadata is false');
  } catch (err) {
    console.error('Error:', err);
  } finally {
    await service.close();
  }
  
  fs.rmSync(tempDir, { recursive: true });
}

test().catch(console.error);