'use client';
import { useEffect, useState, useCallback } from 'react';
import * as Y from 'yjs';

type MetadataValues = boolean | string | number;

export function useIsDocModified(ydoc: Y.Doc, clientId: string) {
  const [isModified, setIsModified] = useState(false);
  useEffect(() => {
    const metadataMap = ydoc.getMap<MetadataValues>('metadata');

    setIsModified(metadataMap.get('isModified') === true);

    const observer = (event: Y.YMapEvent<MetadataValues>, transaction: Y.Transaction) => {
      if (event.keysChanged.has('isModified')) {
        if (transaction.origin !== clientId) {
          setIsModified(metadataMap.get('isModified') === true);
        }
      }
    };

    metadataMap.observe(observer);

    return () => {
      metadataMap.unobserve(observer);
    };
  }, [ydoc, clientId]);
  const setModified = useCallback(
    (status: boolean) => {
      const metadataMap = ydoc.getMap<MetadataValues>('metadata');

      setIsModified(status);

      ydoc.transact(() => {
        metadataMap.set('isModified', status);
      }, clientId);
    },
    [ydoc, clientId]
  );
  return { isModified, setModified };
}
