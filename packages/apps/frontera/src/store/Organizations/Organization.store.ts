import type { RootStore } from '@store/root';

import set from 'lodash/set';
import merge from 'lodash/merge';
import { Channel } from 'phoenix';
import { P, match } from 'ts-pattern';
import { gql } from 'graphql-request';
import { Operation } from '@store/types';
import { makePayload } from '@store/util';
import { Transport } from '@store/transport';
import { rdiffResult } from 'recursive-diff';
import { Store, makeAutoSyncable } from '@store/store';
import { runInAction, makeAutoObservable } from 'mobx';
import { makeAutoSyncableGroup } from '@store/group-store';
import { ActionStore } from '@store/TimelineEvents/Actions/Action.store';

import {
  Market,
  DataSource,
  SocialInput,
  FundingRound,
  Organization,
  OnboardingStatus,
  OrganizationStage,
  SocialUpdateInput,
  LastTouchpointType,
  OrganizationTagInput,
  LinkOrganizationsInput,
  OrganizationUpdateInput,
  OrganizationRelationship,
  OpportunityRenewalLikelihood,
  OpportunityRenewalUpdateAllForOrganizationInput,
} from '@graphql/types';

import { OrganizationsService } from './__service__/Organizations.service';

export class OrganizationStore implements Store<Organization> {
  value: Organization;
  version = 0;
  isLoading = false;
  history: Operation[] = [];
  error: string | null = null;
  channel?: Channel | undefined;
  subscribe = makeAutoSyncable.subscribe;
  sync = makeAutoSyncableGroup.sync;
  load = makeAutoSyncable.load<Organization>();
  update = makeAutoSyncable.update<Organization>();
  private service: OrganizationsService;

  constructor(public root: RootStore, public transport: Transport) {
    this.value = getDefaultValue();
    this.service = OrganizationsService.getInstance(transport);

    makeAutoObservable(this);
    makeAutoSyncable(this, {
      channelName: 'Organization',
      mutator: this.save,
      getId: (d) => d?.metadata?.id,
    });
  }

  get id() {
    return this.value.metadata?.id;
  }
  set id(id: string) {
    this.value.metadata.id = id;
  }

  async invalidate() {
    try {
      this.isLoading = true;
      const { organization } = await this.transport.graphql.request<
        ORGANIZATION_QUERY_RESULT,
        { id: string }
      >(ORGANIZATIONS_QUERY, { id: this.id });

      this.load(organization);
    } catch (err) {
      runInAction(() => {
        this.error = (err as Error)?.message;
      });
    } finally {
      runInAction(() => {
        this.isLoading = false;
      });
    }
  }

  private async updateOwner() {
    try {
      this.isLoading = true;
      await this.transport.graphql.request<unknown, UPDATE_OWNER_PAYLOAD>(
        UPDATE_OWNER_MUTATION,
        {
          organizationId: this.id,
          userId: this.value.owner?.id || '',
        },
      );
    } catch (e) {
      runInAction(() => {
        this.error = (e as Error)?.message;
      });
    } finally {
      runInAction(() => {
        this.isLoading = false;
      });
    }
  }

  private async removeOwner() {
    try {
      this.isLoading = true;
      await this.transport.graphql.request<unknown, REMOVE_OWNER_PAYLOAD>(
        REMOVE_OWNER_MUTATION,
        {
          organizationId: this.id,
        },
      );
    } catch (e) {
      runInAction(() => {
        this.error = (e as Error)?.message;
      });
    } finally {
      runInAction(() => {
        this.isLoading = false;
      });
    }
  }

