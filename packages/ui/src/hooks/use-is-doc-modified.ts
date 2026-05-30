'use client';
import { YDocMetadata, YDocMetadataMap } from '@notopia-uit/lib/yjs';
import { useEffect, useState } from 'react';
import * as Y from 'yjs';

export function useIsDocModified(ydoc: Y.Doc) {
  const [isModified, setIsModified] = useState(false);
  useEffect(() => {
    const metadataMap = ydoc.getMap('metadata') as YDocMetadataMap;

    const update = () => {
      const metadata = metadataMap.get('metadata');
      setIsModified(metadata?.modified === true);
    };

    update();

    const observer = (event: Y.YMapEvent<YDocMetadata>) => {
      if (event.keysChanged.has('metadata')) {
        update();
      }
    };

    metadataMap.observe(observer);

    return () => {
      metadataMap.unobserve(observer);
    };
  }, [ydoc]);
  return { isModified };
}
