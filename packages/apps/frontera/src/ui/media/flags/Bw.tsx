import React from 'react';

import { twMerge } from 'tailwind-merge';

interface IconProps extends React.SVGAttributes<SVGElement> {
  className?: string;
}

export const Bw = ({ className, ...props }: IconProps) => (
  <svg
    fill='none'
    viewBox='0 0 640 480'
    {...props}
    className={twMerge('inline-block size-4', className)}
  >
    <g fillRule='evenodd'>
      <path fill='#00cbff' d='M0 0h640v480H0z' />
      <path fill='#fff' d='M0 160h640v160H0z' />
      <path fill='#000001' d='M0 186h640v108H0z' />
    </g>
  </svg>
);