  private async updateAllOpportunityRenewals() {
    try {
      this.isLoading = true;

      const amount =
        this.value.accountDetails?.renewalSummary?.arrForecast ?? 0;
      const potentialAmount =
        this.value.accountDetails?.renewalSummary?.maxArrForecast ?? 0;
      const rate = (amount / potentialAmount) * 100;

      await this.transport.graphql.request<
        unknown,
        UPDATE_ALL_OPPORTUNITY_RENEWALS_PAYLOAD
      >(UPDATE_ALL_OPPORTUNITY_RENEWAlS_MUTATION, {
        input: {
          organizationId: this.id,
          renewalAdjustedRate: rate,
          renewalLikelihood:
            this.value.accountDetails?.renewalSummary?.renewalLikelihood,
        },
      });
    } catch (err) {
      runInAction(() => {
        this.error = (err as Error)?.message;
      });
    } finally {
      runInAction(() => {
        this.isLoading = false;
      });
    }
  }

  private async updateOrganization(payload: OrganizationUpdateInput) {
    try {
      this.isLoading = true;
      await this.transport.graphql.request<
        unknown,
        UPDATE_ORGANIZATION_PAYLOAD
      >(UPDATE_ORGANIZATION_MUTATION, {
        input: {
          ...payload,
          id: this.id,
          patch: true,
        },
      });
    } catch (err) {
      runInAction(() => {
        this.error = (err as Error)?.message;
      });
    } finally {
      runInAction(() => {
        this.isLoading = false;
      });
    }
  }

  private async updateSocialMedia(index: number) {
    try {
      this.isLoading = true;
      await this.transport.graphql.request<
        unknown,
        UPDATE_SOCIAL_MEDIA_PAYLOAD
      >(UPDATE_SOCIAL_MEDIA_MUTATION, { input: this.value.socialMedia[index] });
    } catch (err) {
      runInAction(() => {
        this.error = (err as Error)?.message;
      });
    } finally {
      runInAction(() => {
        this.isLoading = false;
      });
    }
  }

  private async removeSocialMedia(socialId: string) {
    try {
      this.isLoading = true;
      await this.transport.graphql.request<
        unknown,
        REMOVE_SOCIAL_MEDIA_PAYLOAD
      >(REMOVE_SOCIAL_MEDIA_MUTATION, {
        socialId,
      });
    } catch (err) {
      runInAction(() => {
        this.error = (err as Error)?.message;
      });
    } finally {
      runInAction(() => {
        this.isLoading = false;
      });
    }
  }

  private async addSocialMedia(index: number) {
    try {
      this.isLoading = true;
      const { organization_AddSocial } = await this.transport.graphql.request<
        ADD_SOCIAL_MEDIA_RESPONSE,
        ADD_SOCIAL_MEDIA_PAYLOAD
      >(ADD_SOCIAL_MEDIA_MUTATION, {
        organizationId: this.id,
        input: {
          url: this.value.socialMedia[index].url,
        },
      });

      this.update(
        (org) => {
          org.socialMedia[index].id = organization_AddSocial.id;

          return org;
        },
        { mutate: false },
      );
    } catch (err) {
      runInAction(() => {
        this.error = (err as Error)?.message;
      });
    } finally {
      runInAction(() => {
        this.isLoading = false;
      });
    }
  }

  private async addSubsidiary(subsidiaryId: string) {
    try {
      this.isLoading = true;
      await this.transport.graphql.request<
        { organization_AddSubsidiary: Organization },
        ADD_SUBSIDIARY_TO_ORGANIZATION
      >(ADD_SUBSIDIARY_TO_ORGANIZATION_MUTATION, {
        input: {
          organizationId: this.id,
          subsidiaryId: subsidiaryId,
        },
      });

      runInAction(() => {
        this.root.organizations.value.get(subsidiaryId)?.update(
          (org: Organization) => {
            org.parentCompanies.push({
              organization: this.value,
            });

            return org;
          },
          { mutate: false },
        );
        this.root.organizations.value.get(subsidiaryId)?.invalidate();
      });
    } catch (err) {
      runInAction(() => {
        this.error = (err as Error)?.message;
      });
    } finally {
      runInAction(() => {
        this.isLoading = false;
      });
    }
  }

