import React from 'react';

import { twMerge } from 'tailwind-merge';

interface IconProps extends React.SVGAttributes<SVGElement> {
  className?: string;
}

export const Ye = ({ className, ...props }: IconProps) => (
  <svg
    viewBox='0 0 640 480'
    fill='none'
    {...props}
    className={twMerge('inline-block size-4', className)}
  >
    <g fillRule='evenodd' strokeWidth='1pt'>
      <path fill='#fff' d='M0 0h640v472.8H0z' />
      <path fill='#f10600' d='M0 0h640v157.4H0z' />
      <path fill='#000001' d='M0 322.6h640V480H0z' />
    </g>
  </svg>
);