import { getDevelopmentLintRuns } from './data';
import Graph from './Graph';

export default async function GraphDataDevelopment() {
  const { graphData } = await getDevelopmentLintRuns();

  return <Graph graphData={graphData} />;
}
