import React from 'react';

import { twMerge } from 'tailwind-merge';

interface IconProps extends React.SVGAttributes<SVGElement> {
  className?: string;
}

export const Gd = ({ className, ...props }: IconProps) => (
  <svg
    fill='none'
    viewBox='0 0 640 480'
    {...props}
    className={twMerge('inline-block size-4', className)}
  >
    <defs>
      <g id='gd-c'>
        <g id='gd-b'>
          <path
            id='gd-a'
            d='M0-1v1h.5'
            fill='#fcd116'
            transform='rotate(18 0 -1)'
          />
          <use xlinkHref='#gd-a' transform='scale(-1 1)' />
        </g>
        <use xlinkHref='#gd-b' transform='rotate(72)' />
        <use xlinkHref='#gd-b' transform='rotate(144)' />
        <use xlinkHref='#gd-b' transform='rotate(216)' />
        <use xlinkHref='#gd-b' transform='rotate(288)' />
      </g>
    </defs>
    <path fill='#ce1126' d='M0 0h640v480H0z' />
    <path fill='#007a5e' d='M67.2 67.2h505.6v345.6H67.2z' />
    <path fill='#fcd116' d='M67.2 67.3h505.6L67.2 412.9h505.6z' />
    <circle r='57.6' cx='319.9' cy='240.1' fill='#ce1126' />
    <use
      width='100%'
      height='100%'
      xlinkHref='#gd-c'
      transform='translate(320 240)scale(52.8)'
    />
    <use
      x='-100'
      width='100%'
      height='100%'
      xlinkHref='#gd-d'
      transform='translate(-30.3)'
    />
    <use
      id='gd-d'
      width='100%'
      height='100%'
      xlinkHref='#gd-c'
      transform='translate(320 33.6)scale(31.2)'
    />
    <use
      x='100'
      width='100%'
      height='100%'
      xlinkHref='#gd-d'
      transform='translate(30.3)'
    />
    <path
      fill='#ce1126'
      d='M102.3 240.7a80.4 80.4 0 0 0 33.5 33.2 111 111 0 0 0-11.3-45z'
    />
    <path
      fill='#fcd116'
      d='M90.1 194.7c10.4 21.7-27.1 73.7 35.5 85.9a63.2 63.2 0 0 1-10.9-41.9 70 70 0 0 1 32.5 30.8c16.4-59.5-42-55.8-57.1-74.8'
    />
    <use
      x='-100'
      width='100%'
      height='100%'
      xlinkHref='#gd-d'
      transform='translate(-30.3 414.6)'
    />
    <use
      width='100%'
      height='100%'
      xlinkHref='#gd-c'
      transform='translate(320 448.2)scale(31.2)'
    />
    <use
      x='100'
      width='100%'
      height='100%'
      xlinkHref='#gd-d'
      transform='translate(30.3 414.6)'
    />
  </svg>
);