  private async removeSubsidiary(organizationId: string) {
    try {
      this.isLoading = true;
      await this.transport.graphql.request<{
        organization_RemoveSubsidiary: Organization;
        REMOVE_SUBSIDIARY_FROM_ORGANIZATION: Organization;
      }>(REMOVE_SUBSIDIARY_FROM_ORGANIZATION_MUTATION, {
        organizationId: organizationId,
        subsidiaryId: this.id,
      });

      runInAction(() => {
        this.root.organizations.value.get(organizationId)?.invalidate();
        this.root.organizations.value.get(organizationId)?.update(
          (org: Organization) => {
            org.subsidiaries = [];

            return org;
          },
          { mutate: false },
        );
      });
    } catch (err) {
      runInAction(() => {
        this.error = (err as Error)?.message;
      });
    } finally {
      runInAction(() => {
        this.isLoading = false;
      });
    }
  }

  private async updateOnboardingStatus() {
    try {
      await this.service.updateOnboardingStatus({
        input: {
          organizationId: this.id,
          status:
            this.value?.accountDetails?.onboarding?.status ??
            OnboardingStatus.NotApplicable,
          comments: this.value?.accountDetails?.onboarding?.comments ?? '',
        },
      });
    } catch (e) {
      runInAction(() => {
        this.error = (e as Error)?.message;
      });
    }
  }

  private async addTagsToOrganization(tagId: string, tagName: string) {
    try {
      this.isLoading = true;
      await this.transport.graphql.request<
        unknown,
        ADD_TAGS_TO_ORGANIZATION_PAYLOAD
      >(ADD_TAGS_TO_ORGANIZATION_MUTATION, {
        input: {
          organizationId: this.id,
          tag: {
            id: tagId,
            name: tagName,
          },
        },
      });
    } catch (e) {
      runInAction(() => {
        this.error = (e as Error)?.message;
      });
    } finally {
      runInAction(() => {
        this.isLoading = false;
      });
    }
  }

  private async removeTagsFromOrganization(tagId: string) {
    try {
      this.isLoading = true;
      await this.transport.graphql.request<
        unknown,
        REMOVE_TAGS_FROM_ORGANIZATION_PAYLOAD
      >(REMOVE_TAGS_FROM_ORGANIZATION_MUTATION, {
        input: {
          organizationId: this.id,
          tag: {
            id: tagId,
          },
        },
      });
    } catch (e) {
      runInAction(() => {
        this.error = (e as Error)?.message;
      });
    } finally {
      runInAction(() => {
        this.isLoading = false;
      });
    }
  }

  private async save(operation: Operation) {
    const diff = operation.diff?.[0];
    const type = diff?.op;
    const path = diff?.path;
    const value = diff?.val;
    const oldValue = (diff as rdiffResult & { oldVal: unknown })?.oldVal;

    match(path)
      .with(['owner', ...P.array()], () => {
        if (type === 'update') {
          match(value)
            .with(null, () => {
              this.removeOwner();
            })
            .otherwise(() => {
              this.updateOwner();
            });
        }
      })
      .with(['contracts', ...P.array()], () => {
        // console.log('here in contracts path');
      })
      .with(['accountDetails', 'renewalSummary', ...P.array()], () => {
        this.updateAllOpportunityRenewals();
      })
      .with(['accountDetails', 'onboarding', ...P.array()], () => {
        this.updateOnboardingStatus();
      })
      .with(['socialMedia', ...P.array()], () => {
        const index = path[1];

        if (type === 'add') {
          this.addSocialMedia(index as number);
        }
        if (type === 'update') {
          this.updateSocialMedia(index as number);
        }
        if (type === 'delete') {
          this.removeSocialMedia(oldValue?.id);
        }
      })
      .with(['subsidiaries', ...P.array()], () => {
        if (type === 'add') {
          this.addSubsidiary(
            value[0]?.organization?.metadata?.id ||
              value?.organization?.metadata?.id,
          );
        }
      })
      .with(['parentCompanies', ...P.array()], () => {
        if (type === 'delete') {
          this.removeSubsidiary(oldValue?.organization?.metadata?.id);
        }
      })
      .with(['tags', ...P.array()], () => {
        if (type === 'add') {
          this.addTagsToOrganization(value.id, value.name);
        }
        if (type === 'delete') {
          this.removeTagsFromOrganization(oldValue.id);
        }
      })

      .otherwise(() => {
        const payload = makePayload<OrganizationUpdateInput>(operation);
        this.updateOrganization(payload);
      });
  }

