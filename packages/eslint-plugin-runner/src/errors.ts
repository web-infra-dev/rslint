/**
 * Stable error code for "the worker subsystem (WorkerPool / IpcClient) was
 * used after it closed". Callers in a DIFFERENT module realm — notably the
 * CJS-bundled VS Code extension, which can't `instanceof` this ESM-only
 * class — detect it structurally via {@link isWorkerClosedError} keyed on
 * this code, never on the (human-facing, changeable) message text.
 */
export const WORKER_CLOSED_CODE = 'ERR_RSLINT_WORKER_CLOSED';

/** Thrown when a WorkerPool / IpcClient is used after being closed. */
export class WorkerClosedError extends Error {
  readonly code = WORKER_CLOSED_CODE;

  constructor(message: string) {
    super(message);
    this.name = 'WorkerClosedError';
  }
}

/**
 * Realm-safe check for {@link WorkerClosedError}. Matches on the stable
 * `code` so it works across the ESM-runner / CJS-extension boundary, where
 * `instanceof` cannot (separate class identities, no static import).
 */
export function isWorkerClosedError(err: unknown): boolean {
  return (
    typeof err === 'object' &&
    err !== null &&
    (err as { code?: unknown }).code === WORKER_CLOSED_CODE
  );
}
