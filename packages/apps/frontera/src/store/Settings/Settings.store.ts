import type { RootStore } from '@store/root';

import { makeAutoObservable } from 'mobx';
import { TransportLayer } from '@store/transport';

import { Slack } from './Slack.store';
import { Google } from './Google.store';
import { FeaturesStore } from './Features.store';
import { IntegrationsStore } from './Integrations.store';
import { TenantStore } from './Tenant.store';

export interface OAuthToken {
  scope: string;
  expiresAt: Date;
  idToken: string;
  accessToken: string;
  refreshToken: string;
  providerAccountId: string;
}

export class SettingsStore {
  slack: Slack;
  google: Google;
  tenant: TenantStore;
  features: FeaturesStore;
  integrations: IntegrationsStore;
  isLoading = false;
  error: string | null = null;

  constructor(
    private rootStore: RootStore,
    private transportLayer: TransportLayer,
  ) {
    this.slack = new Slack(this.rootStore, this.transportLayer);
    this.google = new Google(this.rootStore, this.transportLayer);
    this.features = new FeaturesStore(this.rootStore, this.transportLayer);
    this.tenant = new TenantStore(this.rootStore, this.transportLayer);
    this.integrations = new IntegrationsStore(
      this.rootStore,
      this.transportLayer,
    );
    makeAutoObservable(this);
  }

  get isBootstrapping() {
    return (
      this.slack.isLoading ||
      this.google.isLoading ||
      this.tenant.isLoading ||
      this.features.isLoading ||
      this.integrations.isLoading
    );
  }
  get bootstrapError() {
    return (
      this.slack.error ||
      this.google.error ||
      this.tenant.error ||
      this.features.error ||
      this.integrations.error
    );
  }
  get isBootstrapped() {
    return (
      this.slack.isBootstrapped &&
      this.google.isBootstrapped &&
      this.tenant.isBootstrapped &&
      this.features.isBootstrapped &&
      this.integrations.isBootstrapped
    );
  }

  async bootstrap() {
    if (this.isBootstrapped) return;

    await Promise.all([
      await this.slack.load(),
      await this.google.load(),
      await this.tenant.bootstrap(),
      await this.features.load(),
      await this.integrations.load(),
    ]);
  }

  async revokeAccess(
    payload: unknown,
    options?: {
      onError?: (err: Error) => void;
      onSuccess?: (res: unknown) => void;
    },
  ) {
    try {
      this.isLoading = true;

      const res = this.transportLayer.http.post('ua/revoke', payload);
      options?.onSuccess?.(res);
    } catch (err) {
      this.error = (err as Error)?.message;
      options?.onError?.(err as Error);
    } finally {
      this.isLoading = false;
    }
  }
}