  init(data: Organization) {
    const output = merge(this.value, data);

    const contracts = data.contracts?.map((item) => {
      this.root.contracts.load([item]);

      return this.root.contracts.value.get(item?.metadata?.id)?.value;
    });

    const parentCompanies = data.parentCompanies?.map((item) => {
      this.root.organizations.load([item.organization]);

      return {
        ...item,
        organization: this.root.organizations.value.get(
          item.organization.metadata.id,
        )?.value,
      };
    });

    const subsidiaries = data.subsidiaries?.map((item) => {
      this.root.organizations.load([item.organization]);

      return {
        ...item,
        organization: this.root.organizations.value.get(
          item.organization.metadata.id,
        )?.value,
      };
    });

    const contacts = data.contacts?.content?.map((item) => {
      this.root.contacts.load([item]);

      const contactStore = this.root.contacts.value.get(item.id);
      if (contactStore) {
        contactStore.organizationId = this.id;
      }

      return contactStore?.value;
    });

    contacts && set(output, 'contacts.content', contacts);
    contracts && set(output, 'contracts', contracts);
    subsidiaries && set(output, 'subsidiaries', subsidiaries);
    parentCompanies && set(output, 'parentCompanies', parentCompanies);

    return output;
  }
}

type ORGANIZATION_QUERY_RESULT = {
  organization: Organization;
};
const ORGANIZATIONS_QUERY = gql`
  query Organization($id: ID!) {
    organization(id: $id) {
      name
      metadata {
        id
        created
      }
      parentCompanies {
        organization {
          metadata {
            id
          }
          name
        }
      }
      contracts {
        metadata {
          id
        }
      }
      owner {
        id
        firstName
        lastName
        name
      }
      stage
      description
      industry
      website
      domains
      isCustomer
      logo
      icon
      relationship
      leadSource
      tags {
        id
        name
        createdAt
        updatedAt
        appSource
      }
      valueProposition
      socialMedia {
        id
        url
      }
      employees
      yearFounded
      accountDetails {
        churned
        ltv
        renewalSummary {
          arrForecast
          maxArrForecast
          renewalLikelihood
          nextRenewalDate
        }
        onboarding {
          status
          comments
          updatedAt
        }
      }
      locations {
        id
        name
        country
        region
        locality
        zip
        street
        postalCode
        houseNumber
        rawAddress
      }
      subsidiaries {
        organization {
          metadata {
            id
          }
          name
          parentCompanies {
            organization {
              name
              metadata {
                id
              }
            }
          }
        }
      }
      parentCompanies {
        organization {
          metadata {
            id
          }
        }
      }
      lastTouchpoint {
        lastTouchPointTimelineEventId
        lastTouchPointAt
        lastTouchPointType
        lastTouchPointTimelineEvent {
          __typename
          ... on PageView {
            id
          }
          ... on Issue {
            id
            createdAt
            updatedAt
          }
          ... on LogEntry {
            id
            createdBy {
              lastName
              firstName
            }
          }
          ... on Note {
            id
            createdBy {
              firstName
              lastName
            }
          }
          ... on InteractionEvent {
            id
            channel
            eventType
            externalLinks {
              type
            }
            sentBy {
              __typename
              ... on EmailParticipant {
                type
                emailParticipant {
                  id
                  email
                  rawEmail
                }
              }
              ... on ContactParticipant {
                contactParticipant {
                  id
                  name
                  firstName
                  lastName
                }
              }
              ... on JobRoleParticipant {
                jobRoleParticipant {
                  contact {
                    id
                    name
                    firstName
                    lastName
                  }
                }
              }
              ... on UserParticipant {
                userParticipant {
                  id
                  firstName
                  lastName
                }
              }
            }
          }
          ... on Analysis {
            id
          }
          ... on Meeting {
            id
            name
            attendedBy {
              __typename
            }
          }
          ... on Action {
            id
            actionType
            createdAt
            source
            actionType
            createdBy {
              id
              firstName
              lastName
            }
          }
        }
      }
    }
  }
`;
type UPDATE_OWNER_PAYLOAD = {
  userId: string;
  organizationId: string;
};

