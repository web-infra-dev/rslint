import React from 'react';
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from 'recharts';

import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from './ui/chart';

const CHART_CONFIG = {
  percentPassing: {
    color: 'hsl(var(--chart))',
  },
} satisfies ChartConfig;

const DATE_FORMAT = new Intl.DateTimeFormat(undefined, {
  month: 'short',
  year: 'numeric',
});

function tooltipFormatter(_value, _name, item) {
  let data = item.payload;
  let gitHash = data.gitHash.slice(0, 7);
  let progress = `${data.passing} / ${data.total}`;
  return (
    <>
      <p>{`${new Date(data.date).toLocaleString('default', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      })}\n→ ${gitHash}`}</p>
      <p>{`${data.percent}%  (${progress})`}</p>
    </>
  );
}

function formatDateTick(dateSinceEpoch: number) {
  return DATE_FORMAT.format(dateSinceEpoch);
}

export default function TestPassingGraph({ graphData }) {
  // compute this manually, otherwise AreaChart will always start at zero
  const minPercent = Math.floor(Math.min(...graphData.map(d => d.percent)));
  return (
    <section
      aria-hidden
      aria-label="Chart showing the progress of tests passing for Turbopack over time"
      className="Graph"
    >
      <ChartContainer config={CHART_CONFIG} className="h-60 w-full">
        <AreaChart
          accessibilityLayer
          data={graphData}
          margin={{
            // avoids clipping decenders in text on the x axis
            bottom: 12,
          }}
        >
          <CartesianGrid
            vertical={false}
            stroke="var(--chart-axis)"
            strokeWidth={2}
            strokeDasharray="0.5rem"
            strokeOpacity={1}
            syncWithTicks
          />
          <XAxis
            dataKey="date"
            tickLine={false}
            axisLine={true}
            tickMargin={20}
            tickFormatter={formatDateTick}
            minTickGap={80}
            stroke="var(--chart-axis)"
            strokeWidth={2}
          />
          <YAxis
            type="number"
            tickLine={false}
            domain={[minPercent, 100]}
            allowDataOverflow
            stroke="var(--chart-axis)"
            strokeWidth={2}
            padding={{
              // avoids clipping the stroke at the 100% line
              top: 12,
              bottom: 12,
            }}
            unit="%"
          />
          <ChartTooltip
            cursor={false}
            formatter={tooltipFormatter}
            content={<ChartTooltipContent indicator="dot" />}
          />
          {/* Ideally the fill would use a linear gradient like shadcn/ui's
           example charts do, but that doesn't play well with allowDataOverflow,
           so just use a solid color instead. */}
          <Area
            dataKey="percent"
            type="stepAfter"
            fill="var(--chart-area)"
            fillOpacity={0.8}
            stroke="var(--chart)"
            strokeWidth={3}
            stackId="a"
            baseValue={50}
          />
        </AreaChart>
      </ChartContainer>
    </section>
  );
}
