import React, { useState } from 'react';

// Simple tooltip component using CSS
const Tooltip: React.FC<{ text: string; children: React.ReactNode }> = ({ text, children }) => (
  <span className="relative group">
    {children}
    <span className="absolute left-0 bottom-full mb-1 px-2 py-1 text-xs text-white bg-gray-800 rounded whitespace-pre-line opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-50">
      {text}
    </span>
  </span>
);

// Enum field configuration - maps field name to enum name
// Note: We support both camelCase and PascalCase field names for flexibility
export const ENUM_FIELDS: Record<string, string> = {
  // camelCase variants (used by Node info)
  kind: 'SyntaxKind',
  flags: 'NodeFlags',
  modifierFlags: 'ModifierFlags',
  symbolFlags: 'SymbolFlags',
  typeFlags: 'TypeFlags',
  // PascalCase variants (used by Type/Symbol/FlowNode info from Go)
  Flags: 'Flags',
  ObjectFlags: 'ObjectFlags',
  CheckFlags: 'CheckFlags',
};

// Value to enum key mapping (passed from parent with actual resolved names)
export interface EnumValueMap {
  [fieldName: string]: string; // fieldName -> resolved enum key string
}

interface ObjectTreeProps {
  data: unknown;
  name?: string;
  defaultExpanded?: boolean;
  enumValues?: EnumValueMap;
}

interface TreeNodeProps {
  name: string;
  value: unknown;
  defaultExpanded?: boolean;
  depth?: number;
  enumValues?: EnumValueMap;
}

function getValueType(value: unknown): string {
  if (value === null) return 'null';
  if (value === undefined) return 'undefined';
  if (Array.isArray(value)) return 'array';
  return typeof value;
}

function formatString(str: string): string {
  return str
    .replace(/\\/g, '\\\\')
    .replace(/\n/g, '\\n')
    .replace(/\r/g, '\\r')
    .replace(/\t/g, '\\t');
}

function getValuePreview(value: unknown, maxLen = 50): string {
  const type = getValueType(value);
  switch (type) {
    case 'null':
      return 'null';
    case 'undefined':
      return 'undefined';
    case 'string': {
      const formatted = formatString(value as string);
      return `"${formatted.length > maxLen ? formatted.slice(0, maxLen) + '...' : formatted}"`;
    }
    case 'number':
    case 'boolean':
      return String(value);
    case 'array':
      return `Array(${(value as unknown[]).length})`;
    case 'object':
      return `{${Object.keys(value as object).length} keys}`;
    case 'function':
      return 'ƒ()';
    default:
      return String(value);
  }
}

function isExpandable(value: unknown): boolean {
  if (value === null || value === undefined) return false;
  const type = typeof value;
  return type === 'object' || Array.isArray(value);
}

