import 'yjs';
import { Map } from 'yjs';

export interface YDocMetadata {
  modified: boolean;
}

export class YDocMetadataMap extends Map<YDocMetadata> {}
