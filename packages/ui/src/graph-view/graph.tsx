'use client';

import { useEffect, useRef } from 'react';

import { renderGraph } from './graph-render';

import styles from './graph.module.css';

export interface D3Config {
  drag: boolean;
  zoom: boolean;
  depth: number;
  scale: number;
  repelForce: number;
  centerForce: number;
  linkDistance: number;
  fontSize: number;
  opacityScale: number;
  removeTags: string[];
  showTags: boolean;
  focusOnHover?: boolean;
  enableRadial?: boolean;
}

interface GraphOptions {
  localGraph: Partial<D3Config>;
  globalGraph: Partial<D3Config>;
}

interface GraphProps {
  data: GraphData;
  currentSlug?: string;
  options?: Partial<GraphOptions>;
  className?: string;
  onNodeClick?: (nodeId: string, nodeType: 'note' | 'tag') => void;
}

export interface GraphNode {
  id: string;
  name: string;
  type: 'note' | 'tag';
  weight?: number;
}

export interface GraphLink {
  source: string;
  target: string;
}

export interface GraphData {
  nodes: GraphNode[];
  links: GraphLink[];
}

const defaultOptions: GraphOptions = {
  localGraph: {
    drag: true,
    zoom: true,
    depth: 1,
    scale: 1.1,
    repelForce: 0.5,
    centerForce: 0.3,
    linkDistance: 30,
    fontSize: 0.6,
    opacityScale: 1,
    showTags: true,
    removeTags: [],
    focusOnHover: false,
    enableRadial: false,
  },
  globalGraph: {
    drag: true,
    zoom: true,
    depth: -1,
    scale: 0.9,
    repelForce: 0.5,
    centerForce: 0.2,
    linkDistance: 30,
    fontSize: 0.6,
    opacityScale: 1,
    showTags: true,
    removeTags: [],
    focusOnHover: true,
    enableRadial: true,
  },
};

export default function Graph({
  data,
  currentSlug = '',
  options,
  className = '',
  onNodeClick,
}: GraphProps) {
  const localGraphRef = useRef<HTMLDivElement>(null);
  const globalGraphRef = useRef<HTMLDivElement>(null);
  const cleanupRef = useRef<(() => void) | null>(null);

  const localGraph = { ...defaultOptions.localGraph, ...options?.localGraph };
  const globalGraph = { ...defaultOptions.globalGraph, ...options?.globalGraph };

  useEffect(() => {
    if (localGraphRef.current) {
      renderGraph(localGraphRef.current, currentSlug, data, localGraph, onNodeClick).then(
        (cleanup) => {
          cleanupRef.current = cleanup;
        }
      );
    }

    return () => {
      if (cleanupRef.current) {
        cleanupRef.current();
      }
    };
  }, [data, currentSlug, localGraph, onNodeClick]);

  return (
    <div className={`${styles.graph} ${className} size-full`}>
      <div className={styles.graphOuter}>
        <div
          ref={localGraphRef}
          className={styles.graphContainer}
          data-cfg={JSON.stringify(localGraph)}
        />
      </div>
    </div>
  );
}
