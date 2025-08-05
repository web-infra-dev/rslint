import { window, OutputChannel } from 'vscode';

export enum LogLevel {
  TRACE = 0,
  DEBUG = 1,
  INFO = 2,
  WARN = 3,
  ERROR = 4,
}

export class Logger {
  private static instance: Logger;
  private outputChannel: OutputChannel;
  private logLevel: LogLevel = LogLevel.INFO;

  private constructor() {
    this.outputChannel = window.createOutputChannel('Rslint');
  }

  public static getInstance(): Logger {
    if (!Logger.instance) {
      Logger.instance = new Logger();
    }
    return Logger.instance;
  }

  public setLogLevel(level: LogLevel): void {
    this.logLevel = level;
  }

  public trace(message: string, ...args: unknown[]): void {
    this.log(LogLevel.TRACE, message, ...args);
  }

  public debug(message: string, ...args: unknown[]): void {
    this.log(LogLevel.DEBUG, message, ...args);
  }

  public info(message: string, ...args: unknown[]): void {
    this.log(LogLevel.INFO, message, ...args);
  }

  public warn(message: string, ...args: unknown[]): void {
    this.log(LogLevel.WARN, message, ...args);
  }

  public error(
    message: string,
    error?: Error | unknown,
    ...args: unknown[]
  ): void {
    if (error instanceof Error) {
      this.log(
        LogLevel.ERROR,
        `${message}: ${error.message}`,
        error.stack,
        ...args,
      );
    } else if (error) {
      this.log(LogLevel.ERROR, message, error, ...args);
    } else {
      this.log(LogLevel.ERROR, message, ...args);
    }
  }

  public show(): void {
    this.outputChannel.show();
  }

  public dispose(): void {
    this.outputChannel.dispose();
  }

  private log(level: LogLevel, message: string, ...args: unknown[]): void {
    if (level < this.logLevel) {
      return;
    }

    const timestamp = new Date().toISOString();
    const levelName = LogLevel[level].padEnd(5);
    const formattedMessage = this.formatMessage(message, ...args);

    this.outputChannel.appendLine(
      `[${timestamp}] ${levelName} ${formattedMessage}`,
    );
  }

  private formatMessage(message: string, ...args: unknown[]): string {
    if (args.length === 0) {
      return message;
    }

    const formattedArgs = args.map(arg => {
      if (typeof arg === 'string') {
        return arg;
      }
      if (arg instanceof Error) {
        return `\n${arg.stack || arg.message}`;
      }
      return JSON.stringify(arg, null, 2);
    });

    return `${message} ${formattedArgs.join(' ')}`;
  }
}

// Export a default instance for convenience
export const logger = Logger.getInstance();
