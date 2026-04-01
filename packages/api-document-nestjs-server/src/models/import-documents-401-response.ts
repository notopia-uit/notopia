

/**
 * The error response body returned when JWT validation or OPA authorization fails.
 */
export interface ImportDocuments401Response { 
  /**
   * The category of the error encountered during the middleware lifecycle.
   */
  type: ImportDocuments401ResponseType;
  /**
   * A descriptive message providing technical context for the failure.
   */
  details: string;
  /**
   * An optional, developer-defined message, often populated by OPA policy violations.
   */
  custom_message: string | null;
}
  export enum ImportDocuments401ResponseType {
    ExtractToken = 'ExtractToken',
    VerifyToken = 'VerifyToken',
    FetchJwks = 'FetchJWKS',
    Opa = 'OPA'
  };



