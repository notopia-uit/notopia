'use client';

import dynamic from 'next/dynamic';
import type { ComponentProps } from 'react';

import type EditorComponent from './editor';

export const Editor = dynamic<ComponentProps<typeof EditorComponent>>(() => import('./editor'), {
  ssr: false,
});
