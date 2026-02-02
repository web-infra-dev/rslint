import React, { useMemo, useState } from 'react';
import dagre from 'dagre';
import type { FlowGraph, FlowGraphNode } from './types';

interface FlowGraphViewProps {
  graph: FlowGraph;
  onNodeHover?: (node: FlowGraphNode | null) => void;
  onNodeClick?: (node: FlowGraphNode) => void;
}

// Node dimensions
const NODE_MIN_WIDTH = 100;
const NODE_HEIGHT = 50;
const NODE_PADDING = 16;
const NODE_MARGIN = 24;

// Extended dagre node type
interface DagreNode {
  x: number;
  y: number;
  width: number;
  height: number;
  node: FlowGraphNode;
}

// Clean up flag name (remove ast.FlowFlags prefix)
function cleanFlagName(name: string): string {
  return name.replace(/^ast\.FlowFlags/, '').replace(/^FlowFlags/, '');
}

// Get flow type label (shown at bottom)
function getFlowTypeLabel(node: FlowGraphNode): string {
  if (node.flagNames && node.flagNames.length > 0) {
    return node.flagNames.map(cleanFlagName).join(' | ');
  }
  return `flags: ${node.flags}`;
}

// Get text/expression label (shown at top)
function getTextLabel(node: FlowGraphNode): string {
  return node.nodeText || '';
}

function getTooltipContent(node: FlowGraphNode): string {
  const lines: string[] = [];

  // Text
  if (node.nodeText) {
    lines.push(`Text: "${node.nodeText}"`);
  }

  // Flags
  if (node.flagNames && node.flagNames.length > 0) {
    lines.push(`Flags: ${node.flagNames.join(' | ')}`);
  } else {
    lines.push(`Flags: ${node.flags}`);
  }

  // Node info
  if (node.nodeKindName) {
    lines.push(`Kind: ${node.nodeKindName}`);
  }
  if (node.nodePos !== undefined && node.nodeEnd !== undefined) {
    lines.push(`Range: [${node.nodePos}, ${node.nodeEnd})`);
  }

  return lines.join('\n');
}

// Estimate text width (rough approximation for monospace)
function estimateTextWidth(text: string, fontSize: number): number {
  return text.length * fontSize * 0.6;
}

