import { Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import * as jose from 'jose';
import { Traceable } from 'nestjs-otel';

import { InvalidAuthenticationTokenException } from '#/authentication/authentication.exception';
import { User } from '#/common/user';
import { AuthenticationConfig } from '#/config/config';
import { AUTHENTICATION_CONFIG } from '#/config/config.factory';

@Injectable()
@Traceable()
export class AuthenticationService {
  private readonly keySets: ReturnType<typeof jose.createRemoteJWKSet>[];
  private readonly issuers?: string[];
  private readonly audiences?: string[];

  private readonly logger = new Logger(AuthenticationService.name);

  constructor(configService: ConfigService) {
    const authenticationConfig = configService.get<AuthenticationConfig>(AUTHENTICATION_CONFIG)!;
    const jwksUrls = authenticationConfig.jwksUrls.map((url) => new URL(url));
    if (jwksUrls.length === 0) {
      throw new Error('NOTOPIA_DOCUMENT_AUTHENTICATION_JWKS_URLS must contain at least one URL');
    }
    this.keySets = jwksUrls.map((url) => jose.createRemoteJWKSet(url));
    this.issuers = authenticationConfig.issuers;
    this.audiences = authenticationConfig.audiences;
    console.info(
      `AuthenticationService initialized with JWKS URLs: ${jwksUrls.join(', ')}, Issuers: ${this.issuers?.join(', ')}, Audiences: ${this.audiences?.join(', ')}`
    );
  }

  async validateToken(token: string): Promise<User> {
    try {
      this.logger.debug(`Validating token: ${token}`);
      let options: jose.JWTVerifyOptions | undefined;
      if (this.issuers || this.audiences) {
        options = {
          issuer: this.issuers,
          audience: this.audiences,
        };
      }
      const { payload } = await jose.jwtVerify(token, this.multiTenantJWKS.bind(this), options);
      return {
        id: payload.sub!,
        email: payload.email as string,
        groups: payload.groups as string[],
        roles: payload.roles as string[],
      };
    } catch (error) {
      this.logger.warn(`Failed to validate token: ${token}`);
      throw new InvalidAuthenticationTokenException(token, error);
    }
  }

  private async multiTenantJWKS(
    protectedHeader?: jose.JWSHeaderParameters,
    token?: jose.FlattenedJWSInput
  ): Promise<jose.CryptoKey> {
    for (const getKey of this.keySets) {
      try {
        this.logger.debug(`Trying JWKS for token: ${JSON.stringify(token)}`);
        return await getKey(protectedHeader, token);
      } catch (err) {
        if (err instanceof jose.errors.JWKSNoMatchingKey) {
          this.logger.debug(
            `No matching key found in this JWKS for token: ${JSON.stringify(token)}`
          );
          continue;
        }
        this.logger.log(`Error fetching key from JWKS for token: ${JSON.stringify(token)}`);
        throw err;
      }
    }
    this.logger.warn(`No matching key found in any JWKS for token: ${JSON.stringify(token)}`);
    throw new Error('ERR_JWKS_NO_MATCHING_KEY: Key not found in any configured JWKS');
  }
}
