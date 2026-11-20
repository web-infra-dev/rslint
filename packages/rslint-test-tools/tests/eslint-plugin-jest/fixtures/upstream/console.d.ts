// Model methods rather than function-valued properties or any. The unbound
// checks must distinguish the safe built-in log method from an ordinary method.
interface TestConsole {
  log(...values: unknown[]): void;
  customMethod(): void;
}

declare const console: TestConsole;

// The fixture project uses moduleDetection: force to keep a case-local console
// separate from this ambient declaration, matching sourceType: module.
