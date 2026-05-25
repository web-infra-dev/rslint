import React from 'react';
import { Input } from '@components/ui/input';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from '@components/ui/select';
import { RotateCcw } from 'lucide-react';

export interface TableSelectorProps {
  onSearchChange: (value: string) => void;
  onGroupChange: (value: string) => void;
  onStatusChange: (value: string) => void;
  onPresetChange: (value: string) => void;
  searchValue?: string;
  groupValue?: string;
  statusValue?: string;
  presetValue?: string;
  groups?: string[];
  statuses?: Array<{ value: string; label: string }>;
  presetOptions?: Array<{ value: string; label: string }>;
}

export const CancelSymbol = 'cancel';

function SelectWithCancel(props: {
  value: string;
  onValueChange: (value: string) => void;
  placeholder: string;
  label: string;
  items: Array<{ value: string; label: string }>;
}) {
  return (
    <Select value={props.value} onValueChange={props.onValueChange}>
      <SelectTrigger className="w-[180px] cursor-pointer">
        <SelectValue placeholder={props.placeholder} />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          <SelectLabel>{props.label}</SelectLabel>
          {props.items.map((item) => (
            <SelectItem
              className="cursor-pointer"
              key={item.value}
              value={item.value}
            >
              {item.label}
            </SelectItem>
          ))}
          <SelectSeparator />
          <SelectItem className="cursor-pointer" value={CancelSymbol}>
            Cancel
          </SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
  );
}

export const TableSelector: React.FC<TableSelectorProps> = ({
  onSearchChange,
  onGroupChange,
  onStatusChange,
  onPresetChange,
  searchValue = '',
  groupValue = '',
  statusValue = '',
  presetValue = '',
  groups = ['typescript-eslint'],
  statuses = [
    { value: 'full', label: 'Full' },
    { value: 'partial-impl', label: 'Partial Implemented' },
    { value: 'partial-test', label: 'Partial Tested' },
    { value: 'total', label: 'Total' },
  ],
  presetOptions = [],
}) => {
  // Clear all selectors
  const handleClearAll = () => {
    onSearchChange('');
    onGroupChange('');
    onStatusChange('');
    onPresetChange('');
  };

  return (
    <div className="flex flex-row gap-2 items-end justify-between flex-wrap">
      <Input
        type="text"
        placeholder="Search rules..."
        className="w-full max-w-sm"
        value={searchValue}
        onChange={(e) => onSearchChange(e.target.value)}
      />

      <SelectWithCancel
        value={groupValue}
        onValueChange={onGroupChange}
        placeholder="Select a Group"
        label="Groups"
        items={groups.map((g) => ({ value: g, label: g }))}
      />

      <SelectWithCancel
        value={statusValue}
        onValueChange={onStatusChange}
        placeholder="Select a Status"
        label="Status"
        items={statuses}
      />

      {presetOptions.length > 0 && (
        <SelectWithCancel
          value={presetValue}
          onValueChange={onPresetChange}
          placeholder="Preset"
          label="Preset"
          items={presetOptions}
        />
      )}

      {/* cancel for selectors */}
      <div
        onClick={handleClearAll}
        className="h-10 w-10 flex items-center justify-center cursor-pointer hover:bg-gray-100 rounded-md transition-colors"
        title="Reset all filters"
      >
        <RotateCcw className="h-4 w-4 text-gray-500 hover:text-gray-700" />
      </div>
    </div>
  );
};
