import { getLinter } from './bundler';
import { getProductionLintRuns } from './data';
import { ProgressBar } from './ProgressBar';

export default async function ProgressBarProduction() {
  const { mostRecent } = await getProductionLintRuns();

  if (!mostRecent) {
    return null;
  }

  return (
    <ProgressBar linter={getLinter()} mostRecent={mostRecent} dev={false} />
  );
}