const UPDATE_OWNER_MUTATION = gql`
  mutation setOrganizationOwner($organizationId: ID!, $userId: ID!) {
    organization_SetOwner(organizationId: $organizationId, userId: $userId) {
      id
    }
  }
`;
type REMOVE_OWNER_PAYLOAD = {
  organizationId: string;
};

const REMOVE_OWNER_MUTATION = gql`
  mutation setOrganizationOwner($organizationId: ID!) {
    organization_UnsetOwner(organizationId: $organizationId) {
      id
    }
  }
`;
type UPDATE_ALL_OPPORTUNITY_RENEWALS_PAYLOAD = {
  input: OpportunityRenewalUpdateAllForOrganizationInput;
};

const UPDATE_ALL_OPPORTUNITY_RENEWAlS_MUTATION = gql`
  mutation bulkUpdateOpportunityRenewal(
    $input: OpportunityRenewalUpdateAllForOrganizationInput!
  ) {
    opportunityRenewal_UpdateAllForOrganization(input: $input) {
      metadata {
        id
      }
    }
  }
`;
type UPDATE_ORGANIZATION_PAYLOAD = {
  input: OrganizationUpdateInput;
};

const UPDATE_ORGANIZATION_MUTATION = gql`
  mutation updateOrganization($input: OrganizationUpdateInput!) {
    organization_Update(input: $input) {
      metadata {
        id
      }
    }
  }
`;

type UPDATE_SOCIAL_MEDIA_PAYLOAD = {
  input: SocialUpdateInput;
};

const UPDATE_SOCIAL_MEDIA_MUTATION = gql`
  mutation updateSocial($input: SocialUpdateInput!) {
    social_Update(input: $input) {
      id
      url
    }
  }
`;

type REMOVE_SOCIAL_MEDIA_PAYLOAD = {
  socialId: string;
};

const REMOVE_SOCIAL_MEDIA_MUTATION = gql`
  mutation removeSocial($socialId: ID!) {
    social_Remove(socialId: $socialId) {
      result
    }
  }
`;

type ADD_SOCIAL_MEDIA_PAYLOAD = {
  input: SocialInput;
  organizationId: string;
};

type ADD_SOCIAL_MEDIA_RESPONSE = {
  organization_AddSocial: {
    id: string;
    url: string;
  };
};

const ADD_SOCIAL_MEDIA_MUTATION = gql`
  mutation addSocial($organizationId: ID!, $input: SocialInput!) {
    organization_AddSocial(organizationId: $organizationId, input: $input) {
      id
      url
    }
  }
`;

type ADD_SUBSIDIARY_TO_ORGANIZATION = {
  input: LinkOrganizationsInput;
};

const ADD_SUBSIDIARY_TO_ORGANIZATION_MUTATION = gql`
  mutation addSubsidiaryToOrganization($input: LinkOrganizationsInput!) {
    organization_AddSubsidiary(input: $input) {
      metadata {
        id
      }
      subsidiaries {
        organization {
          metadata {
            id
          }
          name
          locations {
            id
            address
          }
        }
      }
    }
  }
`;

