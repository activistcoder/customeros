import React, { useMemo } from 'react';

import { UseMutationResult } from '@tanstack/react-query';

import { cn } from '@ui/utils/cn';
import { useDisclosure } from '@ui/utils';
import { Edit03 } from '@ui/media/icons/Edit03';
import { DotLive } from '@ui/media/icons/DotLive';
import { XSquare } from '@ui/media/icons/XSquare';
import { RefreshCw02 } from '@ui/media/icons/RefreshCw02';
import { DotsVertical } from '@ui/media/icons/DotsVertical';
import { Exact, ContractStatus, ContractUpdateInput } from '@graphql/types';
import { Menu, MenuItem, MenuList, MenuButton } from '@ui/overlay/Menu/Menu';
import { GetContractsQuery } from '@organization/src/graphql/getContracts.generated';
import { UpdateContractMutation } from '@organization/src/graphql/updateContract.generated';
import {
  ContractEndModal,
  ContractStartModal,
} from '@organization/src/components/Tabs/panels/AccountPanel/ContractNew/ChangeContractStatusModals';

interface ContractStatusSelectProps {
  renewsAt?: string;
  contractId: string;
  status: ContractStatus;
  serviceStarted?: string;
  organizationName: string;
  nextInvoiceDate?: string;
  onOpenEditModal: () => void;

  onUpdateContract: UseMutationResult<
    UpdateContractMutation,
    unknown,
    Exact<{ input: ContractUpdateInput }>,
    { previousEntries: GetContractsQuery | undefined }
  >;
}

export const ContractMenu: React.FC<ContractStatusSelectProps> = ({
  status,
  renewsAt,
  contractId,
  organizationName,
  serviceStarted,
  onUpdateContract,
  nextInvoiceDate,
  onOpenEditModal,
}) => {
  const {
    onOpen: onOpenEndModal,
    onClose,
    isOpen,
  } = useDisclosure({
    id: 'end-contract-modal',
  });
  const {
    onOpen: onOpenStartModal,
    onClose: onCloseStartModal,
    isOpen: isStartModalOpen,
  } = useDisclosure({
    id: 'start-contract-modal',
  });

  const getStatusDisplay = useMemo(() => {
    let icon, text;
    switch (status) {
      case ContractStatus.Live:
        icon = <XSquare color='gray.500' mr={1} />;
        text = 'End contract...';
        break;
      case ContractStatus.Draft:
      case ContractStatus.Ended:
        icon = <DotLive color='gray.500' mr={1} />;
        text = 'Make live';
        break;
      case ContractStatus.OutOfContract:
        icon = <RefreshCw02 color='gray.500' mr={2} />;
        text = 'Renew contract';
        break;
      default:
        icon = null;
        text = null;
    }

    return (
      <>
        {icon}
        {text}
      </>
    );
  }, [status]);

  return (
    <>
      <Menu>
        <MenuButton
          className={cn(`flex items-center max-h-5 `, {
            'text-gray-800':
              status === ContractStatus.Draft ||
              status === ContractStatus.Ended,
            'text-primary-800': status === ContractStatus.Live,
            'text-warning-800': status === ContractStatus.OutOfContract,
            'border-gray-800':
              status === ContractStatus.Draft ||
              status === ContractStatus.Ended,
            'border-primary-800': status === ContractStatus.Live,
            'border-warning-800': status === ContractStatus.OutOfContract,
            'bg-gray-50':
              status === ContractStatus.Draft ||
              status === ContractStatus.Ended,
            'bg-primary-50': status === ContractStatus.Live,
            'bg-warning-50': status === ContractStatus.OutOfContract,
          })}
        >
          <DotsVertical color='gray.400' />
        </MenuButton>
        <MenuList align='end' side='bottom'>
          <MenuItem onClick={onOpenEditModal} className='flex items-center'>
            <Edit03 mr={2} color='gray.500' />
            Edit contract
          </MenuItem>
          <MenuItem
            className='flex items-center'
            onClick={
              status === ContractStatus.Live ? onOpenEndModal : onOpenStartModal
            }
          >
            {getStatusDisplay}
          </MenuItem>
        </MenuList>
      </Menu>

      <ContractEndModal
        isOpen={isOpen}
        onClose={onClose}
        contractId={contractId}
        organizationName={organizationName}
        renewsAt={renewsAt}
        serviceStarted={serviceStarted}
        onUpdateContract={onUpdateContract}
        nextInvoiceDate={nextInvoiceDate}
      />
      <ContractStartModal
        isOpen={isStartModalOpen}
        onClose={onCloseStartModal}
        contractId={contractId}
        organizationName={organizationName}
        serviceStarted={serviceStarted}
        onUpdateContract={onUpdateContract}
      />
    </>
  );
};