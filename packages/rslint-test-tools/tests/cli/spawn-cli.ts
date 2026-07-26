import { spawn } from 'node:child_process';

const CLI_PROCESS_TIMEOUT_MS = 180_000;
const CLI_OUTPUT_LIMIT_BYTES = 16 * 1024 * 1024;
const EXPECTED_CLI_EXIT_CODES = new Set([0, 1, 2]);
const RUNTIME_FATAL_MARKER =
  /^(?:panic:|fatal error:|FATAL ERROR:|fatal runtime error:|thread ['"].*['"](?: \([^)]*\))? panicked at\b)|\bERR_WORKER_OUT_OF_MEMORY\b/m;

export interface CliProcessResult {
  exitCode: number;
  stdout: string;
  stderr: string;
}

export function validateCliClose(input: {
  timedOut: boolean;
  spawnError?: Error;
  outputOverflow?: boolean;
  code: number | null;
  signal: NodeJS.Signals | null;
  stdout?: string;
  stderr?: string;
}): string | undefined {
  if (input.timedOut) {
    return `timed out after ${CLI_PROCESS_TIMEOUT_MS}ms`;
  }
  if (input.spawnError) {
    return `failed to spawn: ${input.spawnError.message}`;
  }
  if (input.outputOverflow) {
    return `output exceeded ${CLI_OUTPUT_LIMIT_BYTES} bytes`;
  }
  if (input.signal !== null || input.code === null) {
    return `closed abnormally (code=${String(input.code)}, signal=${String(input.signal)})`;
  }
  if (!EXPECTED_CLI_EXIT_CODES.has(input.code)) {
    return `closed with unsupported exit code ${input.code}`;
  }
  const fatalMarker = input.stderr?.match(RUNTIME_FATAL_MARKER)?.[0];
  if (fatalMarker) {
    return `emitted runtime-fatal marker ${JSON.stringify(fatalMarker)}`;
  }
  return undefined;
}

export function runCliProcess(
  command: string,
  args: string[],
  options: {
    cwd?: string;
    env?: NodeJS.ProcessEnv;
  } = {},
): Promise<CliProcessResult> {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd: options.cwd,
      env: options.env,
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    const stdoutChunks: Buffer[] = [];
    const stderrChunks: Buffer[] = [];
    let spawnError: Error | undefined;
    let timedOut = false;
    let outputBytes = 0;
    let outputOverflow = false;

    child.stdout.on('data', (data: Buffer) => {
      outputBytes += data.length;
      if (outputBytes > CLI_OUTPUT_LIMIT_BYTES) {
        outputOverflow = true;
        child.kill('SIGKILL');
        return;
      }
      stdoutChunks.push(data);
    });
    child.stderr.on('data', (data: Buffer) => {
      outputBytes += data.length;
      if (outputBytes > CLI_OUTPUT_LIMIT_BYTES) {
        outputOverflow = true;
        child.kill('SIGKILL');
        return;
      }
      stderrChunks.push(data);
    });
    child.on('error', (error) => {
      spawnError = error;
    });

    const timer = setTimeout(() => {
      timedOut = true;
      child.kill('SIGKILL');
    }, CLI_PROCESS_TIMEOUT_MS);
    timer.unref();

    child.on('close', (code, signal) => {
      clearTimeout(timer);
      const stdout = Buffer.concat(stdoutChunks).toString('utf8');
      const stderr = Buffer.concat(stderrChunks).toString('utf8');
      const commandLine = [command, ...args].join(' ');
      const failure = validateCliClose({
        timedOut,
        spawnError,
        outputOverflow,
        code,
        signal,
        stderr,
      });
      if (failure) {
        reject(
          new Error(
            `CLI process ${failure}: ${commandLine}\nstdout:\n${stdout}\nstderr:\n${stderr}`,
          ),
        );
        return;
      }
      resolve({ exitCode: code!, stdout, stderr });
    });
  });
}
