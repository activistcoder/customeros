import { File04 } from '@ui/media/icons/File04';
import { FeaturedIcon } from '@ui/media/Icon/FeaturedIcon2';
import HalfCirclePattern from '@shared/assets/HalfCirclePattern';

export const Empty = () => {
  return (
    <div className='h-[100vh] mx-auto flex items-center'>
      <div className='flex flex-col'>
        <div className='flex relative'>
          <FeaturedIcon
            size='lg'
            colorScheme='primary'
            className='absolute top-[24%] right-[46%]'
          >
            <File04 boxSize='5' />
          </FeaturedIcon>
          <HalfCirclePattern />
        </div>
        <div className='flex flex-col text-center items-center translate-y-[-120px]'>
          <p className='text-gray-900 text-md font-semibold'>
            Awaiting your invoices
          </p>
          <p className='text-sm text-gray-600 my-1'>
            Create your first contract with services. Once issued, <br />
            invoices will appear here.
          </p>
        </div>
      </div>
    </div>
  );
};