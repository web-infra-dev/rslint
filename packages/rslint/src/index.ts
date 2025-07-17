import { LintResponse, RSLintService } from './service.ts'
export async function lint(tsconfig: string): Promise<LintResponse> {
  const service = new RSLintService();
  const result = await service.lint({
    tsconfig
  })
  await service.close();
  return result;
}

