import React from 'react';
import { useForm } from 'react-inverted-form';

import { useDebounce } from 'rooks';
import { useQueryClient } from '@tanstack/react-query';
import { useBankAccountsQuery } from '@settings/graphql/getBankAccounts.generated';
import { useUpdateBankAccountMutation } from '@settings/graphql/updateBankAccount.generated';
import { BankNameInput } from '@settings/components/Tabs/panels/BillingPanel/components/BankNameInput';
import { useBankTransferSelectionContext } from '@settings/components/Tabs/panels/BillingPanel/context/BankTransferSelectionContext';

import { FormInput } from '@ui/form/Input/FormInput';
import { FormMaskInput } from '@ui/form/Input/FormMaskInput';
import { getGraphQLClient } from '@shared/util/getGraphQLClient';
import { Card, CardHeader, CardContent } from '@ui/presentation/Card/Card';
import { Currency, BankAccount, BankAccountUpdateInput } from '@graphql/types';
import { FormAutoresizeTextarea } from '@ui/form/Textarea/FormAutoresizeTextarea';

import { BankTransferMenu } from './BankTransferMenu';
import { BankTransferCurrencySelect } from './BankTransferCurrencySelect';

const bankOptions = {
  mask: [
    {
      mask: 'XX 00 0000 0000 0000 0000 0000 0000',
      definitions: {
        X: /[A-Za-z]/,
        '0': /[0-9]/,
      },
    },
    {
      mask: '00 0000 0000 0000 0000 0000 0000 0000',
      definitions: {
        '0': /[0-9]/,
      },
    },
  ],
};

const sortCodeOptions = {
  mask: '00-00-00',
  definitions: {
    '0': /[0-9]/,
  },
};

