'use client';
import { useRouter, useSearchParams } from 'next/navigation';

import { getGraphQLClient } from '@shared/util/getGraphQLClient';
import { useGetInvoiceQuery } from '@shared/graphql/getInvoice.generated';
import { InvoicePreviewModalContent } from '@shared/components/Invoice/InvoicePreviewModal';
import {
  Modal,
  ModalPortal,
  ModalContent,
  ModalOverlay,
} from '@ui/overlay/Modal/Modal';

export const Preview = () => {
  const router = useRouter();
  const client = getGraphQLClient();
  const searchParams = useSearchParams();
  const invoiceId = searchParams?.get('preview');

  const onOpenChange = () => {
    const newParams = new URLSearchParams(searchParams?.toString());
    newParams.delete('preview');
    window.history.pushState({}, '', `?${newParams.toString()}`);
    router.push(`?${newParams.toString()}`);
  };

  const { data, isLoading, isError } = useGetInvoiceQuery(
    client,
    {
      id: invoiceId ?? '',
    },
    {
      enabled: !!invoiceId,
    },
  );

  return (
    <Modal open={!!invoiceId} onOpenChange={onOpenChange}>
      <ModalPortal>
        <ModalOverlay />
        {/* width and height of A4 */}
        <ModalContent className='max-w-[794px] h-[1123px]'>
          <InvoicePreviewModalContent
            data={data}
            isError={isError}
            isFetching={isLoading}
          />
        </ModalContent>
      </ModalPortal>
    </Modal>
  );
};