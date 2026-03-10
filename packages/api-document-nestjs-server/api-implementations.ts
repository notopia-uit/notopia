import { Type } from '@nestjs/common';
import { DocumentApi } from './api';
import { RevisionApi } from './api';

/**
 * Provide this type to {@link ApiModule} to provide your API implementations
**/
export type ApiImplementations = {
  documentApi: Type<DocumentApi>
  revisionApi: Type<RevisionApi>
};
