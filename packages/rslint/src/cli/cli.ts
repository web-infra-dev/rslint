import { runRustCLI } from '../rust/cli.js';

export type RunCLIOptions = {
  /** Arguments in the shape of Node.js process.argv. */
  argv?: string[];
};

/** Retains the historical wrapper signature; execution is in the Rust addon. */
export async function run(
  _binPath: string,
  argv: string[],
  startTime: number,
): Promise<number> {
  return runRustCLI(argv, { startTime });
}

export async function runCLI({
  argv = process.argv,
}: RunCLIOptions = {}): Promise<void> {
  const startTime = Date.now();
  process.exitCode = await runRustCLI(argv.slice(2), { startTime });
}
