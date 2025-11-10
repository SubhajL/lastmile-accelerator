import type { SmtpConfig } from '../types.js';

export function createSmtpTransporter(config: SmtpConfig, deps: { nodemailer: { createTransporter: (opts: any) => any } }) {
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
