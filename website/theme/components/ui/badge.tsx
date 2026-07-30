import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';

import { cn } from '@/theme/lib/utils';

const badgeVariants = cva(
  'inline-flex w-fit shrink-0 items-center justify-center rounded-xs px-2 py-1 text-xs font-medium whitespace-nowrap',
  {
    variants: {
      variant: {
        secondary:
          'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
        version:
          'bg-emerald-200/80 text-black dark:bg-emerald-700/80 dark:text-white',
        unreleased: 'bg-amber-100 text-black dark:bg-zinc-700 dark:text-white',
      },
    },
    defaultVariants: {
      variant: 'secondary',
    },
  },
);

function Badge({
  className,
  variant,
  ...props
}: React.ComponentProps<'span'> & VariantProps<typeof badgeVariants>) {
  return (
    <span
      data-slot="badge"
      className={cn(badgeVariants({ variant }), className)}
      {...props}
    />
  );
}

export { Badge, badgeVariants };
