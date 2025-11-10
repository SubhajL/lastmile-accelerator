import { describe, it, expect, vi } from 'vitest';
import { createSmtpTransporter } from './transporter.js';

const mockNodemailer = {
  createTransporter: vi.fn()
};

describe('channels/transporter', () => {
  it('creates transporter with SMTP config', () => {
    const config = {
      host: 'smtp.example.com',
      port: 587,
      user: 'user@example.com',
      password: 'secret',
      from: 'noreply@example.com',
      secure: false
    };

    const transporter = createSmtpTransporter(config, { nodemailer: mockNodemailer as any });

    expect(mockNodemailer.createTransporter).toHaveBeenCalledWith({
      host: 'smtp.example.com',
      port: 587,
      secure: false,
      auth: {
        user: 'user@example.com',
        pass: 'secret'
      }
    });
  });

  it('sets secure true for port 465', () => {
    const config = {
      host: 'smtp.gmail.com',
      port: 465,
      user: 'user@gmail.com',
      password: 'secret',
      from: 'noreply@gmail.com',
      secure: true
    };

    createSmtpTransporter(config, { nodemailer: mockNodemailer as any });

    expect(mockNodemailer.createTransporter).toHaveBeenCalledWith({
      host: 'smtp.gmail.com',
      port: 465,
      secure: true,
      auth: {
        user: 'user@gmail.com',
        pass: 'secret'
      }
    });
  });
});