const TreeNode: React.FC<TreeNodeProps> = ({
  name,
  value,
  defaultExpanded = false,
  depth = 0,
  enumValues,
}) => {
  const [expanded, setExpanded] = useState(defaultExpanded);
  const type = getValueType(value);
  const expandable = isExpandable(value);

  // Check if this field is an enum field
  const enumName = ENUM_FIELDS[name];
  const enumKey = enumValues?.[name];
  // Don't show tooltip for flags with value 0
  const flagFields = ['flags', 'modifierFlags', 'symbolFlags', 'typeFlags', 'Flags', 'ObjectFlags', 'CheckFlags'];
  const isZeroFlag = flagFields.includes(name) && value === 0;
  const hasEnumInfo = !!(enumName && enumKey) && !isZeroFlag;
  // For flags, show the resolved flag names directly; for kind/other, show enum prefix
  const tooltipText = hasEnumInfo
    ? (flagFields.includes(name) ? enumKey : `${enumName}.${enumKey}`)
    : undefined;

  const toggleExpand = () => {
    if (expandable) {
      setExpanded(!expanded);
    }
  };

  const renderValue = () => {
    if (!expandable) {
      const preview = getValuePreview(value);
      const valueSpan = (
        <span
          className={`whitespace-pre-wrap break-all ${
            type === 'string' ? 'text-blue-700' :
            type === 'number' || type === 'boolean' ? 'text-blue-600' :
            type === 'null' || type === 'undefined' ? 'text-gray-400 italic' :
            'text-gray-600'
          } ${hasEnumInfo ? 'underline decoration-dotted cursor-help' : ''}`}
        >
          {preview}
        </span>
      );

      if (hasEnumInfo && tooltipText) {
        return <Tooltip text={tooltipText}>{valueSpan}</Tooltip>;
      }
      return valueSpan;
    }

    if (!expanded) {
      return (
        <span className="ml-1 text-gray-400">
          {getValuePreview(value)}
        </span>
      );
    }

    const entries = Array.isArray(value)
      ? (value as unknown[]).map((v, i) => [String(i), v] as [string, unknown])
      : Object.entries(value as object);

    return (
      <div className="ml-3.5 border-l border-gray-200 pl-2">
        {entries.map(([key, val]) => (
          <TreeNode
            key={key}
            name={key}
            value={val}
            defaultExpanded={false}
            depth={depth + 1}
            enumValues={enumValues}
          />
        ))}
      </div>
    );
  };

  return (
    <div>
      <div
        className={`flex items-start py-0.5 ${expandable ? 'cursor-pointer hover:bg-gray-100 rounded' : ''}`}
        onClick={toggleExpand}
      >
        {expandable ? (
          <span className={`inline-flex items-center justify-center w-3.5 h-[18px] text-[8px] text-gray-500 transition-transform duration-100 flex-shrink-0 ${expanded ? 'rotate-90' : ''}`}>
            ▶
          </span>
        ) : (
          <span className="w-3.5 flex-shrink-0" />
        )}
        <span className="text-amber-700">{name}</span>
        <span className="text-gray-800 mr-1">:</span>
        {!expanded && renderValue()}
      </div>
      {expanded && expandable && renderValue()}
    </div>
  );
};

export const ObjectTree: React.FC<ObjectTreeProps> = ({
  data,
  name,
  defaultExpanded = true,
  enumValues,
}) => {
  const [expanded, setExpanded] = useState(defaultExpanded);
  const expandable = isExpandable(data);
  const type = getValueType(data);

  if (!expandable) {
    return (
      <div className="font-mono text-xs leading-relaxed">
        <div className="flex items-start py-0.5">
          {name && (
            <>
              <span className="w-3.5 flex-shrink-0" />
              <span className="text-amber-700">{name}</span>
              <span className="text-gray-800 mr-1">:</span>
            </>
          )}
          <span
            className={`whitespace-pre-wrap break-all ${
              type === 'string' ? 'text-blue-700' :
              type === 'number' || type === 'boolean' ? 'text-blue-600' :
              'text-gray-400 italic'
            }`}
          >
            {getValuePreview(data)}
          </span>
        </div>
      </div>
    );
  }

  const entries = Array.isArray(data)
    ? (data as unknown[]).map((v, i) => [String(i), v] as [string, unknown])
    : Object.entries(data as object);

  return (
    <div className="font-mono text-xs leading-relaxed">
      <div
        className="flex items-start py-0.5 cursor-pointer hover:bg-gray-100 rounded"
        onClick={() => setExpanded(!expanded)}
      >
        <span className={`inline-flex items-center justify-center w-3.5 h-[18px] text-[8px] text-gray-500 transition-transform duration-100 flex-shrink-0 ${expanded ? 'rotate-90' : ''}`}>
          ▶
        </span>
        {name && (
          <>
            <span className="text-amber-700">{name}</span>
            <span className="text-gray-800 mr-1">:</span>
          </>
        )}
        {!expanded && (
          <span className="ml-1 text-gray-400">{getValuePreview(data)}</span>
        )}
      </div>
      {expanded && (
        <div className="ml-3.5 border-l border-gray-200 pl-2">
          {entries.map(([key, val]) => (
            <TreeNode
              key={key}
              name={key}
              value={val}
              defaultExpanded={false}
              depth={1}
              enumValues={enumValues}
            />
          ))}
        </div>
      )}
    </div>
  );
};
