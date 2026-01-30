import React, { useState } from 'react';
import type {
  TypeInfo,
  SymbolInfo,
  SignatureInfo,
  NodeInfo,
  NodeListMeta,
  IndexInfo,
} from './types';
import {
  SignatureInfoView,
  getSignatureTitle,
  getSignaturePreview,
} from './SignatureInfoView';
import { LazySymbolView } from './LazySymbolView';
import { LazyTypeView } from './LazyTypeView';
import { LazyNodeView } from './LazyNodeView';
import { ControlledCollapsibleItem } from './CollapsibleItem';

// Lazy Array Display for Type arrays
interface LazyTypeArrayProps {
  label: string;
  items: TypeInfo[];
}

export const LazyTypeArray: React.FC<LazyTypeArrayProps> = ({
  label,
  items,
}) => {
  const [expanded, setExpanded] = useState(false);

  if (!items || items.length === 0) return null;

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
                <LazyTypeView
                  key={index}
                  pos={item.pos}
                  label={String(index)}
                  preview={item.typeString}
                  shallow={item}
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

// Lazy Array Display for Symbol arrays
interface LazySymbolArrayProps {
  label: string;
  items: SymbolInfo[];
}

export const LazySymbolArray: React.FC<LazySymbolArrayProps> = ({
  label,
  items,
}) => {
  const [expanded, setExpanded] = useState(false);

  if (!items || items.length === 0) return null;

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
                <LazySymbolView
                  key={index}
                  pos={item.pos}
                  label={String(index)}
                  preview={`name: "${item.name}", flags: ${item.flags}`}
                  shallow={item}
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

// Signature Array Display - shows full info inline (no lazy loading since signatures can't be fetched by position)
interface LazySignatureArrayProps {
  label: string;
  items: SignatureInfo[];
}

export const LazySignatureArray: React.FC<LazySignatureArrayProps> = ({
  label,
  items,
}) => {
  const [expanded, setExpanded] = useState(false);

  if (!items || items.length === 0) return null;

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
                <InlineSignatureView
                  key={index}
                  label={String(index)}
                  info={item}
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

// Inline Signature View - displays full signature info without fetching
interface InlineSignatureViewProps {
  label: string;
  info: SignatureInfo;
}

const InlineSignatureView: React.FC<InlineSignatureViewProps> = ({
  label,
  info,
}) => {
  const [expanded, setExpanded] = useState(false);
  const handleToggle = () => setExpanded(!expanded);
  const title = getSignatureTitle(info);
  const preview = getSignaturePreview(info);

  return (
    <ControlledCollapsibleItem
      label={label}
      title={title}
      preview={preview}
      expanded={expanded}
      onToggle={handleToggle}
    >
      <SignatureInfoView info={info} />
    </ControlledCollapsibleItem>
  );
};

// Lazy Array Display for Node arrays
interface LazyNodeArrayProps {
  label: string;
  items: NodeInfo[];
  listMeta?: NodeListMeta;
}

export const LazyNodeArray: React.FC<LazyNodeArrayProps> = ({
  label,
  items,
  listMeta,
}) => {
  const [expanded, setExpanded] = useState(false);

  if (!items || items.length === 0) return null;

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
            {/* Array items first */}
            {items.map((item, index) =>
              item ? (
                <LazyNodeView
                  key={index}
                  pos={item.pos}
                  label={String(index)}
                  kindName={item.kindName?.replace('ast.', '')}
                  preview={
                    item.pos >= 0 ? `pos: ${item.pos}, end: ${item.end}` : ''
                  }
                  shallow={item}
                />
              ) : null,
            )}
            {/* Display list metadata after items */}
            {listMeta && (
              <>
                <div className="flex items-center py-0.5">
                  <span className="w-4 flex-shrink-0" />
                  <span className="text-purple-600">Pos</span>
                  <span className="text-gray-500">:</span>
                  <span className="text-gray-700 ml-1">{listMeta.pos}</span>
                </div>
                <div className="flex items-center py-0.5">
                  <span className="w-4 flex-shrink-0" />
                  <span className="text-purple-600">End</span>
                  <span className="text-gray-500">:</span>
                  <span className="text-gray-700 ml-1">{listMeta.end}</span>
                </div>
                <div className="flex items-center py-0.5">
                  <span className="w-4 flex-shrink-0" />
                  <span className="text-purple-600">HasTrailingComma</span>
                  <span className="text-gray-500">:</span>
                  <span
                    className={`ml-1 ${listMeta.hasTrailingComma ? 'text-green-600' : 'text-gray-500'}`}
                  >
                    {listMeta.hasTrailingComma ? 'true' : 'false'}
                  </span>
                </div>
              </>
            )}
          </div>
          <div className="text-gray-500 ml-4">{']'}</div>
        </>
      )}
    </div>
  );
};

// Index Info Array Display - shows index signature information
interface IndexInfoArrayProps {
  label: string;
  items: IndexInfo[];
}

export const IndexInfoArray: React.FC<IndexInfoArrayProps> = ({
  label,
  items,
}) => {
  const [expanded, setExpanded] = useState(false);

  if (!items || items.length === 0) return null;

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
                <IndexInfoView key={index} label={String(index)} info={item} />
              ) : null,
            )}
          </div>
          <div className="text-gray-500 ml-4">{']'}</div>
        </>
      )}
    </div>
  );
};

// Index Info View - displays a single index signature
interface IndexInfoViewProps {
  label: string;
  info: IndexInfo;
}

const IndexInfoView: React.FC<IndexInfoViewProps> = ({ label, info }) => {
  const [expanded, setExpanded] = useState(false);
  const handleToggle = () => setExpanded(!expanded);
  const title = 'IndexInfo';
  const preview = `[${info.keyType?.typeString ?? '?'}]: ${info.valueType?.typeString ?? '?'}${info.isReadonly ? ' (readonly)' : ''}`;

  return (
    <ControlledCollapsibleItem
      label={label}
      title={title}
      preview={preview}
      expanded={expanded}
      onToggle={handleToggle}
    >
      {info.isReadonly && (
        <div className="flex items-center py-0.5">
          <span className="w-4 flex-shrink-0" />
          <span className="text-purple-600">IsReadonly</span>
          <span className="text-gray-500">:</span>
          <span className="text-green-600 ml-1">true</span>
        </div>
      )}
      {info.keyType && (
        <LazyTypeView
          pos={info.keyType.pos}
          label="KeyType"
          preview={info.keyType.typeString}
          shallow={info.keyType}
        />
      )}
      {info.valueType && (
        <LazyTypeView
          pos={info.valueType.pos}
          label="ValueType"
          preview={info.valueType.typeString}
          shallow={info.valueType}
        />
      )}
    </ControlledCollapsibleItem>
  );
};
