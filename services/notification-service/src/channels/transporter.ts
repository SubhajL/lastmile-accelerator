import type { SmtpConfig } from '../types.js';

type TransportOptions = { host: string; port: number; secure: boolean; auth: { user: string; pass: string } };

export function createSmtpTransporter(config: SmtpConfig, deps: { nodemailer: { createTransporter: (opts: TransportOptions) => unknown } }) {
  const secure = config.port === 465 ? true : config.secure;
  const transportOptions = {
    host: config.host,
    port: config.port,
    secure,
    auth: {
      user: config.user,
      pass: config.password
    }
  };
  return deps.nodemailer.createTransporter(transportOptions);
}
