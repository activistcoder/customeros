import { useParams } from 'react-router-dom';
import { useField } from 'react-inverted-form';
import { useMemo, useState, useEffect, forwardRef, useCallback } from 'react';
import {
  GroupBase,
  OptionProps,
  SelectInstance,
  OptionsOrGroups,
  MultiValueProps,
  components as reactSelectComponents,
} from 'react-select';

import { SelectOption } from '@ui/utils/types';
import { Copy01 } from '@ui/media/icons/Copy01';
import { IconButton } from '@ui/form/IconButton/IconButton';
import { getName } from '@spaces/utils/getParticipantsName';
import { Contact, ComparisonOperator } from '@graphql/types';
import { getGraphQLClient } from '@shared/util/getGraphQLClient';
import { useCopyToClipboard } from '@shared/hooks/useCopyToClipboard';
import { emailRegex } from '@organization/components/Timeline/PastZone/events/email/utils';
import {
  FormSelectProps,
  MultiCreatableSelect,
} from '@ui/form/MultiCreatableSelect/MultiCreatableSelect';
import { MultiValueWithActionMenu } from '@shared/components/EmailMultiCreatableSelect/MultiValueWithActionMenu';
import {
  GetContactsEmailListDocument,
  useGetContactsEmailListQuery,
} from '@organization/graphql/getContactsEmailList.generated';

type ExistingContact = { id: string; label: string; value?: string | null };
export const EmailFormMultiCreatableSelect = forwardRef<
  SelectInstance,
  FormSelectProps & { navigateAfterAddingToPeople: boolean }
>(({ name, formId, navigateAfterAddingToPeople, ...rest }, ref) => {
  const client = getGraphQLClient();
  const organizationId = useParams()?.id as string;
  const [existingContacts, setExistingContacts] = useState<
    Array<ExistingContact>
  >([]);

  const { data } = useGetContactsEmailListQuery(client, {
    id: organizationId,
    pagination: {
      page: 1,
      limit: 100,
    },
  });

  useEffect(() => {
    if (data?.organization?.contacts?.content?.length) {
      const organizationContacts = (
        (data?.organization?.contacts?.content || []) as Array<Contact>
      )
        .map((e: Contact) => {
          if (e.emails.some((e) => !!e.email)) {
            return e.emails.map((email) => ({
              id: e.id,
              value: email.email,
              label: `${e.firstName} ${e.lastName}`,
            }));
          }

          return [
            {
              id: e.id,
              label: getName(e),
              value: '',
            },
          ];
        })
        .flat();
      setExistingContacts(organizationContacts);
    }
  }, [data]);

  const { getInputProps } = useField(name, formId);
  const { id, onChange, onBlur, value } = getInputProps();
  const [_, copyToClipboard] = useCopyToClipboard();

  const handleBlur = (stringVal: string) => {
    if (stringVal && emailRegex.test(stringVal)) {
      onBlur([...value, { label: stringVal, value: stringVal }]);

      return;
    }
    onBlur(value);
  };

  const getFilteredSuggestions = async (
    filterString: string,
    callback: (options: OptionsOrGroups<unknown, GroupBase<unknown>>) => void,
  ) => {
    try {
      const results = await client.request<{
        organization: {
          contacts: { content: Contact[] };
        };
      }>(GetContactsEmailListDocument, {
        id: organizationId,
        pagination: {
          page: 1,
          limit: 5,
        },
        where: {
          OR: [
            {
              filter: {
                property: 'FIRST_NAME',
                value: filterString,
                operation: ComparisonOperator.Contains,
              },
            },
            {
              filter: {
                property: 'LAST_NAME',
                value: filterString,
                operation: ComparisonOperator.Contains,
              },
            },
            {
              filter: {
                property: 'NAME',
                value: filterString,
                operation: ComparisonOperator.Contains,
              },
            },
          ],
        },
      });
      const options: OptionsOrGroups<unknown, GroupBase<unknown>> = (
        results?.organization?.contacts?.content || []
      )
        .map((e: Contact) => {
          if (e.emails.some((e) => !!e.email)) {
            return e.emails.map((email) => ({
              value: email.email,
              label: `${e.firstName} ${e.lastName}`,
            }));
          }

          return [
            {
              label: getName(e),
              value: '',
            },
          ];
        })
        .flat();

      callback(options);
    } catch (error) {
      callback([]);
    }
  };

  const Option = useCallback((rest: OptionProps<SelectOption>) => {
    const fullLabel =
      rest?.data?.label &&
      rest?.data?.value &&
      `${rest.data.label} - ${rest.data.value}`;
    const emailOnly =
      !rest?.data?.label && rest?.data?.value && `${rest.data.value}`;

    const noEmail = rest?.data?.label && !rest?.data?.value && (
      <p>
        {rest.data.label} -
        <span className='text-gray-500 ml-1'>[No email for this contact]</span>
      </p>
    );

    return (
      <reactSelectComponents.Option {...rest}>
        {fullLabel || emailOnly || noEmail}
        {rest?.isFocused && (
          <IconButton
            className='h-5 p-0 self-end float-end'
            aria-label='Copy'
            size='xs'
            variant='ghost'
            icon={<Copy01 className='size-3 text-gray-500' />}
            onClick={(e) => {
              e.stopPropagation();
              copyToClipboard(rest.data.value, 'Email copied');
            }}
          />
        )}
      </reactSelectComponents.Option>
    );
  }, []);

  const MultiValue = useCallback(
    (multiValueProps: MultiValueProps<SelectOption>) => {
      return (
        <MultiValueWithActionMenu
          {...multiValueProps}
          name={name}
          formId={formId}
          navigateAfterAddingToPeople={navigateAfterAddingToPeople}
          existingContacts={existingContacts}
        />
      );
    },
    [name, formId, navigateAfterAddingToPeople],
  );

  const components = useMemo(
    () => ({
      MultiValueRemove: () => null,
      LoadingIndicator: () => null,
      MultiValue,
    }),
    [MultiValue],
  );

  return (
    <MultiCreatableSelect
      ref={ref}
      id={id}
      formId={formId}
      name={name}
      value={value}
      classNames={{
        multiValueLabel: () =>
          'multiValueClass px-2 bg-transparent text-sm shadow-md border font-semibold rounded-lg border-gray-200 max-h-[12rem] cursor-pointer z-50',
      }}
      onBlur={(e) => handleBlur(e.target.value)}
      onChange={onChange}
      Option={Option}
      components={components}
      loadOptions={(inputValue: string, callback) => {
        getFilteredSuggestions(inputValue, callback);
      }}
      {...rest}
    />
  );
});