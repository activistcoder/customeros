import type { RootStore } from '@store/root';

import { makeAutoObservable } from 'mobx';
import { TransportLayer } from '@store/transport';

interface Field {
  name: string;
  label: string;
  textarea?: boolean;
}

export interface Integration {
  key: string;
  name: string;
  icon: string;
  fields: Field[];
  identifier: string;
  state: 'INACTIVE' | 'ACTIVE';
  isFromIntegrationApp?: boolean;
}

export class IntegrationsStore {
  value: Record<string, Integration> = {};
  isMutating = false;
  isBootstrapped = false;
  isBootstrapping = false;
  error: string | null = null;

  constructor(
    private rootStore: RootStore,
    private transportLayer: TransportLayer,
  ) {
    makeAutoObservable(this);
  }

  get isLoading() {
    return !this.isBootstrapped;
  }

  async load() {
    try {
      this.isBootstrapping = true;
      const { data } = await this.transportLayer.http.get('/sa/integrations');
      this.value = data;
      this.isBootstrapped = true;
    } catch (err) {
      this.error = (err as Error).message;
    } finally {
      this.isBootstrapping = false;
    }
  }

  async update(identifier: string, payload: unknown) {
    Object.assign(this.value, {
      [identifier]: { state: 'ACTIVE' },
    });

    try {
      this.isMutating = true;
      this.transportLayer.http.post('/sa/integration', {
        [identifier]: payload,
      });
      this.rootStore.uiStore.toastSuccess(
        'Settings updated successfully!',
        'integration-settings-update',
      );
    } catch (err) {
      delete this.value[identifier];
      this.error = (err as Error).message;
      this.rootStore.uiStore.toastError(
        `We couldn't update the Settings.`,
        'integration-settings-update-failed',
      );
    } finally {
      this.isMutating = false;
    }
  }

  async delete(identifier: string) {
    const integration = { ...this.value[identifier] };

    if (identifier in this.value) {
      delete this.value[identifier];
    }

    try {
      this.isMutating = true;
      this.transportLayer.http.delete(`/sa/integration/${identifier}`);
      this.rootStore.uiStore.toastSuccess(
        'Settings updated successfully!',
        'integration-settings-delete',
      );
    } catch (err) {
      this.value[identifier] = integration;
      this.error = (err as Error).message;
      this.rootStore.uiStore.toastError(
        `We couldn't update the Settings.`,
        'integration-settings-delete-failed',
      );
    } finally {
      this.isMutating = false;
    }
  }
}