import { RevisionTrigger } from './revision-trigger';


export interface Revision { 
  readonly id: string;
  name: string | null;
  readonly createdAt: string;
  /**
   * Who the revision is attributed to. Null for revisions created before authors were recorded
   */
  readonly authorId: string | null;
  trigger: RevisionTrigger;
  /**
   * Users who edited the document since the previous revision, most recent last
   */
  readonly contributorIds: Array<string>;
}