export const FlowGraphView: React.FC<FlowGraphViewProps> = ({
  graph,
  onNodeHover,
  onNodeClick,
}) => {
  const [hoveredNode, setHoveredNode] = useState<FlowGraphNode | null>(null);

  // Build layout using dagre
  const layout = useMemo(() => {
    if (!graph || graph.nodes.length === 0) return null;

    const g = new dagre.graphlib.Graph();
    g.setGraph({
      rankdir: 'BT', // Bottom to top (antecedents point upward)
      nodesep: NODE_MARGIN,
      ranksep: NODE_MARGIN * 1.5,
      marginx: NODE_MARGIN,
      marginy: NODE_MARGIN,
    });
    g.setDefaultEdgeLabel(() => ({}));

    // Add nodes with dynamic width
    for (const node of graph.nodes) {
      const flowType = getFlowTypeLabel(node);
      const textLabel = getTextLabel(node);
      const flowTypeWidth = estimateTextWidth(flowType, 11);
      const textWidth = estimateTextWidth(textLabel, 12);
      const width = Math.max(
        NODE_MIN_WIDTH,
        Math.max(flowTypeWidth, textWidth) + NODE_PADDING * 2,
      );

      g.setNode(String(node.id), {
        width,
        height: NODE_HEIGHT,
        node,
      });
    }

    // Add edges (from antecedent to current)
    for (const edge of graph.edges) {
      g.setEdge(String(edge.from), String(edge.to));
    }

    // Calculate layout
    dagre.layout(g);

    // Extract positioned nodes and edges
    const nodes: Array<{
      x: number;
      y: number;
      width: number;
      node: FlowGraphNode;
    }> = [];
    const edges: Array<{ points: Array<{ x: number; y: number }> }> = [];

    g.nodes().forEach(id => {
      const n = g.node(id) as DagreNode | undefined;
      if (n) {
        nodes.push({ x: n.x, y: n.y, width: n.width, node: n.node });
      }
    });

    g.edges().forEach(e => {
      const edge = g.edge(e);
      if (edge && edge.points) {
        edges.push({ points: edge.points });
      }
    });

    // Calculate SVG dimensions
    const graphInfo = g.graph();
    const width = (graphInfo.width || 200) + NODE_MARGIN * 2;
    const height = (graphInfo.height || 200) + NODE_MARGIN * 2;

    return { nodes, edges, width, height };
  }, [graph]);

  if (!layout) {
    return (
      <div className="text-center text-xs text-gray-400 py-4">
        No flow graph data
      </div>
    );
  }

  const handleMouseEnter = (node: FlowGraphNode) => {
    setHoveredNode(node);
    onNodeHover?.(node);
  };

  const handleMouseLeave = () => {
    setHoveredNode(null);
    onNodeHover?.(null);
  };

  return (
    <div className="flow-graph-container overflow-auto relative">
      <svg
        width={layout.width}
        height={layout.height}
        className="flow-graph-svg"
        style={{
          fontFamily:
            'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
        }}
      >
        {/* Arrow marker definition */}
        <defs>
          <marker
            id="flow-arrowhead"
            markerWidth="8"
            markerHeight="6"
            refX="7"
            refY="3"
            orient="auto"
          >
            <polygon points="0 0, 8 3, 0 6" fill="#666" />
          </marker>
        </defs>

        {/* Edges */}
        {layout.edges.map((edge, i) => {
          if (edge.points.length < 2) return null;
          const pathData = edge.points
            .map((p, j) => `${j === 0 ? 'M' : 'L'} ${p.x} ${p.y}`)
            .join(' ');
          return (
            <path
              key={i}
              d={pathData}
              fill="none"
              stroke="#666"
              strokeWidth="1"
              markerEnd="url(#flow-arrowhead)"
            />
          );
        })}

        {/* Nodes */}
        {layout.nodes.map(({ x, y, width, node }) => {
          const flowType = getFlowTypeLabel(node);
          const textLabel = getTextLabel(node);
          const isHovered = hoveredNode?.id === node.id;
          const hasText = textLabel.length > 0;

          return (
            <g
              key={node.id}
              transform={`translate(${x - width / 2}, ${y - NODE_HEIGHT / 2})`}
              className="cursor-pointer"
              onMouseEnter={() => handleMouseEnter(node)}
              onMouseLeave={handleMouseLeave}
              onClick={() => onNodeClick?.(node)}
            >
              <title>{getTooltipContent(node)}</title>
              {/* Node background */}
              <rect
                width={width}
                height={NODE_HEIGHT}
                rx="3"
                fill={isHovered ? '#f0f9ff' : '#fff'}
                stroke={isHovered ? '#3b82f6' : '#333'}
                strokeWidth={isHovered ? 1.5 : 1}
              />
              {/* Text/expression label (top) */}
              {hasText && (
                <text
                  x={width / 2}
                  y={18}
                  textAnchor="middle"
                  fill="#111"
                  fontSize="12"
                  fontWeight="500"
                >
                  {textLabel}
                </text>
              )}
              {/* Separator line */}
              {hasText && (
                <line
                  x1={8}
                  y1={NODE_HEIGHT / 2}
                  x2={width - 8}
                  y2={NODE_HEIGHT / 2}
                  stroke="#ddd"
                  strokeWidth="1"
                />
              )}
              {/* Flow type label (bottom) */}
              <text
                x={width / 2}
                y={hasText ? NODE_HEIGHT - 12 : NODE_HEIGHT / 2}
                textAnchor="middle"
                dominantBaseline={hasText ? 'auto' : 'middle'}
                fill="#666"
                fontSize="11"
              >
                {flowType}
              </text>
            </g>
          );
        })}
      </svg>
    </div>
  );
};
