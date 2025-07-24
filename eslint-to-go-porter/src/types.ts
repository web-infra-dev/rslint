export interface RuleInfo {
  name: string;
  ruleUrl: string;
  testUrl: string;
}

export interface ClaudeResponse {
  type: string;
  message?: {
    content: Array<{
      text?: string;
      type?: string;
    }>;
  };
  subtype?: string;
  result?: {
    message?: string;
  };
}

export interface PortingResult {
  ruleName: string;
  success: boolean;
  goCode?: string;
  error?: string;
  testPath?: string;
}