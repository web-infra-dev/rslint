import HeatMapItem from './HeatMapItem';

function getTooltipContent(data) {
  let ruleName = data.test.slice(2);
  return `rule: "${ruleName}"`;
}

export function HeapMap({ lintResults }) {
  let index = 0;
  let testData = {};

  Object.keys(lintResults).forEach(status => {
    const value = lintResults[status];
    if (!value) return;
    value.split('\n\n').forEach(ruleGroup => {
      let lines = ruleGroup.replace(/\n$/, '').split('\n');
      let file = lines[0];
      let rules = lines.slice(1);
      if (!testData[file]) {
        testData[file] = {};
      }
      testData[file][status] = rules.map(test => {
        const tooltipContent = getTooltipContent({ file, test });
        return (
          <HeatMapItem
            key={index++}
            tooltipContent={tooltipContent}
            href={`https://github.com/rslint/rslint/blob/master/${file}`}
            isPassing={status === 'passing'}
          />
        );
      });
    });
  });

  let items = [];
  Object.keys(testData).forEach(file => {
    let testList = testData[file];
    items = items.concat(
      Object.keys(testList).map(status => {
        return testList[status];
      }),
    );
  });

  return <>{items}</>;
}
