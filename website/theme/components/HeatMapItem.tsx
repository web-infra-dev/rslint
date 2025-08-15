import React, { useEffect } from 'react';
import { useTooltip } from './TooltipContext';

// Utility function to join class names
const twJoin = (...classes: (string | undefined | null | false)[]): string => {
  return classes.filter(Boolean).join(' ');
};

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
        'border border-gray-300 dark:border-gray-600 hover:border-gray-500 dark:hover:border-gray-400 cursor-pointer ' +
          'text-[0px] leading-[0px] w-[10px] h-[10px] overflow-hidden ' +
          'transition-all duration-200 hover:scale-110 hover:z-10 relative',
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
