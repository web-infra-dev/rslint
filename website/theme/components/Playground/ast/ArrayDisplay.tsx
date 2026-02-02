import React, { useState } from 'react';

interface ArrayDisplayProps<T> {
  label: string;
  items: T[];
  getItemPreview: (item: T) => string;
  renderItem: (item: T, index: number) => React.ReactNode;
  getItemKindName?: (item: T) => string | undefined; // For Node types
  getItemClickHandler?: (item: T) => (() => void) | undefined;
  defaultExpanded?: boolean;
}

export function ArrayDisplay<T>({
  label,
  items,
  getItemPreview,
  renderItem,
  getItemKindName,
  getItemClickHandler,
  defaultExpanded = false,
}: ArrayDisplayProps<T>) {
  const [expanded, setExpanded] = useState(defaultExpanded);
  const [expandedItems, setExpandedItems] = useState<Set<number>>(new Set());

  if (!items || items.length === 0) {
    return null;
  }

  const handleToggle = () => setExpanded(!expanded);

  const toggleItem = (index: number) => {
    setExpandedItems(prev => {
      const next = new Set(prev);
      if (next.has(index)) {
        next.delete(index);
      } else {
        next.add(index);
      }
      return next;
    });
  };

  return (
    <div className="font-mono text-xs leading-relaxed">
      <div className="flex items-center py-0.5">
        <span
          className="text-[10px] text-gray-500 w-4 flex-shrink-0 text-center cursor-pointer hover:bg-gray-200 rounded"
          onClick={handleToggle}
        >
          {expanded ? '▼' : '▶'}
        </span>
        <span
          className="text-purple-600 cursor-pointer hover:underline"
          onClick={handleToggle}
        >
          {label}
        </span>
        <span className="text-gray-500">:</span>
        {!expanded && (
          <span className="text-gray-400 ml-1">[{items.length} items]</span>
        )}
      </div>
      {expanded && (
        <div className="ml-4 border-l border-gray-200 pl-2">
          {items.map((item, index) => {
            const isItemExpanded = expandedItems.has(index);
            const clickHandler = getItemClickHandler?.(item);
            const kindName = getItemKindName?.(item);
            const handleItemToggle = () => toggleItem(index);
            return (
              <div key={index} className="my-0.5">
                <div className="flex items-center py-0.5">
                  <span
                    className="text-[10px] text-gray-500 w-4 flex-shrink-0 text-center cursor-pointer hover:bg-gray-200 rounded"
                    onClick={handleItemToggle}
                  >
                    {isItemExpanded ? '▼' : '▶'}
                  </span>
                  <span
                    className={`text-cyan-600 cursor-pointer hover:underline ${clickHandler ? 'hover:text-cyan-800' : ''}`}
                    onClick={
                      clickHandler
                        ? e => {
                            e.stopPropagation();
                            clickHandler();
                          }
                        : handleItemToggle
                    }
                  >
                    {index}
                  </span>
                  <span className="text-gray-500">:</span>
                  {kindName && (
                    <span
                      className="text-blue-600 ml-1 underline decoration-dotted cursor-pointer hover:decoration-solid"
                      onClick={handleItemToggle}
                    >
                      {kindName}
                    </span>
                  )}
                  {!isItemExpanded && (
                    <span className="text-gray-400 ml-1 truncate max-w-[300px]">
                      {'{'}
                      {getItemPreview(item)}
                      {'}'}
                    </span>
                  )}
                </div>
                {isItemExpanded && (
                  <div className="ml-4 py-1 border-l border-gray-200 pl-2">
                    {renderItem(item, index)}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
