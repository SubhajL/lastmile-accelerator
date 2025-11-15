declare module '../lib/jwks' {
  export type VerifyJwtArgs = {
    token: string;
    jwksUrl: string;
    issuer?: string;
    audience?: string;
    alg?: string;
    clockSkewSec?: number;
  };

  export function verifyJwt(args: VerifyJwtArgs): Promise<unknown>;
}
