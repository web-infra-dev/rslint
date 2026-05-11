/**
 * Pure helper for reacting to dynamic workspace-folder add/remove
 * events. Lifted out of `Extension.ts` so it can be unit-tested
 * without a live VS Code runtime (Extension's constructor binds
 * Logger + ExtensionContext + Rslint, all of which require the
 * extension host).
 *
 * The handler is intentionally tolerant of per-folder failures:
 * a failed remove of one folder must not block the remove or
 * add of any other folder in the same event. Errors land on the
 * passed-in logger and are surfaced to the user via the VS Code
 * output channel; we never let one mis-configured folder crash
 * the whole extension.
 */
import type { WorkspaceFolder } from 'vscode';

export interface WorkspaceFoldersChangeEventLike {
  readonly added: readonly WorkspaceFolder[];
  readonly removed: readonly WorkspaceFolder[];
}

export interface RslintInstanceManager {
  /** Check whether an instance for the given folder id is already alive. */
  has(id: string): boolean;
  /**
   * Create + start + wire monitoring for a freshly-added folder.
   * Errors thrown here are caught by `applyFolderChange` and logged.
   */
  create(folder: WorkspaceFolder): Promise<void>;
  /**
   * Stop + dispose + drop the instance for a removed folder. Errors
   * thrown here are caught and logged; subsequent folders still get
   * processed.
   */
  remove(id: string): Promise<void>;
}

export interface FolderChangeLogger {
  warn(msg: string): void;
  error(msg: string, err?: unknown): void;
}

/**
 * Stable, unique key for a workspace folder. Uses the folder URI string
 * (globally unique) rather than `folder.name` — which defaults to the
 * basename and collides across folders sharing a last path segment
 * (e.g. two `pkg` folders), causing the second to be skipped / not linted.
 */
export function folderKey(folder: WorkspaceFolder): string {
  return folder.uri.toString();
}

/**
 * Apply one onDidChangeWorkspaceFolders event to the instance manager.
 *
 * Order matters: REMOVED folders are processed before ADDED folders so
 * that an entry whose folder is re-keyed (same name removed + added
 * back) ends up with a fresh, clean instance instead of hitting the
 * "instance already exists" guard.
 */
export async function applyFolderChange(
  event: WorkspaceFoldersChangeEventLike,
  manager: RslintInstanceManager,
  logger: FolderChangeLogger,
): Promise<void> {
  for (const removed of event.removed) {
    try {
      await manager.remove(folderKey(removed));
    } catch (err) {
      logger.error(
        `Failed to stop Rslint instance for removed folder '${removed.name}'`,
        err,
      );
    }
  }
  for (const added of event.added) {
    if (manager.has(folderKey(added))) {
      logger.warn(
        `Workspace folder '${added.name}' added but instance already exists; skipping`,
      );
      continue;
    }
    try {
      await manager.create(added);
    } catch (err) {
      logger.error(
        `Failed to start Rslint for added folder '${added.name}'`,
        err,
      );
    }
  }
}
