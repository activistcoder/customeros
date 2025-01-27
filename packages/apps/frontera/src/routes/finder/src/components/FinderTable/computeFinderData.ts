import { Params } from 'react-router-dom';

import Fuse from 'fuse.js';
import { match } from 'ts-pattern';
import { RootStore } from '@store/root';
import { inPlaceSort } from 'fast-sort';
import { SortingState } from '@tanstack/table-core';
import { TableViewDefStore } from '@store/TableViewDefs/TableViewDef.store';

import { TableIdType, WorkflowType, TableViewType } from '@graphql/types';

import { getFlowFilterFns } from '../Columns/organizations/flowFilters';
import { getFlowsFilterFns, getFlowsColumnSortFn } from '../Columns/flows';
import {
  getOpportunitiesSortFn,
  getOpportunityFilterFns,
} from '../Columns/opportunities';
import {
  getContactSortFn,
  getContactFilterFns,
  getContactDefaultFilterFns,
} from '../Columns/contacts';
import {
  getInvoicesSortFn,
  getInvoiceFilterFns,
  getInvoiceDefaultFilterFns,
} from '../Columns/invoices';
import {
  getContractSortFn,
  getContractFilterFns,
  getContractDefaultFilters,
} from '../Columns/contracts';
import {
  getOrganizationSortFn,
  getOrganizationFilterFns,
  getOrganizationDefaultFilterFns,
} from '../Columns/organizations';

interface ComputeFinderDataOptions {
  searchTerm: string;
  sorting: SortingState;
  tableViewDef?: TableViewDefStore;
  urlParams: Readonly<Params<string>>;
}

