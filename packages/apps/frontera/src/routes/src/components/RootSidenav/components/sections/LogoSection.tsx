import React from 'react';
import { useNavigate } from 'react-router-dom';

import { Image } from '@ui/media/Image/Image';
import { Skeleton } from '@ui/feedback/Skeleton';
import { useStore } from '@shared/hooks/useStore';
import { LogOut01 } from '@ui/media/icons/LogOut01';
import { Settings02 } from '@ui/media/icons/Settings02';
import { ChevronDown } from '@ui/media/icons/ChevronDown';
import { Menu, MenuItem, MenuList, MenuButton } from '@ui/overlay/Menu/Menu';

import logoCustomerOs from '../../../../../src/assets/customer-os-small.png';

export const LogoSection = () => {
  const store = useStore();
  const navigate = useNavigate();

  const isLoading = store.globalCache?.isLoading;

  const handleSignOutClick = () => {
    store.session.clearSession();

    if (store.demoMode) {
      window.location.reload();

      return;
    }
    navigate('/auth/signin');
  };

  return (
    <div className='px-2 pt-2 h-fit mb-2 ml-[10px] cursor-pointer flex justify-flex-start relative'>
      <Menu>
        <MenuButton>
          <div className='flex items-center gap-2'>
            {!isLoading ? (
              <>
                <Image
                  width={20}
                  height={20}
                  alt='CustomerOS'
                  className='logo-image'
                  src={
                    store.settings.tenant.value?.workspaceLogo || logoCustomerOs
                  }
                />
                <span className='font-semibold  text-start w-[120px] overflow-hidden text-ellipsis whitespace-nowrap'>
                  {store.settings.tenant.value?.workspaceName || 'CustomerOS'}
                </span>
                <ChevronDown className='size-3' />
              </>
            ) : (
              <Skeleton className='w-full h-8 mr-2' />
            )}
          </div>
        </MenuButton>
        <MenuList side='bottom' align='center' className='w-[180px] ml-2'>
          <MenuItem className='group' onClick={() => navigate('/settings')}>
            <div className='flex gap-2 items-center'>
              <Settings02 className='group-hover:text-gray-700 text-gray-500' />
              <span>Settings</span>
            </div>
          </MenuItem>
          <MenuItem className='group' onClick={handleSignOutClick}>
            <div className='flex gap-2 items-center'>
              <LogOut01 className='group-hover:text-gray-700 text-gray-500' />
              <span>Sign Out</span>
            </div>
          </MenuItem>
        </MenuList>
      </Menu>
    </div>
  );
};