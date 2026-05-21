'use client';

import { DefaultReactSuggestionItem } from '@blocknote/react';
import { useMemo } from 'react';

export function createLoadingMenuItem(): DefaultReactSuggestionItem {
  return {
    title: 'Loading...',
    subtext: 'Searching',
    onItemClick: () => {},
  };
}

export function createErrorMenuItem(error: Error): DefaultReactSuggestionItem {
  return {
    title: 'Search Error',
    subtext: error.message || 'Failed to fetch results',
    onItemClick: () => {},
  };
}

export function createNoResultsMenuItem(query: string): DefaultReactSuggestionItem {
  return {
    title: `No results for "${query}"`,
    subtext: 'Try a different search term',
    onItemClick: () => {},
  };
}

interface MenuItemsWithStateProps {
  items: DefaultReactSuggestionItem[];
  isLoading: boolean;
  error: Error | null;
  query: string;
}

export function getMenuItemsWithState({
  items,
  isLoading,
  error,
  query,
}: MenuItemsWithStateProps): DefaultReactSuggestionItem[] {
  if (error) {
    return [createErrorMenuItem(error)];
  }

  if (isLoading) {
    return [createLoadingMenuItem()];
  }

  if (items.length === 0 && query) {
    return [createNoResultsMenuItem(query)];
  }

  return items;
}
