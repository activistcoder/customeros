import React from 'react';

import { observer } from 'mobx-react-lite';

import { cn } from '@ui/utils/cn.ts';
import { useStore } from '@shared/hooks/useStore';
import { XSquare } from '@ui/media/icons/XSquare.tsx';
import { BracketsPlus } from '@ui/media/icons/BracketsPlus.tsx';
import { DotsVertical } from '@ui/media/icons/DotsVertical.tsx';
import {
  Menu,
  MenuItem,
  MenuList,
  MenuButton,
} from '@ui/overlay/Menu/Menu.tsx';

interface ServiceItemMenuProps {
  id: string;
  closed?: boolean;
  contractId: string;
  allowAddModification?: boolean;
  handleCloseService: (isClosed: boolean) => void;
}

export const ServiceItemMenu: React.FC<ServiceItemMenuProps> = observer(
  ({ id, contractId, allowAddModification, handleCloseService }) => {
    const store = useStore();
    const contractLineItemsStore = store.contractLineItems;

    return (
      <>
        <Menu>
          <MenuButton
            className={cn(
              `flex items-center max-h-5 p-1 py-2 hover:bg-gray-100 rounded translate-x-2`,
            )}
          >
            <DotsVertical className='text-gray-400' />
          </MenuButton>
          <MenuList align='end' side='bottom' className='p-0'>
            {allowAddModification && (
              <MenuItem
                onClick={() =>
                  contractLineItemsStore?.create({
                    id,
                    contractId,
                  })
                }
                className='flex items-center text-base'
              >
                <BracketsPlus className='mr-2 text-gray-500' />
                Add modification
              </MenuItem>
            )}

            <MenuItem
              onClick={() => handleCloseService(true)}
              className='flex items-center text-base'
            >
              <XSquare className='mr-2 text-gray-500' />
              End the service
            </MenuItem>
          </MenuList>
        </Menu>
      </>
    );
  },
);