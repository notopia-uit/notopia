

export interface RevisionWithContent { 
  readonly id: string;
  name: string | null;
  readonly createdAt: string;
  /**
   * BlockNote model
   */
  content: Array<object>;
}

