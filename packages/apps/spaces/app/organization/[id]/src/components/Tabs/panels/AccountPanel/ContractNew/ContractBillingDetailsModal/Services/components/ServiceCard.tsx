import React, { useState } from 'react';

import { observer } from 'mobx-react-lite';

import { cn } from '@ui/utils/cn';
import { FlipBackward } from '@ui/media/icons/FlipBackward';
import { IconButton } from '@ui/form/IconButton/IconButton';
import { ChevronExpand } from '@ui/media/icons/ChevronExpand';
import { ResizableInput } from '@ui/form/Input/ResizableInput';
import { ChevronCollapse } from '@ui/media/icons/ChevronCollapse';
import { Card, CardHeader, CardContent } from '@ui/presentation/Card/Card';
import ServiceLineItemStore from '@organization/src/components/Tabs/panels/AccountPanel/ContractNew/ContractBillingDetailsModal/stores/Service.store';
import { Highlighter } from '@organization/src/components/Tabs/panels/AccountPanel/ContractNew/ContractBillingDetailsModal/Services/components/highlighters';

import { ServiceItem } from './ServiceItem';
import { ServiceItemMenu } from './ServiceItemMenu';

interface ServiceCardProps {
  currency?: string;
  data: ServiceLineItemStore[];
  type: 'subscription' | 'one-time';
}

export const ServiceCard: React.FC<ServiceCardProps> = observer(
  ({ data, type, currency }) => {
    const [showEnded, setShowEnded] = useState(false);

    const endedServices = data.filter((service) => {
      return service.serviceLineItem?.serviceEnded;
    });

    const liveServices = data.filter(
      (service) => !service.serviceLineItem?.serviceEnded,
    );
    const [description, setDescription] = useState(
      liveServices[0].serviceLineItem?.description || '',
    );

    const isClosed = liveServices[0].serviceLineItem?.closedVersion;

    const handleDescriptionChange = () => {
      liveServices.forEach((service) => {
        service.updateDescription(description);
      });
    };
    const handleCloseChange = (closed: boolean) => {
      liveServices.forEach((service) => {
        service.setIsClosedVersion(closed);
      });
    };

    const descriptionLI = liveServices[0];

    return (
      <Card className='px-3 py-2 mb-2'>
        <CardHeader className={cn('flex justify-between')}>
          <Highlighter
            highlightVersion={descriptionLI?.uiMetadata?.shapeVariant}
            backgroundColor={
              liveServices.length === 1 &&
              descriptionLI?.isNewlyAdded &&
              !isClosed
                ? descriptionLI.uiMetadata?.color
                : undefined
            }
          >
            <ResizableInput
              value={description ?? 'Unnamed'}
              onChange={(e) => setDescription(e.target.value)}
              onBlur={handleDescriptionChange}
              onFocus={(e) => e.target.select()}
              size='xs'
              className={cn(
                'text-base text-gray-500 min-w-2.5 min-h-0 max-h-4 border-none hover:border-none focus:border-none ',
                {
                  'text-gray-400 line-through': isClosed,
                },
              )}
            />
          </Highlighter>

          <div className='flex items-baseline'>
            {endedServices.length > 0 && (
              <IconButton
                aria-label={
                  showEnded ? 'Hide ended services' : 'Show ended services'
                }
                icon={
                  showEnded ? (
                    <ChevronCollapse className='text-inherit' />
                  ) : (
                    <ChevronExpand className='text-inherit' />
                  )
                }
                variant='ghost'
                size='xs'
                className='p-0 px-1 text-gray-400'
                onClick={() => setShowEnded(!showEnded)}
              />
            )}

            {isClosed ? (
              <>
                <IconButton
                  aria-label='Undo'
                  icon={<FlipBackward className='text-gray-400' />}
                  size='xs'
                  className='p-1  max-h-5 hover:bg-gray-100 rounded translate-x-1'
                  variant='ghost'
                  onClick={() => handleCloseChange(false)}
                />
              </>
            ) : (
              <ServiceItemMenu
                id={data[0]?.serviceLineItem?.parentId || ''}
                closed={data[0]?.serviceLineItem?.closedVersion}
                type={type}
                handleCloseService={handleCloseChange}
                allowAddModification={
                  !data.some((e) => e?.serviceLineItem?.isNew)
                }
              />
            )}
          </div>
        </CardHeader>
        <CardContent className='text-sm p-0 gap-y-0.25 flex flex-col'>
          {showEnded &&
            endedServices.map((service, serviceIndex) => (
              <ServiceItem
                key={`ended-service-item-${serviceIndex}`}
                service={service}
                currency={currency}
                isEnded
              />
            ))}
          {liveServices.map((service, serviceIndex) => (
            <ServiceItem
              key={`service-item-${serviceIndex}`}
              currency={currency}
              service={service}
            />
          ))}
        </CardContent>
      </Card>
    );
  },
);