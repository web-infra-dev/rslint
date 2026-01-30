import React, { useState } from 'react';
import type { FlowInfo, FlowGraphNode } from './types';
import { PropRow } from './PropRow';
import { FlagsDisplay } from './FlagsDisplay';
import { LazyNodeView } from './LazyViews';
import { FlowGraphView } from './FlowGraphView';
import { ControlledCollapsibleItem } from './CollapsibleItem';
import { useAstInfoContextOptional } from './AstInfoContext';

interface FlowInfoViewProps {
  info?: FlowInfo;
}

export const FlowInfoView: React.FC<FlowInfoViewProps> = ({ info }) => {
  if (!info) {
    return (
      <div className="py-2 text-center text-xs text-gray-400">
        No flow information
      </div>
    );
  }

  return (
    <div className="font-mono text-xs leading-relaxed">
      {/* Properties */}
      <PropRow
        label="Flags"
        value={<FlagsDisplay flags={info.flags} names={info.flagNames} />}
      />

      {info.node && (
        <LazyNodeView
          pos={info.node.pos}
          label="Node"
          kindName={info.node.kindName?.replace('ast.', '')}
          preview={
            info.node.pos >= 0
              ? `pos: ${info.node.pos}, end: ${info.node.end}`
              : ''
          }
          shallow={info.node}
        />
      )}

      {info.antecedent && (
        <FlowInfoLazy
          label="Antecedent"
          preview={
            info.antecedent.flagNames
              ?.map(e => e.replace('ast.', ''))
              .join(' | ') || `flags: ${info.antecedent.flags}`
          }
          flow={info.antecedent}
        />
      )}

      {info.antecedents && info.antecedents.length > 0 && (
        <FlowInfoArray label="Antecedents" items={info.antecedents} />
      )}
    </div>
  );
};

// Helper to get title for top-level display
export const getFlowTitle = (info: FlowInfo) => {
  return `FlowNode[${info.flagNames?.map(e => e.replace('ast.', '')).join('|') || ''}]`;
};

// Helper to get preview for top-level display
export const getFlowPreview = (info: FlowInfo) => {
  return info.node ? `node: ${info.node.kindName?.replace('ast.', '')}` : '';
};

// Graph visualization component - to be rendered outside the collapsible wrapper
interface FlowGraphSectionProps {
  info?: FlowInfo;
}

export const FlowGraphSection: React.FC<FlowGraphSectionProps> = ({ info }) => {
  const context = useAstInfoContextOptional();

  if (!info?.graph) return null;

  const handleNodeHover = (node: FlowGraphNode | null) => {
    if (node && node.nodePos !== undefined && node.nodeEnd !== undefined) {
      context?.highlightRange(node.nodePos, node.nodeEnd);
    } else {
      context?.clearHighlight();
    }
  };

  return (
    <div className="mt-3 pt-3 border-t border-gray-200">
      <div className="text-gray-500 text-[10px] mb-2">
        Graph: {info.graph.nodes.length} nodes, {info.graph.edges.length} edges
      </div>
      <div className="border border-gray-200 rounded bg-gray-50 p-2">
        <FlowGraphView graph={info.graph} onNodeHover={handleNodeHover} />
      </div>
    </div>
  );
};

// Lazy flow info component - Flow doesn't have pos, so we just render it directly
interface FlowInfoLazyProps {
  label: string;
  preview: string;
  flow: FlowInfo;
}

const FlowInfoLazy: React.FC<FlowInfoLazyProps> = ({
  label,
  preview,
  flow,
}) => {
  const [expanded, setExpanded] = useState(false);
  const handleToggle = () => setExpanded(!expanded);
  const title = getFlowTitle(flow);

  return (
    <ControlledCollapsibleItem
      label={label}
      title={title}
      preview={preview}
      expanded={expanded}
      onToggle={handleToggle}
    >
      <FlowInfoView info={flow} />
    </ControlledCollapsibleItem>
  );
};

// Flow info array
interface FlowInfoArrayProps {
  label: string;
  items: FlowInfo[];
}

const FlowInfoArray: React.FC<FlowInfoArrayProps> = ({ label, items }) => {
  const [expanded, setExpanded] = useState(false);
  const handleToggle = () => setExpanded(!expanded);

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
        {expanded && <span className="text-gray-500 ml-1">{'['}</span>}
      </div>
      {expanded && (
        <>
          <div className="ml-4 border-l border-gray-200 pl-2">
            {items.map((item, index) =>
              item ? (
                <FlowInfoLazy
                  key={index}
                  label={String(index)}
                  preview={getFlowPreview(item)}
                  flow={item}
                />
              ) : null,
            )}
          </div>
          <div className="text-gray-500 ml-4">{']'}</div>
        </>
      )}
    </div>
  );
};
