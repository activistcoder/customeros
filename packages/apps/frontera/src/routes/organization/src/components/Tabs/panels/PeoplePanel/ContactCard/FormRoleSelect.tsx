import { useRef } from 'react';

import { cn } from '@ui/utils/cn';
import { FormSelect } from '@ui/form/Select/FormSelect';
import { SelectOption } from '@shared/types/SelectOptions';
import { useOutsideClick } from '@ui/utils/hooks/useOutsideClick';
import {
  getMultiValueClassNames,
  getMultiValueLabelClassNames,
} from '@ui/form/Select';

import { RoleTag } from './RoleTag';

interface FormRoleSelectProps {
  name: string;
  formId: string;
  isFocused: boolean;
  placeholder?: string;
  isCardOpen?: boolean;
  data: SelectOption<string>[];
  setIsFocused: (isFocused: boolean) => void;
}

const options = [
  {
    value: 'Decision Maker',
    label: 'Decision Maker',
  },
  {
    value: 'Influencer',
    label: 'Influencer',
  },
  {
    value: 'User',
    label: 'User',
  },
  {
    value: 'Stakeholder',
    label: 'Stakeholder',
  },
  {
    value: 'Gatekeeper',
    label: 'Gatekeeper',
  },
  {
    value: 'Champion',
    label: 'Champion',
  },
  {
    value: 'Data Owner',
    label: 'Data Owner',
  },
];

export const FormRoleSelect = ({
  name,
  formId,
  isFocused,
  isCardOpen,
  placeholder,
  data,
  setIsFocused,
}: FormRoleSelectProps) => {
  const ref = useRef<HTMLDivElement>(null);

  useOutsideClick({
    ref,
    handler: () => setIsFocused(false),
  });
  if (isFocused) {
    return (
      <span onClick={(e) => e.stopPropagation()} ref={ref}>
        <FormSelect
          isMulti
          autoFocus
          menuIsOpen
          name={name}
          options={options}
          formId={formId}
          placeholder='Role'
          classNames={{
            multiValue: ({ data }) =>
              getMultiValueClassNames(
                cn({
                  'bg-gray-50 border-gray-200': data.label === 'Data Owner',
                  'bg-rose-50 border-rose-200': data.label === 'Stakeholder',
                  'bg-warning-50 border-warning-200':
                    data.label === 'Gatekeeper',
                  'bg-error-50 border-error-200': data.label === 'Champion',
                  'bg-primary-50 border-primary-200':
                    data.label === 'Decision Maker',
                  'bg-greenLight-50 border-greenLight-200':
                    data.label === 'Influencer',
                  'bg-blueDark-50 border-blueDark-200': data.label === 'User',
                  'border-[1px]': true,
                  'text-sm': true,
                }),
              ),
            multiValueLabel: ({ data }) =>
              getMultiValueLabelClassNames(
                cn({
                  'text-gray-700': data.label === 'Data Owner',
                  'text-rose-700': data.label === 'Stakeholder',
                  'text-warning-700': data.label === 'Gatekeeper',
                  'text-error-700': data.label === 'Champion',
                  'text-primary-700': data.label === 'Decision Maker',
                  'text-greenLight-700': data.label === 'Influencer',
                  'text-blueDark-700': data.label === 'User',
                }),
              ),
          }}
        />
      </span>
    );
  }

  if (!data.length) {
    return (
      <span
        className='hover:border-gray-300 border-b border-transparent cursor-text text-gray-400 transition-colors duration-200 ease-in-out'
        onClick={(e) => {
          if (isCardOpen) {
            e.stopPropagation();
          }
          setIsFocused(true);
        }}
      >
        {placeholder}
      </span>
    );
  }

  return (
    <div
      className='flex gap-1 mt-2 pb-2 flex-wrap'
      onClick={(e) => {
        if (isCardOpen) {
          e.stopPropagation();
        }
        setIsFocused(true);
      }}
    >
      {data.map((e) => (
        <RoleTag key={e.label} label={e.label} />
      ))}
    </div>
  );
};