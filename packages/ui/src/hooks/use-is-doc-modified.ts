'use client';
import { YDocMetadata, YDocMetadataMap } from '@notopia-uit/lib/yjs';
import { useEffect, useState, useRef } from 'react';
import * as Y from 'yjs';

export function useIsDocModified(ydoc: Y.Doc, clientId: string) {
  const [isModified, setIsModified] = useState(false);
  const metadataMapRef = useRef<YDocMetadataMap>(null);
  useEffect(() => {
    const metadataMap = ydoc.getMap('metadata') as YDocMetadataMap;
    metadataMapRef.current = metadataMap;

    const metadata = metadataMap.get('metadata');
    setIsModified(metadata?.modified === true);

    const observer = (event: Y.YMapEvent<YDocMetadata>) => {
      if (event.keysChanged.has('metadata')) {
        setIsModified(metadata?.modified === true);
      }
    };

    metadataMap.observe(observer);

    return () => {
      metadataMap.unobserve(observer);
    };
  }, [ydoc, clientId]);
  return { isModified };
}
