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
      name?: string;
      id?: string;
      input?: any;
      content?: string;
    }>;
  };
  subtype?: string;
  result?: {
    message?: string;
  };
  model?: string;
  cwd?: string;
  num_turns?: number;
  usage?: {
    input_tokens: number;
    output_tokens: number;
  };
}

export interface PortingResult {
  ruleName: string;
  success: boolean;
  error?: string;
  testPath?: string;
}