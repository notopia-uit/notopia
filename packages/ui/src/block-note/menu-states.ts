'use client';

import { DefaultReactSuggestionItem } from '@blocknote/react';

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

interface MenuItemsWithStateProps<T extends DefaultReactSuggestionItem> {
  items: T[];
  isLoading: boolean;
  error: Error | null;
  query: string;
}

export function getMenuItemsWithState<T extends DefaultReactSuggestionItem>({
  items,
  isLoading,
  error,
  query,
}: MenuItemsWithStateProps<T>): T[] {
  if (error) {
    return [createErrorMenuItem(error) as T];
  }

  if (isLoading) {
    return [createLoadingMenuItem() as T];
  }

  if (items.length === 0 && query) {
    return [createNoResultsMenuItem(query) as T];
  }

  return items;
}
