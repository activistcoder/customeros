import { Channel } from 'phoenix';
import { Store } from '@store/store';
import { RootStore } from '@store/root';
import { Transport } from '@store/transport';
import { GroupOperation } from '@store/types';
import { runInAction, makeAutoObservable } from 'mobx';
import { GroupStore, makeAutoSyncableGroup } from '@store/group-store';
import { FlowSequenceStore } from '@store/Sequences/FlowSequence.store';
import { FlowSequenceService } from '@store/Sequences/__service__/FlowSequence.service';
import { CreateSequenceMutationVariables } from '@store/Sequences/__service__/createSequence.generated';

import { FlowSequence, FlowSequenceStatus } from '@graphql/types';

export class FlowSequencesStore implements GroupStore<FlowSequence> {
  version = 0;
  isLoading = false;
  history: GroupOperation[] = [];
  error: string | null = null;
  channel?: Channel | undefined;
  isBootstrapped: boolean = false;
  value: Map<string, Store<FlowSequence>> = new Map();
  sync = makeAutoSyncableGroup.sync;
  subscribe = makeAutoSyncableGroup.subscribe;
  load = makeAutoSyncableGroup.load<FlowSequence>();
  totalElements = 0;
  private service: FlowSequenceService;

  constructor(public root: RootStore, public transport: Transport) {
    makeAutoObservable(this);
    makeAutoSyncableGroup(this, {
      channelName: 'FlowSequences',
      getItemId: (item) => item?.metadata?.id,
      ItemStore: FlowSequenceStore,
    });
    this.service = FlowSequenceService.getInstance(transport);
  }

  toArray() {
    return Array.from(this.value.values());
  }

  toComputedArray(compute: (arr: FlowSequenceStore[]) => FlowSequenceStore[]) {
    const arr = this.toArray().filter(
      (item) => item.value.status !== FlowSequenceStatus.Archived,
    );

    return compute(arr as FlowSequenceStore[]);
  }

