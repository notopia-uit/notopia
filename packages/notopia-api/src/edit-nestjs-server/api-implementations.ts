import { Type } from '@nestjs/common';
import { EditApi } from './api';

/**
 * Provide this type to {@link ApiModule} to provide your API implementations
**/
export type ApiImplementations = {
  editApi: Type<EditApi>
};
