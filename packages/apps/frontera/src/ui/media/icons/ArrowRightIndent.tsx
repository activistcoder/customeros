import React from 'react';

import { twMerge } from 'tailwind-merge';

interface IconProps extends React.SVGAttributes<SVGElement> {
  className?: string;
}

export const ArrowRightIndent = ({ className, ...props }: IconProps) => (
  <svg
    viewBox='0 0 24 24'
    fill='none'
    {...props}
    className={twMerge('inline-block size-4', className)}
  >
    <path
      d='M20 14H4V4M20 14L14 8M20 14L14 20'
      stroke='currentColor'
      strokeWidth='2'
      strokeLinecap='round'
      strokeLinejoin='round'
    />
  </svg>
);