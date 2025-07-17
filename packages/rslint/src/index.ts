
import { LintOptions,RSLintService } from "./service.js";
export async function lint(options: LintOptions) {
  const service = new RSLintService();
  try {
    return await service.lint(options);
  } finally {
    await service.close();
  }
}
