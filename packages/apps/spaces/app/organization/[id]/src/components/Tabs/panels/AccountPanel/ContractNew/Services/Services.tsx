import React from 'react';

import { Plus } from '@ui/media/icons/Plus';
import { ServiceLineItem } from '@graphql/types';
import { IconButton } from '@ui/form/IconButton/IconButton';
import { ServicesList } from '@organization/src/components/Tabs/panels/AccountPanel/ContractNew/Services/ServicesList';

interface Props {
  onModalOpen: () => void;
  currency?: string | null;
  data?: Array<ServiceLineItem> | null;
}

export const Services: React.FC<Props> = ({ data, currency, onModalOpen }) => {
  return (
    <>
      <p className='w-full flex items-center justify-between'>
        {!data?.length && (
          <p className='text-sm font-semibold mt-2'>No services</p>
        )}

        {!data?.length && (
          <IconButton
            size='xs'
            variant='ghost'
            colorScheme='gray'
            aria-label={'Add services'}
            onClick={() => {
              onModalOpen();
            }}
            icon={<Plus boxSize='4' className='text-gray-400' />}
          />
        )}
      </p>

      {data?.length && (
        <ServicesList
          data={data}
          onModalOpen={onModalOpen}
          currency={currency}
        />
      )}
    </>
  );
};