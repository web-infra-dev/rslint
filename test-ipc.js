import { RSLintService } from './packages/rslint/dist/index.js';

async function test() {
  console.log('Starting test...');
  const service = new RSLintService();
  
  try {
    console.log('Sending lint request...');
    const result = await service.lint({
      fileContents: {
        'test.ts': 'const arr = [1, 2, 3]; delete arr[0];'
      }
    });
    console.log('Result:', result);
  } catch (err) {
    console.error('Error:', err);
  } finally {
    await service.close();
  }
}

test().catch(console.error);