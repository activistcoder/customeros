import { observer } from 'mobx-react-lite';
import { ContactsStore } from '@store/Contacts/Contacts.store';
import { OrganizationsStore } from '@store/Organizations/Organizations.store';

import { flags } from '@ui/media/flags';
import { useStore } from '@shared/hooks/useStore';

interface ContactNameCellProps {
  id: string;
  type?: 'contact' | 'organization';
}

export const CountryCell = observer(({ id, type }: ContactNameCellProps) => {
  const { organizations, contacts } = useStore();
  const store: ContactsStore | OrganizationsStore =
    type === 'contact' ? contacts : organizations;
  const itemStore = store.value.get(id);
  const country = itemStore?.country;

  const enrichedItem = itemStore?.value.enrichDetails;

  const enrichingStatus =
    !enrichedItem?.enrichedAt &&
    enrichedItem?.requestedAt &&
    !enrichedItem?.failedAt;

  if (!country) {
    return (
      <div className='text-gray-400'>
        {enrichingStatus ? 'Enriching...' : 'Not set'}
      </div>
    );
  }
  const alpha2 = itemStore?.value?.locations?.[0]?.countryCodeA2;

  return (
    <div className='flex items-center'>
      <div className='flex items-center'>{alpha2 && flags[alpha2]}</div>
      <span className='ml-2 overflow-hidden overflow-ellipsis whitespace-nowrap'>
        {country}
      </span>
    </div>
  );
});
