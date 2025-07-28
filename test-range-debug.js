import { RSLintService } from './packages/rslint/src/service.ts';

console.log('Testing range reporting directly...');

const service = new RSLintService();

try {
  const result = await service.lint({
    fileContents: {
      'test.ts': `
export class XXXX {
  public constructor(readonly value: string) {}
}
      `
    },
    ruleOptions: {
      'explicit-member-accessibility': JSON.stringify([{
        accessibility: 'off',
        overrides: {
          parameterProperties: 'explicit',
        },
      }])
    }
  });

  console.log('Result:', JSON.stringify(result, null, 2));
  
  if (result.diagnostics.length > 0) {
    const diag = result.diagnostics[0];
    console.log(`Range: line ${diag.range.start.line}, col ${diag.range.start.column} - ${diag.range.end.column}`);
    console.log(`Expected: line 3, col 22 - 36`);
  }
} catch (error) {
  console.error('Error:', error);
} finally {
  await service.close();
}