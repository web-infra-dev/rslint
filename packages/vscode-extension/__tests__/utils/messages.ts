import * as vscode from 'vscode';

export interface RecordedMessage {
  readonly level: 'warning' | 'error' | 'information';
  readonly message: string;
  readonly items: readonly unknown[];
}

type MessageMethod =
  'showWarningMessage' | 'showErrorMessage' | 'showInformationMessage';

const levels: Record<MessageMethod, RecordedMessage['level']> = {
  showWarningMessage: 'warning',
  showErrorMessage: 'error',
  showInformationMessage: 'information',
};

/**
 * Records notifications instead of showing them. Test files and the extension
 * bundle share one per-extension `vscode` API object, so replacing its window
 * methods observes the prompts the running extension shows. Each notification
 * resolves to `answer`, which by default dismisses it.
 */
export function recordMessages(
  answer: (message: RecordedMessage) => string | undefined = () => undefined,
): {
  readonly messages: RecordedMessage[];
  restore(): void;
} {
  const messages: RecordedMessage[] = [];
  const window = vscode.window as unknown as Record<MessageMethod, unknown>;
  const originals = new Map<MessageMethod, unknown>();
  for (const method of Object.keys(levels) as MessageMethod[]) {
    originals.set(method, window[method]);
    window[method] = (message: string, ...items: unknown[]) => {
      const recorded = { level: levels[method], message, items };
      messages.push(recorded);
      return Promise.resolve(answer(recorded));
    };
  }
  return {
    messages,
    restore() {
      for (const [method, original] of originals) window[method] = original;
    },
  };
}

export async function waitForMessages(
  messages: readonly RecordedMessage[],
  predicate: (messages: readonly RecordedMessage[]) => boolean,
  description: string,
  timeoutMs = 60_000,
): Promise<void> {
  const deadline = Date.now() + timeoutMs;
  while (!predicate(messages)) {
    if (Date.now() >= deadline) {
      throw new Error(
        `Timed out waiting for ${description}; recorded: ${JSON.stringify(messages)}`,
      );
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
}
