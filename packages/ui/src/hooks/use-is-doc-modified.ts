'use client';
import { useEffect, useState, useCallback, useRef } from 'react';
import * as Y from 'yjs';
import { YDocMetadata } from '@notopia-uit/lib/yjs';

export function useIsDocModified(ydoc: Y.Doc, clientId: string) {
  const [isModified, setIsModified] = useState(false);
  const metadataMapRef = useRef<Y.Map<YDocMetadata>>(null);
  useEffect(() => {
    const metadataMap = ydoc.getMap<YDocMetadata>('metadata');
    metadataMapRef.current = metadataMap;

    const metadata = metadataMap.get('modified');
    setIsModified(metadata?.modified === true);

    const observer = (event: Y.YMapEvent<YDocMetadata>, transaction: Y.Transaction) => {
      if (event.keysChanged.has('modified')) {
        if (transaction.origin !== clientId) {
          const metadata = metadataMap.get('modified');
          setIsModified(metadata?.modified === true);
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
      const metadataMap = metadataMapRef.current ?? ydoc.getMap<YDocMetadata>('metadata');

      setIsModified(status);

      ydoc.transact(() => {
        metadataMap.set('modified', { modified: status });
      }, clientId);
    },
    [ydoc, clientId]
  );
  return { isModified, setModified };
}
