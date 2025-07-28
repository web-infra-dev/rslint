import { lint } from '../rslint/dist/index.js';
import * as path from 'path';

async function testEmoji() {
  console.log('Testing emoji handling in ban-ts-comment...');
  
  const result = await lint({
    fileContents: {
      '/test.ts': `
// @ts-expect-error 👨‍👩‍👧‍👦
console.log('test');
      `.trim()
    },
    workingDirectory: process.cwd(),
    ruleOptions: {
      'ban-ts-comment': JSON.stringify({
        'ts-expect-error': 'allow-with-description',
        'minimumDescriptionLength': 3
      })
    }
  });
  
  console.log('Result:', JSON.stringify(result, null, 2));
  console.log('Expected: Error because single emoji is only 1 character');
  
  const result2 = await lint({
    fileContents: {
      '/test2.ts': `
// @ts-expect-error 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦
console.log('test');
      `.trim()
    },
    workingDirectory: process.cwd(),
    ruleOptions: {
      'ban-ts-comment': JSON.stringify({
        'ts-expect-error': 'allow-with-description',
        'minimumDescriptionLength': 3
      })
    }
  });
  
  console.log('\nResult2:', JSON.stringify(result2, null, 2));
  console.log('Expected: No error because 3 emojis = 3 characters');
}

testEmoji().catch(console.error);