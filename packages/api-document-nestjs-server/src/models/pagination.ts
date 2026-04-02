

export interface Pagination { 
  /**
   * Current page number
   */
  page: number;
  /**
   * Number of items per page
   */
  limit: number;
  /**
   * Total number of items
   */
  total: number;
  /**
   * Total number of pages
   */
  totalPages: number;
  /**
   * Whether there is a next page
   */
  hasNext: boolean;
  /**
   * Whether there is a previous page
   */
  hasPrev: boolean;
}

