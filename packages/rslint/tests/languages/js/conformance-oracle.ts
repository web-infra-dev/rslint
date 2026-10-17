import * as espree from 'espree';
import * as typescriptParser from '@typescript-eslint/parser';
import type { ConformanceCase } from './conformance-cases.js';
import type { ESTreeNode } from '../../../src/languages/js/types.js';

interface ReferenceToken {
  type: string;
  value: string;
  range: [number, number];
  loc: {
    start: { line: number; column: number };
    end: { line: number; column: number };
  };
  regex?: { pattern: string; flags: string };
}
interface ReferenceProgram extends ESTreeNode {
  tokens: ReferenceToken[];
  comments: ReferenceToken[];
}

export function referenceParser(fixture: ConformanceCase) {
  return /\.tsx?$/.test(fixture.filePath) ? typescriptParser : espree;
}

export function parseReference(fixture: ConformanceCase) {
  return referenceParser(fixture).parse(fixture.text.replace(/^\ufeff/, ''), {
    filePath: fixture.filePath,
    ecmaVersion: 'latest',
    sourceType: fixture.sourceType,
    ecmaFeatures: {
      jsx: !fixture.filePath.endsWith('.ts'),
    },
    range: true,
    loc: true,
    tokens: true,
    comment: true,
  }) as unknown as ReferenceProgram;
}
