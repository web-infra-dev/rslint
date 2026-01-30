import React, { useState, useCallback, useEffect, useRef } from 'react';

interface ResizableSplitPaneProps {
  left: React.ReactNode;
  right: React.ReactNode;
  storageKey?: string;
  defaultLeftWidth?: number;
  minLeftWidth?: number;
  minRightWidth?: number;
}

export const ResizableSplitPane: React.FC<ResizableSplitPaneProps> = ({
  left,
  right,
  storageKey = 'resizable-split-pane-width',
  defaultLeftWidth = 50,
  minLeftWidth = 20,
  minRightWidth = 20,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [leftWidthPercent, setLeftWidthPercent] = useState(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem(storageKey);
      if (saved) {
        const parsed = parseFloat(saved);
        if (
          !isNaN(parsed) &&
          parsed >= minLeftWidth &&
          parsed <= 100 - minRightWidth
        ) {
          return parsed;
        }
      }
    }
    return defaultLeftWidth;
  });
  const [isDragging, setIsDragging] = useState(false);

  // Save to localStorage when width changes
  useEffect(() => {
    if (typeof window !== 'undefined') {
      localStorage.setItem(storageKey, leftWidthPercent.toString());
    }
  }, [leftWidthPercent, storageKey]);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const handleMouseMove = useCallback(
    (e: MouseEvent) => {
      if (!isDragging || !containerRef.current) return;

      const containerRect = containerRef.current.getBoundingClientRect();
      const newLeftWidth =
        ((e.clientX - containerRect.left) / containerRect.width) * 100;

      // Clamp the value between min widths
      const clampedWidth = Math.max(
        minLeftWidth,
        Math.min(100 - minRightWidth, newLeftWidth),
      );
      setLeftWidthPercent(clampedWidth);
    },
    [isDragging, minLeftWidth, minRightWidth],
  );

  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
  }, []);

  // Add and remove event listeners
  useEffect(() => {
    if (isDragging) {
      window.addEventListener('mousemove', handleMouseMove);
      window.addEventListener('mouseup', handleMouseUp);
      // Prevent text selection while dragging
      document.body.style.userSelect = 'none';
      document.body.style.cursor = 'col-resize';
    }

    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
      document.body.style.userSelect = '';
      document.body.style.cursor = '';
    };
  }, [isDragging, handleMouseMove, handleMouseUp]);

  return (
    <div
      ref={containerRef}
      className="relative flex h-full w-full overflow-hidden"
    >
      {/* Left pane - min-w-0 prevents flex item from expanding beyond its container */}
      <div
        className="h-full min-w-0 overflow-auto"
        style={{ width: `calc(${leftWidthPercent}% - 2px)`, flexShrink: 0 }}
      >
        {left}
      </div>

      {/* Divider */}
      <div
        className={`relative h-full cursor-col-resize bg-gray-200 hover:bg-blue-400 ${
          isDragging ? 'bg-blue-500' : ''
        }`}
        style={{ width: '4px', flexShrink: 0 }}
        onMouseDown={handleMouseDown}
      >
        {/* Wider hit area for easier grabbing */}
        <div className="absolute inset-y-0 -left-1 -right-1" />
      </div>

      {/* Right pane - min-w-0 prevents flex item from expanding beyond its container */}
      <div className="h-full min-w-0 flex-1 overflow-auto">{right}</div>
    </div>
  );
};
