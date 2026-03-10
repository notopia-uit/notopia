import { Pagination } from './pagination';
import { Revision } from './revision';


export interface GetRevisions200Response { 
  data: Array<Revision>;
  pagination: Pagination;
}

