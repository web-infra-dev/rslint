import React, { useEffect } from 'react';
import { twJoin } from 'tailwind-merge';
import { useTooltip } from './TooltipContext';

function HeatMapItem({ tooltipContent, href, isPassing }) {
  const { onMouseOver, onMouseOut } = useTooltip();
  const handleMouseOver = event => {
    onMouseOver(event, tooltipContent, isPassing ? 'passing' : 'failing');
  };

  return (
    // biome-ignore lint/a11y/useAnchorContent: aria-label is sufficient
    <a
      aria-label={`${tooltipContent} is ${isPassing ? 'passing' : 'failing'}`}
      className={twJoin(
        'border border-background hover:border-foreground cursor-default ' +
          'text-[0px] leading-[0px] w-[10px] h-[10px] overflow-hidden',
        isPassing ? 'bg-passing-square' : 'bg-failing-square',
      )}
      href={href}
      onFocus={handleMouseOver}
      onBlur={onMouseOut}
      onMouseOver={handleMouseOver}
      onMouseOut={onMouseOut}
    />
  );
}

export default React.memo(HeatMapItem);
