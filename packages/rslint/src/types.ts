/**
 * Shared types for rslint IPC protocol across all environments
 */
export interface Position {
  line: number;
  column: number;
}

export interface Range {
  start: Position;
  end: Position;
}

export interface Diagnostic {
  ruleName: string;
  message: string;
  messageId: string;
  filePath: string;
  range: Range;
  severity?: string;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  suggestions: any[];
}

export interface LintResponse {
  diagnostics: Diagnostic[];
  errorCount: number;
  fileCount: number;
  ruleCount: number;
  duration: string;
}

export interface LintOptions {
  files?: string[];
  config?: string; // Path to rslint.json config file
  workingDirectory?: string;
  ruleOptions?: Record<string, string>;
  fileContents?: Record<string, string>; // Map of file paths to their contents for VFS
  languageOptions?: LanguageOptions; // Override languageOptions from config file
}

export interface LanguageOptions {
  parserOptions?: ParserOptions;
}

export interface ParserOptions {
  projectService?: boolean;
  project?: string[] | string;
}

export interface ApplyFixesRequest {
  fileContent: string; // Current content of the file
  diagnostics: Diagnostic[]; // Diagnostics with fixes to apply
}

export interface ApplyFixesResponse {
  fixedContent: string[]; // The content after applying fixes (array of intermediate versions)
  wasFixed: boolean; // Whether any fixes were actually applied
  appliedCount: number; // Number of fixes that were applied
  unappliedCount: number; // Number of fixes that couldn't be applied
}

export interface RSlintOptions {
  rslintPath?: string;
  workingDirectory?: string;
}

export interface PendingMessage {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  resolve: (data: any) => void;
  reject: (error: Error) => void;
}

export interface IpcMessage {
  id: number;
  kind: string;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data: any;
}

// Service interface that all implementations must follow
export interface RslintServiceInterface {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  sendMessage(kind: string, data: any): Promise<any>;
  terminate(): void;
}
