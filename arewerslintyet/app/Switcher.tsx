'use client';
import { usePathname, useRouter } from 'next/navigation';
import { useEffect } from 'react';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

export default function Switcher() {
  const pathname = usePathname();
  const router = useRouter();
  const isProduction = pathname === '/';

  useEffect(() => {
    router.prefetch(pathname === 'development' ? '/' : '/build');
  }, [pathname, router.prefetch]);

  return (
    <Select
      onValueChange={value => {
        router.push(value === 'development' ? '/dev' : '/');
      }}
      defaultValue={isProduction ? 'production' : 'development'}
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
