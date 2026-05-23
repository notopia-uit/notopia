

export interface Pagination { 
  /**
   * Current page number, starting from 1
   */
  page: number;
  /**
   * Number of items in the current page
   */
  currentTotal: number;
  /**
   * Total items across all pages
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

