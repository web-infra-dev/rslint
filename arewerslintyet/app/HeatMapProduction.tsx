import { getProductionLintResults } from './data';
import { HeapMap } from './HeatMap';

export default async function HeatMapProduction() {
  const data = await getProductionLintResults();

  if (!data) {
    return null;
  }

  return (
    <section className="HeatMap">
      <HeapMap lintResults={data} />
    </section>
  );
}
