

export interface Revision { 
  readonly id: string;
  name: string | null;
  /**
   * BlockNote model
   */
  content: Array<object> | null;
  readonly createdAt: string;
}

