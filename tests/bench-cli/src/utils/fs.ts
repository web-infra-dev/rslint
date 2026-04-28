import { stat } from 'node:fs/promises';

export async function assertExists(filePath: string): Promise<void> {
  try {
    await stat(filePath);
  } catch (error) {
    throw new Error(`Missing benchmark fixture: ${filePath}`, { cause: error });
  }
}
