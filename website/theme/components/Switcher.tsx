import React, { useState } from 'react';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './ui/select';

export default function Switcher() {
  const [environment, setEnvironment] = useState('production');

  return (
    <Select
      onValueChange={value => {
        setEnvironment(value);
        // In RSPress, we'll handle this differently or make it informational only
      }}
      defaultValue={environment}
    >
      <SelectTrigger className="min-w-34 bg-background">
        <SelectValue placeholder="" />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          <SelectItem value="development">Development</SelectItem>
          <SelectItem value="production">Production</SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
  );
}
