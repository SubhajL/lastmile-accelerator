export type ZipUploadSession = {
  projectId: string;
  bucket: string;
  objectKey: string;
  expiresAtUnixSeconds: number;
};

export type ZipUploadSessionStore = {
  putSession: (args: { uploadId: string; session: ZipUploadSession; ttlSeconds: number }) => Promise<void>;
  getSession: (uploadId: string) => Promise<ZipUploadSession | null>;
  deleteSession: (uploadId: string) => Promise<void>;
  close: () => Promise<void>;
};

export function createInMemoryZipUploadSessionStore(): ZipUploadSessionStore {
  const sessions = new Map<string, ZipUploadSession>();

  return {
    async putSession(args) {
      sessions.set(args.uploadId, args.session);
    },
    async getSession(uploadId) {
      return sessions.get(uploadId) ?? null;
    },
    async deleteSession(uploadId) {
      sessions.delete(uploadId);
    },
    async close() {},
  };
}

