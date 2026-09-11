export const SOURCE_FILE_NAMES = [
  'index.ts',
  'index.js',
  'index.tsx',
  'index.jsx',
] as const;

export type SourceFileName = (typeof SOURCE_FILE_NAMES)[number];

export const DEFAULT_SOURCE_FILE_NAME: SourceFileName = 'index.ts';

export function isSourceFileName(value: unknown): value is SourceFileName {
  return SOURCE_FILE_NAMES.some((fileName) => fileName === value);
}

export function sourceFileLanguage(fileName: SourceFileName) {
  return fileName.endsWith('.js') || fileName.endsWith('.jsx')
    ? 'javascript'
    : 'typescript';
}
