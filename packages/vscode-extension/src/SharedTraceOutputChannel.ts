import type { OutputChannel, ViewColumn } from 'vscode';

const lineBreak = /\r\n|\r|\n/g;

function encodeLabelValue(value: string): string {
  return JSON.stringify(value)
    .replaceAll('\u2028', '\\u2028')
    .replaceAll('\u2029', '\\u2029');
}

/**
 * Build a readable label from the two values that form a runtime identity.
 * JSON string encoding keeps control characters from forging extra log lines.
 */
export function createRuntimeTraceLabel(
  workspaceUri: string,
  coreIdentity: string,
): string {
  return `workspace=${encodeLabelValue(workspaceUri)} core=${encodeLabelValue(coreIdentity)}`;
}

/**
 * Coordinates runtime-scoped views over one extension-owned output channel.
 * It owns presentation only; vscode-languageclient remains the sole authority
 * for the window-wide trace level and its live configuration updates.
 */
export class SharedTraceOutputChannel {
  private atLineStart = true;
  private activeLabel: string | undefined;

  public constructor(private readonly channel: OutputChannel) {}

  public forRuntime(label: string): OutputChannel {
    if (label.length === 0) {
      throw new Error('A runtime trace label must not be empty');
    }
    return new RuntimeTraceOutputChannel(this, label);
  }

  public append(label: string, value: string, appendLine: boolean): void {
    const text = appendLine ? `${value}\n` : value;
    if (text.length === 0) return;
    this.channel.append(this.render(label, text));
  }

  public replace(label: string, value: string): void {
    this.resetLineState();
    this.channel.replace(value.length === 0 ? '' : this.render(label, value));
  }

  public clear(): void {
    this.resetLineState();
    this.channel.clear();
  }

  public show(
    columnOrPreserveFocus?: ViewColumn | boolean,
    preserveFocus?: boolean,
  ): void {
    if (
      columnOrPreserveFocus === undefined ||
      typeof columnOrPreserveFocus === 'boolean'
    ) {
      this.channel.show(columnOrPreserveFocus);
      return;
    }
    this.channel.show(columnOrPreserveFocus, preserveFocus);
  }

  public hide(): void {
    this.channel.hide();
  }

  public get name(): string {
    return this.channel.name;
  }

  private render(label: string, value: string): string {
    const rendered: string[] = [];
    if (!this.atLineStart && this.activeLabel !== label) {
      // OutputChannel.append() can leave a physical line open. Never allow a
      // second runtime to inherit that unattributed partial line.
      rendered.push('\n');
      this.resetLineState();
    }

    let start = 0;
    for (const match of value.matchAll(lineBreak)) {
      const index = match.index;
      rendered.push(
        this.prefixForLine(label),
        value.slice(start, index),
        match[0],
      );
      this.resetLineState();
      start = index + match[0].length;
    }

    if (start < value.length) {
      rendered.push(this.prefixForLine(label), value.slice(start));
    }
    return rendered.join('');
  }

  private prefixForLine(label: string): string {
    if (!this.atLineStart) return '';
    this.atLineStart = false;
    this.activeLabel = label;
    return `[${label}] `;
  }

  private resetLineState(): void {
    this.atLineStart = true;
    this.activeLabel = undefined;
  }
}

class RuntimeTraceOutputChannel implements OutputChannel {
  private disposed = false;

  public constructor(
    private readonly shared: SharedTraceOutputChannel,
    private readonly label: string,
  ) {}

  public get name(): string {
    return this.shared.name;
  }

  public append(value: string): void {
    if (this.disposed) return;
    this.shared.append(this.label, value, false);
  }

  public appendLine(value: string): void {
    if (this.disposed) return;
    this.shared.append(this.label, value, true);
  }

  public replace(value: string): void {
    if (this.disposed) return;
    this.shared.replace(this.label, value);
  }

  public clear(): void {
    if (this.disposed) return;
    this.shared.clear();
  }

  public show(preserveFocus?: boolean): void;
  public show(column?: ViewColumn, preserveFocus?: boolean): void;
  public show(
    columnOrPreserveFocus?: ViewColumn | boolean,
    preserveFocus?: boolean,
  ): void {
    if (this.disposed) return;
    this.shared.show(columnOrPreserveFocus, preserveFocus);
  }

  public hide(): void {
    if (this.disposed) return;
    this.shared.hide();
  }

  public dispose(): void {
    // The view has no independent VS Code resource. The Extension disposes the
    // shared underlying channel only after every runtime has closed.
    this.disposed = true;
  }
}
