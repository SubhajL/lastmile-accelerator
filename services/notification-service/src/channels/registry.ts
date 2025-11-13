import { createEmailChannel } from './email.js';
import type { ChannelAdapter } from '../notifications/dispatcher.js';
import type { EmailChannelOptions } from './email.js';

export interface ChannelRegistry {
  get(channel: string): ChannelAdapter | undefined;
}

export function createChannelRegistry(adapters: Partial<Record<string, ChannelAdapter>>): ChannelRegistry {
  return {
    get(channel: string) {
      return adapters[channel];
    }
  };
}

export function createDefaultChannelRegistry(opts: { email: EmailChannelOptions }): ChannelRegistry {
  const email = createEmailChannel(opts.email);
  return createChannelRegistry({ email });
}