export const computeFinderData = (
  store: RootStore,
  options: ComputeFinderDataOptions,
) => {
  const { searchTerm, sorting, tableViewDef, urlParams } = options;

  if (!tableViewDef) return [];

  const tableType =
    tableViewDef?.value.tableType || TableViewType.Organizations;

  const getWorkFlow = store.workFlows
    .toArray()
    .filter((wf) => wf.value.type === WorkflowType.IdealCustomerProfile);

  const getWorkFlowId = getWorkFlow.map((wf) => wf.value.id);
  const workFlow = store.workFlows.getByType(getWorkFlowId[0]);

  return match(tableType)
    .with(TableViewType.Organizations, () =>
      store.organizations?.toComputedArray((arr) => {
        const defaultFilters = getOrganizationDefaultFilterFns(
          tableViewDef?.getDefaultFilters(),
        );
        const filters = getOrganizationFilterFns(tableViewDef?.getFilters());
        const flowFilters = getFlowFilterFns(workFlow?.getFilters());

        // TODO: Should filter only once for both defaultFilters and filters
        if (defaultFilters) {
          arr = arr.filter((v) => defaultFilters.every((fn) => fn(v)));
        }

        if (flowFilters.length && store.ui.isFilteringICP) {
          arr = arr.filter((v) => !flowFilters.every((fn) => fn(v)));
        }

        if (filters) {
          arr = arr.filter((v) => filters.every((fn) => fn(v)));
        }

        if (tableType) {
          const columnId = sorting[0]?.id;
          const isDesc = sorting[0]?.desc;

          arr = inPlaceSort(arr)?.[isDesc ? 'desc' : 'asc'](
            getOrganizationSortFn(columnId),
          );
        }

        if (searchTerm) {
          arr = new Fuse(arr, {
            keys: [
              {
                name: 'name',
                getFn: (o) => o.value.name,
              },
              {
                name: 'webiste',
                getFn: (o) => o.value?.website || '',
              },
              {
                name: 'socials',
                getFn: (o) => o.value.socialMedia?.[0]?.url,
              },
            ],
            threshold: 0.4,
            isCaseSensitive: false,
          })
            .search(removeAccents(searchTerm), { limit: 40 })
            .map((r) => r.item);
        }

        return arr;
      }),
    )
    .with(TableViewType.Contacts, () =>
      store.contacts?.toComputedArray((arr) => {
        if (tableViewDef?.value.tableId === TableIdType.FlowContacts) {
          const currentFlowId = urlParams?.id as string;

          arr = arr.filter(
            (v) =>
              v.hasFlows &&
              (currentFlowId ? v.getFlowById(currentFlowId) : true),
          );
        }

        const defaultFilters = getContactDefaultFilterFns(
          tableViewDef?.getDefaultFilters(),
        );

        const filters = getContactFilterFns(tableViewDef?.getFilters());

        if (defaultFilters) {
          arr = arr.filter((v) => defaultFilters.every((fn) => fn(v)));
        }

        if (filters) {
          arr = arr.filter((v) => filters.every((fn) => fn(v)));
        }

        if (tableType) {
          const columnId = sorting[0]?.id;
          const isDesc = sorting[0]?.desc;

          arr = inPlaceSort(arr)?.[isDesc ? 'desc' : 'asc'](
            getContactSortFn(columnId),
          );
        }

        if (searchTerm) {
          arr = new Fuse(arr, {
            keys: [
              { name: 'name', getFn: (o) => o.name },
              {
                name: 'organization',
                getFn: (o) => o.value?.organizations.content?.[0]?.name,
              },
              {
                name: 'email',
                getFn: (o) => o.value?.emails?.[0]?.email || '',
              },
            ],
            threshold: 0.3,
            isCaseSensitive: false,
          })
            .search(removeAccents(searchTerm), { limit: 40 })
            .map((r) => r.item);
        }

        return arr;
      }),
    )
    .with(TableViewType.Contracts, () =>
      store.contracts?.toComputedArray((arr) => {
        const defaultFilters = getContractDefaultFilters(
          tableViewDef?.getDefaultFilters(),
        );

        const filters = getContractFilterFns(tableViewDef?.getFilters());

        if (defaultFilters) {
          arr = arr.filter((v) => defaultFilters.every((fn) => fn(v)));
        }

        if (filters) {
          arr = arr.filter((v) => filters.every((fn) => fn(v)));
        }

        if (tableType) {
          const columnId = sorting[0]?.id;
          const isDesc = sorting[0]?.desc;

          arr = inPlaceSort(arr)?.[isDesc ? 'desc' : 'asc'](
            getContractSortFn(columnId),
          );
        }

        if (searchTerm) {
          arr = new Fuse(arr, {
            keys: ['value.name'],
            threshold: 0.3,
            isCaseSensitive: false,
          })
            .search(removeAccents(searchTerm), { limit: 40 })
            .map((r) => r.item);
        }

        return arr;
      }),
    )
    .with(TableViewType.Invoices, () =>
      store.invoices.toComputedArray((arr) => {
        const defaultFilters = getInvoiceDefaultFilterFns(
          tableViewDef?.getDefaultFilters(),
        );
        const filters = getInvoiceFilterFns(tableViewDef?.getFilters());

        if (defaultFilters) {
          arr = arr.filter((v) => defaultFilters.every((fn) => fn(v)));
        }

        if (filters) {
          arr = arr.filter((v) => filters.every((fn) => fn(v)));
        }

        if (tableType) {
          const columnId = sorting[0]?.id;
          const isDesc = sorting[0]?.desc;

          arr = inPlaceSort(arr)?.[isDesc ? 'desc' : 'asc'](
            getInvoicesSortFn(columnId),
          );
        }

        if (searchTerm) {
          const normalizedSearchTerm = removeAccents(searchTerm);

          arr = new Fuse(arr, {
            keys: ['value.contract.contractName'],
            threshold: 0.3,
            isCaseSensitive: false,
          })
            .search(normalizedSearchTerm, { limit: 40 })
            .map((r) => r.item);
        }

        return arr;
      }),
    )
    .with(TableViewType.Flow, () =>
      store.flows.toComputedArray((arr) => {
        if (tableType !== TableViewType.Flow) return arr;

        const filters = getFlowsFilterFns(tableViewDef?.getFilters());

        if (filters) {
          arr = arr.filter((v) => filters.every((fn) => fn(v)));
        }

        if (tableType) {
          const columnId = sorting[0]?.id;
          const isDesc = sorting[0]?.desc;

          arr = inPlaceSort(arr)?.[isDesc ? 'desc' : 'asc'](
            getFlowsColumnSortFn(columnId),
          );
        }

        if (searchTerm) {
          const normalizedSearchTerm = removeAccents(searchTerm);

          arr = new Fuse(arr, {
            keys: ['value.name'],
            threshold: 0.3,
            isCaseSensitive: false,
          })
            .search(normalizedSearchTerm, { limit: 40 })
            .map((r) => r.item);
        }

        return arr.filter((e) => e.value.status !== 'ARCHIVED');
      }),
    )
    .with(TableViewType.Opportunities, () =>
      store.opportunities.toComputedArray((arr) => {
        if (tableType !== TableViewType.Opportunities) return arr;
        arr = arr.filter((opp) => opp.value.internalType === 'NBO');

        const filters = getOpportunityFilterFns(tableViewDef?.getFilters());

        if (filters) {
          arr = arr.filter((v) => filters.every((fn) => fn(v)));
        }

        if (tableType) {
          const columnId = sorting[0]?.id;
          const isDesc = sorting[0]?.desc;

          arr = inPlaceSort(arr)?.[isDesc ? 'desc' : 'asc'](
            getOpportunitiesSortFn(columnId),
          );
        }

        if (searchTerm) {
          const normalizedSearchTerm = removeAccents(searchTerm);

          arr = new Fuse(arr, {
            keys: ['value.name', 'organization.value.name', 'owner.name'],
            threshold: 0.3,
            isCaseSensitive: false,
          })
            .search(normalizedSearchTerm, { limit: 40 })
            .map((r) => r.item);
        }

        return arr;
      }),
    )
    .otherwise(() => []);
};

function removeAccents(str: string) {
  return str
    .toLowerCase()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '');
}
