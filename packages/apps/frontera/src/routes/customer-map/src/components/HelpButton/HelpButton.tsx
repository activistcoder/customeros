import { cn } from '@ui/utils/cn';
import { HelpCircle } from '@ui/media/icons/HelpCircle';
import { IconButton } from '@ui/form/IconButton/IconButton';

interface HelpButtonProps {
  isOpen: boolean;
  onOpen: () => void;
}

export const HelpButton = ({ isOpen, onOpen }: HelpButtonProps) => {
  return (
    <IconButton
      size='xs'
      variant='ghost'
      onClick={onOpen}
      aria-label='Help'
      className={cn(
        isOpen ? 'opacity-100' : 'opacity-0  group-hover:opacity-100',
      )}
      icon={<HelpCircle className='text-gray-500' />}
    />
  );
};