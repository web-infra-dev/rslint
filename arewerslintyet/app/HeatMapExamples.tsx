import { getRuleExamplesResults } from './data';
import HeatMapItem from './HeatMapItem';

export async function HeapMapExamples() {
  const ruleExamplesResult = await getRuleExamplesResults();

  if (!ruleExamplesResult || Object.keys(ruleExamplesResult).length === 0) {
    return null;
  }

  let items = [];
  for (const ruleName in ruleExamplesResult) {
    const isPassing = ruleExamplesResult[ruleName];
    items.push(
      <HeatMapItem
        key={ruleName}
        tooltipContent={ruleName}
        href={`https://github.com/rslint/rslint/blob/master/crates/rslint_core/src/groups/rules/${ruleName}.rs`}
        isPassing={isPassing}
      />,
    );
  }

  return (
    <>
      <h2 className="text-4xl my-2">Rule Examples</h2>
      <section className="HeatMap">{items}</section>
    </>
  );
}