const REMOVE_SUBSIDIARY_FROM_ORGANIZATION_MUTATION = gql`
  mutation removeSubsidiaryToOrganization(
    $organizationId: ID!
    $subsidiaryId: ID!
  ) {
    organization_RemoveSubsidiary(
      organizationId: $organizationId
      subsidiaryId: $subsidiaryId
    ) {
      id
      subsidiaries {
        organization {
          id
          name
          locations {
            id
            address
          }
        }
      }
    }
  }
`;

type ADD_TAGS_TO_ORGANIZATION_PAYLOAD = {
  input: OrganizationTagInput;
};

const ADD_TAGS_TO_ORGANIZATION_MUTATION = gql`
  mutation addTagsToOrganization($input: OrganizationTagInput!) {
    organization_AddTag(input: $input) {
      accepted
    }
  }
`;

type REMOVE_TAGS_FROM_ORGANIZATION_PAYLOAD = {
  input: OrganizationTagInput;
};

const REMOVE_TAGS_FROM_ORGANIZATION_MUTATION = gql`
  mutation removeTagFromOrganization($input: OrganizationTagInput!) {
    organization_RemoveTag(input: $input) {
      accepted
    }
  }
`;

const getDefaultValue = (): Organization => ({
  name: 'Unnamed',
  metadata: {
    id: crypto.randomUUID(),
    created: new Date().toISOString(),
    lastUpdated: new Date().toISOString(),
    appSource: DataSource.Openline,
    source: DataSource.Openline,
    sourceOfTruth: DataSource.Openline,
  },
  owner: null,
  contactCount: 0,
  contacts: {
    content: [],
    totalElements: 0,
    totalPages: 0,
  },
  customerOsId: '',
  customFields: [],
  domains: [],
  emails: [],
  externalLinks: [],
  industry: '',
  fieldSets: [],
  hide: false,
  inboundCommsCount: 0,
  issueSummaryByStatus: [],
  jobRoles: [],
  locations: [],
  orders: [],
  outboundCommsCount: 0,
  phoneNumbers: [],
  parentCompanies: [],
  socialMedia: [],
  stage: OrganizationStage.Target,
  tags: [],
  subsidiaries: [],
  suggestedMergeTo: [],
  timelineEvents: [],
  website: '',
  timelineEventsTotalCount: 0,
  accountDetails: {
    onboarding: {
      status: OnboardingStatus.NotStarted,
      comments: '',
      updatedAt: '',
    },
    ltv: 0,
    churned: new Date().toISOString(),
    renewalSummary: {
      arrForecast: 0,
      maxArrForecast: 0,
      renewalLikelihood: OpportunityRenewalLikelihood.HighRenewal,
      nextRenewalDate: '',
    },
  },
  contracts: [],
  customId: '',
  description: '',
  employees: 0,
  employeeGrowthRate: '',
  // entityTemplate: {} -> ignored | unused
  headquarters: '',
  isCustomer: false,
  logo: '',
  industryGroup: '',
  lastFundingAmount: '',
  lastFundingRound: FundingRound.PreSeed,
  lastTouchpoint: {
    lastTouchPointTimelineEventId: crypto.randomUUID(),
    lastTouchPointAt: new Date().toISOString(),
    lastTouchPointType: LastTouchpointType.ActionCreated,
    lastTouchPointTimelineEvent: ActionStore.getDefaultValue(),
  }, // nested defaults ignored for now -> should be converted into a Store
  lastTouchPointTimelineEventId: '',
  leadSource: '',
  market: Market.B2B,
  notes: '',
  public: false,
  relationship: OrganizationRelationship.Prospect,
  slackChannelId: '',
  stageLastUpdated: '',
  subIndustry: '',
  targetAudience: '',
  valueProposition: '',
  yearFounded: 0,
  // deprecated field -> needed because they're required in the TS type
  id: '',
  appSource: '',
  source: DataSource.Na,
  socials: [],
  createdAt: '',
  sourceOfTruth: DataSource.Na,
  subsidiaryOf: [],
  updatedAt: '',
});