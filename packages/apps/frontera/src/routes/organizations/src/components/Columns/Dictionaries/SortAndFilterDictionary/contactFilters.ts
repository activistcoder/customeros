import { match } from 'ts-pattern';
import { Store } from '@store/store.ts';
import { FilterItem } from '@store/types.ts';

import { Tag, Contact, ColumnViewType } from '@graphql/types';

export const getContactFilterFn = (filter: FilterItem | undefined | null) => {
  const noop = (_row: Store<Contact>) => true;
  if (!filter) return noop;

  return match(filter)
    .with({ property: 'STAGE' }, (filter) => (row: Store<Contact>) => {
      const filterValues = filter?.value;

      if (!filterValues) return false;

      const hasOrgWithMatchingStage = row.value?.organizations.content.every(
        (o) => {
          const stage = row.root?.organizations?.value.get(o.metadata.id)?.value
            ?.stage;

          return filterValues.includes(stage);
        },
      );

      return hasOrgWithMatchingStage;
    })
    .with({ property: 'RELATIONSHIP' }, (filter) => (row: Store<Contact>) => {
      const filterValues = filter?.value;
      if (!filterValues) return false;
      const hasOrgWithMatchingRelationship =
        row.value?.organizations.content.every((o) => {
          const stage = row.root?.organizations?.value.get(o.metadata.id)?.value
            ?.relationship;

          return filterValues.includes(stage);
        });

      return hasOrgWithMatchingRelationship;
    })
    .with(
      { property: ColumnViewType.ContactsName },
      (filter) => (row: Store<Contact>) => {
        const filterValue = filter?.value;
        if (!filter.active) return true;

        if (!row.value?.name?.length && filter.includeEmpty) return true;
        if (!filterValue || !row.value?.name?.length) return false;

        return filterValue.toLowerCase().includes(row.value.name.toLowerCase());
      },
    )
    .with(
      { property: ColumnViewType.ContactsOrganization },
      (filter) => (row: Store<Contact>) => {
        const filterValues = filter?.value;

        if (!filter.active) return true;

        const orgs = row.value?.organizations?.content?.map((o) =>
          o.name.toLowerCase().trim(),
        );

        return orgs?.some((e) => e.includes(filterValues));
      },
    )
    .with(
      { property: ColumnViewType.ContactsEmails },
      (filter) => (row: Store<Contact>) => {
        const filterValues = filter?.value;
        if (!filter.active) return true;

        if (!filterValues || filter.operation === 'EQ') return false;

        return row.value?.emails?.some(
          (e) =>
            e.email &&
            filterValues.some(
              (value: string) => e.email && value.includes(e.email),
            ),
        );
      },
    )
    .with({ property: 'EMAIL_VERIFIED' }, (filter) => (row: Store<Contact>) => {
      const filterValue = filter?.value;

      if (!filter.active) return true;
      if (row.value?.emails?.length === 0) return false;

      return row.value?.emails?.every((e) => {
        const { validated, isReachable, isValidSyntax } =
          e.emailValidationDetails;

        if (filterValue === 'verified') {
          return validated && isReachable !== 'invalid' && isValidSyntax;
        }

        return !validated || isReachable === 'invalid' || !isValidSyntax;
      });
    })
    .with(
      { property: ColumnViewType.ContactsPhoneNumbers },
      (filter) => (row: Store<Contact>) => {
        const filterValues = filter?.value;

        if (!filterValues) return false;

        return row.value?.phoneNumbers?.some((e) =>
          filterValues.includes(e.e164),
        );
      },
    )

    .with(
      { property: ColumnViewType.ContactsLinkedin },
      (filter) => (row: Store<Contact>) => {
        if (!filter.active) return true;
        const filterValue = filter?.value;

        // specific logic for linkedin
        const linkedInUrl = row.value.socials?.find(
          (v: { id: string; url: string }) => v.url.includes('linkedin'),
        )?.url;

        if (!linkedInUrl && filter.includeEmpty) return true;

        return linkedInUrl && linkedInUrl.includes(filterValue);
      },
    )
    .with(
      { property: ColumnViewType.ContactsCity },
      (filter) => (row: Store<Contact>) => {
        if (!filter.active) return true;
        const filterValue = filter?.value;
        const cities = row.value.locations?.map((l) => l.locality);

        if (!cities.length && filter.includeEmpty) return true;

        return row.value.locations?.some((l) =>
          l?.locality?.includes(filterValue),
        );
      },
    )
    .with(
      { property: ColumnViewType.ContactsPersona },
      (filter) => (row: Store<Contact>) => {
        if (!filter.active) return true;
        const tags = row.value.tags?.map((l: Tag) => l.name);

        if (!tags?.length && filter.includeEmpty) return true;

        return tags?.some((tag: string) =>
          tag.toLowerCase().trim().includes(filter.value.toLowerCase().trim()),
        );
      },
    )

    .otherwise(() => noop);
};