declare module '@aws-sdk/client-s3' {
  export class S3Client {
    constructor(config?: any);
    send: (cmd: any) => Promise<any>;
  }
  export class PutObjectCommand {
    constructor(input: any);
  }
}

declare module '@kubernetes/client-node' {
  export class KubeConfig {
    loadFromDefault(): void;
    makeApiClient<T>(api: any): T;
  }
  export class BatchV1Api {
    createNamespacedJob(namespace: string, manifest: any): Promise<any>;
    readNamespacedJob(name: string, namespace: string): Promise<any>;
  }
  export class CoreV1Api {}
  export interface V1Job {}
}

declare module 'selenium-webdriver' {
  export class Builder {
    forBrowser(browser: string): this;
    usingServer(url: string): this;
    build(): Promise<any>;
  }
}
