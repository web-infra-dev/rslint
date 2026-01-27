// Utility functions
export function isEven(n: number): boolean {
  return n % 2 === 0;
}

export function capitalize(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

export type Result<T, E> = { ok: true; value: T } | { ok: false; error: E };

export function tryCatch<T>(fn: () => T): Result<T, Error> {
  try {
    return { ok: true, value: fn() };
  } catch (error) {
    return { ok: false, error: error as Error };
  }
}
