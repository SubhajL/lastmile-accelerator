import { describe, it, expect, vi } from 'vitest';
import { createDbClient, healthCheck } from './client.js';
import type { PostgresConfig } from '../types.js';
import type { Pool } from 'pg';

describe('db/client', () => {
  const mockConfig: PostgresConfig = {
    host: 'localhost',
    port: 5432,
    database: 'test',
    user: 'test',
    password: 'test',
    maxConnections: 10
  };

  describe('createDbClient', () => {
    it('should create database pool with provided config', () => {
      const pool = createDbClient(mockConfig);

      expect(pool).toBeDefined();
      expect(pool.options.host).toBe('localhost');
      expect(pool.options.port).toBe(5432);
      expect(pool.options.database).toBe('test');
      expect(pool.options.user).toBe('test');
      expect(pool.options.max).toBe(10);
    });

    it('should configure idle timeout and connection timeout', () => {
      const pool = createDbClient(mockConfig);

      expect(pool.options.idleTimeoutMillis).toBe(30000);
      expect(pool.options.connectionTimeoutMillis).toBe(5000);
    });
  });

  describe('healthCheck', () => {
    it('should return true on successful query', async () => {
      const mockPool = {
        query: vi.fn().mockResolvedValue({ rows: [{ result: 1 }] })
      } as unknown as Pool;

      const result = await healthCheck(mockPool);

      expect(result).toBe(true);
      expect(mockPool.query).toHaveBeenCalledWith('SELECT 1 as result');
    });

    it('should return false on query failure', async () => {
      const mockPool = {
        query: vi.fn().mockRejectedValue(new Error('Connection failed'))
      } as unknown as Pool;

      const result = await healthCheck(mockPool);

      expect(result).toBe(false);
    });
  });
});