  async bootstrap() {
    if (this.root.demoMode) {
      this.isBootstrapped = true;

      return;
    }

    if (this.isBootstrapped || this.isLoading) return;

    try {
      const { sequences } = await this.service.getSequences();

      runInAction(() => {
        this.load(sequences);
      });
      runInAction(() => {
        this.isBootstrapped = true;
        this.totalElements = sequences.length;
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

  async invalidate() {
    this.isLoading = true;
  }

  async create(
    payload: CreateSequenceMutationVariables['input'],
    options?: { onSuccess?: (serverId: string) => void },
  ) {
    const newSequence = new FlowSequenceStore(this.root, this.transport);
    const tempId = newSequence.value.metadata?.id;

    newSequence.value = {
      ...newSequence.value,
      ...payload,
    };

    let serverId: string | undefined;

    this.value.set(tempId, newSequence);

    try {
      const { flow_sequence_Create } = await this.service.createSequence({
        input: payload,
      });

      runInAction(() => {
        serverId = flow_sequence_Create?.metadata.id;
        newSequence.setId(serverId);

        this.value.set(serverId, newSequence);
        this.value.delete(tempId);

        this.sync({ action: 'APPEND', ids: [serverId] });
      });
    } catch (e) {
      runInAction(() => {
        this.error = (e as Error)?.message;
      });
    } finally {
      serverId && options?.onSuccess?.(serverId);
      setTimeout(() => {
        if (serverId) {
          this.value.get(serverId)?.invalidate();
          this.root.flows.bootstrap();
        }
      }, 1000);
    }
  }

  archive = async (id: string, options?: { onSuccess?: () => void }) => {
    this.isLoading = true;

    const flow = this.value.get(id);

    try {
      const { flow_sequence_ChangeStatus } =
        await this.service.updateSequenceStatus({
          id,
          stage: FlowSequenceStatus.Archived,
        });

      if (flow_sequence_ChangeStatus.metadata.id) {
        runInAction(() => {
          flow?.update(
            (seq) => {
              seq.status = FlowSequenceStatus.Archived;

              return seq;
            },
            { mutate: false },
          );

          this.sync({
            action: 'INVALIDATE',
            ids: [id],
          });
        });
        this.root.ui.toastSuccess(
          `Sequence archived`,
          'archive-sequence-success',
        );
      }
    } catch (err) {
      runInAction(() => {
        this.error = (err as Error).message;
        this.root.ui.toastError(
          `We couldn't archive this sequence`,
          'archive-view-error',
        );
      });
    } finally {
      this.isLoading = false;
      options?.onSuccess?.();
    }
  };

  archiveMany = async (ids: string[], options?: { onSuccess?: () => void }) => {
    this.isLoading = true;

    try {
      const results = await Promise.all(
        ids.map((id) =>
          this.service.updateSequenceStatus({
            id,
            stage: FlowSequenceStatus.Archived,
          }),
        ),
      );

      const successfulIds = results.map(
        ({ flow_sequence_ChangeStatus }) =>
          flow_sequence_ChangeStatus?.metadata?.id,
      );

      runInAction(() => {
        successfulIds.forEach((id) => {
          this.value
            .get(id)
            ?.update(
              (seq) => ({ ...seq, status: FlowSequenceStatus.Archived }),
              { mutate: false },
            );
        });

        if (successfulIds.length > 0) {
          this.sync({ action: 'INVALIDATE', ids: successfulIds });
          this.root.ui.toastSuccess(
            `${successfulIds.length} sequences archived`,
            'archive-sequences-success',
          );
        }
      });
    } catch (err) {
      this.error = (err as Error).message;
      this.root.ui.toastError(
        "We couldn't archive these sequences",
        'archive-sequences-error',
      );
    } finally {
      this.isLoading = false;
      options?.onSuccess?.();
    }
  };

  public linkContacts = async (sequenceId: string, contactIds: string[]) => {
    this.isLoading = true;

    try {
      const results = await Promise.all(
        contactIds.map(async (id) => {
          const contactStore = this.root.contacts.value.get(id);
          const emailId = this.root.contacts.value.get(id)?.emailId ?? '';

          if (contactStore?.sequence) {
            await this.service.unlinkContact({
              sequenceId: contactStore.sequence.id,
              contactId: id,
              emailId,
            });
          }

          return this.service.linkContact({
            sequenceId,
            contactId: id,
            emailId,
          });
        }),
      );

      const successfulIds = results.map(
        ({ flow_sequence_LinkContact }) =>
          flow_sequence_LinkContact?.metadata?.id,
      );

      runInAction(() => {
        if (successfulIds.length > 0) {
          this.root.contacts.sync({ action: 'INVALIDATE', ids: contactIds });
          this.root.ui.toastSuccess(
            `${successfulIds.length} contacts added to '${
              this.value.get(sequenceId)?.value?.name
            }'`,
            'archive-sequences-success',
          );
        }
      });
    } catch (err) {
      this.error = (err as Error).message;
      this.root.ui.toastError(
        "We couldn't add those contacts to a sequence",
        'add-contacts-sequences-error',
      );
    } finally {
      this.isLoading = false;
    }
  };

  public unlinkContacts = async (contactIds: string[]) => {
    this.isLoading = true;

    try {
      const results = await Promise.all(
        contactIds.map((id) => {
          const contactStore = this.root.contacts.value.get(id);
          const emailId = contactStore?.emailId ?? '';
          const sequenceId = contactStore?.sequence?.id;

          if (!sequenceId) return false;

          return this.service.unlinkContact({
            sequenceId,
            contactId: id,
            emailId,
          });
        }),
      );

      const resultsData = results.map(
        (result) => result && result?.flow_sequence_UnlinkContact?.result,
      );

      runInAction(() => {
        if (resultsData.length > 0) {
          this.root.contacts.sync({ action: 'INVALIDATE', ids: contactIds });
          this.root.ui.toastSuccess(
            `${contactIds.length} contacts removed from their sequences`,
            'remove-contacts-from-sequences-success',
          );
        }
      });
    } catch (err) {
      this.error = (err as Error).message;
      this.root.ui.toastError(
        "We couldn't remove a contact from a sequence",
        'remove-contacts-sequences-error',
      );
    } finally {
      this.isLoading = false;
    }
  };
}