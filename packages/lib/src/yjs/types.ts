import 'yjs';
import { Doc, Map } from 'yjs';

declare module 'yjs' {
  export interface Doc {
    /* The shape is {value: boolean}, don't get other key
     */
    getIsModifiedMap(): Map<boolean>;
    setIsModified(isModified: boolean): void;
    getIsModified(): boolean;
  }
}

Doc.prototype.getIsModifiedMap = function (): Map<boolean> {
  return this.getMap<boolean>('isModified');
};

Doc.prototype.setIsModified = function (isModified: boolean): void {
  this.getIsModifiedMap().set('value', isModified);
};

Doc.prototype.getIsModified = function (): boolean {
  return this.getIsModifiedMap().get('value') ?? false;
};
