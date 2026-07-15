import * as vscode from 'vscode';

export const rslintDiagnosticSource = 'rslint';

export function getRslintDiagnostics(
  documentOrUri: vscode.TextDocument | vscode.Uri,
): vscode.Diagnostic[] {
  const uri =
    documentOrUri instanceof vscode.Uri ? documentOrUri : documentOrUri.uri;
  return vscode.languages
    .getDiagnostics(uri)
    .filter((diagnostic) => diagnostic.source === rslintDiagnosticSource);
}

function describeDiagnostics(
  diagnostics: readonly vscode.Diagnostic[],
): string {
  if (diagnostics.length === 0) return '<none>';
  return diagnostics
    .map(
      (diagnostic) =>
        `${diagnostic.source ?? '<no source>'}: ${diagnostic.message}`,
    )
    .join(' | ');
}

/** Subscribe before the initial read, filter by source, and reject on timeout. */
export function waitForRslintDiagnostics(
  document: vscode.TextDocument,
  predicate: (diagnostics: vscode.Diagnostic[]) => boolean = (diagnostics) =>
    diagnostics.length > 0,
  timeoutMs = 60_000,
): Promise<vscode.Diagnostic[]> {
  return new Promise<vscode.Diagnostic[]>((resolve, reject) => {
    const uriString = document.uri.toString();
    let settled = false;
    let timer: ReturnType<typeof setTimeout> | undefined;

    const finish = (
      diagnostics: vscode.Diagnostic[] | undefined,
      error?: unknown,
    ): void => {
      if (settled) return;
      settled = true;
      subscription.dispose();
      if (timer) clearTimeout(timer);
      if (error) reject(error);
      else resolve(diagnostics ?? []);
    };
    const check = (): void => {
      const diagnostics = getRslintDiagnostics(document);
      try {
        if (predicate(diagnostics)) finish(diagnostics);
      } catch (error) {
        finish(undefined, error);
      }
    };
    const subscription = vscode.languages.onDidChangeDiagnostics((event) => {
      if (event.uris.some((uri) => uri.toString() === uriString)) check();
    });

    timer = setTimeout(() => {
      const diagnostics = getRslintDiagnostics(document);
      finish(
        undefined,
        new Error(
          `Timed out after ${timeoutMs}ms waiting for rslint diagnostics for ${document.uri.toString()}. ` +
            `Last rslint diagnostics: ${describeDiagnostics(diagnostics)}`,
        ),
      );
    }, timeoutMs);
    check();
  });
}

export function waitForRslintDiagnosticsToChange(
  document: vscode.TextDocument,
  previousCount: number,
  timeoutMs = 30_000,
): Promise<vscode.Diagnostic[]> {
  return waitForRslintDiagnostics(
    document,
    (diagnostics) => diagnostics.length !== previousCount,
    timeoutMs,
  );
}

export function waitForRslintDiagnosticsCount(
  document: vscode.TextDocument,
  expectedCount: number,
  timeoutMs = 30_000,
): Promise<vscode.Diagnostic[]> {
  return waitForRslintDiagnostics(
    document,
    (diagnostics) => diagnostics.length === expectedCount,
    timeoutMs,
  );
}