export const BankTransferCard = ({
  account,
  existingCurrencies,
}: {
  account: BankAccount;
  existingCurrencies: Array<string>;
}) => {
  const formId = `bank-transfer-form-${account.metadata.id}`;
  const queryKey = useBankAccountsQuery.getKey();
  const queryClient = useQueryClient();
  const { setFocusAccount, setHoverAccount } =
    useBankTransferSelectionContext();

  const client = getGraphQLClient();
  const { mutate } = useUpdateBankAccountMutation(client, {
    onSuccess: () => {},
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });

  const updateBankAccountDebounced = useDebounce(
    (variables: Partial<BankAccountUpdateInput>) => {
      mutate({
        input: {
          id: account.metadata.id,
          ...variables,
        },
      });
    },
    500,
  );
  useForm<BankAccount>({
    formId,
    defaultValues: account,
    debug: true,
    stateReducer: (state, action, next) => {
      if (action.type === 'FIELD_CHANGE') {
        switch (action.payload.name) {
          case 'bic':
          case 'sortCode':
          case 'routingNumber':
          case 'bankName':
            updateBankAccountDebounced({
              [action.payload.name]: action.payload.value,
              currency: account.currency,
            });

            return next;

          case 'iban':
          case 'accountNumber':
            updateBankAccountDebounced({
              [action.payload.name]: action.payload.value?.toUpperCase(),
              currency: account.currency,
            });

            return {
              ...next,
              values: {
                ...next.values,
                [action.payload.name]: action.payload.value?.toUpperCase(),
              },
            };
          case 'currency':
            mutate({
              input: {
                id: account.metadata.id,
                currency: action.payload.value?.value,
                sortCode: '',
                iban: '',
                routingNumber: '',
                accountNumber: '',
                bic: '',
              },
            });

            return next;

          default: {
            return next;
          }
        }
      }

      return next;
    },
  });

  return (
    <>
      <Card
        className='py-2 px-4 rounded-lg border-[1px] border-gray-200'
        onMouseEnter={() => setHoverAccount(account as BankAccount)}
        onMouseLeave={() => setHoverAccount(null)}
      >
        <CardHeader className='p-0 pb-1 flex justify-between'>
          <BankNameInput formId={formId} metadata={account.metadata} />

          <div className='flex'>
            <BankTransferCurrencySelect
              existingCurrencies={existingCurrencies}
              currency={account.currency}
              formId={formId}
            />

            <BankTransferMenu id={account?.metadata?.id} />
          </div>
        </CardHeader>
        <CardContent className='p-0 gap-2'>
          {account.currency !== 'USD' && account.currency !== 'GBP' && (
            <>
              <FormMaskInput
                options={{ opts: bankOptions }}
                label='IBAN'
                placeholder='IBAN #'
                labelProps={{ className: 'text-sm mb-0 font-semibold' }}
                name='iban'
                className='mb-1'
                formId={formId}
                onFocus={() => setFocusAccount(account as BankAccount)}
                onBlur={() => setFocusAccount(null)}
              />
              <FormInput
                autoComplete='off'
                label='BIC/Swift'
                placeholder='BIC/Swift'
                labelProps={{ className: 'text-sm mb-0 font-semibold' }}
                name='bic'
                formId={formId}
                onFocus={() => setFocusAccount(account as BankAccount)}
                onBlur={() => setFocusAccount(null)}
              />
            </>
          )}
          <div className='flex pb-1 gap-2'>
            {account.currency === 'GBP' && (
              <>
                <FormMaskInput
                  options={{ opts: sortCodeOptions }}
                  autoComplete='off'
                  label='Sort code'
                  placeholder='Sort code'
                  labelProps={{ className: 'text-sm mb-0 font-semibold' }}
                  name='sortCode'
                  formId={formId}
                  className='max-w-[80px]'
                  onFocus={() => setFocusAccount(account as BankAccount)}
                  onBlur={() => setFocusAccount(null)}
                />
                <FormMaskInput
                  options={{ opts: bankOptions }}
                  autoComplete='off'
                  label='Account number'
                  placeholder='Bank account #'
                  labelProps={{ className: 'text-sm mb-0 font-semibold' }}
                  name='accountNumber'
                  formId={formId}
                  onFocus={() => setFocusAccount(account as BankAccount)}
                  onBlur={() => setFocusAccount(null)}
                />
              </>
            )}
          </div>
          {account.currency === 'USD' && (
            <>
              <FormInput
                autoComplete='off'
                className='mb-1'
                label='Routing number'
                placeholder='Routing number'
                labelProps={{ className: 'text-sm mb-0 font-semibold' }}
                name='routingNumber'
                formId={formId}
                onFocus={() => setFocusAccount(account as BankAccount)}
                onBlur={() => setFocusAccount(null)}
              />
              <FormMaskInput
                options={{ opts: bankOptions }}
                autoComplete='off'
                label='Account number'
                placeholder='Bank account #'
                labelProps={{ className: 'text-sm mb-0 font-semibold' }}
                name='accountNumber'
                formId={formId}
                onFocus={() => setFocusAccount(account as BankAccount)}
                onBlur={() => setFocusAccount(null)}
              />
            </>
          )}
          {account.allowInternational &&
            (account.currency === 'USD' || account.currency === 'GBP') && (
              <FormInput
                autoComplete='off'
                label='BIC/Swift'
                placeholder='BIC/Swift'
                labelProps={{ className: 'text-sm mb-0 font-semibold' }}
                name='bic'
                formId={formId}
                onFocus={() => setFocusAccount(account as BankAccount)}
                onBlur={() => setFocusAccount(null)}
              />
            )}

          {(account.allowInternational ||
            ![Currency.Gbp, Currency.Usd, Currency.Eur].includes(
              account?.currency as Currency,
            )) && (
            <FormAutoresizeTextarea
              autoComplete='off'
              label='Other details'
              placeholder='Other details'
              name='otherDetails'
              formId={formId}
              onFocus={() => setFocusAccount(account as BankAccount)}
              onBlur={() => setFocusAccount(null)}
            />
          )}
        </CardContent>
      </Card>
    </>
  );
};