/**
 * Run one asynchronous startup step within a shared absolute deadline.
 * A never-settling operation rejects at the deadline instead of hanging the
 * Extension Host before Mocha (and its per-test timeouts) has started.
 */
export async function runBeforeDeadline<T>(
  operation: () => T | PromiseLike<T>,
  deadline: number,
  description: string,
): Promise<T> {
  const remainingMs = deadline - Date.now();
  if (remainingMs <= 0) {
    throw new Error(
      `Timed out waiting for ${description}: the shared startup deadline has expired`,
    );
  }

  let timer: ReturnType<typeof setTimeout> | undefined;
  try {
    return await Promise.race([
      Promise.resolve().then(operation),
      new Promise<never>((_, reject) => {
        timer = setTimeout(
          () =>
            reject(
              new Error(
                `Timed out waiting for ${description} before the shared startup deadline`,
              ),
            ),
          remainingMs,
        );
      }),
    ]);
  } finally {
    if (timer) clearTimeout(timer);
  }
}
