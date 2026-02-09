

export interface ModelError { 
  /**
   * Error code
   */
  code: string;
  /**
   * Human-readable error message
   */
  message: string;
  /**
   * URL with more information about the error
   */
  more_info?: string;
}

