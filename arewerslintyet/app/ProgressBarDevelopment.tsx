import { getLinter } from './bundler';
import { getDevelopmentLintRuns } from './data';
import { ProgressBar } from './ProgressBar';

export default async function ProgressBarDevelopment() {
  const { mostRecent } = await getDevelopmentLintRuns();

  if (!mostRecent) {
    return null;
  }

  return (
    <ProgressBar linter={getLinter()} mostRecent={mostRecent} dev={true} />
  );
}
