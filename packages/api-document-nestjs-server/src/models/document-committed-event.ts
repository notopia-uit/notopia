

export interface DocumentCommittedEvent { 
  id: string;
  /**
   * BlockNote model
   */
  content: Array<object>;
  userId: string;
  tags: Array<string>;
  outgoingLinkIds: Array<string>;
}

