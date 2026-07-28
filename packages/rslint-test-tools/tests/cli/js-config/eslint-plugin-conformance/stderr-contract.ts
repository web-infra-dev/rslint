/**
 * Validate stderr without treating worker-dependent warning multiplicity as
 * a stable contract. Every distinct expected line is required, all other
 * lines are rejected, and the annotated-case count remains a fail-closed upper
 * bound against warning amplification.
 */
export function validateConformanceStderrContract(
  actualLines: readonly string[],
  expectedLines: readonly string[],
): string | undefined {
  const countOccurrences = (lines: readonly string[]): Map<string, number> => {
    const counts = new Map<string, number>();
    for (const line of lines) {
      counts.set(line, (counts.get(line) ?? 0) + 1);
    }
    return counts;
  };
  const actualCounts = countOccurrences(actualLines);
  const expectedMaximumCounts = countOccurrences(expectedLines);
  const violations: string[] = [];

  for (const [line, expectedAtMost] of expectedMaximumCounts) {
    const actual = actualCounts.get(line) ?? 0;
    if (actual === 0) {
      violations.push(`missing required stderr line ${JSON.stringify(line)}`);
    } else if (actual > expectedAtMost) {
      violations.push(
        `stderr line ${JSON.stringify(line)} occurred ${actual} times; expected at most ${expectedAtMost}`,
      );
    }
  }

  for (const [line, actual] of actualCounts) {
    if (!expectedMaximumCounts.has(line)) {
      violations.push(
        `unexpected stderr line ${JSON.stringify(line)} (${actual} occurrence${actual === 1 ? '' : 's'})`,
      );
    }
  }

  return violations.length > 0 ? violations.join('\n') : undefined;
}
